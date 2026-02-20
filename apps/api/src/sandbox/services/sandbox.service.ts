/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ForbiddenException, Injectable, Logger, NotFoundException, ConflictException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Not, Repository, LessThan, In, JsonContains, FindOptionsWhere, ILike } from 'typeorm'
import { Sandbox } from '../entities/sandbox.entity'
import { CreateSandboxDto } from '../dto/create-sandbox.dto'
import { ResizeSandboxDto } from '../dto/resize-sandbox.dto'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { RunnerService } from './runner.service'
import { SandboxError } from '../../exceptions/sandbox-error.exception'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { Cron, CronExpression } from '@nestjs/schedule'
import { BackupState } from '../enums/backup-state.enum'
import { Snapshot } from '../entities/snapshot.entity'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../constants/sandbox.constants'
import { SandboxWarmPoolService } from './sandbox-warm-pool.service'
import { EventEmitter2, OnEvent } from '@nestjs/event-emitter'
import { WarmPoolEvents } from '../constants/warmpool-events.constants'
import { WarmPoolTopUpRequested } from '../events/warmpool-topup-requested.event'
import { Runner } from '../entities/runner.entity'
import { Organization } from '../../organization/entities/organization.entity'
import { SandboxEvents } from '../constants/sandbox-events.constants'
import { SandboxStateUpdatedEvent } from '../events/sandbox-state-updated.event'
import { BuildInfo } from '../entities/build-info.entity'
import { generateBuildInfoHash as generateBuildSnapshotRef } from '../entities/build-info.entity'
import { SandboxBackupCreatedEvent } from '../events/sandbox-backup-created.event'
import { SandboxDestroyedEvent } from '../events/sandbox-destroyed.event'
import { SandboxStartedEvent } from '../events/sandbox-started.event'
import { SandboxStoppedEvent } from '../events/sandbox-stopped.event'
import { SandboxArchivedEvent } from '../events/sandbox-archived.event'
import { OrganizationService } from '../../organization/services/organization.service'
import { OrganizationEvents } from '../../organization/constants/organization-events.constant'
import { OrganizationSuspendedSandboxStoppedEvent } from '../../organization/events/organization-suspended-sandbox-stopped.event'
import { TypedConfigService } from '../../config/typed-config.service'
import { WarmPool } from '../entities/warm-pool.entity'
import { SandboxDto, SandboxVolume } from '../dto/sandbox.dto'
import { isValidUuid } from '../../common/utils/uuid'
import { RunnerAdapterFactory } from '../runner-adapter/runnerAdapter'
import { validateNetworkAllowList } from '../utils/network-validation.util'
import { OrganizationUsageService } from '../../organization/services/organization-usage.service'
import { SshAccess } from '../entities/ssh-access.entity'
import { SshAccessDto, SshAccessValidationDto } from '../dto/ssh-access.dto'
import { VolumeService } from './volume.service'
import { PaginatedList } from '../../common/interfaces/paginated-list.interface'
import {
  SandboxSortField,
  SandboxSortDirection,
  DEFAULT_SANDBOX_SORT_FIELD,
  DEFAULT_SANDBOX_SORT_DIRECTION,
} from '../dto/list-sandboxes-query.dto'
import { createRangeFilter } from '../../common/utils/range-filter'
import { LogExecution } from '../../common/decorators/log-execution.decorator'
import {
  UPGRADE_TIER_MESSAGE,
  ARCHIVE_SANDBOXES_MESSAGE,
  PER_SANDBOX_LIMIT_MESSAGE,
} from '../../common/constants/error-messages'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { customAlphabet as customNanoid, nanoid, urlAlphabet } from 'nanoid'
import { WithInstrumentation } from '../../common/decorators/otel.decorator'
import { validateMountPaths, validateSubpaths } from '../utils/volume-mount-path-validation.util'
import { SandboxRepository } from '../repositories/sandbox.repository'
import { PortPreviewUrlDto, SignedPortPreviewUrlDto } from '../dto/port-preview-url.dto'
import { RegionService } from '../../region/services/region.service'
import { DefaultRegionRequiredException } from '../../organization/exceptions/DefaultRegionRequiredException'
import { SnapshotService } from './snapshot.service'
import { RegionType } from '../../region/enums/region-type.enum'
import { SandboxCreatedEvent } from '../events/sandbox-create.event'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import {
  SANDBOX_LOOKUP_CACHE_TTL_MS,
  sandboxLookupCacheKeyById,
  sandboxLookupCacheKeyByName,
} from '../utils/sandbox-lookup-cache.util'
import { SandboxLookupCacheInvalidationService } from './sandbox-lookup-cache-invalidation.service'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { CheckpointService } from './checkpoint.service'
import { CheckpointState } from '../enums/checkpoint-state.enum'

const DEFAULT_CPU = 1
const DEFAULT_MEMORY = 1
const DEFAULT_DISK = 3
const DEFAULT_GPU = 0

@Injectable()
export class SandboxService {
  private readonly logger = new Logger(SandboxService.name)

  constructor(
    private readonly sandboxRepository: SandboxRepository,
    @InjectRepository(Snapshot)
    private readonly snapshotRepository: Repository<Snapshot>,
    @InjectRepository(Runner)
    private readonly runnerRepository: Repository<Runner>,
    @InjectRepository(BuildInfo)
    private readonly buildInfoRepository: Repository<BuildInfo>,
    @InjectRepository(SshAccess)
    private readonly sshAccessRepository: Repository<SshAccess>,
    private readonly runnerService: RunnerService,
    private readonly volumeService: VolumeService,
    private readonly configService: TypedConfigService,
    private readonly warmPoolService: SandboxWarmPoolService,
    private readonly eventEmitter: EventEmitter2,
    private readonly organizationService: OrganizationService,
    private readonly runnerAdapterFactory: RunnerAdapterFactory,
    private readonly organizationUsageService: OrganizationUsageService,
    private readonly redisLockProvider: RedisLockProvider,
    @InjectRedis() private readonly redis: Redis,
    private readonly regionService: RegionService,
    private readonly snapshotService: SnapshotService,
    private readonly sandboxLookupCacheInvalidationService: SandboxLookupCacheInvalidationService,
    private readonly dockerRegistryService: DockerRegistryService,
    private readonly checkpointService: CheckpointService,
  ) {}

  /**
   * Central place for Sandbox partial updates that also invalidates lookup cache.
   */
  async updateById(
    sandboxId: string,
    patch: Partial<Omit<Sandbox, 'id' | 'createdAt' | 'updatedAt'>>,
    sandboxInfo?: Pick<Sandbox, 'id' | 'name' | 'organizationId'>,
  ): Promise<void> {
    const result = await this.sandboxRepository.update(sandboxId, patch)
    if (!result.affected) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    // Best-effort cache invalidation (only when sandboxInfo is provided)
    if (!sandboxInfo) {
      return
    }

    const nextOrganizationId =
      typeof patch.organizationId === 'string' ? patch.organizationId : sandboxInfo.organizationId
    const nextName = typeof patch.name === 'string' ? patch.name : sandboxInfo.name

    this.sandboxLookupCacheInvalidationService.invalidate({
      sandboxId: sandboxInfo.id,
      organizationId: nextOrganizationId,
      name: nextName,
      previousOrganizationId: sandboxInfo.organizationId,
      previousName: sandboxInfo.name,
    })
  }

  protected getLockKey(id: string): string {
    return `sandbox:${id}:state-change`
  }

  private assertSandboxNotErrored(sandbox: Sandbox): void {
    if ([SandboxState.ERROR, SandboxState.BUILD_FAILED].includes(sandbox.state)) {
      throw new SandboxError('Sandbox is in an errored state')
    }
  }

