/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Injectable,
  NotFoundException,
  ConflictException,
  ForbiddenException,
  BadRequestException,
  Logger,
} from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository, Not, In, IsNull, Raw, Like, JsonContains } from 'typeorm'
import { Snapshot } from '../entities/snapshot.entity'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { CreateSnapshotDto } from '../dto/create-snapshot.dto'
import { BuildInfo } from '../entities/build-info.entity'
import { generateBuildInfoHash as generateBuildSnapshotRef } from '../entities/build-info.entity'
import { OnEvent } from '@nestjs/event-emitter'
import { SandboxEvents } from '../constants/sandbox-events.constants'
import { SandboxCreatedEvent } from '../events/sandbox-create.event'
import { Organization } from '../../organization/entities/organization.entity'
import { OrganizationService } from '../../organization/services/organization.service'
import { SnapshotRunner } from '../entities/snapshot-runner.entity'
import { Sandbox } from '../entities/sandbox.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { OrganizationEvents } from '../../organization/constants/organization-events.constant'
import { OrganizationSuspendedSnapshotDeactivatedEvent } from '../../organization/events/organization-suspended-snapshot-deactivated.event'
import { SnapshotRunnerState } from '../enums/snapshot-runner-state.enum'
import { PaginatedList } from '../../common/interfaces/paginated-list.interface'

const IMAGE_NAME_REGEX = /^[a-zA-Z0-9_.\-:]+(\/[a-zA-Z0-9_.\-:]+)*$/
@Injectable()
export class SnapshotService {
  private readonly logger = new Logger(SnapshotService.name)

  constructor(
    @InjectRepository(Sandbox)
    private readonly sandboxRepository: Repository<Sandbox>,
    @InjectRepository(Snapshot)
    private readonly snapshotRepository: Repository<Snapshot>,
    @InjectRepository(BuildInfo)
    private readonly buildInfoRepository: Repository<BuildInfo>,
    @InjectRepository(SnapshotRunner)
    private readonly snapshotRunnerRepository: Repository<SnapshotRunner>,
    private readonly organizationService: OrganizationService,
  ) {}

  private validateImageName(name: string): string | null {
    if (!name.includes(':') || name.endsWith(':') || /:\s*$/.test(name)) {
      return 'Image name must include a tag (e.g., ubuntu:22.04)'
    }

    if (name.endsWith(':latest')) {
      return 'Images with tag ":latest" are not allowed'
    }

    if (!IMAGE_NAME_REGEX.test(name)) {
      return 'Invalid image name format. Must be lowercase, may contain digits, dots, dashes, and single slashes between components'
    }

    return null
  }

  private validateSnapshotName(name: string): string | null {
    if (!IMAGE_NAME_REGEX.test(name)) {
      return 'Invalid snapshot name format. May contain letters, digits, dots, colons, and dashes'
    }

    return null
  }

  async createSnapshot(organization: Organization, createSnapshotDto: CreateSnapshotDto, general = false) {
    const nameValidationError = this.validateSnapshotName(createSnapshotDto.name)
    if (nameValidationError) {
      throw new BadRequestException(nameValidationError)
    }

    if (createSnapshotDto.imageName) {
      const imageValidationError = this.validateImageName(createSnapshotDto.imageName)
      if (imageValidationError) {
        throw new BadRequestException(imageValidationError)
      }
    }

    this.organizationService.assertOrganizationIsNotSuspended(organization)

    // check if the organization has reached the snapshot quota
    const snapshots = await this.snapshotRepository.find({
      where: { organizationId: organization.id },
    })

    if (snapshots.length >= organization.snapshotQuota) {
      throw new ForbiddenException('Reached the maximum number of snapshots in the organization')
    }

    await this.validateOrganizationMaxQuotas(
      organization,
      createSnapshotDto.cpu,
      createSnapshotDto.memory,
      createSnapshotDto.disk,
    )

    try {
      const snapshot = this.snapshotRepository.create({
        organizationId: organization.id,
        ...createSnapshotDto,
        mem: createSnapshotDto.memory, // Map memory to mem
        state: createSnapshotDto.buildInfo ? SnapshotState.BUILD_PENDING : SnapshotState.PENDING,
        general,
      })

      if (createSnapshotDto.buildInfo) {
        const buildSnapshotRef = generateBuildSnapshotRef(
          createSnapshotDto.buildInfo.dockerfileContent,
          createSnapshotDto.buildInfo.contextHashes,
        )

        // Check if buildInfo with the same snapshotRef already exists
        const existingBuildInfo = await this.buildInfoRepository.findOne({
          where: { snapshotRef: buildSnapshotRef },
        })

        if (existingBuildInfo) {
          snapshot.buildInfo = existingBuildInfo
        } else {
          const buildInfoEntity = this.buildInfoRepository.create({
            ...createSnapshotDto.buildInfo,
          })
          await this.buildInfoRepository.save(buildInfoEntity)
          snapshot.buildInfo = buildInfoEntity
        }
      }

      return await this.snapshotRepository.save(snapshot)
    } catch (error) {
      if (error.code === '23505') {
        // PostgreSQL unique violation error code
        throw new ConflictException(
          `Snapshot with name "${createSnapshotDto.name}" already exists for this organization`,
        )
      }
      throw error
    }
  }

