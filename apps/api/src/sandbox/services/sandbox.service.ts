/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  ForbiddenException,
  Inject,
  Injectable,
  Logger,
  NotFoundException,
  ConflictException,
  HttpException,
  HttpStatus,
} from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Not, Repository, LessThan, In, JsonContains, FindOptionsWhere, ILike } from 'typeorm'
import { Sandbox } from '../entities/sandbox.entity'
import { SandboxFork } from '../entities/sandbox-fork.entity'
import { CreateSandboxDto } from '../dto/create-sandbox.dto'
import { CreateSandboxSnapshotDto } from '../dto/create-sandbox-snapshot.dto'
import { ForkSandboxDto } from '../dto/fork-sandbox.dto'
import { ResizeSandboxDto } from '../dto/resize-sandbox.dto'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { isRegistryBasedSandboxClass } from '../utils/sandbox-class.util'
import { OpenFeature } from '@openfeature/server-sdk'
import { FeatureFlags } from '../../common/constants/feature-flags'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { resolveGpuTypePreferences } from '../utils/gpu-type-preferences.util'
import { RunnerService } from './runner.service'
import { SandboxError } from '../../exceptions/sandbox-error.exception'
import { StateChangeInProgressError } from '../../exceptions/state-change-in-progress.exception'
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
import { isApiRecoverableError } from '../constants/errors-for-recovery'
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
import { RunnerAdapter, RunnerAdapterFactory } from '../runner-adapter/runnerAdapter'
import { validateDomainAllowList, validateNetworkAllowList } from '../utils/network-validation.util'
import { OrganizationUsageService } from '../../organization/services/organization-usage.service'
import { SshAccess } from '../entities/ssh-access.entity'
import { SshAccessDto, SshAccessValidationDto } from '../dto/ssh-access.dto'
import { VolumeService } from './volume.service'
import { PaginatedList } from '../../common/interfaces/paginated-list.interface'
import {
  SandboxSortFieldDeprecated,
  SandboxSortDirectionDeprecated,
  DEFAULT_SANDBOX_SORT_FIELD_DEPRECATED,
  DEFAULT_SANDBOX_SORT_DIRECTION_DEPRECATED,
} from '../dto/list-sandboxes-query.deprecated.dto'
import { createRangeFilter } from '../../common/utils/range-filter'
import { LogExecution } from '../../common/decorators/log-execution.decorator'
import {
  UPGRADE_TIER_MESSAGE,
  ARCHIVE_SANDBOXES_MESSAGE,
  PER_SANDBOX_LIMIT_MESSAGE,
} from '../../common/constants/error-messages'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { getStateChangeLockKey } from '../utils/lock-key.util'
import { customAlphabet as customNanoid, nanoid, urlAlphabet } from 'nanoid'
import { WithInstrumentation } from '../../common/decorators/otel.decorator'
import { validateMountPaths, validateSubpaths } from '../utils/volume-mount-path-validation.util'
import { isEphemeral } from '../utils/ephemeral.util'
import { SandboxRepository } from '../repositories/sandbox.repository'
import { SnapshotRepository } from '../repositories/snapshot.repository'
import { PortPreviewUrlDto, SignedPortPreviewUrlDto } from '../dto/port-preview-url.dto'
import { RegionService } from '../../region/services/region.service'
import { DefaultRegionRequiredException } from '../../organization/exceptions/DefaultRegionRequiredException'
import { SnapshotService } from './snapshot.service'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { DockerRegistry } from '../../docker-registry/entities/docker-registry.entity'
import { RegionType } from '../../region/enums/region-type.enum'
import { getEffectivePerSandboxLimits } from '../../organization/utils/sandbox-limits.util'
import { RegionQuotaDto } from '../../organization/dto/region-quota.dto'
import { SandboxCreatedEvent } from '../events/sandbox-create.event'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import {
  SANDBOX_LOOKUP_CACHE_TTL_MS,
  SANDBOX_ORG_ID_CACHE_TTL_MS,
  TOOLBOX_PROXY_URL_CACHE_TTL_S,
  sandboxLookupCacheKeyById,
  sandboxLookupCacheKeyByName,
  sandboxOrgIdCacheKeyById,
  sandboxOrgIdCacheKeyByName,
  toolboxProxyUrlCacheKey,
} from '../utils/sandbox-lookup-cache.util'
import { SandboxLookupCacheInvalidationService } from './sandbox-lookup-cache-invalidation.service'
import { Region } from '../../region/entities/region.entity'
import { SandboxActivityService } from './sandbox-activity.service'
import { ListSandboxesResponseDto } from '../dto/list-sandboxes-response.dto'
import { ListSandboxesQueryDto } from '../dto/list-sandboxes-query.dto'
import { SANDBOX_SEARCH_ADAPTER } from '../constants/sandbox-tokens'
import { SandboxSearchAdapter } from '../interfaces/sandbox-search.interface'

const DEFAULT_CPU = 1
const DEFAULT_MEMORY = 1
const DEFAULT_DISK = 3
const DEFAULT_GPU = 0

@Injectable()
export class SandboxService {
  private readonly logger = new Logger(SandboxService.name)

