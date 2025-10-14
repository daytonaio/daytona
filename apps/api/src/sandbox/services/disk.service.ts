/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ForbiddenException, Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository, Not } from 'typeorm'
import { Disk } from '../entities/disk.entity'
import { Sandbox } from '../entities/sandbox.entity'
import { DiskState } from '../enums/disk-state.enum'
import { SandboxState } from '../enums/sandbox-state.enum'
import { CreateDiskDto } from '../dto/create-disk.dto'
import { v4 as uuidv4 } from 'uuid'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { Organization } from '../../organization/entities/organization.entity'
import { OrganizationService } from '../../organization/services/organization.service'
import { OrganizationUsageService } from '../../organization/services/organization-usage.service'

@Injectable()
export class DiskService {
  private readonly logger = new Logger(DiskService.name)

  constructor(
    @InjectRepository(Disk)
    private readonly diskRepository: Repository<Disk>,
    @InjectRepository(Sandbox)
    private readonly sandboxRepository: Repository<Sandbox>,
    private readonly organizationService: OrganizationService,
    private readonly organizationUsageService: OrganizationUsageService,
  ) {}

  private async validateOrganizationQuotas(
    organization: Organization,
    addedDiskCount: number,
  ): Promise<{
    pendingDiskCountIncremented: boolean
  }> {
    // validate usage quotas
    await this.organizationUsageService.incrementPendingDiskUsage(organization.id, addedDiskCount)

    const usageOverview = await this.organizationUsageService.getSandboxUsageOverview(organization.id)

    try {
      if (usageOverview.currentDiskUsage + usageOverview.pendingDiskUsage > organization.totalDiskQuota) {
        throw new ForbiddenException(`Disk quota exceeded. Maximum allowed: ${organization.totalDiskQuota}`)
      }
    } catch (error) {
      await this.rollbackPendingUsage(organization.id, addedDiskCount)
      throw error
    }

    return {
      pendingDiskCountIncremented: true,
    }
  }

  async rollbackPendingUsage(organizationId: string, pendingDiskCountIncrement?: number): Promise<void> {
    if (!pendingDiskCountIncrement) {
      return
    }

    try {
      await this.organizationUsageService.decrementPendingDiskUsage(organizationId, pendingDiskCountIncrement)
    } catch (error) {
      this.logger.error(`Error rolling back pending disk usage: ${error}`)
    }
  }

  async create(organization: Organization, createDiskDto: CreateDiskDto): Promise<Disk> {
    let pendingDiskCountIncrement: number | undefined

    try {
      this.organizationService.assertOrganizationIsNotSuspended(organization)

      const newDiskCount = 1

      const { pendingDiskCountIncremented } = await this.validateOrganizationQuotas(organization, newDiskCount)

      if (pendingDiskCountIncremented) {
        pendingDiskCountIncrement = newDiskCount
      }

      const disk = new Disk()

      // Generate ID
      disk.id = uuidv4()

      // Set name and size from DTO
      disk.name = createDiskDto.name
      disk.size = createDiskDto.size

      // Check if disk with same name already exists for organization
      const existingDisk = await this.diskRepository.findOne({
        where: {
          organizationId: organization.id,
          name: disk.name,
          state: Not(DiskState.STORED), // Consider STORED as deleted state
        },
      })

      if (existingDisk) {
        throw new BadRequestError(`Disk with name ${disk.name} already exists`)
      }

      disk.organizationId = organization.id
      disk.state = DiskState.FRESH

      const savedDisk = await this.diskRepository.save(disk)
      this.logger.debug(`Created disk ${savedDisk.id} for organization ${organization.id}`)
      return savedDisk
    } catch (error) {
      await this.rollbackPendingUsage(organization.id, pendingDiskCountIncrement)
      throw error
    }
  }

  async delete(diskId: string): Promise<void> {
    const disk = await this.diskRepository.findOne({
      where: {
        id: diskId,
      },
    })

    if (!disk) {
      throw new NotFoundException(`Disk with ID ${diskId} not found`)
    }

    if (disk.state !== DiskState.READY) {
      throw new BadRequestError(`Disk must be in '${DiskState.READY}' state in order to be deleted`)
    }

    // Update state to mark as stored (deleted)
    disk.state = DiskState.STORED
    await this.diskRepository.save(disk)
    this.logger.debug(`Marked disk ${diskId} for deletion`)
  }

