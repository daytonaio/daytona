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
import { Repository, Not, In, Raw, ILike, FindOptionsWhere } from 'typeorm'
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
import { OrganizationUsageService } from '../../organization/services/organization-usage.service'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { SnapshotSortDirection, SnapshotSortField } from '../dto/list-snapshots-query.dto'
import { PER_SANDBOX_LIMIT_MESSAGE } from '../../common/constants/error-messages'

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
    private readonly organizationUsageService: OrganizationUsageService,
    private readonly redisLockProvider: RedisLockProvider,
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
    let pendingSnapshotCountIncrement: number | undefined

    try {
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

      const newSnapshotCount = 1

      const { pendingSnapshotCountIncremented } = await this.validateOrganizationQuotas(
        organization,
        newSnapshotCount,
        createSnapshotDto.cpu,
        createSnapshotDto.memory,
        createSnapshotDto.disk,
      )

      if (pendingSnapshotCountIncremented) {
        pendingSnapshotCountIncrement = newSnapshotCount
      }

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
    } catch (error) {
      await this.rollbackPendingUsage(organization.id, pendingSnapshotCountIncrement)
      throw error
    }
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

  async getAllSnapshots(
    organizationId: string,
    page = 1,
    limit = 10,
    filters?: { name?: string },
    sort?: { field?: SnapshotSortField; direction?: SnapshotSortDirection },
  ): Promise<PaginatedList<Snapshot>> {
    const pageNum = Number(page)
    const limitNum = Number(limit)

    const { name } = filters || {}
    const { field: sortField, direction: sortDirection } = sort || {}

    const baseFindOptions: FindOptionsWhere<Snapshot> = {
      ...(name ? { name: ILike(`%${name}%`) } : {}),
    }

    // Retrieve all snapshots belonging to the organization as well as all general snapshots
    const where: FindOptionsWhere<Snapshot>[] = [
      {
        ...baseFindOptions,
        organizationId,
      },
      {
        ...baseFindOptions,
        general: true,
        hideFromUsers: false,
      },
    ]

    const [items, total] = await this.snapshotRepository.findAndCount({
      where,
      order: {
        general: 'ASC', // Sort general snapshots last
        [sortField]: {
          direction: sortDirection,
          nulls: 'LAST',
        },
        ...(sortField !== SnapshotSortField.CREATED_AT && { createdAt: 'DESC' }),
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

  private async validateOrganizationQuotas(
    organization: Organization,
    addedSnapshotCount: number,
    cpu?: number,
    memory?: number,
    disk?: number,
  ): Promise<{
    pendingSnapshotCountIncremented: boolean
  }> {
    // validate per-sandbox quotas
    if (cpu && cpu > organization.maxCpuPerSandbox) {
      throw new ForbiddenException(
        `CPU request ${cpu} exceeds maximum allowed per sandbox (${organization.maxCpuPerSandbox}).\n${PER_SANDBOX_LIMIT_MESSAGE}`,
      )
    }
    if (memory && memory > organization.maxMemoryPerSandbox) {
      throw new ForbiddenException(
        `Memory request ${memory}GB exceeds maximum allowed per sandbox (${organization.maxMemoryPerSandbox}GB).\n${PER_SANDBOX_LIMIT_MESSAGE}`,
      )
    }
    if (disk && disk > organization.maxDiskPerSandbox) {
      throw new ForbiddenException(
        `Disk request ${disk}GB exceeds maximum allowed per sandbox (${organization.maxDiskPerSandbox}GB).\n${PER_SANDBOX_LIMIT_MESSAGE}`,
      )
    }

    // validate usage quotas
    await this.organizationUsageService.incrementPendingSnapshotUsage(organization.id, addedSnapshotCount)

    const usageOverview = await this.organizationUsageService.getSnapshotUsageOverview(organization.id)

    try {
      if (usageOverview.currentSnapshotUsage + usageOverview.pendingSnapshotUsage > organization.snapshotQuota) {
        throw new ForbiddenException(`Snapshot quota exceeded. Maximum allowed: ${organization.snapshotQuota}`)
      }
    } catch (error) {
      await this.rollbackPendingUsage(organization.id, addedSnapshotCount)
      throw error
    }

    return {
      pendingSnapshotCountIncremented: true,
    }
  }

  async rollbackPendingUsage(organizationId: string, pendingSnapshotCountIncrement?: number): Promise<void> {
    if (!pendingSnapshotCountIncrement) {
      return
    }

    try {
      await this.organizationUsageService.decrementPendingSnapshotUsage(organizationId, pendingSnapshotCountIncrement)
    } catch (error) {
      this.logger.error(`Error rolling back pending snapshot usage: ${error}`)
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

  async activateSnapshot(snapshotId: string, organization: Organization): Promise<Snapshot> {
    const lockKey = `snapshot:${snapshotId}:activate`
    await this.redisLockProvider.waitForLock(lockKey, 60)

    let pendingSnapshotCountIncrement: number | undefined

    try {
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

      this.organizationService.assertOrganizationIsNotSuspended(organization)

      const activatedSnapshotCount = 1

      const { pendingSnapshotCountIncremented } = await this.validateOrganizationQuotas(
        organization,
        activatedSnapshotCount,
        snapshot.cpu,
        snapshot.mem,
        snapshot.disk,
      )

      if (pendingSnapshotCountIncremented) {
        pendingSnapshotCountIncrement = activatedSnapshotCount
      }

      snapshot.state = SnapshotState.ACTIVE
      snapshot.lastUsedAt = new Date()
      return await this.snapshotRepository.save(snapshot)
    } catch (error) {
      await this.rollbackPendingUsage(organization.id, pendingSnapshotCountIncrement)
      throw error
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
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

    try {
      const countActiveSnapshots = await this.snapshotRepository.count({
        where: {
          state: SnapshotState.ACTIVE,
          internalName: snapshot.internalName,
        },
      })

      if (countActiveSnapshots === 0) {
        // Set associated SnapshotRunner records to REMOVING state
        const result = await this.snapshotRunnerRepository.update(
          { snapshotRef: snapshot.internalName },
          { state: SnapshotRunnerState.REMOVING },
        )
        this.logger.debug(
          `Deactivated snapshot ${snapshot.id} and marked ${result.affected} SnapshotRunners for removal`,
        )
      }
    } catch (error) {
      this.logger.error(`Deactivated snapshot ${snapshot.id}, but failed to mark snapshot runners for removal`, error)
    }
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
