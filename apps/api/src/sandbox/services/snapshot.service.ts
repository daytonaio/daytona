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
import { v4 as uuidv4 } from 'uuid'
import { Snapshot } from '../entities/snapshot.entity'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { CreateSnapshotDto } from '../dto/create-snapshot.dto'
import { BuildInfo } from '../entities/build-info.entity'
import { generateBuildInfoHash as generateBuildSnapshotRef } from '../entities/build-info.entity'
import { EventEmitter2, OnEvent } from '@nestjs/event-emitter'
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
import { DockerRegistryService, ImageDetails } from '../../docker-registry/services/docker-registry.service'
import { DefaultRegionRequiredException } from '../../organization/exceptions/DefaultRegionRequiredException'
import { Region } from '../../region/entities/region.entity'
import { RunnerState } from '../enums/runner-state.enum'
import { OnAsyncEvent } from '../../common/decorators/on-async-event.decorator'
import { RunnerEvents } from '../constants/runner-events'
import { RunnerDeletedEvent } from '../events/runner-deleted.event'
import { SnapshotRegion } from '../entities/snapshot-region.entity'
import { RegionType } from '../../region/enums/region-type.enum'
import { SnapshotEvents } from '../constants/snapshot-events'
import { SnapshotCreatedEvent } from '../events/snapshot-created.event'
import { RunnerService } from './runner.service'
import { RegionService } from '../../region/services/region.service'
import { Transactional } from 'typeorm-transactional'