  async toggleSnapshotState(snapshotId: string, enabled: boolean) {
    const snapshot = await this.snapshotRepository.findOne({
      where: { id: snapshotId },
    })

    if (!snapshot) {
      throw new NotFoundException(`Snapshot ${snapshotId} not found`)
    }

    snapshot.enabled = enabled
    return await this.snapshotRepository.save(snapshot)
  }

  async removeSnapshot(snapshotId: string) {
    const snapshot = await this.snapshotRepository.findOne({
      where: { id: snapshotId },
    })

    if (!snapshot) {
      throw new NotFoundException(`Snapshot ${snapshotId} not found`)
    }
    if (snapshot.general) {
      throw new ForbiddenException('You cannot delete a general snapshot')
    }
    snapshot.state = SnapshotState.REMOVING
    await this.snapshotRepository.save(snapshot)
  }

  async getAllSnapshots(organizationId: string, page = 1, limit = 10): Promise<PaginatedList<Snapshot>> {
    const pageNum = Number(page)
    const limitNum = Number(limit)

    const [items, total] = await this.snapshotRepository.findAndCount({
      // Retrieve all snapshots belonging to the organization as well as all general snapshots
      where: [{ organizationId }, { general: true, hideFromUsers: false }],
      order: {
        general: 'ASC', // Sort general snapshots last
        lastUsedAt: {
          direction: 'DESC',
          nulls: 'LAST',
        },
        createdAt: 'DESC',
      },
      skip: (pageNum - 1) * limitNum,
      take: limitNum,
    })

    return {
      items,
      total,
      page: pageNum,
      totalPages: Math.ceil(total / limit),
    }
  }

  async getSnapshot(snapshotId: string): Promise<Snapshot> {
    const snapshot = await this.snapshotRepository.findOne({
      where: { id: snapshotId },
    })

    if (!snapshot) {
      throw new NotFoundException(`Snapshot ${snapshotId} not found`)
    }

    return snapshot
  }

  async getSnapshotByName(snapshotName: string, organizationId: string): Promise<Snapshot> {
    const snapshot = await this.snapshotRepository.findOne({
      where: { name: snapshotName, organizationId },
    })

    if (!snapshot) {
      //  check if the snapshot is general
      const generalSnapshot = await this.snapshotRepository.findOne({
        where: { name: snapshotName, general: true },
      })
      if (generalSnapshot) {
        return generalSnapshot
      }

      throw new NotFoundException(`Snapshot with name ${snapshotName} not found`)
    }

    return snapshot
  }

  async setSnapshotGeneralStatus(snapshotId: string, general: boolean) {
    const snapshot = await this.snapshotRepository.findOne({
      where: { id: snapshotId },
    })

    if (!snapshot) {
      throw new NotFoundException(`Snapshot ${snapshotId} not found`)
    }

    snapshot.general = general
    return await this.snapshotRepository.save(snapshot)
  }