  async findOne(diskId: string): Promise<Disk> {
    const disk = await this.diskRepository.findOne({
      where: { id: diskId },
    })

    if (!disk) {
      throw new NotFoundException(`Disk with ID ${diskId} not found`)
    }

    return disk
  }

  async findAll(organizationId: string): Promise<Disk[]> {
    return this.diskRepository.find({
      where: {
        organizationId,
        state: Not(DiskState.STORED), // Exclude stored (deleted) disks
      },
      order: {
        createdAt: 'DESC',
      },
    })
  }

  async findByName(organizationId: string, name: string): Promise<Disk> {
    const disk = await this.diskRepository.findOne({
      where: {
        organizationId,
        name,
        state: Not(DiskState.STORED),
      },
    })

    if (!disk) {
      throw new NotFoundException(`Disk with name ${name} not found`)
    }

    return disk
  }

  async attachToSandbox(diskId: string, sandboxId: string): Promise<Disk> {
    // Find the disk
    const disk = await this.diskRepository.findOne({
      where: { id: diskId },
    })

    if (!disk) {
      throw new NotFoundException(`Disk with ID ${diskId} not found`)
    }

    // Validate disk state
    if (disk.state !== DiskState.STORED && disk.state !== DiskState.FRESH) {
      throw new BadRequestError(`Disk must be in '${DiskState.STORED}' or '${DiskState.FRESH}' state to be attached`)
    }

    // Check if disk is already attached to another sandbox
    if (disk.sandboxId && disk.sandboxId !== sandboxId) {
      throw new BadRequestError(`Disk is already attached to sandbox ${disk.sandboxId}`)
    }

    // Find the sandbox
    const sandbox = await this.sandboxRepository.findOne({
      where: { id: sandboxId },
    })

    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    // Validate sandbox state
    if (sandbox.state !== SandboxState.STOPPED) {
      throw new BadRequestError(`Sandbox must be in '${SandboxState.STOPPED}' state to attach a disk`)
    }

    // Check if sandbox already has a disk attached
    const existingAttachedDisk = await this.diskRepository.findOne({
      where: {
        sandboxId: sandboxId,
        state: DiskState.ATTACHED,
      },
    })

    if (existingAttachedDisk) {
      throw new BadRequestError(`Sandbox already has a disk attached (${existingAttachedDisk.id})`)
    }

    // Attach the disk
    disk.sandboxId = sandboxId
    disk.state = DiskState.ATTACHED

    const savedDisk = await this.diskRepository.save(disk)
    this.logger.debug(`Attached disk ${diskId} to sandbox ${sandboxId}`)
    return savedDisk
  }

  async detachFromSandbox(diskId: string): Promise<Disk> {
    // Find the disk
    const disk = await this.diskRepository.findOne({
      where: { id: diskId },
    })

    if (!disk) {
      throw new NotFoundException(`Disk with ID ${diskId} not found`)
    }

    // Validate disk state
    if (disk.state !== DiskState.ATTACHED) {
      throw new BadRequestError(`Disk must be in '${DiskState.ATTACHED}' state to be detached`)
    }

    // Validate sandbox state if it exists
    if (disk.sandboxId) {
      const sandbox = await this.sandboxRepository.findOne({
        where: { id: disk.sandboxId },
      })

      if (sandbox && sandbox.state !== SandboxState.STOPPED) {
        throw new BadRequestError(`Sandbox must be in '${SandboxState.STOPPED}' state to detach a disk`)
      }
    }

    // Detach the disk
    disk.sandboxId = null
    disk.state = DiskState.DETACHED

    const savedDisk = await this.diskRepository.save(disk)
    this.logger.debug(`Detached disk ${diskId} from sandbox ${disk.sandboxId}`)
    return savedDisk
  }

  async detachAllFromSandbox(sandboxId: string): Promise<void> {
    // Find all disks attached to the sandbox
    const attachedDisks = await this.diskRepository.find({
      where: {
        sandboxId: sandboxId,
        state: DiskState.ATTACHED,
      },
    })

    // Detach each disk
    for (const disk of attachedDisks) {
      disk.sandboxId = null
      disk.state = DiskState.DETACHED
      await this.diskRepository.save(disk)
      this.logger.debug(`Auto-detached disk ${disk.id} from sandbox ${sandboxId}`)
    }
  }
}