  private async validateOrganizationQuotas(
    organization: Organization,
    regionId: string,
    cpu: number,
    memory: number,
    disk: number,
    excludeSandboxId?: string,
  ): Promise<{
    pendingCpuIncremented: boolean
    pendingMemoryIncremented: boolean
    pendingDiskIncremented: boolean
  }> {
    // validate per-sandbox quotas
    if (cpu > organization.maxCpuPerSandbox) {
      throw new ForbiddenException(
        `CPU request ${cpu} exceeds maximum allowed per sandbox (${organization.maxCpuPerSandbox}).\n${PER_SANDBOX_LIMIT_MESSAGE}`,
      )
    }
    if (memory > organization.maxMemoryPerSandbox) {
      throw new ForbiddenException(
        `Memory request ${memory}GB exceeds maximum allowed per sandbox (${organization.maxMemoryPerSandbox}GB).\n${PER_SANDBOX_LIMIT_MESSAGE}`,
      )
    }
    if (disk > organization.maxDiskPerSandbox) {
      throw new ForbiddenException(
        `Disk request ${disk}GB exceeds maximum allowed per sandbox (${organization.maxDiskPerSandbox}GB).\n${PER_SANDBOX_LIMIT_MESSAGE}`,
      )
    }

    const region = await this.regionService.findOne(regionId)
    if (!region) {
      throw new NotFoundException('Region not found')
    }

    // e.g. region belonging to an organization
    if (!region.enforceQuotas) {
      return {
        pendingCpuIncremented: false,
        pendingMemoryIncremented: false,
        pendingDiskIncremented: false,
      }
    }

    const regionQuota = await this.organizationService.getRegionQuota(organization.id, regionId)

    if (!regionQuota) {
      if (region.regionType === RegionType.SHARED) {
        // region is public, but the organization does not have a quota for it
        throw new ForbiddenException(`Region ${regionId} is not available to the organization`)
      } else {
        // region is not public, respond as if the region was not found
        throw new NotFoundException('Region not found')
      }
    }

    // validate usage quotas
    const {
      cpuIncremented: pendingCpuIncremented,
      memoryIncremented: pendingMemoryIncremented,
      diskIncremented: pendingDiskIncremented,
    } = await this.organizationUsageService.incrementPendingSandboxUsage(
      organization.id,
      regionId,
      cpu,
      memory,
      disk,
      excludeSandboxId,
    )

    const usageOverview = await this.organizationUsageService.getSandboxUsageOverview(
      organization.id,
      regionId,
      excludeSandboxId,
    )

    try {
      const upgradeTierMessage = UPGRADE_TIER_MESSAGE(this.configService.getOrThrow('dashboardUrl'))

      if (usageOverview.currentCpuUsage + usageOverview.pendingCpuUsage > regionQuota.totalCpuQuota) {
        throw new ForbiddenException(
          `Total CPU limit exceeded. Maximum allowed: ${regionQuota.totalCpuQuota}.\n${upgradeTierMessage}`,
        )
      }

      if (usageOverview.currentMemoryUsage + usageOverview.pendingMemoryUsage > regionQuota.totalMemoryQuota) {
        throw new ForbiddenException(
          `Total memory limit exceeded. Maximum allowed: ${regionQuota.totalMemoryQuota}GiB.\n${upgradeTierMessage}`,
        )
      }

      if (usageOverview.currentDiskUsage + usageOverview.pendingDiskUsage > regionQuota.totalDiskQuota) {
        throw new ForbiddenException(
          `Total disk limit exceeded. Maximum allowed: ${regionQuota.totalDiskQuota}GiB.\n${ARCHIVE_SANDBOXES_MESSAGE}\n${upgradeTierMessage}`,
        )
      }
    } catch (error) {
      await this.rollbackPendingUsage(
        organization.id,
        regionId,
        pendingCpuIncremented ? cpu : undefined,
        pendingMemoryIncremented ? memory : undefined,
        pendingDiskIncremented ? disk : undefined,
      )
      throw error
    }

    return {
      pendingCpuIncremented,
      pendingMemoryIncremented,
      pendingDiskIncremented,
    }
  }

  async rollbackPendingUsage(
    organizationId: string,
    regionId: string,
    pendingCpuIncrement?: number,
    pendingMemoryIncrement?: number,
    pendingDiskIncrement?: number,
  ): Promise<void> {
    if (!pendingCpuIncrement && !pendingMemoryIncrement && !pendingDiskIncrement) {
      return
    }

    try {
      await this.organizationUsageService.decrementPendingSandboxUsage(
        organizationId,
        regionId,
        pendingCpuIncrement,
        pendingMemoryIncrement,
        pendingDiskIncrement,
      )
    } catch (error) {
      this.logger.error(`Error rolling back pending sandbox usage: ${error}`)
    }
  }

  async archive(sandboxIdOrName: string, organizationId?: string): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    this.assertSandboxNotErrored(sandbox)

    if (String(sandbox.state) !== String(sandbox.desiredState)) {
      throw new SandboxError('State change in progress')
    }

    if (sandbox.state !== SandboxState.STOPPED) {
      throw new SandboxError('Sandbox is not stopped')
    }

    if (sandbox.pending) {
      throw new SandboxError('Sandbox state change in progress')
    }

    if (sandbox.autoDeleteInterval === 0) {
      throw new SandboxError('Ephemeral sandboxes cannot be archived')
    }

    sandbox.state = SandboxState.ARCHIVING
    sandbox.desiredState = SandboxDesiredState.ARCHIVED
    await this.sandboxRepository.saveWhere(sandbox, { pending: false, state: SandboxState.STOPPED })