  private async validateOrganizationMaxQuotas(
    organization: Organization,
    cpu?: number,
    memory?: number,
    disk?: number,
  ): Promise<void> {
    if (cpu && cpu > organization.maxCpuPerSandbox) {
      throw new ForbiddenException(
        `CPU request ${cpu} exceeds maximum allowed per sandbox (${organization.maxCpuPerSandbox})`,
      )
    }
    if (memory && memory > organization.maxMemoryPerSandbox) {
      throw new ForbiddenException(
        `Memory request ${memory}GB exceeds maximum allowed per sandbox (${organization.maxMemoryPerSandbox}GB)`,
      )
    }
    if (disk && disk > organization.maxDiskPerSandbox) {
      throw new ForbiddenException(
        `Disk request ${disk}GB exceeds maximum allowed per sandbox (${organization.maxDiskPerSandbox}GB)`,
      )
    }
  }

  @OnEvent(SandboxEvents.CREATED)
  private async handleSandboxCreatedEvent(event: SandboxCreatedEvent) {
    if (!event.sandbox.snapshot) {
      return
    }

    const snapshot = await this.getSnapshotByName(event.sandbox.snapshot, event.sandbox.organizationId)
    snapshot.lastUsedAt = event.sandbox.createdAt
    await this.snapshotRepository.save(snapshot)
  }

  async activateSnapshot(snapshotId: string): Promise<Snapshot> {
    const snapshot = await this.snapshotRepository.findOne({
      where: { id: snapshotId },
    })

    if (!snapshot) {
      throw new NotFoundException(`Snapshot ${snapshotId} not found`)
    }

    if (snapshot.state === SnapshotState.ACTIVE) {
      throw new BadRequestException(`Snapshot ${snapshotId} is already active`)
    }

    if (snapshot.state !== SnapshotState.INACTIVE) {
      throw new BadRequestException(`Snapshot ${snapshotId} cannot be activated - it is in ${snapshot.state} state`)
    }

    snapshot.state = SnapshotState.ACTIVE
    snapshot.lastUsedAt = new Date()
    return await this.snapshotRepository.save(snapshot)
  }

  async canCleanupImage(imageName: string): Promise<boolean> {
    const snapshot = await this.snapshotRepository.findOne({
      where: {
        state: Not(In([SnapshotState.ERROR, SnapshotState.BUILD_FAILED])),
        internalName: imageName,
      },
    })

    if (snapshot) {
      return false
    }

    const sandbox = await this.sandboxRepository.findOne({
      where: [
        {
          existingBackupSnapshots: Raw((alias) => `${alias} @> '[{"snapshotName":"${imageName}"}]'::jsonb`),
        },
        {
          existingBackupSnapshots: Raw((alias) => `${alias} @> '[{"imageName":"${imageName}"}]'::jsonb`),
        },
        {
          backupSnapshot: imageName,
        },
      ],
    })

    if (sandbox && sandbox.state !== SandboxState.DESTROYED) {
      return false
    }

    return true
  }

  async deactivateSnapshot(snapshotId: string): Promise<void> {
    const snapshot = await this.snapshotRepository.findOne({
      where: { id: snapshotId },
    })

    if (!snapshot) {
      throw new NotFoundException(`Snapshot ${snapshotId} not found`)
    }

    if (snapshot.state === SnapshotState.INACTIVE) {
      return
    }

    snapshot.state = SnapshotState.INACTIVE
    await this.snapshotRepository.save(snapshot)

    // Set associated SnapshotRunner records to REMOVING state
    const result = await this.snapshotRunnerRepository.update(
      { snapshotRef: snapshot.internalName },
      { state: SnapshotRunnerState.REMOVING },
    )

    this.logger.debug(`Deactivated snapshot ${snapshot.id} and marked ${result.affected} SnapshotRunners for removal`)
  }

  @OnEvent(OrganizationEvents.SUSPENDED_SNAPSHOT_DEACTIVATED)
  async handleSuspendedOrganizationSnapshotDeactivated(event: OrganizationSuspendedSnapshotDeactivatedEvent) {
    await this.deactivateSnapshot(event.snapshotId).catch((error) => {
      //  log the error for now, but don't throw it as it will be retried
      this.logger.error(
        `Error deactivating snapshot from suspended organization. SnapshotId: ${event.snapshotId}: `,
        error,
      )
    })
  }
}