const IMAGE_NAME_REGEX = /^[a-zA-Z0-9_.\-:]+(\/[a-zA-Z0-9_.\-:]+)*(@sha256:[a-f0-9]{64})?$/
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
    @InjectRepository(Region)
    private readonly regionRepository: Repository<Region>,
    @InjectRepository(SnapshotRegion)
    private readonly snapshotRegionRepository: Repository<SnapshotRegion>,
    private readonly organizationService: OrganizationService,
    private readonly organizationUsageService: OrganizationUsageService,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly runnerService: RunnerService,
    private readonly regionService: RegionService,
    private readonly dockerRegistryService: DockerRegistryService,
    private readonly eventEmitter: EventEmitter2,
  ) {}

  private validateImageName(name: string): string | null {
    // Check for digest format (@sha256:hash)
    if (name.includes('@sha256:')) {
      const [imageName, digest] = name.split('@sha256:')
      if (!imageName || !digest || !/^[a-f0-9]{64}$/.test(digest)) {
        return 'Invalid digest format. Must be image@sha256:64_hex_characters'
      }
      return null
    }

    // Handle tag format (image:tag)
    if (!name.includes(':') || name.endsWith(':') || /:\s*$/.test(name)) {
      return 'Image name must include a tag (e.g., ubuntu:22.04) or digest (@sha256:...)'
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

  private processEntrypoint(entrypoint?: string[]): string[] | undefined {
    if (!entrypoint || entrypoint.length === 0) {
      return undefined
    }

    // Filter out empty strings from the array
    const filteredEntrypoint = entrypoint.filter((cmd) => cmd && cmd.trim().length > 0)

    return filteredEntrypoint.length > 0 ? filteredEntrypoint : undefined
  }

  private async readySnapshotRunnerExists(ref: string, regionId: string): Promise<boolean> {
    return await this.snapshotRunnerRepository
      .createQueryBuilder('sr')
      .innerJoin('runner', 'r', 'r.id::text = sr."runnerId"::text')
      .where('sr."snapshotRef" = :ref', { ref })
      .andWhere('sr.state = :snapshotRunnerState', { snapshotRunnerState: SnapshotRunnerState.READY })
      .andWhere('r.region = :regionId', { regionId })
      .andWhere('r.state = :runnerState', { runnerState: RunnerState.READY })
      .andWhere('r.unschedulable = false')
      .getExists()
  }

  @Transactional()
  async createFromPull(organization: Organization, createSnapshotDto: CreateSnapshotDto, general = false) {
    if (!organization.defaultRegionId) {
      throw new DefaultRegionRequiredException()
    }

    const regionId = await this.getValidatedOrDefaultRegionId(organization, createSnapshotDto.regionId)

    let pendingSnapshotCountIncrement: number | undefined

    if (!createSnapshotDto.imageName) {
      throw new BadRequestException('Must specify an image name')
    }

    try {
      let entrypoint = createSnapshotDto.entrypoint
      let ref: string | undefined = undefined
      let state: SnapshotState = SnapshotState.PENDING

      const nameValidationError = this.validateSnapshotName(createSnapshotDto.name)
      if (nameValidationError) {
        throw new BadRequestException(nameValidationError)
      }

      const imageValidationError = this.validateImageName(createSnapshotDto.imageName)
      if (imageValidationError) {
        throw new BadRequestException(imageValidationError)
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

      let imageDetails: ImageDetails | undefined = undefined

      try {
        imageDetails = await this.dockerRegistryService.getImageDetails(createSnapshotDto.imageName, organization.id)
      } catch (error) {
        this.logger.warn(`Could not get image details for ${createSnapshotDto.imageName}: ${error}`)
      }

      if (imageDetails) {
        if (imageDetails?.sizeGB > organization.maxSnapshotSize) {
          throw new ForbiddenException(
            `Image size ${imageDetails.sizeGB} exceeds the maximum allowed snapshot size (${organization.maxSnapshotSize})`,
          )
        }

        if ((!entrypoint || entrypoint.length === 0) && imageDetails) {
          if (imageDetails.entrypoint && imageDetails.entrypoint.length > 0) {
            entrypoint = imageDetails.entrypoint
          } else {
            entrypoint = ['sleep', 'infinity']
          }
        }

        const defaultInternalRegistry = await this.dockerRegistryService.getDefaultInternalRegistry()
        const hash =
          imageDetails.digest && imageDetails.digest.startsWith('sha256:')
            ? imageDetails.digest.substring('sha256:'.length)
            : imageDetails.digest
        ref = `${defaultInternalRegistry.url.replace(/^https?:\/\//, '')}/${defaultInternalRegistry.project}/daytona-${hash}:daytona`

        const exists = await this.readySnapshotRunnerExists(ref, regionId)

        if (exists) {
          state = SnapshotState.ACTIVE
        }
      }

      try {
        const snapshotId = uuidv4()

        const snapshot = this.snapshotRepository.create({
          id: snapshotId,
          organizationId: organization.id,
          ...createSnapshotDto,
          entrypoint: this.processEntrypoint(entrypoint),
          mem: createSnapshotDto.memory, // Map memory to mem
          state,
          ref,
          general,
          snapshotRegions: [{ snapshotId, regionId }],
        })

        await this.eventEmitter.emitAsync(SnapshotEvents.CREATED, new SnapshotCreatedEvent(snapshot))

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

  @Transactional()
  async createFromBuildInfo(organization: Organization, createSnapshotDto: CreateSnapshotDto, general = false) {
    if (!organization.defaultRegionId) {
      throw new DefaultRegionRequiredException()
    }

    const regionId = await this.getValidatedOrDefaultRegionId(organization, createSnapshotDto.regionId)

    let pendingSnapshotCountIncrement: number | undefined
    let entrypoint: string[] | undefined = undefined

    try {
      const nameValidationError = this.validateSnapshotName(createSnapshotDto.name)
      if (nameValidationError) {
        throw new BadRequestException(nameValidationError)
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

      entrypoint = this.getEntrypointFromDockerfile(createSnapshotDto.buildInfo.dockerfileContent)

      const snapshotId = uuidv4()

      const snapshot = this.snapshotRepository.create({
        id: snapshotId,
        organizationId: organization.id,
        ...createSnapshotDto,
        entrypoint: this.processEntrypoint(entrypoint),
        mem: createSnapshotDto.memory, // Map memory to mem
        state: SnapshotState.PENDING,
        general,
        snapshotRegions: [{ snapshotId, regionId }],
      })

      const buildSnapshotRef = generateBuildSnapshotRef(
        createSnapshotDto.buildInfo.dockerfileContent,
        createSnapshotDto.buildInfo.contextHashes,
      )

      // Check if buildInfo with the same snapshotRef already exists
      const existingBuildInfo = await this.buildInfoRepository.findOne({
        where: { snapshotRef: buildSnapshotRef },
      })

      if (
        existingBuildInfo &&
        // Update lastUsed once per minute at most
        (await this.redisLockProvider.lock(`build-info:${existingBuildInfo.snapshotRef}:update`, 60))
      ) {
        snapshot.buildInfo = existingBuildInfo
        existingBuildInfo.lastUsedAt = new Date()
        await this.buildInfoRepository.save(existingBuildInfo)
      } else {
        const buildInfoEntity = this.buildInfoRepository.create({
          ...createSnapshotDto.buildInfo,
        })
        await this.buildInfoRepository.save(buildInfoEntity)
        snapshot.buildInfo = buildInfoEntity
      }

      const defaultInternalRegistry = await this.dockerRegistryService.getDefaultInternalRegistry()
      snapshot.ref = `${defaultInternalRegistry.url.replace(/^(https?:\/\/)/, '')}/${defaultInternalRegistry.project}/${buildSnapshotRef}`

      const exists = await this.readySnapshotRunnerExists(snapshot.ref, regionId)

      if (exists) {
        snapshot.state = SnapshotState.ACTIVE
      }

      try {
        await this.eventEmitter.emitAsync(SnapshotEvents.CREATED, new SnapshotCreatedEvent(snapshot))

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
      relations: ['snapshotRegions'],
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

  async getBuildLogsUrl(snapshot: Snapshot): Promise<string> {
    if (!snapshot.initialRunnerId) {
      throw new NotFoundException(`Snapshot ${snapshot.id} has no initial runner`)
    }

    const runner = await this.runnerService.findOne(snapshot.initialRunnerId)
    if (!runner) {
      throw new NotFoundException(`Initial runner for snapshot ${snapshot.id} not found`)
    }

    const region = await this.regionService.findOne(runner.region, true)

    if (!region) {
      throw new NotFoundException(`Region for initial runner for snapshot ${snapshot.id} not found`)
    }

    if (!region.proxyUrl) {
      throw new NotFoundException(`Region for initial runner for snapshot ${snapshot.id} has no proxy URL`)
    }

    return region.proxyUrl + '/snapshots/' + snapshot.id + '/build-logs'
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

    // Update once per minute at most
    if (!(await this.redisLockProvider.lock(`snapshot:${event.sandbox.snapshot}:update-last-used`, 60))) {
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
        ref: imageName,
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
          ref: snapshot.ref,
        },
      })

      if (countActiveSnapshots === 0) {
        // Set associated SnapshotRunner records to REMOVING state
        const result = await this.snapshotRunnerRepository.update(
          { snapshotRef: snapshot.ref },
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

  // TODO: revise/cleanup
  getEntrypointFromDockerfile(dockerfileContent: string): string[] {
    // Match ENTRYPOINT with either a string or JSON array
    const entrypointMatch = dockerfileContent.match(/ENTRYPOINT\s+(.*)/)
    if (entrypointMatch) {
      const rawEntrypoint = entrypointMatch[1].trim()
      try {
        // Try parsing as JSON array
        const parsed = JSON.parse(rawEntrypoint)
        if (Array.isArray(parsed)) {
          return parsed
        }
      } catch {
        // Fallback: it's probably a plain string
        return [rawEntrypoint.replace(/["']/g, '')]
      }
    }

    return ['sleep', 'infinity']
  }

  /**
   * Validates and returns a region ID for snapshot availability.
   *
   * @param organization - The organization which is creating the snapshot.
   * @param regionId - The requested region ID. If omitted, the organization's default region is used.
   * @returns The validated region ID
   * @throws {NotFoundException} If the requested region is not available to the organization
   */
  private async getValidatedOrDefaultRegionId(organization: Organization, regionId?: string): Promise<string> {
    if (!regionId) {
      return organization.defaultRegionId
    }

    const region = await this.regionRepository.findOne({
      where: { id: regionId },
    })

    if (!region) {
      throw new NotFoundException('Region not found')
    }

    const availableRegions = await this.organizationService.listAvailableRegions(organization.id)

    if (!availableRegions.some((r) => r.id === regionId)) {
      if (region.regionType === RegionType.SHARED) {
        // region is public, but the organization does not have a quota for it
        throw new ForbiddenException(`Region ${regionId} is not available to the organization`)
      } else {
        // region is not public, respond as if the region was not found
        throw new NotFoundException('Region not found')
      }
    }

    return regionId
  }

  /**
   * @param snapshotId
   * @returns The regions where the snapshot is configured to be propagated to.
   */
  async getSnapshotRegions(snapshotId: string): Promise<Region[]> {
    return await this.regionRepository
      .createQueryBuilder('r')
      .innerJoin('snapshot_region', 'sr', 'sr."regionId" = r.id')
      .where('sr."snapshotId" = :snapshotId', { snapshotId })
      .getMany()
  }

  /**
   * @param snapshotId - The ID of the snapshot.
   * @param regionId - The ID of the region.
   * @returns true if the snapshot is available in the region, false otherwise.
   */
  async isAvailableInRegion(snapshotId: string, regionId: string): Promise<boolean> {
    return await this.snapshotRegionRepository.exists({
      where: {
        snapshotId,
        regionId,
      },
    })
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

  @OnAsyncEvent({
    event: RunnerEvents.DELETED,
  })
  async handleRunnerDeletedEvent(payload: RunnerDeletedEvent): Promise<void> {
    await payload.entityManager.update(
      SnapshotRunner,
      { runnerId: payload.runnerId },
      { state: SnapshotRunnerState.REMOVING },
    )
  }
}