    this.eventEmitter.emit(SandboxEvents.ARCHIVED, new SandboxArchivedEvent(sandbox))
    return sandbox
  }

  async createForWarmPool(warmPoolItem: WarmPool): Promise<Sandbox> {
    const sandbox = new Sandbox(warmPoolItem.target)

    sandbox.organizationId = SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION

    sandbox.class = warmPoolItem.class
    sandbox.snapshot = warmPoolItem.snapshot
    //  TODO: default user should be configurable
    sandbox.osUser = 'daytona'
    sandbox.env = warmPoolItem.env || {}

    sandbox.cpu = warmPoolItem.cpu
    sandbox.gpu = warmPoolItem.gpu
    sandbox.mem = warmPoolItem.mem
    sandbox.disk = warmPoolItem.disk

    const snapshot = await this.snapshotRepository.findOne({
      where: [
        { organizationId: sandbox.organizationId, name: sandbox.snapshot, state: SnapshotState.ACTIVE },
        { general: true, name: sandbox.snapshot, state: SnapshotState.ACTIVE },
      ],
    })
    if (!snapshot) {
      throw new BadRequestError(`Snapshot ${sandbox.snapshot} not found while creating warm pool sandbox`)
    }

    const runner = await this.runnerService.getRandomAvailableRunner({
      regions: [sandbox.region],
      sandboxClass: sandbox.class,
      snapshotRef: snapshot.ref,
    })

    sandbox.runnerId = runner.id
    sandbox.pending = true

    await this.sandboxRepository.insert(sandbox)
    return sandbox
  }

  async createFromSnapshot(
    createSandboxDto: CreateSandboxDto,
    organization: Organization,
    useSandboxResourceParams_deprecated?: boolean,
  ): Promise<SandboxDto> {
    let pendingCpuIncrement: number | undefined
    let pendingMemoryIncrement: number | undefined
    let pendingDiskIncrement: number | undefined

    const regionId = await this.getValidatedOrDefaultRegionId(organization, createSandboxDto.target)

    try {
      const sandboxClass = this.getValidatedOrDefaultClass(createSandboxDto.class)

      let snapshotIdOrName = createSandboxDto.snapshot

      if (!createSandboxDto.snapshot?.trim()) {
        snapshotIdOrName = this.configService.getOrThrow('defaultSnapshot')
      }

      const snapshotFilter: FindOptionsWhere<Snapshot>[] = [
        { organizationId: organization.id, name: snapshotIdOrName },
        { general: true, name: snapshotIdOrName },
      ]

      if (isValidUuid(snapshotIdOrName)) {
        snapshotFilter.push(
          { organizationId: organization.id, id: snapshotIdOrName },
          { general: true, id: snapshotIdOrName },
        )
      }

      const snapshots = await this.snapshotRepository.find({
        where: snapshotFilter,
      })

      if (snapshots.length === 0) {
        throw new BadRequestError(
          `Snapshot ${snapshotIdOrName} not found. Did you add it through the Daytona Dashboard?`,
        )
      }

      let snapshot = snapshots.find((s) => s.state === SnapshotState.ACTIVE)

      if (!snapshot) {
        snapshot = snapshots[0]
      }

      if (!(await this.snapshotService.isAvailableInRegion(snapshot.id, regionId))) {
        throw new BadRequestError(`Snapshot ${snapshotIdOrName} is not available in region ${regionId}`)
      }

      if (snapshot.state !== SnapshotState.ACTIVE) {
        throw new BadRequestError(`Snapshot ${snapshotIdOrName} is ${snapshot.state}`)
      }

      if (!snapshot.ref) {
        throw new BadRequestError('Snapshot ref is not defined')
      }

      let cpu = snapshot.cpu
      let mem = snapshot.mem
      let disk = snapshot.disk
      let gpu = snapshot.gpu

      // Remove the deprecated behavior in a future release
      if (useSandboxResourceParams_deprecated) {
        if (createSandboxDto.cpu) {
          cpu = createSandboxDto.cpu
        }
        if (createSandboxDto.memory) {
          mem = createSandboxDto.memory
        }
        if (createSandboxDto.disk) {
          disk = createSandboxDto.disk
        }
        if (createSandboxDto.gpu) {
          gpu = createSandboxDto.gpu
        }
      }

      this.organizationService.assertOrganizationIsNotSuspended(organization)

      const { pendingCpuIncremented, pendingMemoryIncremented, pendingDiskIncremented } =
        await this.validateOrganizationQuotas(organization, regionId, cpu, mem, disk)

      if (pendingCpuIncremented) {
        pendingCpuIncrement = cpu
      }
      if (pendingMemoryIncremented) {
        pendingMemoryIncrement = mem
      }
      if (pendingDiskIncremented) {
        pendingDiskIncrement = disk
      }

      if (!createSandboxDto.volumes || createSandboxDto.volumes.length === 0) {
        const warmPoolSandbox = await this.warmPoolService.fetchWarmPoolSandbox({
          organizationId: organization.id,
          snapshot: snapshotIdOrName,
          target: regionId,
          class: createSandboxDto.class,
          cpu: cpu,
          mem: mem,
          disk: disk,
          gpu: gpu,
          osUser: createSandboxDto.user,
          env: createSandboxDto.env,
          state: SandboxState.STARTED,
        })

        if (warmPoolSandbox) {
          return await this.assignWarmPoolSandbox(warmPoolSandbox, createSandboxDto, organization)
        }
      } else {
        const volumeIdOrNames = createSandboxDto.volumes.map((v) => v.volumeId)
        await this.volumeService.validateVolumes(organization.id, volumeIdOrNames)
      }

      const runner = await this.runnerService.getRandomAvailableRunner({
        regions: [regionId],
        sandboxClass,
        snapshotRef: snapshot.ref,
      })

      const sandbox = new Sandbox(regionId, createSandboxDto.name)

      sandbox.organizationId = organization.id

      //  TODO: make configurable
      sandbox.class = sandboxClass
      sandbox.snapshot = snapshot.name
      //  TODO: default user should be configurable
      sandbox.osUser = createSandboxDto.user || 'daytona'
      sandbox.env = createSandboxDto.env || {}
      sandbox.labels = createSandboxDto.labels || {}

      sandbox.cpu = cpu
      sandbox.gpu = gpu
      sandbox.mem = mem
      sandbox.disk = disk

      sandbox.public = createSandboxDto.public || false

      if (createSandboxDto.networkBlockAll !== undefined) {
        sandbox.networkBlockAll = createSandboxDto.networkBlockAll
      }

      if (createSandboxDto.networkAllowList !== undefined) {
        sandbox.networkAllowList = this.resolveNetworkAllowList(createSandboxDto.networkAllowList)
      }

      if (createSandboxDto.autoStopInterval !== undefined) {
        sandbox.autoStopInterval = this.resolveAutoStopInterval(createSandboxDto.autoStopInterval)
      }

      if (createSandboxDto.autoArchiveInterval !== undefined) {
        sandbox.autoArchiveInterval = this.resolveAutoArchiveInterval(createSandboxDto.autoArchiveInterval)
      }

      if (createSandboxDto.autoDeleteInterval !== undefined) {
        sandbox.autoDeleteInterval = createSandboxDto.autoDeleteInterval
      }

      if (createSandboxDto.volumes !== undefined) {
        sandbox.volumes = this.resolveVolumes(createSandboxDto.volumes)
      }

      sandbox.runnerId = runner.id
      sandbox.pending = true

      await this.sandboxRepository.insert(sandbox)

      this.eventEmitter.emit(SandboxEvents.CREATED, new SandboxCreatedEvent(sandbox))

      return SandboxDto.fromSandbox(sandbox)
    } catch (error) {
      await this.rollbackPendingUsage(
        organization.id,
        regionId,
        pendingCpuIncrement,
        pendingMemoryIncrement,
        pendingDiskIncrement,
      )

      if (error.code === '23505') {
        throw new ConflictException(`Sandbox with name ${createSandboxDto.name} already exists`)
      }

      throw error
    }
  }

  private async assignWarmPoolSandbox(
    warmPoolSandbox: Sandbox,
    createSandboxDto: CreateSandboxDto,
    organization: Organization,
  ): Promise<SandboxDto> {
    if (createSandboxDto.name) {
      warmPoolSandbox.name = createSandboxDto.name
    }

    warmPoolSandbox.public = createSandboxDto.public || false
    warmPoolSandbox.labels = createSandboxDto.labels || {}
    warmPoolSandbox.organizationId = organization.id
    warmPoolSandbox.createdAt = new Date()

    if (createSandboxDto.autoStopInterval !== undefined) {
      warmPoolSandbox.autoStopInterval = this.resolveAutoStopInterval(createSandboxDto.autoStopInterval)
    }

    if (createSandboxDto.autoArchiveInterval !== undefined) {
      warmPoolSandbox.autoArchiveInterval = this.resolveAutoArchiveInterval(createSandboxDto.autoArchiveInterval)
    }

    if (createSandboxDto.autoDeleteInterval !== undefined) {
      warmPoolSandbox.autoDeleteInterval = createSandboxDto.autoDeleteInterval
    }

    if (createSandboxDto.networkBlockAll !== undefined) {
      warmPoolSandbox.networkBlockAll = createSandboxDto.networkBlockAll
    }
    if (createSandboxDto.networkAllowList !== undefined) {
      warmPoolSandbox.networkAllowList = this.resolveNetworkAllowList(createSandboxDto.networkAllowList)
    }

    if (!warmPoolSandbox.runnerId) {
      throw new SandboxError('Runner not found for warm pool sandbox')
    }

    if (
      createSandboxDto.networkBlockAll !== undefined ||
      createSandboxDto.networkAllowList !== undefined ||
      organization.sandboxLimitedNetworkEgress
    ) {
      const runner = await this.runnerService.findOne(warmPoolSandbox.runnerId)
      if (!runner) {
        throw new NotFoundException(`Runner with ID ${warmPoolSandbox.runnerId} not found`)
      }

      const runnerAdapter = await this.runnerAdapterFactory.create(runner)
      await runnerAdapter.updateNetworkSettings(
        warmPoolSandbox.id,
        createSandboxDto.networkBlockAll,
        createSandboxDto.networkAllowList,
        organization.sandboxLimitedNetworkEgress,
      )
    }

    const result = await this.sandboxRepository.save(warmPoolSandbox)

    // Treat this as a newly started sandbox
    this.eventEmitter.emit(
      SandboxEvents.STATE_UPDATED,
      new SandboxStateUpdatedEvent(warmPoolSandbox, SandboxState.STARTED, SandboxState.STARTED),
    )
    return SandboxDto.fromSandbox(result)
  }

  async createFromBuildInfo(createSandboxDto: CreateSandboxDto, organization: Organization): Promise<SandboxDto> {
    let pendingCpuIncrement: number | undefined
    let pendingMemoryIncrement: number | undefined
    let pendingDiskIncrement: number | undefined

    const regionId = await this.getValidatedOrDefaultRegionId(organization, createSandboxDto.target)

    try {
      const sandboxClass = this.getValidatedOrDefaultClass(createSandboxDto.class)

      const cpu = createSandboxDto.cpu || DEFAULT_CPU
      const mem = createSandboxDto.memory || DEFAULT_MEMORY
      const disk = createSandboxDto.disk || DEFAULT_DISK
      const gpu = createSandboxDto.gpu || DEFAULT_GPU

      this.organizationService.assertOrganizationIsNotSuspended(organization)

      const { pendingCpuIncremented, pendingMemoryIncremented, pendingDiskIncremented } =
        await this.validateOrganizationQuotas(organization, regionId, cpu, mem, disk)

      if (pendingCpuIncremented) {
        pendingCpuIncrement = cpu
      }
      if (pendingMemoryIncremented) {
        pendingMemoryIncrement = mem
      }
      if (pendingDiskIncremented) {
        pendingDiskIncrement = disk
      }

      if (createSandboxDto.volumes && createSandboxDto.volumes.length > 0) {
        const volumeIdOrNames = createSandboxDto.volumes.map((v) => v.volumeId)
        await this.volumeService.validateVolumes(organization.id, volumeIdOrNames)
      }

      const sandbox = new Sandbox(regionId, createSandboxDto.name)

      sandbox.organizationId = organization.id

      sandbox.class = sandboxClass
      sandbox.osUser = createSandboxDto.user || 'daytona'
      sandbox.env = createSandboxDto.env || {}
      sandbox.labels = createSandboxDto.labels || {}

      sandbox.cpu = cpu
      sandbox.gpu = gpu
      sandbox.mem = mem
      sandbox.disk = disk
      sandbox.public = createSandboxDto.public || false

      if (createSandboxDto.networkBlockAll !== undefined) {
        sandbox.networkBlockAll = createSandboxDto.networkBlockAll
      }

      if (createSandboxDto.networkAllowList !== undefined) {
        sandbox.networkAllowList = this.resolveNetworkAllowList(createSandboxDto.networkAllowList)
      }

      if (createSandboxDto.autoStopInterval !== undefined) {
        sandbox.autoStopInterval = this.resolveAutoStopInterval(createSandboxDto.autoStopInterval)
      }

      if (createSandboxDto.autoArchiveInterval !== undefined) {
        sandbox.autoArchiveInterval = this.resolveAutoArchiveInterval(createSandboxDto.autoArchiveInterval)
      }

      if (createSandboxDto.autoDeleteInterval !== undefined) {
        sandbox.autoDeleteInterval = createSandboxDto.autoDeleteInterval
      }

      if (createSandboxDto.volumes !== undefined) {
        sandbox.volumes = this.resolveVolumes(createSandboxDto.volumes)
      }

      const buildInfoSnapshotRef = generateBuildSnapshotRef(
        createSandboxDto.buildInfo.dockerfileContent,
        createSandboxDto.buildInfo.contextHashes,
      )

      // Check if buildInfo with the same snapshotRef already exists
      const existingBuildInfo = await this.buildInfoRepository.findOne({
        where: { snapshotRef: buildInfoSnapshotRef },
      })

      if (existingBuildInfo) {
        sandbox.buildInfo = existingBuildInfo
        if (await this.redisLockProvider.lock(`build-info:${existingBuildInfo.snapshotRef}:update`, 60)) {
          await this.buildInfoRepository.update(sandbox.buildInfo.snapshotRef, { lastUsedAt: new Date() })
        }
      } else {
        const buildInfoEntity = this.buildInfoRepository.create({
          ...createSandboxDto.buildInfo,
        })
        await this.buildInfoRepository.save(buildInfoEntity)
        sandbox.buildInfo = buildInfoEntity
      }

      let runner: Runner

      try {
        const declarativeBuildScoreThreshold = this.configService.get('runnerScore.thresholds.declarativeBuild')
        runner = await this.runnerService.getRandomAvailableRunner({
          regions: [sandbox.region],
          sandboxClass: sandbox.class,
          snapshotRef: sandbox.buildInfo.snapshotRef,
          ...(declarativeBuildScoreThreshold !== undefined && {
            availabilityScoreThreshold: declarativeBuildScoreThreshold,
          }),
        })
        sandbox.runnerId = runner.id
      } catch (error) {
        if (
          error instanceof BadRequestError == false ||
          error.message !== 'No available runners' ||
          !sandbox.buildInfo
        ) {
          throw error
        }
        sandbox.state = SandboxState.PENDING_BUILD
      }

      sandbox.pending = true

      await this.sandboxRepository.insert(sandbox)

      this.eventEmitter.emit(SandboxEvents.CREATED, new SandboxCreatedEvent(sandbox))

      return SandboxDto.fromSandbox(sandbox)
    } catch (error) {
      await this.rollbackPendingUsage(
        organization.id,
        regionId,
        pendingCpuIncrement,
        pendingMemoryIncrement,
        pendingDiskIncrement,
      )

      if (error.code === '23505') {
        throw new ConflictException(`Sandbox with name ${createSandboxDto.name} already exists`)
      }

      throw error
    }
  }

  async createBackup(sandboxIdOrName: string, organizationId?: string): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    if (sandbox.autoDeleteInterval === 0) {
      throw new SandboxError('Ephemeral sandboxes cannot be backed up')
    }

    if (![BackupState.COMPLETED, BackupState.NONE].includes(sandbox.backupState)) {
      throw new SandboxError('Sandbox backup is already in progress')
    }

    this.eventEmitter.emit(SandboxEvents.BACKUP_CREATED, new SandboxBackupCreatedEvent(sandbox))

    return sandbox
  }

  async createCheckpoint(
    sandboxIdOrName: string,
    organizationId: string,
    checkpointName: string,
  ): Promise<SandboxDto> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    if (sandbox.pending) {
      throw new SandboxError('Sandbox state change in progress')
    }

    if (![SandboxState.STARTED, SandboxState.STOPPED].includes(sandbox.state)) {
      throw new SandboxError('Sandbox must be started or stopped to create a checkpoint')
    }

    if (!sandbox.runnerId) {
      throw new SandboxError('Sandbox has no runner assigned')
    }

    const runner = await this.runnerRepository.findOne({
      where: { id: sandbox.runnerId },
    })

    if (!runner) {
      throw new SandboxError('Runner not found')
    }

    const internalRegistry = await this.dockerRegistryService.getAvailableInternalRegistry(runner.region)
    if (!internalRegistry) {
      throw new SandboxError('No internal registry available')
    }

    const checkpoint = await this.checkpointService.createPending(sandbox, checkpointName, sandbox.runnerId)

    sandbox.pending = true
    await this.sandboxRepository.saveWhere(sandbox, { pending: false })

    try {
      const runnerAdapter = await this.runnerAdapterFactory.create(runner)
      await runnerAdapter.createCheckpoint(sandbox.id, checkpointName, organizationId, internalRegistry)

      sandbox.checkpoint = checkpoint
      sandbox.pending = false
      await this.sandboxRepository.save(sandbox)

      return SandboxDto.fromSandbox(sandbox)
    } catch (error) {
      sandbox.pending = false
      await this.sandboxRepository.save(sandbox)
      await this.checkpointService.markError(checkpoint.id, error.message || 'Failed to create checkpoint')
      throw error
    }
  }

  async createFromCheckpoint(
    checkpointId: string,
    organization: Organization,
    name?: string,
  ): Promise<SandboxDto> {
    let pendingCpuIncrement: number | undefined
    let pendingMemoryIncrement: number | undefined
    let pendingDiskIncrement: number | undefined

    const checkpoint = await this.checkpointService.getCheckpoint(checkpointId, organization.id)

    if (checkpoint.state !== CheckpointState.ACTIVE) {
      throw new BadRequestError(`Checkpoint is not active (state: ${checkpoint.state})`)
    }

    if (!checkpoint.ref) {
      throw new BadRequestError('Checkpoint has no image ref')
    }

    const regionId = checkpoint.region

    try {
      this.organizationService.assertOrganizationIsNotSuspended(organization)

      const { pendingCpuIncremented, pendingMemoryIncremented, pendingDiskIncremented } =
        await this.validateOrganizationQuotas(organization, regionId, checkpoint.cpu, checkpoint.mem, checkpoint.disk)

      if (pendingCpuIncremented) {
        pendingCpuIncrement = checkpoint.cpu
      }
      if (pendingMemoryIncremented) {
        pendingMemoryIncrement = checkpoint.mem
      }
      if (pendingDiskIncremented) {
        pendingDiskIncrement = checkpoint.disk
      }

      const sandboxClass = checkpoint.class

      const runner = await this.runnerService.getRandomAvailableRunner({
        regions: [regionId],
        sandboxClass,
        snapshotRef: checkpoint.ref,
      })

      const sandbox = new Sandbox(regionId, name)
      sandbox.organizationId = organization.id
      sandbox.class = sandboxClass
      sandbox.snapshot = checkpoint.ref
      sandbox.checkpoint = checkpoint
      sandbox.osUser = checkpoint.osUser
      sandbox.env = checkpoint.env || {}
      sandbox.labels = checkpoint.labels || {}
      sandbox.cpu = checkpoint.cpu
      sandbox.gpu = checkpoint.gpu
      sandbox.mem = checkpoint.mem
      sandbox.disk = checkpoint.disk
      sandbox.public = checkpoint.public
      sandbox.networkBlockAll = checkpoint.networkBlockAll
      sandbox.networkAllowList = checkpoint.networkAllowList
      sandbox.autoStopInterval = checkpoint.autoStopInterval
      sandbox.autoArchiveInterval = checkpoint.autoArchiveInterval
      sandbox.autoDeleteInterval = checkpoint.autoDeleteInterval
      sandbox.volumes = checkpoint.volumes || []
      sandbox.runnerId = runner.id
      sandbox.pending = true

      await this.sandboxRepository.insert(sandbox)

      this.eventEmitter.emit(SandboxEvents.CREATED, new SandboxCreatedEvent(sandbox))

      return SandboxDto.fromSandbox(sandbox)
    } catch (error) {
      await this.rollbackPendingUsage(
        organization.id,
        regionId,
        pendingCpuIncrement,
        pendingMemoryIncrement,
        pendingDiskIncrement,
      )

      if (error.code === '23505') {
        throw new ConflictException(`Sandbox with that name already exists`)
      }

      throw error
    }
  }

  async findAllDeprecated(
    organizationId: string,
    labels?: { [key: string]: string },
    includeErroredDestroyed?: boolean,
  ): Promise<Sandbox[]> {
    const baseFindOptions: FindOptionsWhere<Sandbox> = {
      organizationId,
      ...(labels ? { labels: JsonContains(labels) } : {}),
    }

    const where: FindOptionsWhere<Sandbox>[] = [
      {
        ...baseFindOptions,
        state: Not(In([SandboxState.DESTROYED, SandboxState.ERROR, SandboxState.BUILD_FAILED])),
      },
      {
        ...baseFindOptions,
        state: In([SandboxState.ERROR, SandboxState.BUILD_FAILED]),
        ...(includeErroredDestroyed ? {} : { desiredState: Not(SandboxDesiredState.DESTROYED) }),
      },
    ]

    return this.sandboxRepository.find({ where })
  }

  async findAll(
    organizationId: string,
    page = 1,
    limit = 10,
    filters?: {
      id?: string
      name?: string
      labels?: { [key: string]: string }
      includeErroredDestroyed?: boolean
      states?: SandboxState[]
      snapshots?: string[]
      regionIds?: string[]
      minCpu?: number
      maxCpu?: number
      minMemoryGiB?: number
      maxMemoryGiB?: number
      minDiskGiB?: number
      maxDiskGiB?: number
      lastEventAfter?: Date
      lastEventBefore?: Date
    },
    sort?: {
      field?: SandboxSortField
      direction?: SandboxSortDirection
    },
  ): Promise<PaginatedList<Sandbox>> {
    const pageNum = Number(page)
    const limitNum = Number(limit)

    const {
      id,
      name,
      labels,
      includeErroredDestroyed,
      states,
      snapshots,
      regionIds,
      minCpu,
      maxCpu,
      minMemoryGiB,
      maxMemoryGiB,
      minDiskGiB,
      maxDiskGiB,
      lastEventAfter,
      lastEventBefore,
    } = filters || {}

    const { field: sortField = DEFAULT_SANDBOX_SORT_FIELD, direction: sortDirection = DEFAULT_SANDBOX_SORT_DIRECTION } =
      sort || {}

    const baseFindOptions: FindOptionsWhere<Sandbox> = {
      organizationId,
      ...(id ? { id: ILike(`${id}%`) } : {}),
      ...(name ? { name: ILike(`${name}%`) } : {}),
      ...(labels ? { labels: JsonContains(labels) } : {}),
      ...(snapshots ? { snapshot: In(snapshots) } : {}),
      ...(regionIds ? { region: In(regionIds) } : {}),
    }

    baseFindOptions.cpu = createRangeFilter(minCpu, maxCpu)
    baseFindOptions.mem = createRangeFilter(minMemoryGiB, maxMemoryGiB)
    baseFindOptions.disk = createRangeFilter(minDiskGiB, maxDiskGiB)
    baseFindOptions.lastActivityAt = createRangeFilter(lastEventAfter, lastEventBefore)

    const statesToInclude = (states || Object.values(SandboxState)).filter((state) => state !== SandboxState.DESTROYED)
    const errorStates = [SandboxState.ERROR, SandboxState.BUILD_FAILED]

    const nonErrorStatesToInclude = statesToInclude.filter((state) => !errorStates.includes(state))
    const errorStatesToInclude = statesToInclude.filter((state) => errorStates.includes(state))

    const where: FindOptionsWhere<Sandbox>[] = []

    if (nonErrorStatesToInclude.length > 0) {
      where.push({
        ...baseFindOptions,
        state: In(nonErrorStatesToInclude),
      })
    }

    if (errorStatesToInclude.length > 0) {
      where.push({
        ...baseFindOptions,
        state: In(errorStatesToInclude),
        ...(includeErroredDestroyed ? {} : { desiredState: Not(SandboxDesiredState.DESTROYED) }),
      })
    }

    const [items, total] = await this.sandboxRepository.findAndCount({
      where,
      order: {
        [sortField]: {
          direction: sortDirection,
          nulls: 'LAST',
        },
        ...(sortField !== SandboxSortField.CREATED_AT && { createdAt: 'DESC' }),
      },
      skip: (pageNum - 1) * limitNum,
      take: limitNum,
    })

    return {
      items,
      total,
      page: pageNum,
      totalPages: Math.ceil(total / limitNum),
    }
  }

  private getExpectedDesiredStateForState(state: SandboxState): SandboxDesiredState | undefined {
    switch (state) {
      case SandboxState.STARTED:
        return SandboxDesiredState.STARTED
      case SandboxState.STOPPED:
        return SandboxDesiredState.STOPPED
      case SandboxState.ARCHIVED:
        return SandboxDesiredState.ARCHIVED
      case SandboxState.DESTROYED:
        return SandboxDesiredState.DESTROYED
      default:
        return undefined
    }
  }

  private hasValidDesiredState(state: SandboxState): boolean {
    return this.getExpectedDesiredStateForState(state) !== undefined
  }

  async findByRunnerId(
    runnerId: string,
    states?: SandboxState[],
    skipReconcilingSandboxes?: boolean,
  ): Promise<Sandbox[]> {
    const where: FindOptionsWhere<Sandbox> = { runnerId }
    if (states && states.length > 0) {
      // Validate that all states have corresponding desired states
      states.forEach((state) => {
        if (!this.hasValidDesiredState(state)) {
          throw new BadRequestError(`State ${state} does not have a corresponding desired state`)
        }
      })
      where.state = In(states)
    }

    let sandboxes = await this.sandboxRepository.find({ where })

    if (skipReconcilingSandboxes) {
      sandboxes = sandboxes.filter((sandbox) => {
        const expectedDesiredState = this.getExpectedDesiredStateForState(sandbox.state)
        return expectedDesiredState !== undefined && expectedDesiredState === sandbox.desiredState
      })
    }

    return sandboxes
  }

  async findOneByIdOrName(
    sandboxIdOrName: string,
    organizationId: string,
    returnDestroyed?: boolean,
  ): Promise<Sandbox> {
    const stateFilter = returnDestroyed ? {} : { state: Not(SandboxState.DESTROYED) }
    const relations: ['buildInfo'] = ['buildInfo']

    // Try lookup by ID first
    let sandbox = await this.sandboxRepository.findOne({
      where: {
        id: sandboxIdOrName,
        organizationId,
        ...stateFilter,
      },
      relations,
      cache: {
        id: sandboxLookupCacheKeyById({ organizationId, returnDestroyed, sandboxId: sandboxIdOrName }),
        milliseconds: SANDBOX_LOOKUP_CACHE_TTL_MS,
      },
    })

    // Fallback to lookup by name
    if (!sandbox) {
      sandbox = await this.sandboxRepository.findOne({
        where: {
          name: sandboxIdOrName,
          organizationId,
          ...stateFilter,
        },
        relations,
        cache: {
          id: sandboxLookupCacheKeyByName({ organizationId, returnDestroyed, sandboxName: sandboxIdOrName }),
          milliseconds: SANDBOX_LOOKUP_CACHE_TTL_MS,
        },
      })
    }

    if (
      !sandbox ||
      (!returnDestroyed &&
        [SandboxState.ERROR, SandboxState.BUILD_FAILED].includes(sandbox.state) &&
        sandbox.desiredState === SandboxDesiredState.DESTROYED)
    ) {
      throw new NotFoundException(`Sandbox with ID or name ${sandboxIdOrName} not found`)
    }

    return sandbox
  }

  async findOne(sandboxId: string, returnDestroyed?: boolean): Promise<Sandbox> {
    const sandbox = await this.sandboxRepository.findOne({
      where: {
        id: sandboxId,
        ...(returnDestroyed ? {} : { state: Not(SandboxState.DESTROYED) }),
      },
    })

    if (
      !sandbox ||
      (!returnDestroyed &&
        [SandboxState.ERROR, SandboxState.BUILD_FAILED].includes(sandbox.state) &&
        sandbox.desiredState === SandboxDesiredState.DESTROYED)
    ) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    return sandbox
  }

  async getOrganizationId(sandboxIdOrName: string, organizationId?: string): Promise<string> {
    let sandbox = await this.sandboxRepository.findOne({
      where: {
        id: sandboxIdOrName,
        ...(organizationId ? { organizationId: organizationId } : {}),
      },
      select: ['organizationId'],
      loadEagerRelations: false,
    })

    if (!sandbox && organizationId) {
      sandbox = await this.sandboxRepository.findOne({
        where: {
          name: sandboxIdOrName,
          organizationId: organizationId,
        },
        select: ['organizationId'],
        loadEagerRelations: false,
      })
    }

    if (!sandbox || !sandbox.organizationId) {
      throw new NotFoundException(`Sandbox with ID or name ${sandboxIdOrName} not found`)
    }

    return sandbox.organizationId
  }

  async getRunnerId(sandboxId: string): Promise<string | null> {
    const sandbox = await this.sandboxRepository.findOne({
      where: {
        id: sandboxId,
      },
      select: ['runnerId'],
      loadEagerRelations: false,
    })

    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    return sandbox.runnerId || null
  }

  async getRegionId(sandboxId: string): Promise<string> {
    const sandbox = await this.sandboxRepository.findOne({
      where: {
        id: sandboxId,
      },
      select: ['region'],
      loadEagerRelations: false,
    })

    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    return sandbox.region
  }

  async getPortPreviewUrl(sandboxIdOrName: string, organizationId: string, port: number): Promise<PortPreviewUrlDto> {
    if (port < 1 || port > 65535) {
      throw new BadRequestError('Invalid port')
    }

    const proxyDomain = this.configService.getOrThrow('proxy.domain')
    const proxyProtocol = this.configService.getOrThrow('proxy.protocol')

    const where: FindOptionsWhere<Sandbox> = {
      organizationId: organizationId,
      state: Not(SandboxState.DESTROYED),
    }

    const sandbox = await this.sandboxRepository.findOne({
      where: [
        {
          id: sandboxIdOrName,
          ...where,
        },
        {
          name: sandboxIdOrName,
          ...where,
        },
      ],
      cache: {
        id: `sandbox:${sandboxIdOrName}:organization:${organizationId}`,
        milliseconds: 1000,
      },
    })

    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID or name ${sandboxIdOrName} not found`)
    }

    let url = `${proxyProtocol}://${port}-${sandbox.id}.${proxyDomain}`

    const region = await this.regionService.findOne(sandbox.region, true)
    if (region && region.proxyUrl) {
      // Insert port and sandbox.id into the custom proxy URL
      url = region.proxyUrl.replace(/(https?:\/)(\/)/, `$1/${port}-${sandbox.id}.`)
    }

    return {
      sandboxId: sandbox.id,
      url,
      token: sandbox.authToken,
    }
  }

  async getSignedPortPreviewUrl(
    sandboxIdOrName: string,
    organizationId: string,
    port: number,
    expiresInSeconds = 60,
  ): Promise<SignedPortPreviewUrlDto> {
    if (port < 1 || port > 65535) {
      throw new BadRequestError('Invalid port')
    }

    if (expiresInSeconds < 1 || expiresInSeconds > 60 * 60 * 24) {
      throw new BadRequestError('expiresInSeconds must be between 1 second and 24 hours')
    }

    const proxyDomain = this.configService.getOrThrow('proxy.domain')
    const proxyProtocol = this.configService.getOrThrow('proxy.protocol')

    const where: FindOptionsWhere<Sandbox> = {
      organizationId: organizationId,
      state: Not(SandboxState.DESTROYED),
    }

    const sandbox = await this.sandboxRepository.findOne({
      where: [
        {
          id: sandboxIdOrName,
          ...where,
        },
        {
          name: sandboxIdOrName,
          ...where,
        },
      ],
      cache: {
        id: `sandbox:${sandboxIdOrName}:organization:${organizationId}`,
        milliseconds: 1000,
      },
    })

    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID or name ${sandboxIdOrName} not found`)
    }

    const token = customNanoid(urlAlphabet.replace('_', '').replace('-', ''))(16).toLocaleLowerCase()

    const lockKey = `sandbox:signed-preview-url-token:${port}:${token}`
    await this.redis.setex(lockKey, expiresInSeconds, sandbox.id)

    let url = `${proxyProtocol}://${port}-${token}.${proxyDomain}`

    const region = await this.regionService.findOne(sandbox.region, true)
    if (region && region.proxyUrl) {
      // Insert port and sandbox.id into the custom proxy URL
      url = region.proxyUrl.replace(/(https?:\/)(\/)/, `$1/${port}-${token}.`)
    }

    return {
      sandboxId: sandbox.id,
      port,
      token,
      url,
    }
  }

  async getSandboxIdFromSignedPreviewUrlToken(token: string, port: number): Promise<string> {
    const lockKey = `sandbox:signed-preview-url-token:${port}:${token}`
    const sandboxId = await this.redis.get(lockKey)
    if (!sandboxId) {
      throw new ForbiddenException('Invalid or expired token')
    }
    return sandboxId
  }

  async expireSignedPreviewUrlToken(
    sandboxIdOrName: string,
    organizationId: string,
    token: string,
    port: number,
  ): Promise<void> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)
    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID or name ${sandboxIdOrName} not found`)
    }

    const lockKey = `sandbox:signed-preview-url-token:${port}:${token}`
    await this.redis.del(lockKey)
  }

  async destroy(sandboxIdOrName: string, organizationId?: string): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    if (sandbox.pending) {
      throw new SandboxError('Sandbox state change in progress')
    }

    sandbox.applyDesiredDestroyedState()
    await this.sandboxRepository.saveWhere(sandbox, { pending: false, state: sandbox.state })

    this.eventEmitter.emit(SandboxEvents.DESTROYED, new SandboxDestroyedEvent(sandbox))
    return sandbox
  }

  async start(sandboxIdOrName: string, organization: Organization): Promise<Sandbox> {
    let pendingCpuIncrement: number | undefined
    let pendingMemoryIncrement: number | undefined
    let pendingDiskIncrement: number | undefined

    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organization.id)

    try {
      if (sandbox.state === SandboxState.STARTED && sandbox.desiredState === SandboxDesiredState.STARTED) {
        return sandbox
      }

      this.assertSandboxNotErrored(sandbox)

      if (String(sandbox.state) !== String(sandbox.desiredState)) {
        // Allow start of stopped | archived and archiving | archived sandboxes
        if (
          sandbox.desiredState !== SandboxDesiredState.ARCHIVED ||
          (sandbox.state !== SandboxState.STOPPED && sandbox.state !== SandboxState.ARCHIVING)
        ) {
          throw new SandboxError('State change in progress')
        }
      }

      if (![SandboxState.STOPPED, SandboxState.ARCHIVED, SandboxState.ARCHIVING].includes(sandbox.state)) {
        throw new SandboxError('Sandbox is not in valid state')
      }

      if (sandbox.pending) {
        throw new SandboxError('Sandbox state change in progress')
      }

      this.organizationService.assertOrganizationIsNotSuspended(organization)

      const { pendingCpuIncremented, pendingMemoryIncremented, pendingDiskIncremented } =
        await this.validateOrganizationQuotas(
          organization,
          sandbox.region,
          sandbox.cpu,
          sandbox.mem,
          sandbox.disk,
          sandbox.id,
        )

      if (pendingCpuIncremented) {
        pendingCpuIncrement = sandbox.cpu
      }
      if (pendingMemoryIncremented) {
        pendingMemoryIncrement = sandbox.mem
      }
      if (pendingDiskIncremented) {
        pendingDiskIncrement = sandbox.disk
      }

      sandbox.pending = true
      sandbox.desiredState = SandboxDesiredState.STARTED
      sandbox.authToken = nanoid(32).toLocaleLowerCase()

      await this.sandboxRepository.saveWhere(sandbox, { pending: false, state: sandbox.state })

      this.eventEmitter.emit(SandboxEvents.STARTED, new SandboxStartedEvent(sandbox))

      return sandbox
    } catch (error) {
      await this.rollbackPendingUsage(
        organization.id,
        sandbox.region,
        pendingCpuIncrement,
        pendingMemoryIncrement,
        pendingDiskIncrement,
      )
      throw error
    }
  }

  async stop(sandboxIdOrName: string, organizationId?: string): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    this.assertSandboxNotErrored(sandbox)

    if (String(sandbox.state) !== String(sandbox.desiredState)) {
      throw new SandboxError('State change in progress')
    }

    if (sandbox.state !== SandboxState.STARTED) {
      throw new SandboxError('Sandbox is not started')
    }

    if (sandbox.pending) {
      throw new SandboxError('Sandbox state change in progress')
    }

    sandbox.pending = true
    //  if auto-delete interval is 0, delete the sandbox immediately
    if (sandbox.autoDeleteInterval === 0) {
      sandbox.desiredState = SandboxDesiredState.DESTROYED
    } else {
      sandbox.desiredState = SandboxDesiredState.STOPPED
    }

    await this.sandboxRepository.saveWhere(sandbox, { pending: false, state: sandbox.state })

    if (sandbox.autoDeleteInterval === 0) {
      this.eventEmitter.emit(SandboxEvents.DESTROYED, new SandboxDestroyedEvent(sandbox))
    } else {
      this.eventEmitter.emit(SandboxEvents.STOPPED, new SandboxStoppedEvent(sandbox))
    }
    return sandbox
  }

  async recover(sandboxIdOrName: string, organization: Organization): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organization.id)

    if (sandbox.state !== SandboxState.ERROR) {
      throw new BadRequestError('Sandbox must be in error state to recover')
    }

    if (sandbox.pending) {
      throw new SandboxError('Sandbox state change in progress')
    }

    // Validate runner exists
    if (!sandbox.runnerId) {
      throw new NotFoundException(`Sandbox with ID ${sandbox.id} does not have a runner`)
    }
    const runner = await this.runnerRepository.findOneBy({ id: sandbox.runnerId })
    if (!runner) {
      throw new NotFoundException(`Runner with ID ${sandbox.runnerId} not found`)
    }

    if (runner.apiVersion === '2') {
      // TODO: we need "recovering" state that can be set after calling recover
      // Once in recovering, we abort further processing and let the manager/job handler take care of it
      // (Also, since desiredState would be STARTED, we need to check the quota)
      throw new ForbiddenException('Recovering sandboxes with runner API version 2 is not supported')
    }

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    try {
      await runnerAdapter.recoverSandbox(sandbox)
    } catch (error) {
      if (error instanceof Error && error.message.includes('storage cannot be further expanded')) {
        const errorMsg = `Sandbox storage cannot be further expanded. Maximum expansion of ${(sandbox.disk * 0.1).toFixed(2)}GB (10% of original ${sandbox.disk.toFixed(2)}GB) has been reached. Please contact support for further assistance.`
        throw new ForbiddenException(errorMsg)
      }
      throw error
    }

    // Clear error state
    sandbox.state = SandboxState.STOPPED
    sandbox.desiredState = SandboxDesiredState.STOPPED
    sandbox.errorReason = null
    sandbox.recoverable = false
    await this.sandboxRepository.saveWhere(sandbox, { state: SandboxState.ERROR })

    // Now that sandbox is in STOPPED state, use the normal start flow
    // This handles quota validation, pending usage, event emission, etc.
    return await this.start(sandbox.id, organization)
  }

  async resize(sandboxIdOrName: string, resizeDto: ResizeSandboxDto, organization: Organization): Promise<Sandbox> {
    let pendingCpuIncrement: number | undefined
    let pendingMemoryIncrement: number | undefined
    let pendingDiskIncrement: number | undefined

    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organization.id)

    try {
      // Validate sandbox is in a valid state for resize
      if (sandbox.state !== SandboxState.STARTED && sandbox.state !== SandboxState.STOPPED) {
        throw new BadRequestError('Sandbox must be in started or stopped state to resize')
      }

      if (sandbox.pending) {
        throw new SandboxError('Sandbox state change in progress')
      }

      // If no resize parameters provided, throw error
      if (resizeDto.cpu === undefined && resizeDto.memory === undefined && resizeDto.disk === undefined) {
        throw new BadRequestError('No resource changes specified - sandbox is already at the desired configuration')
      }

      // Disk resize requires stopped sandbox (cold resize only)
      if (resizeDto.disk !== undefined && sandbox.state !== SandboxState.STOPPED) {
        throw new BadRequestError('Disk resize can only be performed on a stopped sandbox')
      }

      // Hot resize (sandbox is running): only CPU and memory can be increased
      const isHotResize = sandbox.state === SandboxState.STARTED

      // Validate hot resize constraints
      if (isHotResize) {
        if (resizeDto.cpu !== undefined && resizeDto.cpu < sandbox.cpu) {
          throw new BadRequestError('Sandbox must be in stopped state to decrease the number of CPU cores')
        }

        if (resizeDto.memory !== undefined && resizeDto.memory < sandbox.mem) {
          throw new BadRequestError('Sandbox must be in stopped state to decrease memory')
        }
      }

      // Disk can only be increased (never decreased)
      if (resizeDto.disk !== undefined && resizeDto.disk < sandbox.disk) {
        throw new BadRequestError('Sandbox disk size cannot be decreased')
      }

      // Calculate new resource values
      const newCpu = resizeDto.cpu ?? sandbox.cpu
      const newMem = resizeDto.memory ?? sandbox.mem
      const newDisk = resizeDto.disk ?? sandbox.disk

      // Throw if nothing actually changes
      if (newCpu === sandbox.cpu && newMem === sandbox.mem && newDisk === sandbox.disk) {
        throw new BadRequestError('No resource changes specified - sandbox is already at the desired configuration')
      }

      // Validate organization quotas for the new resource values
      this.organizationService.assertOrganizationIsNotSuspended(organization)

      // Validate per-sandbox quotas with total new values
      if (newCpu > organization.maxCpuPerSandbox) {
        throw new ForbiddenException(
          `CPU request ${newCpu} exceeds maximum allowed per sandbox (${organization.maxCpuPerSandbox}).\n${PER_SANDBOX_LIMIT_MESSAGE}`,
        )
      }
      if (newMem > organization.maxMemoryPerSandbox) {
        throw new ForbiddenException(
          `Memory request ${newMem}GB exceeds maximum allowed per sandbox (${organization.maxMemoryPerSandbox}GB).\n${PER_SANDBOX_LIMIT_MESSAGE}`,
        )
      }
      if (newDisk > organization.maxDiskPerSandbox) {
        throw new ForbiddenException(
          `Disk request ${newDisk}GB exceeds maximum allowed per sandbox (${organization.maxDiskPerSandbox}GB).\n${PER_SANDBOX_LIMIT_MESSAGE}`,
        )
      }

      // For cold resize, cpu/memory don't affect quota until sandbox is STARTED.
      // For hot resize, track all deltas (positive reserves quota, negative frees quota for others).
      const cpuDeltaForQuota = isHotResize ? newCpu - sandbox.cpu : 0
      const memDeltaForQuota = isHotResize ? newMem - sandbox.mem : 0
      const diskDeltaForQuota = newDisk - sandbox.disk // Disk only increases (validated at start of method)

      // Validate and track pending for any non-zero quota changes
      if (cpuDeltaForQuota !== 0 || memDeltaForQuota !== 0 || diskDeltaForQuota !== 0) {
        const { pendingCpuIncremented, pendingMemoryIncremented, pendingDiskIncremented } =
          await this.validateOrganizationQuotas(
            organization,
            sandbox.region,
            cpuDeltaForQuota,
            memDeltaForQuota,
            diskDeltaForQuota,
          )

        if (pendingCpuIncremented) {
          pendingCpuIncrement = cpuDeltaForQuota
        }
        if (pendingMemoryIncremented) {
          pendingMemoryIncrement = memDeltaForQuota
        }
        if (pendingDiskIncremented) {
          pendingDiskIncrement = diskDeltaForQuota
        }
      }

      // Get runner and validate before changing state
      if (!sandbox.runnerId) {
        throw new BadRequestError('Sandbox has no runner assigned')
      }

      const runner = await this.runnerService.findOne(sandbox.runnerId)
      if (!runner) {
        throw new NotFoundException(`Runner with ID ${sandbox.runnerId} not found`)
      }

      // Capture the previous state before transitioning to RESIZING (STARTED or STOPPED)
      const previousState =
        sandbox.state === SandboxState.STARTED
          ? SandboxState.STARTED
          : sandbox.state === SandboxState.STOPPED
            ? SandboxState.STOPPED
            : null

      if (!previousState) {
        throw new BadRequestError('Sandbox must be in started or stopped state to resize')
      }

      // Now transition to RESIZING state
      sandbox.state = SandboxState.RESIZING
      await this.sandboxRepository.saveWhere(sandbox, { pending: false, state: previousState })

      try {
        const runnerAdapter = await this.runnerAdapterFactory.create(runner)

        await runnerAdapter.resizeSandbox(sandbox.id, resizeDto.cpu, resizeDto.memory, resizeDto.disk)

        // For V0 runners, update resources immediately (subscriber emits STATE_UPDATED)
        // For V2 runners, job handler will update resources on completion
        if (runner.apiVersion === '0') {
          sandbox.cpu = newCpu
          sandbox.mem = newMem
          sandbox.disk = newDisk
          sandbox.state = previousState
          await this.sandboxRepository.saveWhere(sandbox, { state: SandboxState.RESIZING })

          // Apply the usage change (increments current, decrements pending)
          // Only apply deltas for quotas that were validated/pending-incremented
          await this.organizationUsageService.applyResizeUsageChange(
            organization.id,
            sandbox.region,
            cpuDeltaForQuota,
            memDeltaForQuota,
            diskDeltaForQuota,
          )
        }

        return await this.findOneByIdOrName(sandbox.id, organization.id)
      } catch (error) {
        // Return to previous state on error
        sandbox.state = previousState
        await this.sandboxRepository.saveWhere(sandbox, { state: SandboxState.RESIZING })

        throw error
      }
    } catch (error) {
      await this.rollbackPendingUsage(
        organization.id,
        sandbox.region,
        pendingCpuIncrement,
        pendingMemoryIncrement,
        pendingDiskIncrement,
      )
      throw error
    }
  }

  async updatePublicStatus(sandboxIdOrName: string, isPublic: boolean, organizationId?: string): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    sandbox.public = isPublic
    await this.sandboxRepository.save(sandbox)

    return sandbox
  }

  async updateLastActivityAt(sandboxId: string, lastActivityAt: Date): Promise<void> {
    // Prevent spamming updates
    const lockKey = `sandbox:update-last-activity:${sandboxId}`
    const acquired = await this.redisLockProvider.lock(lockKey, 45)
    if (!acquired) {
      return
    }

    await this.updateById(sandboxId, { lastActivityAt })
  }

  async getToolboxProxyUrl(sandboxId: string): Promise<string> {
    const sandbox = await this.findOne(sandboxId)

    const region = await this.regionService.findOne(sandbox.region, true)

    if (region && region.toolboxProxyUrl) {
      return region.toolboxProxyUrl + '/toolbox'
    }

    return this.configService.getOrThrow('proxy.toolboxUrl')
  }

  async getBuildLogsUrl(sandboxIdOrName: string, organizationId: string): Promise<string> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    if (!sandbox.buildInfo?.snapshotRef) {
      throw new NotFoundException(`Sandbox ${sandboxIdOrName} has no build info`)
    }

    const region = await this.regionService.findOne(sandbox.region, true)

    if (!region) {
      throw new NotFoundException(`Region for runner for sandbox ${sandboxIdOrName} not found`)
    }

    if (!region.proxyUrl) {
      return `${this.configService.getOrThrow('proxy.protocol')}://${this.configService.getOrThrow('proxy.domain')}/sandboxes/${sandbox.id}/build-logs`
    }

    return region.proxyUrl + '/sandboxes/' + sandbox.id + '/build-logs'
  }

  private async getValidatedOrDefaultRegionId(organization: Organization, regionIdOrName?: string): Promise<string> {
    if (!organization.defaultRegionId) {
      throw new DefaultRegionRequiredException()
    }

    regionIdOrName = regionIdOrName?.trim()

    if (!regionIdOrName) {
      return organization.defaultRegionId
    }

    const region =
      (await this.regionService.findOneByName(regionIdOrName, organization.id)) ??
      (await this.regionService.findOneByName(regionIdOrName, null)) ??
      (await this.regionService.findOne(regionIdOrName))

    if (!region) {
      throw new NotFoundException('Region not found')
    }

    return region.id
  }

  private getValidatedOrDefaultClass(sandboxClass: SandboxClass): SandboxClass {
    if (!sandboxClass) {
      return SandboxClass.SMALL
    }

    if (Object.values(SandboxClass).includes(sandboxClass)) {
      return sandboxClass
    } else {
      throw new BadRequestError('Invalid class')
    }
  }

  async replaceLabels(
    sandboxIdOrName: string,
    labels: { [key: string]: string },
    organizationId?: string,
  ): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    // Replace all labels
    sandbox.labels = labels
    await this.sandboxRepository.save(sandbox)

    return sandbox
  }

  @Cron(CronExpression.EVERY_10_MINUTES, { name: 'cleanup-destroyed-sandboxes' })
  @LogExecution('cleanup-destroyed-sandboxes')
  @WithInstrumentation()
  async cleanupDestroyedSandboxes() {
    const twentyFourHoursAgo = new Date()
    twentyFourHoursAgo.setHours(twentyFourHoursAgo.getHours() - 24)

    const destroyedSandboxs = await this.sandboxRepository.delete({
      state: SandboxState.DESTROYED,
      updatedAt: LessThan(twentyFourHoursAgo),
    })

    if (destroyedSandboxs.affected > 0) {
      this.logger.debug(`Cleaned up ${destroyedSandboxs.affected} destroyed sandboxes`)
    }
  }

  @Cron(CronExpression.EVERY_10_MINUTES, { name: 'cleanup-build-failed-sandboxes' })
  @LogExecution('cleanup-build-failed-sandboxes')
  @WithInstrumentation()
  async cleanupBuildFailedSandboxes() {
    const twentyFourHoursAgo = new Date()
    twentyFourHoursAgo.setHours(twentyFourHoursAgo.getHours() - 24)

    const destroyedSandboxs = await this.sandboxRepository.delete({
      state: SandboxState.BUILD_FAILED,
      desiredState: SandboxDesiredState.DESTROYED,
      updatedAt: LessThan(twentyFourHoursAgo),
    })

    if (destroyedSandboxs.affected > 0) {
      this.logger.debug(`Cleaned up ${destroyedSandboxs.affected} build failed sandboxes`)
    }
  }

  async setAutostopInterval(sandboxIdOrName: string, interval: number, organizationId?: string): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    sandbox.autoStopInterval = this.resolveAutoStopInterval(interval)
    await this.sandboxRepository.save(sandbox)

    return sandbox
  }

  async setAutoArchiveInterval(sandboxIdOrName: string, interval: number, organizationId?: string): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    sandbox.autoArchiveInterval = this.resolveAutoArchiveInterval(interval)
    await this.sandboxRepository.save(sandbox)

    return sandbox
  }

  async setAutoDeleteInterval(sandboxIdOrName: string, interval: number, organizationId?: string): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    sandbox.autoDeleteInterval = interval
    await this.sandboxRepository.save(sandbox)

    return sandbox
  }

  async updateNetworkSettings(
    sandboxIdOrName: string,
    networkBlockAll?: boolean,
    networkAllowList?: string,
    organizationId?: string,
  ): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    if (networkBlockAll !== undefined) {
      sandbox.networkBlockAll = networkBlockAll
    }

    if (networkAllowList !== undefined) {
      sandbox.networkAllowList = this.resolveNetworkAllowList(networkAllowList)
    }

    await this.sandboxRepository.save(sandbox)

    // Update network settings on the runner
    if (sandbox.runnerId) {
      const runner = await this.runnerService.findOne(sandbox.runnerId)
      if (runner) {
        const runnerAdapter = await this.runnerAdapterFactory.create(runner)
        await runnerAdapter.updateNetworkSettings(sandbox.id, networkBlockAll, networkAllowList)
      }
    }

    return sandbox
  }

  // used by internal services to update the state of a sandbox to resolve domain and runner state mismatch
  // notably, when a sandbox instance stops or errors on the runner, the domain state needs to be updated to reflect the actual state
  async updateState(
    sandboxId: string,
    newState: SandboxState,
    recoverable = false,
    errorReason?: string,
  ): Promise<void> {
    const sandbox = await this.sandboxRepository.findOne({
      where: { id: sandboxId },
    })

    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    if (sandbox.state === newState) {
      this.logger.debug(`Sandbox ${sandboxId} is already in state ${newState}`)
      return
    }

    //  only allow updating the state of started | stopped sandboxes
    if (![SandboxState.STARTED, SandboxState.STOPPED].includes(sandbox.state)) {
      throw new BadRequestError('Sandbox is not in a valid state to be updated')
    }

    if (sandbox.desiredState == SandboxDesiredState.DESTROYED) {
      this.logger.debug(`Sandbox ${sandboxId} is already DESTROYED, skipping state update`)
      return
    }

    const oldState = sandbox.state
    const oldDesiredState = sandbox.desiredState
    sandbox.state = newState
    sandbox.recoverable = false
    if (errorReason !== undefined) {
      sandbox.errorReason = errorReason
      if (newState === SandboxState.ERROR) {
        sandbox.recoverable = recoverable
      }
    }
    //  we need to update the desired state to match the new state
    const desiredState = this.getExpectedDesiredStateForState(newState)
    if (desiredState) {
      sandbox.desiredState = desiredState
    }
    await this.sandboxRepository.saveWhere(sandbox, { pending: false, state: oldState, desiredState: oldDesiredState })
  }

  @OnEvent(WarmPoolEvents.TOPUP_REQUESTED)
  private async createWarmPoolSandbox(event: WarmPoolTopUpRequested) {
    await this.createForWarmPool(event.warmPool)
  }

  @Cron(CronExpression.EVERY_MINUTE, { name: 'handle-unschedulable-runners' })
  @LogExecution('handle-unschedulable-runners')
  @WithInstrumentation()
  private async handleUnschedulableRunners() {
    const runners = await this.runnerRepository.find({ where: { unschedulable: true } })

    if (runners.length === 0) {
      return
    }

    //  find all sandboxes that are using the unschedulable runners and have organizationId = '00000000-0000-0000-0000-000000000000'
    const sandboxes = await this.sandboxRepository.find({
      where: {
        runnerId: In(runners.map((runner) => runner.id)),
        organizationId: '00000000-0000-0000-0000-000000000000',
        state: SandboxState.STARTED,
        desiredState: Not(SandboxDesiredState.DESTROYED),
      },
    })

    if (sandboxes.length === 0) {
      return
    }

    const destroyPromises = sandboxes.map((sandbox) => this.destroy(sandbox.id))
    const results = await Promise.allSettled(destroyPromises)

    // Log any failed sandbox destructions
    results.forEach((result, index) => {
      if (result.status === 'rejected') {
        this.logger.error(`Failed to destroy sandbox ${sandboxes[index].id}: ${result.reason}`)
      }
    })
  }

  async isSandboxPublic(sandboxId: string): Promise<boolean> {
    const sandbox = await this.sandboxRepository.findOne({
      where: { id: sandboxId },
    })

    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    return sandbox.public
  }

  @OnEvent(OrganizationEvents.SUSPENDED_SANDBOX_STOPPED)
  async handleSuspendedSandboxStopped(event: OrganizationSuspendedSandboxStoppedEvent) {
    await this.stop(event.sandboxId).catch((error) => {
      //  log the error for now, but don't throw it as it will be retried
      this.logger.error(`Error stopping sandbox from suspended organization. SandboxId: ${event.sandboxId}: `, error)
    })
  }

  private resolveAutoStopInterval(autoStopInterval: number): number {
    if (autoStopInterval < 0) {
      throw new BadRequestError('Auto-stop interval must be non-negative')
    }

    return autoStopInterval
  }

  private resolveAutoArchiveInterval(autoArchiveInterval: number): number {
    if (autoArchiveInterval < 0) {
      throw new BadRequestError('Auto-archive interval must be non-negative')
    }

    const maxAutoArchiveInterval = this.configService.getOrThrow('maxAutoArchiveInterval')

    if (autoArchiveInterval === 0) {
      return maxAutoArchiveInterval
    }

    return Math.min(autoArchiveInterval, maxAutoArchiveInterval)
  }

  private resolveNetworkAllowList(networkAllowList: string): string {
    try {
      validateNetworkAllowList(networkAllowList)
    } catch (error) {
      throw new BadRequestError(error instanceof Error ? error.message : 'Invalid network allow list')
    }

    return networkAllowList
  }

  private resolveVolumes(volumes: SandboxVolume[]): SandboxVolume[] {
    try {
      validateMountPaths(volumes)
    } catch (error) {
      throw new BadRequestError(error instanceof Error ? error.message : 'Invalid volume mount configuration')
    }

    try {
      validateSubpaths(volumes)
    } catch (error) {
      throw new BadRequestError(error instanceof Error ? error.message : 'Invalid volume subpath configuration')
    }

    return volumes
  }

  async createSshAccess(
    sandboxIdOrName: string,
    expiresInMinutes = 60,
    organizationId?: string,
  ): Promise<SshAccessDto> {
    //  check if sandbox exists
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    // Revoke any existing SSH access for this sandbox
    await this.revokeSshAccess(sandbox.id)

    const sshAccess = new SshAccess()
    sshAccess.sandboxId = sandbox.id
    // Generate a safe token that can't doesn't have _ or - to avoid CLI issues
    sshAccess.token = customNanoid(urlAlphabet.replace('_', '').replace('-', ''))(32)
    sshAccess.expiresAt = new Date(Date.now() + expiresInMinutes * 60 * 1000)

    await this.sshAccessRepository.save(sshAccess)

    const region = await this.regionService.findOne(sandbox.region, true)
    if (region && region.sshGatewayUrl) {
      return SshAccessDto.fromSshAccess(sshAccess, region.sshGatewayUrl)
    }

    return SshAccessDto.fromSshAccess(sshAccess, this.configService.getOrThrow('sshGateway.url'))
  }

  async revokeSshAccess(sandboxIdOrName: string, token?: string, organizationId?: string): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    if (token) {
      // Revoke specific SSH access by token
      await this.sshAccessRepository.delete({ sandboxId: sandbox.id, token })
    } else {
      // Revoke all SSH access for the sandbox
      await this.sshAccessRepository.delete({ sandboxId: sandbox.id })
    }

    return sandbox
  }

  async validateSshAccess(token: string): Promise<SshAccessValidationDto> {
    const sshAccess = await this.sshAccessRepository.findOne({
      where: {
        token,
      },
      relations: ['sandbox'],
    })

    if (!sshAccess) {
      return { valid: false, sandboxId: null }
    }

    // Check if token is expired
    const isExpired = sshAccess.expiresAt < new Date()
    if (isExpired) {
      return { valid: false, sandboxId: null }
    }

    // Get runner information if sandbox exists
    if (sshAccess.sandbox && sshAccess.sandbox.runnerId) {
      const runner = await this.runnerRepository.findOne({
        where: { id: sshAccess.sandbox.runnerId },
      })

      if (runner) {
        return {
          valid: true,
          sandboxId: sshAccess.sandbox.id,
        }
      }
    }

    return { valid: true, sandboxId: sshAccess.sandbox.id }
  }

  async updateSandboxBackupState(
    sandboxId: string,
    backupState: BackupState,
    backupSnapshot?: string | null,
    backupRegistryId?: string | null,
    backupErrorReason?: string | null,
  ): Promise<void> {
    const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
      id: sandboxId,
    })
    const originalState = sandboxToUpdate.state
    const originalRunnerId = sandboxToUpdate.runnerId

    sandboxToUpdate.setBackupState(backupState, backupSnapshot, backupRegistryId, backupErrorReason)

    const updateData: Partial<Sandbox> = {
      backupState: sandboxToUpdate.backupState,
      backupSnapshot: sandboxToUpdate.backupSnapshot,
      backupRegistryId: sandboxToUpdate.backupRegistryId,
      backupErrorReason: sandboxToUpdate.backupErrorReason,
      lastBackupAt: sandboxToUpdate.lastBackupAt,
      existingBackupSnapshots: sandboxToUpdate.existingBackupSnapshots,
    }

    if (sandboxToUpdate.state !== originalState) {
      updateData.state = sandboxToUpdate.state
    }

    if (sandboxToUpdate.runnerId !== originalRunnerId) {
      updateData.runnerId = sandboxToUpdate.runnerId
    }

    await this.updateById(sandboxId, updateData, sandboxToUpdate)
  }
}