  constructor(
    private readonly sandboxRepository: SandboxRepository,
    private readonly snapshotRepository: SnapshotRepository,
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
    private readonly sandboxActivityService: SandboxActivityService,
    private readonly dockerRegistryService: DockerRegistryService,
    @InjectRepository(SandboxFork)
    private readonly sandboxForkRepository: Repository<SandboxFork>,
    @Inject(SANDBOX_SEARCH_ADAPTER)
    private readonly sandboxSearchAdapter: SandboxSearchAdapter,
  ) {}

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
    region: Region,
    sandboxClass: SandboxClass,
    cpu: number,
    memory: number,
    disk: number,
    gpu: number,
    ephemeral: boolean,
    excludeSandboxId?: string,
    regionQuota?: RegionQuotaDto | null,
    // Controls which per-sandbox limits table is used (GPU-specific vs non-GPU).
    // Defaults to `gpu > 0`, which is correct for create/start/fork/archive paths where
    // `gpu` is the absolute sandbox GPU allocation. Resize passes `gpu = 0` (no GPU delta)
    // but still needs GPU-aware per-sandbox limits when the sandbox itself is a GPU sandbox.
    gpuEnabled: boolean = gpu > 0,
  ): Promise<{
    pendingCpuIncremented: boolean
    pendingMemoryIncremented: boolean
    pendingDiskIncremented: boolean
    pendingGpuIncremented: boolean
  }> {
    if (region.enforceQuotas) {
      if (!regionQuota) {
        regionQuota = await this.organizationService.getRegionQuota(organization.id, region.id, sandboxClass)
      }

      if (!regionQuota) {
        if (region.regionType === RegionType.SHARED) {
          // region is public, but the organization does not have a quota for it
          throw new ForbiddenException(
            `Region ${region.id} is not available to the organization for class ${sandboxClass}`,
          )
        } else {
          // region is not public, respond as if the region was not found
          throw new NotFoundException('Region not found')
        }
      }
    }

    // validate per-sandbox quotas
    const { maxCpuPerSandbox, maxMemoryPerSandbox, maxDiskPerSandbox, maxDiskPerNonEphemeralSandbox } =
      getEffectivePerSandboxLimits(organization, regionQuota, gpuEnabled)

    if (cpu > maxCpuPerSandbox) {
      throw new BadRequestError(
        `CPU request ${cpu} exceeds maximum allowed per sandbox (${maxCpuPerSandbox}).\n${PER_SANDBOX_LIMIT_MESSAGE}`,
      )
    }
    if (memory > maxMemoryPerSandbox) {
      throw new BadRequestError(
        `Memory request ${memory}GB exceeds maximum allowed per sandbox (${maxMemoryPerSandbox}GB).\n${PER_SANDBOX_LIMIT_MESSAGE}`,
      )
    }
    if (disk > maxDiskPerSandbox) {
      throw new BadRequestError(
        `Disk request ${disk}GB exceeds maximum allowed per sandbox (${maxDiskPerSandbox}GB).\n${PER_SANDBOX_LIMIT_MESSAGE}`,
      )
    }

    if (!ephemeral && maxDiskPerNonEphemeralSandbox !== null) {
      if (maxDiskPerNonEphemeralSandbox === 0) {
        throw new BadRequestError('Only ephemeral sandboxes are permitted in this region')
      }
      if (disk > maxDiskPerNonEphemeralSandbox) {
        throw new BadRequestError(
          `Disk request ${disk}GB exceeds maximum allowed per non-ephemeral sandbox (${maxDiskPerNonEphemeralSandbox}GB).\n${PER_SANDBOX_LIMIT_MESSAGE}`,
        )
      }
    }

    // e.g. region belonging to an organization
    if (!region.enforceQuotas) {
      return {
        pendingCpuIncremented: false,
        pendingMemoryIncremented: false,
        pendingDiskIncremented: false,
        pendingGpuIncremented: false,
      }
    }

    // Fail fast: requesting GPU in a region with no GPU quota at all.
    if (gpu > 0 && regionQuota.totalGpuQuota === 0) {
      throw new BadRequestError(`Total GPU limit exceeded. Maximum allowed: ${regionQuota.totalGpuQuota}.`)
    }

    // validate usage quotas
    const {
      cpuIncremented: pendingCpuIncremented,
      memoryIncremented: pendingMemoryIncremented,
      diskIncremented: pendingDiskIncremented,
      gpuIncremented: pendingGpuIncremented,
    } = await this.organizationUsageService.incrementPendingSandboxUsage(
      organization.id,
      region.id,
      sandboxClass,
      cpu,
      memory,
      disk,
      gpu,
      excludeSandboxId,
    )

    const usageOverview = await this.organizationUsageService.getSandboxUsageOverview(
      organization.id,
      region.id,
      sandboxClass,
      excludeSandboxId,
    )

    try {
      const upgradeTierMessage = UPGRADE_TIER_MESSAGE(this.configService.getOrThrow('dashboardUrl'))

      if (usageOverview.currentCpuUsage + usageOverview.pendingCpuUsage > regionQuota.totalCpuQuota) {
        throw new BadRequestError(
          `Total CPU limit exceeded. Maximum allowed: ${regionQuota.totalCpuQuota}.\n${upgradeTierMessage}`,
        )
      }

      if (usageOverview.currentMemoryUsage + usageOverview.pendingMemoryUsage > regionQuota.totalMemoryQuota) {
        throw new BadRequestError(
          `Total memory limit exceeded. Maximum allowed: ${regionQuota.totalMemoryQuota}GiB.\n${upgradeTierMessage}`,
        )
      }

      if (usageOverview.currentDiskUsage + usageOverview.pendingDiskUsage > regionQuota.totalDiskQuota) {
        throw new BadRequestError(
          `Total disk limit exceeded. Maximum allowed: ${regionQuota.totalDiskQuota}GiB.\n${ARCHIVE_SANDBOXES_MESSAGE}\n${upgradeTierMessage}`,
        )
      }

      if (usageOverview.currentGpuUsage + usageOverview.pendingGpuUsage > regionQuota.totalGpuQuota) {
        throw new BadRequestError(
          `Total GPU limit exceeded. Maximum allowed: ${regionQuota.totalGpuQuota}.\n${upgradeTierMessage}`,
        )
      }
    } catch (error) {
      await this.rollbackPendingUsage(
        organization.id,
        region.id,
        sandboxClass,
        pendingCpuIncremented ? cpu : undefined,
        pendingMemoryIncremented ? memory : undefined,
        pendingDiskIncremented ? disk : undefined,
        pendingGpuIncremented ? gpu : undefined,
      )
      throw error
    }

    return {
      pendingCpuIncremented,
      pendingMemoryIncremented,
      pendingDiskIncremented,
      pendingGpuIncremented,
    }
  }

  async rollbackPendingUsage(
    organizationId: string,
    regionId: string,
    sandboxClass: SandboxClass,
    pendingCpuIncrement?: number,
    pendingMemoryIncrement?: number,
    pendingDiskIncrement?: number,
    pendingGpuIncrement?: number,
  ): Promise<void> {
    if (!pendingCpuIncrement && !pendingMemoryIncrement && !pendingDiskIncrement && !pendingGpuIncrement) {
      return
    }

    try {
      await this.organizationUsageService.decrementPendingSandboxUsage(
        organizationId,
        regionId,
        sandboxClass,
        pendingCpuIncrement,
        pendingMemoryIncrement,
        pendingDiskIncrement,
        pendingGpuIncrement,
      )
    } catch (error) {
      this.logger.error(`Error rolling back pending sandbox usage: ${error}`)
    }
  }

  async archive(sandboxIdOrName: string, organizationId?: string): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    this.assertSandboxNotErrored(sandbox)

    if (String(sandbox.state) !== String(sandbox.desiredState)) {
      throw new StateChangeInProgressError()
    }

    if (sandbox.state !== SandboxState.STOPPED) {
      throw new SandboxError('Sandbox is not stopped')
    }

    if (sandbox.pending) {
      throw new StateChangeInProgressError()
    }

    if (isEphemeral(sandbox)) {
      throw new SandboxError('Ephemeral sandboxes cannot be archived')
    }

    const updateData: Partial<Sandbox> = {
      state: SandboxState.ARCHIVING,
      desiredState: SandboxDesiredState.ARCHIVED,
    }

    const updatedSandbox = await this.sandboxRepository.updateWhere(sandbox.id, {
      updateData,
      whereCondition: { pending: false, state: SandboxState.STOPPED },
    })

    this.eventEmitter.emit(SandboxEvents.ARCHIVED, new SandboxArchivedEvent(updatedSandbox))
    return updatedSandbox
  }

  async createForWarmPool(warmPoolItem: WarmPool): Promise<Sandbox> {
    const sandbox = new Sandbox({ region: warmPoolItem.target })

    sandbox.organizationId = SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION

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

    sandbox.gpuType = snapshot.gpuType ?? null

    let gpuRunnerAssignmentLockKey: string | undefined

    sandbox.sandboxClass = snapshot.sandboxClass

    try {
      // Same per-region GPU runner assignment serialization as createFromSnapshot.
      if (sandbox.gpu > 0) {
        const key = `gpu-runner-assignment:${sandbox.region}`
        await this.redisLockProvider.waitForLock(key, 60, 30000)
        gpuRunnerAssignmentLockKey = key
      }

      const runner = await this.runnerService.getRandomAvailableRunner({
        regions: [sandbox.region],
        sandboxClass: sandbox.sandboxClass,
        snapshotRef: snapshot.ref,
        gpu: sandbox.gpu,
        gpuType: sandbox.gpuType ?? null,
      })

      sandbox.runnerId = runner.id
      sandbox.pending = true

      await this.sandboxRepository.insert(sandbox)

      if (gpuRunnerAssignmentLockKey) {
        const key = gpuRunnerAssignmentLockKey
        gpuRunnerAssignmentLockKey = undefined
        await this.redisLockProvider
          .unlock(key)
          .catch((err) => this.logger.error('Failed to release GPU runner assignment lock', err))
      }

      return sandbox
    } finally {
      if (gpuRunnerAssignmentLockKey) {
        await this.redisLockProvider
          .unlock(gpuRunnerAssignmentLockKey)
          .catch((err) => this.logger.error('Failed to release GPU runner assignment lock', err))
      }
    }
  }

  async createFromSnapshot(createSandboxDto: CreateSandboxDto, organization: Organization): Promise<SandboxDto> {
    let pendingCpuIncrement: number | undefined
    let pendingMemoryIncrement: number | undefined
    let pendingDiskIncrement: number | undefined
    let pendingGpuIncrement: number | undefined
    let gpuRunnerAssignmentLockKey: string | undefined
    let sandboxClass: SandboxClass | undefined

    const region = await this.getValidatedOrDefaultRegion(organization, createSandboxDto.target)

    try {
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

      if (!(await this.snapshotService.isAvailableInRegion(snapshot.id, region.id))) {
        throw new BadRequestError(`Snapshot ${snapshotIdOrName} is not available in region ${region.id}`)
      }

      if (snapshot.state !== SnapshotState.ACTIVE) {
        throw new BadRequestError(`Snapshot ${snapshotIdOrName} is ${snapshot.state}`)
      }

      if (!snapshot.ref) {
        throw new BadRequestError('Snapshot ref is not defined')
      }

      const cpu = snapshot.cpu
      const mem = snapshot.mem
      const disk = snapshot.disk
      const gpu = snapshot.gpu
      const gpuType = snapshot.gpuType ?? null

      // GPU sandboxes are always ephemeral.
      if (gpu > 0 && !isEphemeral(createSandboxDto)) {
        throw new BadRequestError('GPU sandboxes must be ephemeral - set autoDeleteInterval to 0')
      }

      if (snapshot.sandboxClass === SandboxClass.ANDROID && !createSandboxDto.linkedSandbox) {
        throw new BadRequestError('Android sandboxes must be linked to another sandbox')
      }

      // Resolve and validate an optional linked sandbox. When set, the new sandbox is pinned
      // to the same runner as the linked sandbox so a local network can be established.
      // Constraints:
      //   - linked sandbox must exist in the same org, be STARTED or STOPPED, and have a runnerId
      //   - linked sandbox cannot itself be linked to another sandbox (no chains)
      //   - new sandbox must be ephemeral (autoDeleteInterval === 0)
      const linkedSandbox = await this.resolveLinkedSandbox(createSandboxDto, organization.id)

      this.organizationService.assertOrganizationIsNotSuspended(organization)

      sandboxClass = snapshot.sandboxClass

      const { pendingCpuIncremented, pendingMemoryIncremented, pendingDiskIncremented, pendingGpuIncremented } =
        await this.validateOrganizationQuotas(
          organization,
          region,
          snapshot.sandboxClass,
          cpu,
          mem,
          disk,
          gpu,
          isEphemeral(createSandboxDto),
        )

      if (pendingCpuIncremented) {
        pendingCpuIncrement = cpu
      }
      if (pendingMemoryIncremented) {
        pendingMemoryIncrement = mem
      }
      if (pendingDiskIncremented) {
        pendingDiskIncrement = disk
      }
      if (pendingGpuIncremented) {
        pendingGpuIncrement = gpu
      }

      // Resolve volume names to UUIDs before runner assignment, so invalid references fail fast
      const resolvedVolumes = await this.resolveVolumes(organization.id, createSandboxDto.volumes)

      // GPU sandboxes are always ephemeral: they get exclusive ownership of a
      // runner for their lifetime and are auto-deleted on first stop. Skip the
      // warm-pool path entirely so we always provision a fresh container on a
      // currently-unoccupied GPU runner.
      if (gpu <= 0 && !linkedSandbox && (!createSandboxDto.volumes || createSandboxDto.volumes.length === 0)) {
        const skipWarmPool = (await this.redis.exists(`warm-pool:skip:${snapshot.id}`)) === 1

        if (!skipWarmPool) {
          const warmPoolSandbox = await this.warmPoolService.fetchWarmPoolSandbox({
            organizationId: organization.id,
            snapshot,
            target: region.id,
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
        }
      }

      // Serialize GPU runner assignment per region: getRunnersAtGpuCapacity reads
      // the DB to find runners at capacity, but the just-assigned runnerId on a
      // concurrent request is not yet persisted, so two concurrent creates can
      // pick the same already-full runner. Hold the lock until the runnerId is
      // written to the DB. Only mark the key as held after acquisition succeeds —
      // otherwise a timed-out waiter would unlock the actual holder in finally.
      if (gpu > 0) {
        const key = `gpu-runner-assignment:${region.id}`
        await this.redisLockProvider.waitForLock(key, 60, 30000)
        gpuRunnerAssignmentLockKey = key
      }

      let runner: Runner
      if (linkedSandbox && linkedSandbox.runnerId) {
        runner = await this.runnerService.findOneOrFail(linkedSandbox.runnerId)

        if (runner.region !== region.id) {
          throw new BadRequestError(
            `Runner hosting linked sandbox is in region ${runner.region}, which does not match requested region ${region.id}`,
          )
        }

        this.runnerService.assertRunnerCanHost(runner)
      } else {
        runner = await this.runnerService.getRandomAvailableRunner({
          regions: [region.id],
          sandboxClass: snapshot.sandboxClass,
          snapshotRef: snapshot.ref,
          gpu,
          gpuType,
        })
      }

      const sandbox = new Sandbox({ region: region.id, name: createSandboxDto.name })

      sandbox.organizationId = organization.id

      sandbox.sandboxClass = snapshot.sandboxClass
      sandbox.snapshot = snapshot.name
      //  TODO: default user should be configurable
      sandbox.osUser = createSandboxDto.user || 'daytona'
      sandbox.env = createSandboxDto.env || {}
      sandbox.labels = createSandboxDto.labels || {}

      sandbox.cpu = cpu
      sandbox.gpu = gpu
      sandbox.gpuType = gpuType
      sandbox.mem = mem
      sandbox.disk = disk

      sandbox.public = createSandboxDto.public || false

      if (createSandboxDto.networkBlockAll !== undefined) {
        sandbox.networkBlockAll = createSandboxDto.networkBlockAll
      }

      if (createSandboxDto.networkAllowList !== undefined) {
        sandbox.networkAllowList = this.resolveNetworkAllowList(createSandboxDto.networkAllowList)
      }

      if (createSandboxDto.domainAllowList !== undefined) {
        sandbox.domainAllowList = this.resolveDomainAllowList(createSandboxDto.domainAllowList)
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

      if (resolvedVolumes !== undefined) {
        sandbox.volumes = resolvedVolumes
      }

      sandbox.runnerId = runner.id
      sandbox.linkedSandboxId = linkedSandbox?.id ?? null
      sandbox.pending = true

      const insertedSandbox = await this.sandboxRepository.insert(sandbox)

      if (gpuRunnerAssignmentLockKey) {
        const key = gpuRunnerAssignmentLockKey
        gpuRunnerAssignmentLockKey = undefined
        await this.redisLockProvider
          .unlock(key)
          .catch((err) => this.logger.error('Failed to release GPU runner assignment lock', err))
      }

      this.eventEmitter
        .emitAsync(SandboxEvents.CREATED, new SandboxCreatedEvent(insertedSandbox))
        .catch((err) => this.logger.error('Failed to emit SandboxCreatedEvent', err))

      return this.toSandboxDto(insertedSandbox)
    } catch (error) {
      if (sandboxClass) {
        await this.rollbackPendingUsage(
          organization.id,
          region.id,
          sandboxClass,
          pendingCpuIncrement,
          pendingMemoryIncrement,
          pendingDiskIncrement,
          pendingGpuIncrement,
        )
      }

      if (error.code === '23505') {
        throw new ConflictException(`Sandbox with name ${createSandboxDto.name} already exists`)
      }

      throw error
    } finally {
      if (gpuRunnerAssignmentLockKey) {
        await this.redisLockProvider
          .unlock(gpuRunnerAssignmentLockKey)
          .catch((err) => this.logger.error('Failed to release GPU runner assignment lock', err))
      }
    }
  }

  /**
   * Validates and resolves the optional linkedSandbox reference on a snapshot-based sandbox create request.
   *
   * Returns the linked Sandbox entity when linking is requested, otherwise undefined.
   *
   * @throws {BadRequestError} If any link precondition is not met.
   * @throws {NotFoundException} If the linked sandbox does not exist for the organization.
   */
  private async resolveLinkedSandbox(
    createSandboxDto: CreateSandboxDto,
    organizationId: string,
  ): Promise<Sandbox | undefined> {
    if (!createSandboxDto.linkedSandbox) {
      return undefined
    }

    if (!isEphemeral(createSandboxDto)) {
      throw new BadRequestError('Linked sandboxes must be ephemeral (set autoDeleteInterval to 0)')
    }

    const linkedSandbox = await this.findOneByIdOrName(createSandboxDto.linkedSandbox, organizationId)

    if (linkedSandbox.linkedSandboxId) {
      throw new BadRequestError(
        `Linked sandbox ${linkedSandbox.id} is itself linked to another sandbox; chained links are not allowed`,
      )
    }

    if (![SandboxState.STARTED, SandboxState.STOPPED].includes(linkedSandbox.state) || !linkedSandbox.runnerId) {
      throw new BadRequestError(
        `Linked sandbox must be in STARTED or STOPPED state with an assigned runner (current: ${linkedSandbox.state})`,
      )
    }

    return linkedSandbox
  }

  private async assignWarmPoolSandbox(
    warmPoolSandbox: Sandbox,
    createSandboxDto: CreateSandboxDto,
    organization: Organization,
  ): Promise<SandboxDto> {
    const now = new Date()
    const updateData: Partial<Sandbox> = {
      public: createSandboxDto.public || false,
      labels: createSandboxDto.labels || {},
      organizationId: organization.id,
      createdAt: now,
    }

    if (createSandboxDto.name) {
      updateData.name = createSandboxDto.name
    }

    if (createSandboxDto.autoStopInterval !== undefined) {
      updateData.autoStopInterval = this.resolveAutoStopInterval(createSandboxDto.autoStopInterval)
    }

    if (createSandboxDto.autoArchiveInterval !== undefined) {
      updateData.autoArchiveInterval = this.resolveAutoArchiveInterval(createSandboxDto.autoArchiveInterval)
    }

    if (warmPoolSandbox.gpu > 0) {
      if (createSandboxDto.autoDeleteInterval !== undefined && createSandboxDto.autoDeleteInterval !== 0) {
        throw new BadRequestError('GPU sandboxes must be ephemeral - autoDeleteInterval must be 0')
      }
      updateData.autoDeleteInterval = 0
    } else if (createSandboxDto.autoDeleteInterval !== undefined) {
      updateData.autoDeleteInterval = createSandboxDto.autoDeleteInterval
    }

    if (createSandboxDto.networkBlockAll !== undefined) {
      updateData.networkBlockAll = createSandboxDto.networkBlockAll
    }

    if (createSandboxDto.networkAllowList !== undefined) {
      updateData.networkAllowList = this.resolveNetworkAllowList(createSandboxDto.networkAllowList)
    }

    if (createSandboxDto.domainAllowList !== undefined) {
      updateData.domainAllowList = this.resolveDomainAllowList(createSandboxDto.domainAllowList)
    }

    if (!warmPoolSandbox.runnerId) {
      throw new SandboxError('Runner not found for warm pool sandbox')
    }

    if (
      createSandboxDto.networkBlockAll !== undefined ||
      createSandboxDto.networkAllowList !== undefined ||
      createSandboxDto.domainAllowList !== undefined ||
      organization.sandboxLimitedNetworkEgress
    ) {
      const runner = await this.runnerService.findOneOrFail(warmPoolSandbox.runnerId)
      const runnerAdapter = await this.runnerAdapterFactory.create(runner)
      await runnerAdapter.updateNetworkSettings(
        warmPoolSandbox.id,
        createSandboxDto.networkBlockAll,
        createSandboxDto.networkAllowList,
        organization.sandboxLimitedNetworkEgress,
        updateData.domainAllowList ?? undefined,
      )
    }

    const updatedSandbox = await this.sandboxRepository.update(warmPoolSandbox.id, {
      updateData,
      entity: warmPoolSandbox,
    })

    // Defensive invalidation of orgId cache since the sandbox moved from unassigned to a real organization
    this.sandboxLookupCacheInvalidationService.invalidateOrgId({
      sandboxId: warmPoolSandbox.id,
      organizationId: organization.id,
      name: warmPoolSandbox.name,
      previousOrganizationId: SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION,
    })

    // Treat this as a newly started sandbox
    this.eventEmitter.emit(
      SandboxEvents.STATE_UPDATED,
      new SandboxStateUpdatedEvent(updatedSandbox, SandboxState.STARTED, SandboxState.STARTED),
    )
    return this.toSandboxDto(updatedSandbox)
  }

  async createFromBuildInfo(createSandboxDto: CreateSandboxDto, organization: Organization): Promise<SandboxDto> {
    let pendingCpuIncrement: number | undefined
    let pendingMemoryIncrement: number | undefined
    let pendingDiskIncrement: number | undefined
    let pendingGpuIncrement: number | undefined
    let gpuRunnerAssignmentLockKey: string | undefined

    const region = await this.getValidatedOrDefaultRegion(organization, createSandboxDto.target)

    try {
      const cpu = createSandboxDto.cpu || DEFAULT_CPU
      const mem = createSandboxDto.memory || DEFAULT_MEMORY
      const disk = createSandboxDto.disk || DEFAULT_DISK
      const gpu = createSandboxDto.gpu || DEFAULT_GPU

      // GPU sandboxes are always ephemeral - delete on first stop.
      if (gpu > 0 && !isEphemeral(createSandboxDto)) {
        throw new BadRequestError('GPU sandboxes must be ephemeral - set autoDeleteInterval to 0')
      }

      this.organizationService.assertOrganizationIsNotSuspended(organization)

      const regionQuota = region.enforceQuotas
        ? await this.organizationService.getRegionQuota(organization.id, region.id, SandboxClass.CONTAINER)
        : null
      const gpuTypePreferences = resolveGpuTypePreferences(gpu, createSandboxDto.gpuType, regionQuota?.allowedGpuTypes)

      const { pendingCpuIncremented, pendingMemoryIncremented, pendingDiskIncremented, pendingGpuIncremented } =
        await this.validateOrganizationQuotas(
          organization,
          region,
          SandboxClass.CONTAINER,
          cpu,
          mem,
          disk,
          gpu,
          isEphemeral(createSandboxDto),
          undefined,
          regionQuota,
        )

      if (pendingCpuIncremented) {
        pendingCpuIncrement = cpu
      }
      if (pendingMemoryIncremented) {
        pendingMemoryIncrement = mem
      }
      if (pendingDiskIncremented) {
        pendingDiskIncrement = disk
      }
      if (pendingGpuIncremented) {
        pendingGpuIncrement = gpu
      }

      // Resolve volume names to UUIDs, failing fast on invalid references
      const resolvedVolumes = await this.resolveVolumes(organization.id, createSandboxDto.volumes)

      const sandbox = new Sandbox({ region: region.id, name: createSandboxDto.name })

      sandbox.organizationId = organization.id

      sandbox.sandboxClass = SandboxClass.CONTAINER
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

      if (createSandboxDto.domainAllowList !== undefined) {
        sandbox.domainAllowList = this.resolveDomainAllowList(createSandboxDto.domainAllowList)
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

      if (resolvedVolumes !== undefined) {
        sandbox.volumes = resolvedVolumes
      }

      if (sandbox.sandboxClass !== SandboxClass.CONTAINER) {
        throw new BadRequestError('Declarative builds are only supported for container-class sandboxes')
      }

      const buildInfoSnapshotRef = generateBuildSnapshotRef(
        createSandboxDto.buildInfo.dockerfileContent,
        createSandboxDto.buildInfo.contextHashes,
      )

      let runner: Runner

      // Serialize GPU runner assignment per region: getRunnersAtGpuCapacity reads
      // the DB to find runners at capacity, but the just-assigned runnerId on a
      // concurrent request is not yet persisted, so two concurrent creates can
      // pick the same already-full runner. Hold the lock until the runnerId is
      // written to the DB. Only mark the key as held after acquisition succeeds —
      // otherwise a timed-out waiter would unlock the actual holder in finally.
      if (sandbox.gpu > 0) {
        const key = `gpu-runner-assignment:${region.id}`
        await this.redisLockProvider.waitForLock(key, 60, 30000)
        gpuRunnerAssignmentLockKey = key
      }

      try {
        const declarativeBuildScoreThreshold = this.configService.get('runnerScore.thresholds.declarativeBuild')
        const maxCpuPerRunner = this.configService.getOrThrow('buildInfo.maxCpuPerRunner')
        const excludedRunnerIds =
          maxCpuPerRunner > 0
            ? await this.runnerService.getRunnersWithMaxBuildInfoSnapshotRefCpu(
                buildInfoSnapshotRef,
                maxCpuPerRunner,
                sandbox.cpu,
              )
            : []
        runner = await this.runnerService.getRandomAvailableRunner({
          regions: [sandbox.region],
          sandboxClass: sandbox.sandboxClass,
          snapshotRef: buildInfoSnapshotRef,
          gpu: sandbox.gpu,
          gpuType: gpuTypePreferences ?? null,
          ...(excludedRunnerIds.length > 0 && { excludedRunnerIds }),
          ...(declarativeBuildScoreThreshold !== undefined && {
            availabilityScoreThreshold: declarativeBuildScoreThreshold,
          }),
        })

        sandbox.runnerId = runner.id
        sandbox.gpuType = sandbox.gpu > 0 ? runner.gpuType : null
      } catch (error) {
        if (
          error instanceof BadRequestError == false ||
          !error.message.startsWith('No available runners') ||
          !createSandboxDto.buildInfo
        ) {
          throw error
        }
        sandbox.state = SandboxState.PENDING_BUILD
      }

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

      sandbox.pending = true

      const insertedSandbox = await this.sandboxRepository.insert(sandbox)

      if (gpuRunnerAssignmentLockKey) {
        const key = gpuRunnerAssignmentLockKey
        gpuRunnerAssignmentLockKey = undefined
        await this.redisLockProvider
          .unlock(key)
          .catch((err) => this.logger.error('Failed to release GPU runner assignment lock', err))
      }

      this.eventEmitter
        .emitAsync(SandboxEvents.CREATED, new SandboxCreatedEvent(insertedSandbox))
        .catch((err) => this.logger.error('Failed to emit SandboxCreatedEvent', err))

      return this.toSandboxDto(insertedSandbox)
    } catch (error) {
      await this.rollbackPendingUsage(
        organization.id,
        region.id,
        SandboxClass.CONTAINER,
        pendingCpuIncrement,
        pendingMemoryIncrement,
        pendingDiskIncrement,
        pendingGpuIncrement,
      )

      if (error.code === '23505') {
        throw new ConflictException(`Sandbox with name ${createSandboxDto.name} already exists`)
      }

      throw error
    } finally {
      if (gpuRunnerAssignmentLockKey) {
        await this.redisLockProvider
          .unlock(gpuRunnerAssignmentLockKey)
          .catch((err) => this.logger.error('Failed to release GPU runner assignment lock', err))
      }
    }
  }

  async createBackup(sandboxIdOrName: string, organizationId?: string): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    if (isEphemeral(sandbox)) {
      throw new SandboxError('Ephemeral sandboxes cannot be backed up')
    }

    if (![BackupState.COMPLETED, BackupState.NONE].includes(sandbox.backupState)) {
      throw new SandboxError('Sandbox backup is already in progress')
    }

    this.eventEmitter.emit(SandboxEvents.BACKUP_CREATED, new SandboxBackupCreatedEvent(sandbox))

    return sandbox
  }

  async forkSandbox(sandboxIdOrName: string, organization: Organization, dto: ForkSandboxDto): Promise<Sandbox> {
    let pendingCpuIncrement: number | undefined
    let pendingMemoryIncrement: number | undefined
    let pendingDiskIncrement: number | undefined
    let pendingGpuIncrement: number | undefined

    const sourceSandbox = await this.findOneByIdOrName(sandboxIdOrName, organization.id)

    if (![SandboxClass.LINUX_VM, SandboxClass.WINDOWS].includes(sourceSandbox.sandboxClass)) {
      throw new HttpException('Forking is not supported for this sandbox', HttpStatus.UNPROCESSABLE_ENTITY)
    }

    const region = await this.regionService.findOne(sourceSandbox.region)
    if (!region) {
      throw new NotFoundException(`Region with ID ${sourceSandbox.region} not found`)
    }

    try {
      if (sourceSandbox.state !== SandboxState.STARTED) {
        throw new BadRequestError('Sandbox must be in started state to fork')
      }

      if (sourceSandbox.pending) {
        throw new StateChangeInProgressError()
      }

      if (!sourceSandbox.runnerId) {
        throw new NotFoundException(`Sandbox with ID ${sourceSandbox.id} does not have a runner`)
      }

      if (sourceSandbox.gpu > 0) {
        throw new HttpException('Forking is not supported for GPU sandboxes', HttpStatus.UNPROCESSABLE_ENTITY)
      }

      const runner = await this.runnerService.findOneOrFail(sourceSandbox.runnerId)

      // Copy all properties from source sandbox to forked sandbox
      const forkedSandbox = new Sandbox({ region: sourceSandbox.region, name: dto.name })
      forkedSandbox.organizationId = organization.id
      forkedSandbox.sandboxClass = sourceSandbox.sandboxClass
      forkedSandbox.snapshot = sourceSandbox.snapshot
      forkedSandbox.osUser = sourceSandbox.osUser
      forkedSandbox.env = { ...sourceSandbox.env }
      forkedSandbox.labels = { ...sourceSandbox.labels }
      forkedSandbox.cpu = sourceSandbox.cpu
      forkedSandbox.mem = sourceSandbox.mem
      forkedSandbox.disk = sourceSandbox.disk
      forkedSandbox.gpu = sourceSandbox.gpu
      forkedSandbox.gpuType = sourceSandbox.gpuType ?? null
      forkedSandbox.public = sourceSandbox.public
      forkedSandbox.autoStopInterval = sourceSandbox.autoStopInterval
      forkedSandbox.autoArchiveInterval = sourceSandbox.autoArchiveInterval
      forkedSandbox.autoDeleteInterval = sourceSandbox.autoDeleteInterval
      forkedSandbox.volumes = sourceSandbox.volumes?.map((volume) => ({ ...volume }))
      forkedSandbox.networkBlockAll = sourceSandbox.networkBlockAll
      forkedSandbox.networkAllowList = sourceSandbox.networkAllowList
      forkedSandbox.runnerId = sourceSandbox.runnerId
      forkedSandbox.pending = true
      forkedSandbox.state = SandboxState.CREATING

      this.organizationService.assertOrganizationIsNotSuspended(organization)

      // Validate organization usage quotas are not exceeded due to new sandbox created by forking
      const { pendingCpuIncremented, pendingMemoryIncremented, pendingDiskIncremented, pendingGpuIncremented } =
        await this.validateOrganizationQuotas(
          organization,
          region,
          forkedSandbox.sandboxClass,
          forkedSandbox.cpu,
          forkedSandbox.mem,
          forkedSandbox.disk,
          forkedSandbox.gpu,
          isEphemeral(forkedSandbox),
        )

      if (pendingCpuIncremented) {
        pendingCpuIncrement = forkedSandbox.cpu
      }
      if (pendingMemoryIncremented) {
        pendingMemoryIncrement = forkedSandbox.mem
      }
      if (pendingDiskIncremented) {
        pendingDiskIncrement = forkedSandbox.disk
      }
      if (pendingGpuIncremented) {
        pendingGpuIncrement = forkedSandbox.gpu
      }

      // Capture state of source sandbox before transitioning to FORKING
      const sourceSandboxInitialState = sourceSandbox.state

      await this.sandboxRepository.updateWhere(sourceSandbox.id, {
        updateData: {
          state: SandboxState.FORKING,
          pending: true,
        },
        whereCondition: {
          state: sourceSandbox.state,
          pending: false,
        },
      })

      let insertedForkedSandbox: Sandbox | undefined

      try {
        insertedForkedSandbox = await this.sandboxRepository.insert(forkedSandbox, sourceSandbox.id)
        const runnerAdapter = await this.runnerAdapterFactory.create(runner)
        await runnerAdapter.forkSandbox(sourceSandbox.id, insertedForkedSandbox.id)
      } catch (error) {
        // Rollback source sandbox to its initial state
        await this.sandboxRepository.updateWhere(sourceSandbox.id, {
          updateData: {
            state: sourceSandboxInitialState,
            pending: false,
          },
          whereCondition: { state: SandboxState.FORKING },
        })

        if (insertedForkedSandbox) {
          try {
            const updateData = Sandbox.getSoftDeleteUpdate(insertedForkedSandbox)
            const destroyedSandbox = await this.sandboxRepository.updateWhere(insertedForkedSandbox.id, {
              updateData,
              whereCondition: { pending: true, state: SandboxState.CREATING },
            })
            this.eventEmitter.emit(SandboxEvents.DESTROYED, new SandboxDestroyedEvent(destroyedSandbox))
          } catch (destroyError) {
            this.logger.error(`Failed to rollback forked sandbox ${insertedForkedSandbox.id}`, destroyError)
          }
        }

        throw error
      }

      this.eventEmitter
        .emitAsync(SandboxEvents.CREATED, new SandboxCreatedEvent(insertedForkedSandbox))
        .catch((err) => this.logger.error('Failed to emit SandboxCreatedEvent', err))

      return insertedForkedSandbox
    } catch (error) {
      await this.rollbackPendingUsage(
        organization.id,
        region.id,
        sourceSandbox.sandboxClass,
        pendingCpuIncrement,
        pendingMemoryIncrement,
        pendingDiskIncrement,
        pendingGpuIncrement,
      )

      if (error.code === '23505') {
        throw new ConflictException('Sandbox with this name already exists')
      }

      throw error
    }
  }

  async getForkChildren(
    organizationId: string,
    sandboxIdOrName: string,
    includeDestroyed?: boolean,
  ): Promise<Sandbox[]> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)
    const forks = await this.sandboxForkRepository.find({
      where: {
        parentId: sandbox.id,
        child: {
          organizationId,
          ...(!includeDestroyed && { state: Not(SandboxState.DESTROYED) }),
        },
      },
      relations: ['child'],
      take: 100,
    })
    return forks.map((f) => f.child)
  }

  async getForkParent(organizationId: string, sandboxIdOrName: string): Promise<Sandbox | null> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)
    const fork = await this.sandboxForkRepository.findOne({
      where: { childId: sandbox.id },
      relations: ['parent'],
    })
    if (!fork || fork.parent.organizationId !== organizationId) {
      return null
    }
    return fork.parent
  }

  async getForkAncestors(organizationId: string, sandboxIdOrName: string): Promise<Sandbox[]> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)
    const ancestors: Sandbox[] = []
    const visitedIds = new Set<string>()
    let currentId = sandbox.id

    while (ancestors.length < 50) {
      const fork = await this.sandboxForkRepository.findOne({
        where: { childId: currentId },
        relations: ['parent'],
      })
      if (!fork || visitedIds.has(fork.parentId) || fork.parent.organizationId !== organizationId) {
        break
      }
      visitedIds.add(fork.parentId)
      ancestors.push(fork.parent)
      currentId = fork.parentId
    }

    return ancestors.reverse()
  }

  async createSnapshotFromSandbox(
    sandboxIdOrName: string,
    organization: Organization,
    dto: CreateSandboxSnapshotDto,
  ): Promise<Sandbox> {
    let pendingSnapshotCountIncrement: number | undefined

    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organization.id)

    const includeMemory = dto.includeMemory ?? false

    try {
      if (![SandboxState.STARTED, SandboxState.STOPPED].includes(sandbox.state)) {
        throw new BadRequestError('Sandbox must be in started or stopped state to create a snapshot')
      }

      if (sandbox.pending) {
        throw new StateChangeInProgressError()
      }

      if (!sandbox.runnerId) {
        throw new NotFoundException(`Sandbox with ID ${sandbox.id} does not have a runner`)
      }

      const runner = await this.runnerService.findOneOrFail(sandbox.runnerId)

      if (sandbox.sandboxClass === SandboxClass.WINDOWS) {
        if (includeMemory && sandbox.state !== SandboxState.STARTED) {
          throw new BadRequestError('Snapshots with memory require the Windows sandbox to be running (STARTED)')
        }
        if (!includeMemory && sandbox.state !== SandboxState.STOPPED) {
          throw new BadRequestError('Filesystem-only snapshots require the Windows sandbox to be stopped (STOPPED)')
        }
      } else if (includeMemory) {
        throw new BadRequestError('includeMemory is only supported for Windows sandboxes')
      }

      this.organizationService.assertOrganizationIsNotSuspended(organization)

      const region = await this.regionService.findOne(sandbox.region)
      if (!region) {
        throw new NotFoundException(`Region with ID ${sandbox.region} not found`)
      }

      let registry: DockerRegistry | undefined
      if (isRegistryBasedSandboxClass(sandbox.sandboxClass)) {
        registry = (await this.dockerRegistryService.getAvailableInternalRegistry(sandbox.region)) ?? undefined
        if (sandbox.sandboxClass === SandboxClass.CONTAINER && !registry) {
          throw new BadRequestError(
            'No internal registry is available for this sandbox region; cannot snapshot a container sandbox',
          )
        }
      }

      const { pendingSnapshotCountIncremented } = await this.snapshotService.validateOrganizationQuotas(
        organization,
        region,
        sandbox.sandboxClass,
        1,
        sandbox.cpu,
        sandbox.mem,
        sandbox.disk,
        sandbox.gpu,
      )

      if (pendingSnapshotCountIncremented) {
        pendingSnapshotCountIncrement = 1
      }

      const updatedSandbox = await this.sandboxRepository.updateWhere(sandbox.id, {
        updateData: {
          state: SandboxState.SNAPSHOTTING,
          pending: true,
        },
        whereCondition: {
          state: sandbox.state,
          pending: false,
        },
      })

      const runnerAdapter = await this.runnerAdapterFactory.create(runner)

      // v2 runners enqueue a SNAPSHOT_SANDBOX job and resolve immediately
      // with `undefined`; the job state handler will persist the resulting
      // Snapshot on completion.
      //
      // v0 runners (Docker, container class) don't have jobs - the adapter
      // performs a synchronous HTTP call to the runner's commit+push
      // endpoint, which can take several minutes. We don't want to block
      // the API request that long, so for v0 we kick the call off in the
      // background and immediately return the SNAPSHOTTING sandbox to the
      // caller. The background promise persists the snapshot or reverts
      // sandbox state on failure.
      if (runner.apiVersion === '0') {
        // Hand off pending-quota ownership to the background driver - it must
        // roll back on failure since the outer try/catch returns successfully
        // here.
        const inheritedPendingIncrement = pendingSnapshotCountIncrement
        pendingSnapshotCountIncrement = undefined

        void this.runV0SnapshotFromSandbox({
          sandbox,
          previousState: sandbox.state,
          snapshotName: dto.name,
          organizationId: organization.id,
          registry,
          runner,
          runnerAdapter,
          pendingSnapshotCountIncrement: inheritedPendingIncrement,
        })
      } else {
        try {
          await runnerAdapter.createSnapshotFromSandbox(sandbox.id, dto.name, organization.id, registry, includeMemory)
        } catch (error) {
          await this.sandboxRepository.updateWhere(sandbox.id, {
            updateData: {
              state: sandbox.state,
              pending: false,
            },
            whereCondition: { state: SandboxState.SNAPSHOTTING },
          })

          throw error
        }
      }

      return updatedSandbox
    } catch (error) {
      await this.snapshotService.rollbackPendingUsage(organization.id, pendingSnapshotCountIncrement)
      throw error
    }
  }

  /**
   * Background driver for v0 (Docker) sandbox snapshotting.
   *
   * v0 runners expose a synchronous HTTP endpoint for "commit + push" that
   * may take several minutes. We run it in a background promise so the
   * user-facing API request can return immediately with state=SNAPSHOTTING,
   * mirroring the v2 (job-driven) UX.
   *
   * Errors are intentionally swallowed - sandbox state is the source of
   * truth. On any failure we restore state and refund the pending quota.
   */
  private async runV0SnapshotFromSandbox(params: {
    sandbox: Sandbox
    previousState: SandboxState
    snapshotName: string
    organizationId: string
    registry?: DockerRegistry
    runner: Runner
    runnerAdapter: RunnerAdapter
    pendingSnapshotCountIncrement?: number
  }): Promise<void> {
    const {
      sandbox,
      previousState,
      snapshotName,
      organizationId,
      registry,
      runner,
      runnerAdapter,
      pendingSnapshotCountIncrement,
    } = params

    let succeeded = false
    try {
      const result = await runnerAdapter.createSnapshotFromSandbox(sandbox.id, snapshotName, organizationId, registry)
      if (!result) {
        throw new Error('runner returned no snapshot result')
      }

      await this.snapshotService.persistSnapshotFromSandbox({
        organizationId,
        name: snapshotName,
        ref: result.ref,
        runnerId: runner.id,
        regionId: sandbox.region,
        sandboxClass: sandbox.sandboxClass,
        cpu: sandbox.cpu,
        gpu: sandbox.gpu,
        gpuType: sandbox.gpuType ?? null,
        mem: sandbox.mem,
        disk: sandbox.disk,
        sizeGB: result.sizeGB,
      })
      succeeded = true
    } catch (error) {
      this.logger.error(`v0 snapshotFromSandbox failed for sandbox ${sandbox.id}:`, error)
    }

    // Always clear SNAPSHOTTING - whether the snapshot was persisted or not,
    // the sandbox itself should return to its previous state.
    await this.sandboxRepository
      .updateWhere(sandbox.id, {
        updateData: { state: previousState, pending: false },
        whereCondition: { state: SandboxState.SNAPSHOTTING },
      })
      .catch((err) => this.logger.error(`Failed to clear SNAPSHOTTING state for sandbox ${sandbox.id}:`, err))

    if (!succeeded) {
      await this.snapshotService
        .rollbackPendingUsage(organizationId, pendingSnapshotCountIncrement)
        .catch((err) => this.logger.error(`Failed to roll back pending snapshot quota for org ${organizationId}:`, err))
    }
  }

  async findAllPaginatedDeprecated(
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
      field?: SandboxSortFieldDeprecated
      direction?: SandboxSortDirectionDeprecated
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

    const {
      field: sortField = DEFAULT_SANDBOX_SORT_FIELD_DEPRECATED,
      direction: sortDirection = DEFAULT_SANDBOX_SORT_DIRECTION_DEPRECATED,
    } = sort || {}

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

    const lastActivityFilter = createRangeFilter(lastEventAfter, lastEventBefore)
    if (lastActivityFilter) {
      baseFindOptions.lastActivityAt = { lastActivityAt: lastActivityFilter }
    }

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
      relations: ['lastActivityAt'],
      order: {
        ...(sortField === SandboxSortFieldDeprecated.LAST_ACTIVITY_AT
          ? { lastActivityAt: { lastActivityAt: { direction: sortDirection, nulls: 'LAST' } } }
          : {
              [sortField]: {
                direction: sortDirection,
                nulls: 'LAST',
              },
            }),
        ...(sortField !== SandboxSortFieldDeprecated.CREATED_AT && { createdAt: 'DESC' }),
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

  /**
   * Search sandboxes
   * @param organizationId - The ID of the organization
   * @param query - The query parameters
   * @returns The paginated list of sandboxes. If cursor is omitted from the query, newest sandboxes will be returned.
   * @throws BadRequestError if the cursor is invalid
   */
  async search(organizationId: string, query: ListSandboxesQueryDto): Promise<ListSandboxesResponseDto> {
    let parsedLabels: { [key: string]: string } | undefined
    if (query.labels) {
      try {
        parsedLabels = JSON.parse(query.labels)
      } catch {
        throw new BadRequestError('Invalid labels JSON format')
      }
    }

    const result = await this.sandboxSearchAdapter.search({
      filters: {
        organizationId,
        idPrefix: query.id,
        namePrefix: query.name,
        labels: parsedLabels,
        includeErroredDeleted: query.includeErroredDeleted,
        states: query.states,
        snapshots: query.snapshots,
        regionIds: query.regionIds,
        sandboxClasses: query.sandboxClasses,
        minCpu: query.minCpu,
        maxCpu: query.maxCpu,
        minMemoryGiB: query.minMemoryGiB,
        maxMemoryGiB: query.maxMemoryGiB,
        minDiskGiB: query.minDiskGiB,
        maxDiskGiB: query.maxDiskGiB,
        isPublic: query.isPublic,
        isRecoverable: query.isRecoverable,
        createdAtAfter: query.createdAtAfter,
        createdAtBefore: query.createdAtBefore,
        lastEventAfter: query.lastEventAfter,
        lastEventBefore: query.lastEventBefore,
      },
      pagination: {
        cursor: query.cursor,
        limit: query.limit,
      },
      sort: {
        field: query.sort,
        direction: query.order,
      },
    })

    const targets = [...new Set(result.items.map((item) => item.target))]
    const toolboxProxyUrlMap = await this.resolveToolboxProxyUrls(targets)

    return {
      items: result.items.map((item) => {
        const url = toolboxProxyUrlMap.get(item.target)
        if (!url) {
          throw new NotFoundException(`Toolbox proxy URL not resolved for region ${item.target}`)
        }
        item.toolboxProxyUrl = url
        return item
      }),
      nextCursor: result.nextCursor,
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
      case SandboxState.PAUSED:
        return SandboxDesiredState.PAUSED
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

    let sandboxes = await this.sandboxRepository.find({ where, relations: ['lastActivityAt'] })

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
    const relations = ['buildInfo', 'lastActivityAt']

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
      relations: ['lastActivityAt'],
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
      cache: {
        id: sandboxOrgIdCacheKeyById({ organizationId, sandboxId: sandboxIdOrName }),
        milliseconds: SANDBOX_ORG_ID_CACHE_TTL_MS,
      },
    })

    if (!sandbox && organizationId) {
      sandbox = await this.sandboxRepository.findOne({
        where: {
          name: sandboxIdOrName,
          organizationId: organizationId,
        },
        select: ['organizationId'],
        cache: {
          id: sandboxOrgIdCacheKeyByName({ organizationId, sandboxName: sandboxIdOrName }),
          milliseconds: SANDBOX_ORG_ID_CACHE_TTL_MS,
        },
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

    if (sandbox.pending && sandbox.state !== SandboxState.PENDING_BUILD) {
      throw new StateChangeInProgressError()
    }

    const forkChildren = await this.sandboxForkRepository.find({
      where: { parentId: sandbox.id },
      relations: ['child'],
    })
    const activeChildren = forkChildren.filter((f) => f.child && f.child.desiredState !== SandboxDesiredState.DESTROYED)
    if (activeChildren.length > 0) {
      throw new BadRequestError(
        'Cannot delete sandbox which has active fork children. The forks must be deleted first.',
      )
    }

    const updateData = Sandbox.getSoftDeleteUpdate(sandbox)

    const updatedSandbox = await this.sandboxRepository.updateWhere(sandbox.id, {
      updateData,
      whereCondition: { pending: sandbox.pending, state: sandbox.state },
    })

    this.eventEmitter.emit(SandboxEvents.DESTROYED, new SandboxDestroyedEvent(updatedSandbox))
    return updatedSandbox
  }

  async start(sandboxIdOrName: string, organization: Organization): Promise<Sandbox> {
    let pendingCpuIncrement: number | undefined
    let pendingMemoryIncrement: number | undefined
    let pendingDiskIncrement: number | undefined
    let pendingGpuIncrement: number | undefined

    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organization.id)

    const region = await this.regionService.findOne(sandbox.region)
    if (!region) {
      throw new NotFoundException(`Region with ID ${sandbox.region} not found`)
    }

    try {
      if (sandbox.state === SandboxState.STARTED && sandbox.desiredState === SandboxDesiredState.STARTED) {
        return sandbox
      }

      this.assertSandboxNotErrored(sandbox)

      const wasPaused = sandbox.state === SandboxState.PAUSED

      if (String(sandbox.state) !== String(sandbox.desiredState)) {
        // Allow start of stopped | archived and archiving | archived sandboxes
        if (
          sandbox.desiredState !== SandboxDesiredState.ARCHIVED ||
          (sandbox.state !== SandboxState.STOPPED && sandbox.state !== SandboxState.ARCHIVING)
        ) {
          throw new StateChangeInProgressError()
        }
      }

      if (
        ![SandboxState.STOPPED, SandboxState.ARCHIVED, SandboxState.ARCHIVING, SandboxState.PAUSED].includes(
          sandbox.state,
        )
      ) {
        throw new SandboxError('Sandbox is not in valid state')
      }

      if (sandbox.pending) {
        throw new StateChangeInProgressError()
      }

      if (wasPaused && ![SandboxClass.LINUX_VM, SandboxClass.WINDOWS].includes(sandbox.sandboxClass)) {
        throw new HttpException('Resuming is not supported for this sandbox', HttpStatus.UNPROCESSABLE_ENTITY)
      }

      this.organizationService.assertOrganizationIsNotSuspended(organization)

      const { pendingCpuIncremented, pendingMemoryIncremented, pendingDiskIncremented, pendingGpuIncremented } =
        await this.validateOrganizationQuotas(
          organization,
          region,
          sandbox.sandboxClass,
          sandbox.cpu,
          sandbox.mem,
          sandbox.disk,
          sandbox.gpu,
          isEphemeral(sandbox),
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
      if (pendingGpuIncremented) {
        pendingGpuIncrement = sandbox.gpu
      }

      const updateData: Partial<Sandbox> = wasPaused
        ? {
            pending: true,
            desiredState: SandboxDesiredState.STARTED,
          }
        : {
            pending: true,
            desiredState: SandboxDesiredState.STARTED,
            authToken: nanoid(32).toLocaleLowerCase(),
          }

      const updatedSandbox = await this.sandboxRepository.updateWhere(sandbox.id, {
        updateData,
        whereCondition: { pending: false, state: sandbox.state },
      })

      this.eventEmitter.emit(SandboxEvents.STARTED, new SandboxStartedEvent(updatedSandbox))

      return updatedSandbox
    } catch (error) {
      await this.rollbackPendingUsage(
        organization.id,
        sandbox.region,
        sandbox.sandboxClass,
        pendingCpuIncrement,
        pendingMemoryIncrement,
        pendingDiskIncrement,
        pendingGpuIncrement,
      )
      throw error
    }
  }

  async stop(sandboxIdOrName: string, organizationId?: string, force?: boolean): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    this.assertSandboxNotErrored(sandbox)

    if (String(sandbox.state) !== String(sandbox.desiredState)) {
      throw new StateChangeInProgressError()
    }

    if (sandbox.state !== SandboxState.STARTED && sandbox.state !== SandboxState.PAUSED) {
      throw new SandboxError('Sandbox is not in a stoppable state')
    }

    if (sandbox.pending) {
      throw new StateChangeInProgressError()
    }

    let updateData: Partial<Sandbox> = {}
    if (isEphemeral(sandbox)) {
      updateData = Sandbox.getSoftDeleteUpdate(sandbox)
    } else {
      updateData.pending = true
      updateData.desiredState = SandboxDesiredState.STOPPED
    }

    const updatedSandbox = await this.sandboxRepository.updateWhere(sandbox.id, {
      updateData,
      whereCondition: { pending: false, state: sandbox.state },
    })

    if (isEphemeral(sandbox)) {
      this.eventEmitter.emit(SandboxEvents.DESTROYED, new SandboxDestroyedEvent(updatedSandbox))
    } else {
      this.eventEmitter.emit(SandboxEvents.STOPPED, new SandboxStoppedEvent(updatedSandbox, force))
    }

    return updatedSandbox
  }

  async pause(sandboxIdOrName: string, organization: Organization): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organization.id)

    if (sandbox.state !== SandboxState.STARTED) {
      throw new BadRequestError('Sandbox must be in started state to pause')
    }

    if (sandbox.pending) {
      throw new StateChangeInProgressError()
    }

    if (![SandboxClass.LINUX_VM, SandboxClass.WINDOWS].includes(sandbox.sandboxClass)) {
      throw new HttpException('Pausing is not supported for this sandbox', HttpStatus.UNPROCESSABLE_ENTITY)
    }

    if (!sandbox.runnerId) {
      throw new NotFoundException(`Sandbox with ID ${sandbox.id} does not have a runner`)
    }

    const runner = await this.runnerService.findOneOrFail(sandbox.runnerId)

    await this.sandboxRepository.updateWhere(sandbox.id, {
      updateData: {
        state: SandboxState.PAUSING,
        desiredState: SandboxDesiredState.PAUSED,
        pending: true,
      },
      whereCondition: {
        state: SandboxState.STARTED,
        pending: false,
      },
    })

    try {
      const runnerAdapter = await this.runnerAdapterFactory.create(runner)
      await runnerAdapter.pauseSandbox(sandbox.id)
    } catch (error) {
      // Rollback to STARTED on error
      await this.sandboxRepository.updateWhere(sandbox.id, {
        updateData: {
          state: SandboxState.STARTED,
          desiredState: SandboxDesiredState.STARTED,
          pending: false,
        },
        whereCondition: { state: SandboxState.PAUSING },
      })
      throw error
    }

    return this.findOneByIdOrName(sandbox.id, organization.id)
  }

  async recover(sandboxIdOrName: string, organization: Organization, skipStart = false): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organization.id)

    if (!sandbox.recoverable) {
      throw new BadRequestError('Sandbox is not in a recoverable state')
    }

    const region = await this.regionService.findOne(sandbox.region)
    if (!region) {
      throw new NotFoundException(`Region with ID ${sandbox.region} not found`)
    }

    // Serialize against concurrent recover calls and the draining-runner manager (which takes
    // the same lock). The pending flag can't be used here: enforceInvariants forces pending=false
    // when state=ERROR (sandbox.entity.ts:390-395), so updateWhere claims don't stick.
    const lockKey = getStateChangeLockKey(sandbox.id)
    if (!(await this.redisLockProvider.lock(lockKey, 60))) {
      throw new StateChangeInProgressError()
    }

    try {
      if (sandbox.state !== SandboxState.ERROR) {
        throw new BadRequestError('Sandbox must be in error state to recover')
      }

      this.organizationService.assertOrganizationIsNotSuspended(organization)

      // API-level recoverable errors (e.g. timeouts) bypass the runner and restore
      // from backup on a new runner, provided a completed backup exists.
      if (
        isApiRecoverableError(sandbox.errorReason) &&
        sandbox.backupState === BackupState.COMPLETED &&
        sandbox.backupSnapshot &&
        sandbox.backupRegistryId
      ) {
        return await this.recoverFromBackup(sandbox, organization, region)
      }

      // Everything else goes to the runner for in-place recovery (e.g. disk expansion).
      return await this.recoverInPlace(sandbox, organization, region, skipStart)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  private async recoverFromBackup(sandbox: Sandbox, organization: Organization, region: Region): Promise<Sandbox> {
    const cooldownKey = `sandbox:recover-from-backup:${sandbox.id}`
    const existing = await this.redis.get(cooldownKey)
    if (existing) {
      throw new ConflictException('Sandbox recovery has been attempted recently. Please try again later')
    }

    // The sandbox will be fully recreated from backup on a new runner, so reserve all resources.
    const { pendingCpuIncremented, pendingMemoryIncremented, pendingDiskIncremented, pendingGpuIncremented } =
      await this.validateOrganizationQuotas(
        organization,
        region,
        sandbox.sandboxClass,
        sandbox.cpu,
        sandbox.mem,
        sandbox.disk,
        sandbox.gpu,
        isEphemeral(sandbox),
        sandbox.id,
      )

    try {
      // Transition the sandbox to ARCHIVED with desired state STARTED so the sync loop
      // picks it up, assigns a new runner, and restores from the completed backup.
      // enforceInvariants will set pending=true and runnerId=null for ARCHIVED state.
      const updateData: Partial<Sandbox> = {
        state: SandboxState.ARCHIVED,
        desiredState: SandboxDesiredState.STARTED,
        errorReason: null,
        recoverable: false,
        authToken: nanoid(32).toLocaleLowerCase(),
        ...(sandbox.runnerId && { prevRunnerId: sandbox.runnerId }),
      }

      const updatedSandbox = await this.sandboxRepository.updateWhere(sandbox.id, {
        updateData,
        whereCondition: { recoverable: true, pending: false, state: SandboxState.ERROR },
      })

      await this.redis.set(cooldownKey, '1', 'EX', 3600)

      this.eventEmitter.emit(SandboxEvents.STARTED, new SandboxStartedEvent(updatedSandbox))

      return updatedSandbox
    } catch (error) {
      await this.rollbackPendingUsage(
        organization.id,
        sandbox.region,
        sandbox.sandboxClass,
        pendingCpuIncremented ? sandbox.cpu : undefined,
        pendingMemoryIncremented ? sandbox.mem : undefined,
        pendingDiskIncremented ? sandbox.disk : undefined,
        pendingGpuIncremented ? sandbox.gpu : undefined,
      )
      throw error
    }
  }

  private async recoverInPlace(
    sandbox: Sandbox,
    organization: Organization,
    region: Region,
    skipStart: boolean,
  ): Promise<Sandbox> {
    if (!sandbox.runnerId) {
      throw new NotFoundException(`Sandbox with ID ${sandbox.id} does not have a runner`)
    }
    const runner = await this.runnerService.findOneOrFail(sandbox.runnerId)
    const willStartOnV2 = runner.apiVersion === '2' && !skipStart

    // ERROR → STOPPED activates disk usage; v2 + !skipStart additionally activates cpu/mem/gpu
    // because there is no trailing start() call to validate them.
    const { pendingCpuIncremented, pendingMemoryIncremented, pendingDiskIncremented, pendingGpuIncremented } =
      await this.validateOrganizationQuotas(
        organization,
        region,
        sandbox.sandboxClass,
        willStartOnV2 ? sandbox.cpu : 0,
        willStartOnV2 ? sandbox.mem : 0,
        sandbox.disk,
        willStartOnV2 ? sandbox.gpu : 0,
        isEphemeral(sandbox),
        sandbox.id,
      )

    try {
      // Normalize desiredState upfront so the job handler can detect mid-job intent changes.
      if (runner.apiVersion === '2') {
        await this.sandboxRepository.updateWhere(sandbox.id, {
          updateData: {
            desiredState: skipStart ? SandboxDesiredState.STOPPED : SandboxDesiredState.STARTED,
            ...(willStartOnV2 && { authToken: nanoid(32).toLocaleLowerCase() }),
          },
          whereCondition: { state: SandboxState.ERROR },
        })
      }

      const runnerAdapter = await this.runnerAdapterFactory.create(runner)

      const backupRegistry = sandbox.backupRegistryId
        ? ((await this.dockerRegistryService.findOne(sandbox.backupRegistryId)) ?? undefined)
        : undefined

      if (sandbox.backupRegistryId && !backupRegistry) {
        this.logger.warn(
          `Backup registry ${sandbox.backupRegistryId} not found for sandbox ${sandbox.id}; proceeding without registry credentials`,
        )
      }

      try {
        await runnerAdapter.recoverSandbox(sandbox, backupRegistry, skipStart)
      } catch (error) {
        if (error instanceof Error && error.message.includes('storage cannot be further expanded')) {
          throw new ForbiddenException(
            `Sandbox storage cannot be further expanded. Maximum expansion of ${(sandbox.disk * 0.1).toFixed(2)}GB (10% of original ${sandbox.disk.toFixed(2)}GB) has been reached. Please contact support for further assistance.`,
          )
        }
        throw error
      }

      // v2: job-completion handler writes the terminal state and chains START_SANDBOX if needed.
      if (runner.apiVersion === '2') {
        return sandbox
      }

      const updateData: Partial<Sandbox> = {
        state: SandboxState.STOPPED,
        desiredState: SandboxDesiredState.STOPPED,
        errorReason: null,
        recoverable: false,
        // Clear transient backup state so the poller won't resume a retry post-recover.
        backupState: BackupState.NONE,
        backupErrorReason: null,
      }

      // Only wipe the snapshot pointer on a failed backup — a COMPLETED one is still valid.
      if (sandbox.backupState === BackupState.ERROR) {
        updateData.backupSnapshot = null
      }

      const updatedSandbox = await this.sandboxRepository.updateWhere(sandbox.id, {
        updateData,
        whereCondition: { recoverable: true, pending: false, state: sandbox.state },
      })

      if (skipStart) {
        return updatedSandbox
      }

      // start() validates cpu/mem with self-excluded so disk doesn't double-count.
      return await this.start(sandbox.id, organization)
    } catch (error) {
      await this.rollbackPendingUsage(
        organization.id,
        sandbox.region,
        sandbox.sandboxClass,
        pendingCpuIncremented ? sandbox.cpu : undefined,
        pendingMemoryIncremented ? sandbox.mem : undefined,
        pendingDiskIncremented ? sandbox.disk : undefined,
        pendingGpuIncremented ? sandbox.gpu : undefined,
      )
      throw error
    }
  }

  async resize(sandboxIdOrName: string, resizeDto: ResizeSandboxDto, organization: Organization): Promise<Sandbox> {
    let pendingCpuIncrement: number | undefined
    let pendingMemoryIncrement: number | undefined
    let pendingDiskIncrement: number | undefined
    let pendingGpuIncrement: number | undefined

    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organization.id)

    const region = await this.regionService.findOne(sandbox.region)
    if (!region) {
      throw new NotFoundException(`Region with ID ${sandbox.region} not found`)
    }

    try {
      // Validate sandbox is in a valid state for resize
      if (sandbox.state !== SandboxState.STARTED && sandbox.state !== SandboxState.STOPPED) {
        throw new BadRequestError('Sandbox must be in started or stopped state to resize')
      }

      if (sandbox.pending) {
        throw new StateChangeInProgressError()
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

      const regionQuota = region.enforceQuotas
        ? await this.organizationService.getRegionQuota(organization.id, region.id, sandbox.sandboxClass)
        : null

      const { maxCpuPerSandbox, maxMemoryPerSandbox, maxDiskPerSandbox, maxDiskPerNonEphemeralSandbox } =
        getEffectivePerSandboxLimits(organization, regionQuota, sandbox.gpu > 0)

      if (newCpu > maxCpuPerSandbox) {
        throw new BadRequestError(
          `CPU request ${newCpu} exceeds maximum allowed per sandbox (${maxCpuPerSandbox}).\n${PER_SANDBOX_LIMIT_MESSAGE}`,
        )
      }
      if (newMem > maxMemoryPerSandbox) {
        throw new BadRequestError(
          `Memory request ${newMem}GB exceeds maximum allowed per sandbox (${maxMemoryPerSandbox}GB).\n${PER_SANDBOX_LIMIT_MESSAGE}`,
        )
      }
      if (newDisk > maxDiskPerSandbox) {
        throw new BadRequestError(
          `Disk request ${newDisk}GB exceeds maximum allowed per sandbox (${maxDiskPerSandbox}GB).\n${PER_SANDBOX_LIMIT_MESSAGE}`,
        )
      }

      if (!isEphemeral(sandbox) && maxDiskPerNonEphemeralSandbox !== null) {
        if (maxDiskPerNonEphemeralSandbox === 0) {
          throw new BadRequestError('Non-ephemeral sandboxes are not permitted in this region')
        }
        if (newDisk > maxDiskPerNonEphemeralSandbox) {
          throw new BadRequestError(
            `Disk request ${newDisk}GB exceeds maximum allowed per non-ephemeral sandbox (${maxDiskPerNonEphemeralSandbox}GB).\n${PER_SANDBOX_LIMIT_MESSAGE}`,
          )
        }
      }

      // For cold resize, cpu/memory don't affect quota until sandbox is STARTED.
      // For hot resize, track all deltas (positive reserves quota, negative frees quota for others).
      const cpuDeltaForQuota = isHotResize ? newCpu - sandbox.cpu : 0
      const memDeltaForQuota = isHotResize ? newMem - sandbox.mem : 0
      const diskDeltaForQuota = newDisk - sandbox.disk // Disk only increases (validated at start of method)

      // Validate and track pending for any non-zero quota changes.
      // Resize never changes GPU allocation — always pass 0 for the GPU delta, but pass
      // `gpuEnabled = sandbox.gpu > 0` so per-sandbox limit checks use the GPU-specific table.
      if (cpuDeltaForQuota !== 0 || memDeltaForQuota !== 0 || diskDeltaForQuota !== 0) {
        const { pendingCpuIncremented, pendingMemoryIncremented, pendingDiskIncremented } =
          await this.validateOrganizationQuotas(
            organization,
            region,
            sandbox.sandboxClass,
            cpuDeltaForQuota,
            memDeltaForQuota,
            diskDeltaForQuota,
            0,
            isEphemeral(sandbox),
            undefined,
            regionQuota,
            sandbox.gpu > 0,
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

      const runner = await this.runnerService.findOneOrFail(sandbox.runnerId)

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
      const updateData: Partial<Sandbox> = {
        state: SandboxState.RESIZING,
      }

      await this.sandboxRepository.updateWhere(sandbox.id, {
        updateData,
        whereCondition: { pending: false, state: previousState },
      })

      try {
        const runnerAdapter = await this.runnerAdapterFactory.create(runner)

        const backupRegistry = sandbox.backupRegistryId
          ? ((await this.dockerRegistryService.findOne(sandbox.backupRegistryId)) ?? undefined)
          : undefined

        if (sandbox.backupRegistryId && !backupRegistry) {
          this.logger.warn(
            `Backup registry ${sandbox.backupRegistryId} not found for sandbox ${sandbox.id}; proceeding without registry credentials`,
          )
        }

        await runnerAdapter.resizeSandbox(sandbox.id, resizeDto.cpu, resizeDto.memory, resizeDto.disk, backupRegistry)

        // For V0 runners, update resources immediately (subscriber emits STATE_UPDATED)
        // For V2 runners, job handler will update resources on completion
        if (runner.apiVersion === '0') {
          const updateData: Partial<Sandbox> = {
            cpu: newCpu,
            mem: newMem,
            disk: newDisk,
            state: previousState,
          }

          await this.sandboxRepository.updateWhere(sandbox.id, {
            updateData,
            whereCondition: { state: SandboxState.RESIZING },
          })

          // Apply the usage change (increments current, decrements pending)
          // Only apply deltas for quotas that were validated/pending-incremented.
          // Resize never changes GPU allocation — always pass 0.
          await this.organizationUsageService.applyResizeUsageChange(
            organization.id,
            sandbox.region,
            sandbox.sandboxClass,
            cpuDeltaForQuota,
            memDeltaForQuota,
            diskDeltaForQuota,
            0,
          )
        }

        return await this.findOneByIdOrName(sandbox.id, organization.id)
      } catch (error) {
        // Return to previous state on error
        const updateData: Partial<Sandbox> = {
          state: previousState,
        }

        await this.sandboxRepository.updateWhere(sandbox.id, {
          updateData,
          whereCondition: { state: SandboxState.RESIZING },
        })

        throw error
      }
    } catch (error) {
      await this.rollbackPendingUsage(
        organization.id,
        sandbox.region,
        sandbox.sandboxClass,
        pendingCpuIncrement,
        pendingMemoryIncrement,
        pendingDiskIncrement,
        pendingGpuIncrement,
      )
      throw error
    }
  }

  async updatePublicStatus(sandboxIdOrName: string, isPublic: boolean, organizationId?: string): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    const updateData: Partial<Sandbox> = {
      public: isPublic,
    }

    return await this.sandboxRepository.update(sandbox.id, {
      updateData,
      entity: sandbox,
    })
  }

  async updateLastActivityAt(sandboxId: string, lastActivityAt: Date): Promise<void> {
    await this.sandboxActivityService.updateLastActivityAt(sandboxId, lastActivityAt)
  }

  async getToolboxProxyUrl(sandboxId: string): Promise<string> {
    const sandbox = await this.findOne(sandboxId)
    return this.resolveToolboxProxyUrl(sandbox.region)
  }

  async toSandboxDto(sandbox: Sandbox): Promise<SandboxDto> {
    const toolboxProxyUrl = await this.resolveToolboxProxyUrl(sandbox.region)
    return SandboxDto.fromSandbox(sandbox, toolboxProxyUrl)
  }

  async toSandboxDtos(sandboxes: Sandbox[]): Promise<SandboxDto[]> {
    const urlMap = await this.resolveToolboxProxyUrls(sandboxes.map((s) => s.region))
    return sandboxes.map((s) => {
      const url = urlMap.get(s.region)
      if (!url) {
        throw new NotFoundException(`Toolbox proxy URL not resolved for region ${s.region}`)
      }
      return SandboxDto.fromSandbox(s, url)
    })
  }

  async resolveToolboxProxyUrl(regionId: string): Promise<string> {
    const cacheKey = toolboxProxyUrlCacheKey(regionId)
    const cached = await this.redis.get(cacheKey)
    if (cached) {
      return cached
    }

    const region = await this.regionService.findOne(regionId)
    const url = region?.toolboxProxyUrl
      ? region.toolboxProxyUrl.replace(/\/+$/, '') + '/toolbox'
      : this.configService.getOrThrow('proxy.toolboxUrl')

    this.redis.setex(cacheKey, TOOLBOX_PROXY_URL_CACHE_TTL_S, url).catch((err) => {
      this.logger.warn(`Failed to cache toolbox proxy URL for region ${regionId}: ${err.message}`)
    })
    return url
  }

  async resolveToolboxProxyUrls(regionIds: string[]): Promise<Map<string, string>> {
    const unique = [...new Set(regionIds)]
    const result = new Map<string, string>()

    const pipeline = this.redis.pipeline()
    for (const id of unique) {
      pipeline.get(toolboxProxyUrlCacheKey(id))
    }
    const cached = await pipeline.exec()

    const uncached: string[] = []
    for (let i = 0; i < unique.length; i++) {
      const err = cached?.[i]?.[0]
      if (err) {
        this.logger.warn(`Failed to get cached toolbox proxy URL for region ${unique[i]}: ${err.message}`)
      }
      const val = cached?.[i]?.[1] as string | null
      if (val) {
        result.set(unique[i], val)
      } else {
        uncached.push(unique[i])
      }
    }

    if (uncached.length > 0) {
      const regions = await this.regionService.findByIds(uncached)
      const regionMap = new Map(regions.map((r) => [r.id, r]))
      const fallback = this.configService.getOrThrow('proxy.toolboxUrl')
      const setPipeline = this.redis.pipeline()
      for (const id of uncached) {
        const region = regionMap.get(id)
        const url = region?.toolboxProxyUrl ? region.toolboxProxyUrl.replace(/\/+$/, '') + '/toolbox' : fallback
        result.set(id, url)
        setPipeline.setex(toolboxProxyUrlCacheKey(id), TOOLBOX_PROXY_URL_CACHE_TTL_S, url)
      }
      const setResults = await setPipeline.exec()
      setResults?.forEach(([err], i) => {
        if (err) {
          this.logger.warn(`Failed to cache toolbox proxy URL for region ${uncached[i]}: ${err.message}`)
        }
      })
    }

    return result
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

  private async getValidatedOrDefaultRegion(organization: Organization, regionIdOrName?: string): Promise<Region> {
    if (!organization.defaultRegionId) {
      throw new DefaultRegionRequiredException()
    }

    regionIdOrName = regionIdOrName?.trim()

    if (!regionIdOrName) {
      const region = await this.regionService.findOne(organization.defaultRegionId)
      if (!region) {
        throw new NotFoundException('Default region not found')
      }
      return region
    }

    const region =
      (await this.regionService.findOneByName(regionIdOrName, organization.id)) ??
      (await this.regionService.findOneByName(regionIdOrName, null)) ??
      (await this.regionService.findOne(regionIdOrName))

    if (!region) {
      throw new NotFoundException('Region not found')
    }

    return region
  }

  async replaceLabels(
    sandboxIdOrName: string,
    labels: { [key: string]: string },
    organizationId?: string,
  ): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    // Replace all labels
    const updateData: Partial<Sandbox> = {
      labels,
    }

    return await this.sandboxRepository.update(sandbox.id, { updateData, entity: sandbox })
  }

  @Cron(CronExpression.EVERY_SECOND, { name: 'cleanup-destroyed-sandboxes' })
  @LogExecution('cleanup-destroyed-sandboxes')
  @WithInstrumentation()
  async cleanupDestroyedSandboxes() {
    const lockKey = 'sandbox:cleanup-destroyed-sandboxes'
    const acquired = await this.redisLockProvider.lock(lockKey, 300)
    if (!acquired) {
      return
    }

    try {
      const twentyFourHoursAgo = new Date()
      twentyFourHoursAgo.setHours(twentyFourHoursAgo.getHours() - 24)

      const destroyedSandboxs = await this.sandboxRepository.delete({
        state: SandboxState.DESTROYED,
        updatedAt: LessThan(twentyFourHoursAgo),
      })

      if (destroyedSandboxs.affected > 0) {
        this.logger.debug(`Cleaned up ${destroyedSandboxs.affected} destroyed sandboxes`)
      }
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @Cron(CronExpression.EVERY_10_MINUTES, { name: 'cleanup-build-failed-sandboxes' })
  @LogExecution('cleanup-build-failed-sandboxes')
  @WithInstrumentation()
  async cleanupBuildFailedSandboxes() {
    const lockKey = 'sandbox:cleanup-build-failed-sandboxes'
    const acquired = await this.redisLockProvider.lock(lockKey, 300)
    if (!acquired) {
      return
    }

    try {
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
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @Cron(CronExpression.EVERY_SECOND, { name: 'cleanup-stale-build-failed-sandboxes' })
  @LogExecution('cleanup-stale-build-failed-sandboxes')
  @WithInstrumentation()
  async cleanupStaleBuildFailedSandboxes() {
    const lockKey = 'sandbox:cleanup-stale-build-failed-sandboxes'
    const acquired = await this.redisLockProvider.lock(lockKey, 300)
    if (!acquired) {
      return
    }

    try {
      const sevenDaysAgo = new Date()
      sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7)

      const result = await this.sandboxRepository.delete({
        state: SandboxState.BUILD_FAILED,
        desiredState: SandboxDesiredState.STARTED,
        updatedAt: LessThan(sevenDaysAgo),
      })

      if (result.affected > 0) {
        this.logger.debug(`Cleaned up ${result.affected} stale build failed sandboxes`)
      }
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @Cron(CronExpression.EVERY_SECOND, { name: 'cleanup-stale-error-sandboxes' })
  @LogExecution('cleanup-stale-error-sandboxes')
  @WithInstrumentation()
  async cleanupStaleErrorSandboxes() {
    const lockKey = 'sandbox:cleanup-stale-error-sandboxes'
    const acquired = await this.redisLockProvider.lock(lockKey, 300)
    if (!acquired) {
      return
    }

    try {
      const sevenDaysAgo = new Date()
      sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7)

      const result = await this.sandboxRepository.delete({
        state: SandboxState.ERROR,
        desiredState: SandboxDesiredState.DESTROYED,
        updatedAt: LessThan(sevenDaysAgo),
      })

      if (result.affected > 0) {
        this.logger.debug(`Cleaned up ${result.affected} stale error sandboxes`)
      }
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @Cron(CronExpression.EVERY_MINUTE, { name: 'recover-stale-snapshotting-sandboxes' })
  @LogExecution('recover-stale-snapshotting-sandboxes')
  @WithInstrumentation()
  async recoverStaleSnapshottingSandboxes() {
    const lockKey = 'sandbox:recover-stale-snapshotting-sandboxes'
    const acquired = await this.redisLockProvider.lock(lockKey, 300)
    if (!acquired) {
      return
    }

    try {
      const timeoutMinutes = this.configService.getOrThrow('sandboxSnapshottingTimeoutMin')
      const cutoff = new Date()
      cutoff.setMinutes(cutoff.getMinutes() - timeoutMinutes)

      const staleSandboxes = await this.sandboxRepository.find({
        where: {
          state: SandboxState.SNAPSHOTTING,
          pending: true,
          updatedAt: LessThan(cutoff),
        },
        take: 100,
      })

      if (staleSandboxes.length === 0) {
        return
      }

      for (const sandbox of staleSandboxes) {
        if (!sandbox.runnerId) {
          continue
        }

        try {
          const runner = await this.runnerService.findOneOrFail(sandbox.runnerId)

          // v2 runners are recovered by the job system — skip them
          if (runner.apiVersion !== '0') {
            continue
          }

          const restoredState =
            sandbox.desiredState === SandboxDesiredState.STARTED ? SandboxState.STARTED : SandboxState.STOPPED

          await this.sandboxRepository.updateWhere(sandbox.id, {
            updateData: { state: restoredState, pending: false },
            whereCondition: { state: SandboxState.SNAPSHOTTING, desiredState: sandbox.desiredState },
          })

          await this.snapshotService
            .rollbackPendingUsage(sandbox.organizationId, 1)
            .catch((err) =>
              this.logger.error(
                `Failed to roll back pending snapshot quota for org ${sandbox.organizationId} during stale recovery:`,
                err,
              ),
            )

          this.logger.warn(
            `Recovered stale SNAPSHOTTING sandbox ${sandbox.id} (v0 runner ${sandbox.runnerId}), restored to ${restoredState}`,
          )
        } catch (error) {
          this.logger.error(`Error recovering stale SNAPSHOTTING sandbox ${sandbox.id}:`, error)
        }
      }
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  async setAutostopInterval(sandboxIdOrName: string, interval: number, organizationId?: string): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    const updateData: Partial<Sandbox> = {
      autoStopInterval: this.resolveAutoStopInterval(interval),
    }

    return await this.sandboxRepository.update(sandbox.id, { updateData, entity: sandbox })
  }

  async setAutoArchiveInterval(sandboxIdOrName: string, interval: number, organizationId?: string): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    const updateData: Partial<Sandbox> = {
      autoArchiveInterval: this.resolveAutoArchiveInterval(interval),
    }

    return await this.sandboxRepository.update(sandbox.id, { updateData, entity: sandbox })
  }

  async setAutoDeleteInterval(sandboxIdOrName: string, interval: number, organizationId?: string): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    if (sandbox.gpu > 0) {
      throw new BadRequestError('GPU sandboxes must remain ephemeral')
    }

    const updateData: Partial<Sandbox> = {
      autoDeleteInterval: interval,
    }

    return await this.sandboxRepository.update(sandbox.id, { updateData, entity: sandbox })
  }

  async updateNetworkSettings(
    sandboxIdOrName: string,
    networkBlockAll?: boolean,
    networkAllowList?: string,
    domainAllowList?: string,
    organizationId?: string,
  ): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    const updateData: Partial<Sandbox> = {}
    let effectiveNetworkBlockAll = sandbox.networkBlockAll
    let effectiveNetworkAllowList = sandbox.networkAllowList
    let effectiveDomainAllowList = sandbox.domainAllowList

    if (domainAllowList !== undefined) {
      if (domainAllowList.trim() === '') {
        updateData.domainAllowList = null
        effectiveDomainAllowList = null
      } else {
        updateData.domainAllowList = this.resolveDomainAllowList(domainAllowList)
        effectiveDomainAllowList = updateData.domainAllowList
      }
    }

    if (networkBlockAll !== undefined) {
      updateData.networkBlockAll = networkBlockAll
      effectiveNetworkBlockAll = networkBlockAll
      if (networkBlockAll === true) {
        updateData.networkAllowList = null
        effectiveNetworkAllowList = null
      }
    }

    if (networkAllowList !== undefined) {
      if (networkAllowList.trim() === '') {
        updateData.networkAllowList = null
        effectiveNetworkAllowList = null
      } else {
        const resolvedNetworkAllowList = this.resolveNetworkAllowList(networkAllowList)
        updateData.networkAllowList = resolvedNetworkAllowList
        updateData.networkBlockAll = false
        effectiveNetworkAllowList = resolvedNetworkAllowList
        effectiveNetworkBlockAll = false
      }
    } else if (networkBlockAll === false) {
      updateData.networkAllowList = null
      effectiveNetworkAllowList = null
    }

    // Update network settings on the runner
    if (sandbox.runnerId) {
      const runner = await this.runnerService.findOne(sandbox.runnerId)
      if (runner) {
        const runnerAdapter = await this.runnerAdapterFactory.create(runner)
        await runnerAdapter.updateNetworkSettings(
          sandbox.id,
          effectiveNetworkBlockAll,
          effectiveNetworkAllowList ?? undefined,
          undefined,
          effectiveDomainAllowList ?? undefined,
        )
      }
    }

    const updatedSandbox = await this.sandboxRepository.update(sandbox.id, { updateData, entity: sandbox })

    return updatedSandbox
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

    const updateData: Partial<Sandbox> = {
      state: newState,
      recoverable: false,
    }

    if (errorReason !== undefined) {
      updateData.errorReason = errorReason
      if (newState === SandboxState.ERROR) {
        updateData.recoverable = recoverable
      }
    }

    //  we need to update the desired state to match the new state
    const desiredState = this.getExpectedDesiredStateForState(newState)
    if (desiredState) {
      updateData.desiredState = desiredState
    }

    if (newState === SandboxState.DESTROYED) {
      updateData.name = Sandbox.getSoftDeleteName(sandbox.name)
    }

    await this.sandboxRepository.updateWhere(sandbox.id, {
      updateData,
      whereCondition: { pending: false, state: oldState, desiredState: oldDesiredState },
    })
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

  /**
   * Cascade-destroys any sandboxes that are linked to the just-destroyed sandbox.
   * Linked sandboxes are co-located on the same runner as their owner and share a
   * runner-local network with it; once the owner is gone the followers lose the
   * network and have no reason to exist (they are always ephemeral by design).
   */
  @OnEvent(SandboxEvents.DESTROYED)
  async handleSandboxDestroyedCascadeLinked(event: SandboxDestroyedEvent) {
    if (!event.sandbox?.id) {
      return
    }

    const followers = await this.sandboxRepository.find({
      where: {
        linkedSandboxId: event.sandbox.id,
        desiredState: Not(SandboxDesiredState.DESTROYED),
      },
    })

    for (const follower of followers) {
      try {
        await this.destroy(follower.id, follower.organizationId)
      } catch (err) {
        this.logger.warn(
          `Failed to cascade-destroy linked follower ${follower.id} after owner ${event.sandbox.id} was destroyed: ${
            err instanceof Error ? err.message : String(err)
          }`,
        )
      }
    }
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

  private resolveDomainAllowList(domainAllowList: string): string {
    try {
      validateDomainAllowList(domainAllowList)
    } catch (error) {
      throw new BadRequestError(error instanceof Error ? error.message : 'Invalid domain allow list')
    }

    return domainAllowList
  }

  // Resolves each volumeId (which may be a volume name) to the volume's UUID — the
  // runner builds a host mount path from this value, so only UUIDs may be stored.
  private async resolveVolumes(
    organizationId: string,
    volumes?: SandboxVolume[],
  ): Promise<SandboxVolume[] | undefined> {
    if (volumes === undefined || volumes.length === 0) {
      return volumes
    }

    const volumeIdOrNames = volumes.map((volume) => volume.volumeId)
    const foundVolumes = await this.volumeService.getVolumesByIdOrName(organizationId, volumeIdOrNames)

    const resolved = volumes.map((volume) => {
      const matchedVolume = foundVolumes.get(volume.volumeId)
      if (matchedVolume === undefined || !isValidUuid(matchedVolume.id)) {
        throw new BadRequestError(`Volume '${volume.volumeId}' could not be resolved to a valid volume ID`)
      }
      return { ...volume, volumeId: matchedVolume.id }
    })

    try {
      validateMountPaths(resolved)
    } catch (error) {
      throw new BadRequestError(error instanceof Error ? error.message : 'Invalid volume mount configuration')
    }

    try {
      validateSubpaths(resolved)
    } catch (error) {
      throw new BadRequestError(error instanceof Error ? error.message : 'Invalid volume subpath configuration')
    }

    return resolved
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
      const runner = await this.runnerService.findOne(sshAccess.sandbox.runnerId)

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
    recoverable?: boolean,
  ): Promise<void> {
    const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
      id: sandboxId,
    })

    const updateData = Sandbox.getBackupStateUpdate(
      sandboxToUpdate,
      backupState,
      backupSnapshot,
      backupRegistryId,
      backupErrorReason,
      recoverable,
    )

    await this.sandboxRepository.update(sandboxId, { updateData, entity: sandboxToUpdate })
  }
}
