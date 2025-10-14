/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ForbiddenException, Injectable, Logger, NotFoundException, ConflictException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Not, Repository, LessThan, In, JsonContains, FindOptionsWhere, ILike } from 'typeorm'
import { Sandbox } from '../entities/sandbox.entity'
import { CreateSandboxDto } from '../dto/create-sandbox.dto'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { RunnerService } from './runner.service'
import { SandboxError } from '../../exceptions/sandbox-error.exception'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { Cron, CronExpression } from '@nestjs/schedule'
import { RunnerState } from '../enums/runner-state.enum'
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
import { SandboxDto } from '../dto/sandbox.dto'
import { isValidUuid } from '../../common/utils/uuid'
import { RunnerAdapterFactory } from '../runner-adapter/runnerAdapter'
import { validateNetworkAllowList } from '../utils/network-validation.util'
import { OrganizationUsageService } from '../../organization/services/organization-usage.service'
import { SshAccess } from '../entities/ssh-access.entity'
import { nanoid } from 'nanoid'
import { SshAccessValidationDto } from '../dto/ssh-access.dto'
import { VolumeService } from './volume.service'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { PaginatedList } from '../../common/interfaces/paginated-list.interface'
import {
  SandboxSortField,
  SandboxSortDirection,
  DEFAULT_SANDBOX_SORT_FIELD,
  DEFAULT_SANDBOX_SORT_DIRECTION,
} from '../dto/list-sandboxes-query.dto'
import { createRangeFilter } from '../../common/utils/range-filter'
import { LogExecution } from '../../common/decorators/log-execution.decorator'
import { LockableEntity } from '../../common/services/lockable-entity.service'

const DEFAULT_CPU = 1
const DEFAULT_MEMORY = 1
const DEFAULT_DISK = 3
const DEFAULT_GPU = 0

@Injectable()
export class SandboxService extends LockableEntity {
  private readonly logger = new Logger(SandboxService.name)

  constructor(
    @InjectRepository(Sandbox)
    private readonly sandboxRepository: Repository<Sandbox>,
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
    redisLockProvider: RedisLockProvider,
  ) {
    super(redisLockProvider)
  }

  protected getLockKey(id: string): string {
    return `sandbox:${id}:state-change`
  }

  private async validateOrganizationQuotas(
    organization: Organization,
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
        `CPU request ${cpu} exceeds maximum allowed per sandbox (${organization.maxCpuPerSandbox})`,
      )
    }
    if (memory > organization.maxMemoryPerSandbox) {
      throw new ForbiddenException(
        `Memory request ${memory}GB exceeds maximum allowed per sandbox (${organization.maxMemoryPerSandbox}GB)`,
      )
    }
    if (disk > organization.maxDiskPerSandbox) {
      throw new ForbiddenException(
        `Disk request ${disk}GB exceeds maximum allowed per sandbox (${organization.maxDiskPerSandbox}GB)`,
      )
    }

    // validate usage quotas
    const {
      cpuIncremented: pendingCpuIncremented,
      memoryIncremented: pendingMemoryIncremented,
      diskIncremented: pendingDiskIncremented,
    } = await this.organizationUsageService.incrementPendingSandboxUsage(
      organization.id,
      cpu,
      memory,
      disk,
      excludeSandboxId,
    )

    const usageOverview = await this.organizationUsageService.getSandboxUsageOverview(organization.id, excludeSandboxId)

    try {
      if (usageOverview.currentCpuUsage + usageOverview.pendingCpuUsage > organization.totalCpuQuota) {
        throw new ForbiddenException(`Total CPU quota exceeded. Maximum allowed: ${organization.totalCpuQuota}`)
      }

      if (usageOverview.currentMemoryUsage + usageOverview.pendingMemoryUsage > organization.totalMemoryQuota) {
        throw new ForbiddenException(
          `Total memory quota exceeded. Maximum allowed: ${organization.totalMemoryQuota}GiB`,
        )
      }

      if (usageOverview.currentDiskUsage + usageOverview.pendingDiskUsage > organization.totalDiskQuota) {
        throw new ForbiddenException(`Total disk quota exceeded. Maximum allowed: ${organization.totalDiskQuota}GiB`)
      }
    } catch (error) {
      await this.rollbackPendingUsage(
        organization.id,
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
        pendingCpuIncrement,
        pendingMemoryIncrement,
        pendingDiskIncrement,
      )
    } catch (error) {
      this.logger.error(`Error rolling back pending sandbox usage: ${error}`)
    }
  }

  async archive(sandboxIdOrName: string, organizationId?: string): Promise<Sandbox> {
    const { id: sandboxId } = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    return this.withLock(sandboxId, 60, async () => {
      // Re-fetch with fresh state after acquiring lock
      const sandbox = await this.findOneByIdOrName(sandboxId, organizationId)

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
      await this.sandboxRepository.save(sandbox)

      this.eventEmitter.emit(SandboxEvents.ARCHIVED, new SandboxArchivedEvent(sandbox))
      return sandbox
    })
  }

  async createForWarmPool(warmPoolItem: WarmPool): Promise<Sandbox> {
    const sandbox = new Sandbox()

    sandbox.organizationId = SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION

    sandbox.region = warmPoolItem.target
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
      region: sandbox.region,
      sandboxClass: sandbox.class,
      snapshotRef: snapshot.internalName,
    })

    sandbox.runnerId = runner.id

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

    try {
      const region = this.getValidatedOrDefaultRegion(createSandboxDto.target)
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

      if (snapshot.state !== SnapshotState.ACTIVE) {
        throw new BadRequestError(`Snapshot ${snapshotIdOrName} is ${snapshot.state}`)
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
        await this.validateOrganizationQuotas(organization, cpu, mem, disk)

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
          target: createSandboxDto.target,
          class: createSandboxDto.class,
          cpu: cpu,
          mem: mem,
          disk: disk,
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
        region,
        sandboxClass,
        snapshotRef: snapshot.internalName,
      })

      const sandbox = new Sandbox(createSandboxDto.name)

      sandbox.organizationId = organization.id

      //  TODO: make configurable
      sandbox.region = region
      sandbox.class = sandboxClass
      sandbox.snapshot = snapshot.name
      //  TODO: default user should be configurable
      sandbox.osUser = createSandboxDto.user || 'daytona'
      sandbox.env = createSandboxDto.env || {}
      sandbox.labels = createSandboxDto.labels || {}
      sandbox.volumes = createSandboxDto.volumes || []

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

      sandbox.runnerId = runner.id

      await this.sandboxRepository.insert(sandbox)
      return SandboxDto.fromSandbox(sandbox, runner.domain)
    } catch (error) {
      if (error.code === '23505') {
        throw new ConflictException(`Sandbox with name ${createSandboxDto.name} already exists`)
      }

      await this.rollbackPendingUsage(
        organization.id,
        pendingCpuIncrement,
        pendingMemoryIncrement,
        pendingDiskIncrement,
      )

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

    const runner = await this.runnerService.findOne(warmPoolSandbox.runnerId)
    if (!runner) {
      throw new NotFoundException(`Runner with ID ${warmPoolSandbox.runnerId} not found`)
    }

    if (
      createSandboxDto.networkBlockAll !== undefined ||
      createSandboxDto.networkAllowList !== undefined ||
      organization.sandboxLimitedNetworkEgress
    ) {
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
    return SandboxDto.fromSandbox(result, runner.domain)
  }

  async createFromBuildInfo(createSandboxDto: CreateSandboxDto, organization: Organization): Promise<SandboxDto> {
    let pendingCpuIncrement: number | undefined
    let pendingMemoryIncrement: number | undefined
    let pendingDiskIncrement: number | undefined

    try {
      const region = this.getValidatedOrDefaultRegion(createSandboxDto.target)
      const sandboxClass = this.getValidatedOrDefaultClass(createSandboxDto.class)

      const cpu = createSandboxDto.cpu || DEFAULT_CPU
      const mem = createSandboxDto.memory || DEFAULT_MEMORY
      const disk = createSandboxDto.disk || DEFAULT_DISK
      const gpu = createSandboxDto.gpu || DEFAULT_GPU

      this.organizationService.assertOrganizationIsNotSuspended(organization)

      const { pendingCpuIncremented, pendingMemoryIncremented, pendingDiskIncremented } =
        await this.validateOrganizationQuotas(organization, cpu, mem, disk)

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

      const sandbox = new Sandbox(createSandboxDto.name)

      sandbox.organizationId = organization.id

      sandbox.region = region
      sandbox.class = sandboxClass
      sandbox.osUser = createSandboxDto.user || 'daytona'
      sandbox.env = createSandboxDto.env || {}
      sandbox.labels = createSandboxDto.labels || {}
      sandbox.volumes = createSandboxDto.volumes || []

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
        await this.buildInfoRepository.update(sandbox.buildInfo.snapshotRef, { lastUsedAt: new Date() })
      } else {
        const buildInfoEntity = this.buildInfoRepository.create({
          ...createSandboxDto.buildInfo,
        })
        await this.buildInfoRepository.save(buildInfoEntity)
        sandbox.buildInfo = buildInfoEntity
      }

      let runner: Runner

      try {
        runner = await this.runnerService.getRandomAvailableRunner({
          region: sandbox.region,
          sandboxClass: sandbox.class,
          snapshotRef: sandbox.buildInfo.snapshotRef,
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

      await this.sandboxRepository.insert(sandbox)
      return SandboxDto.fromSandbox(sandbox, runner?.domain)
    } catch (error) {
      if (error.code === '23505') {
        throw new ConflictException(`Sandbox with name ${createSandboxDto.name} already exists`)
      }

      await this.rollbackPendingUsage(
        organization.id,
        pendingCpuIncrement,
        pendingMemoryIncrement,
        pendingDiskIncrement,
      )

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

    await this.sandboxRepository.update(sandbox.id, {
      backupState: BackupState.PENDING,
    })

    this.eventEmitter.emit(SandboxEvents.BACKUP_CREATED, new SandboxBackupCreatedEvent(sandbox))

    return sandbox
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
      regions?: string[]
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
      regions,
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
      ...(regions ? { region: In(regions) } : {}),
    }

    baseFindOptions.cpu = createRangeFilter(minCpu, maxCpu)
    baseFindOptions.mem = createRangeFilter(minMemoryGiB, maxMemoryGiB)
    baseFindOptions.disk = createRangeFilter(minDiskGiB, maxDiskGiB)
    baseFindOptions.lastActivityAt = createRangeFilter(lastEventAfter, lastEventBefore)

    const filteredStates = (states || Object.values(SandboxState)).filter((state) => state !== SandboxState.DESTROYED)
    const errorStates = [SandboxState.ERROR, SandboxState.BUILD_FAILED]
    const filteredWithoutErrorStates = filteredStates.filter((state) => !errorStates.includes(state))

    const where: FindOptionsWhere<Sandbox>[] = [
      {
        ...baseFindOptions,
        state: In(filteredWithoutErrorStates),
      },
      {
        ...baseFindOptions,
        state: In(errorStates),
        ...(includeErroredDestroyed ? {} : { desiredState: Not(SandboxDesiredState.DESTROYED) }),
      },
    ]

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

  private getDesiredStateForState(state: SandboxState): SandboxDesiredState | undefined {
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
    return this.getDesiredStateForState(state) !== undefined
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
        const expectedDesiredState = this.getDesiredStateForState(sandbox.state)
        return expectedDesiredState !== undefined && expectedDesiredState === sandbox.desiredState
      })
    }

    return sandboxes
  }

  async findOneByIdOrName(
    sandboxIdOrName: string,
    organizationId?: string,
    returnDestroyed?: boolean,
  ): Promise<Sandbox> {
    let sandbox = await this.sandboxRepository.findOne({
      where: {
        id: sandboxIdOrName,
        ...(organizationId ? { organizationId: organizationId } : {}),
        ...(returnDestroyed ? {} : { state: Not(SandboxState.DESTROYED) }),
      },
    })

    if (!sandbox && organizationId) {
      sandbox = await this.sandboxRepository.findOne({
        where: {
          name: sandboxIdOrName,
          organizationId: organizationId,
          ...(returnDestroyed ? {} : { state: Not(SandboxState.DESTROYED) }),
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

  async destroy(sandboxIdOrName: string, organizationId?: string): Promise<Sandbox> {
    // Fetch to get actual ID for consistent locking
    const { id: sandboxId } = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    return this.withLock(sandboxId, 60, async () => {
      // Re-fetch with fresh state after acquiring lock
      const sandbox = await this.findOneByIdOrName(sandboxId, organizationId)

      if (sandbox.pending) {
        throw new SandboxError('Sandbox state change in progress')
      }
      sandbox.pending = true
      sandbox.desiredState = SandboxDesiredState.DESTROYED
      sandbox.backupState = BackupState.NONE
      sandbox.name = 'DESTROYED_' + sandbox.name + '_' + Date.now()
      await this.sandboxRepository.save(sandbox)

      this.eventEmitter.emit(SandboxEvents.DESTROYED, new SandboxDestroyedEvent(sandbox))
      return sandbox
    })
  }

  async start(sandboxIdOrName: string, organization: Organization): Promise<Sandbox> {
    // Fetch to get actual ID for consistent locking
    const { id: sandboxId } = await this.findOneByIdOrName(sandboxIdOrName, organization.id)

    let pendingCpuIncrement: number | undefined
    let pendingMemoryIncrement: number | undefined
    let pendingDiskIncrement: number | undefined

    return this.withLock(sandboxId, 60, async () => {
      try {
        // Re-fetch with fresh state after acquiring lock
        const sandbox = await this.findOneByIdOrName(sandboxId, organization.id)

        if (sandbox.state === SandboxState.STARTED && sandbox.desiredState === SandboxDesiredState.STARTED) {
          return sandbox
        }

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
          await this.validateOrganizationQuotas(organization, sandbox.cpu, sandbox.mem, sandbox.disk, sandbox.id)

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
        await this.sandboxRepository.save(sandbox)

        this.eventEmitter.emit(SandboxEvents.STARTED, new SandboxStartedEvent(sandbox))

        return sandbox
      } catch (error) {
        await this.rollbackPendingUsage(
          organization.id,
          pendingCpuIncrement,
          pendingMemoryIncrement,
          pendingDiskIncrement,
        )
        throw error
      }
    })
  }

  async stop(sandboxIdOrName: string, organizationId?: string): Promise<Sandbox> {
    // Fetch to get actual ID for consistent locking
    const { id: sandboxId } = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    return this.withLock(sandboxId, 60, async () => {
      // Re-fetch with fresh state after acquiring lock
      const sandbox = await this.findOneByIdOrName(sandboxId, organizationId)

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

      await this.sandboxRepository.save(sandbox)

      if (sandbox.autoDeleteInterval === 0) {
        this.eventEmitter.emit(SandboxEvents.DESTROYED, new SandboxDestroyedEvent(sandbox))
      } else {
        this.eventEmitter.emit(SandboxEvents.STOPPED, new SandboxStoppedEvent(sandbox))
      }
      return sandbox
    })
  }

  async updatePublicStatus(sandboxIdOrName: string, isPublic: boolean, organizationId?: string): Promise<Sandbox> {
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    sandbox.public = isPublic
    await this.sandboxRepository.save(sandbox)

    return sandbox
  }

  private getValidatedOrDefaultRegion(region?: string): string {
    if (!region || region.trim().length === 0) {
      return 'us'
    }

    return region.trim()
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

  async updateStateAndDesiredState(sandboxId: string, newState: SandboxState): Promise<void> {
    const lockKey = await this.tryLock(sandboxId, 60)
    if (!lockKey) {
      // Skip update if sandbox is locked (user operation in progress)
      this.logger.debug(`Skipped state update for ${sandboxId} - sandbox is locked`)
      return
    }

    try {
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

      sandbox.state = newState
      const desiredState = this.getDesiredStateForState(newState)
      if (desiredState) {
        sandbox.desiredState = desiredState
      }
      await this.sandboxRepository.save(sandbox)
    } finally {
      await this.unlock(lockKey)
    }
  }

  @OnEvent(WarmPoolEvents.TOPUP_REQUESTED)
  private async createWarmPoolSandbox(event: WarmPoolTopUpRequested) {
    await this.createForWarmPool(event.warmPool)
  }

  @Cron(CronExpression.EVERY_MINUTE, { name: 'handle-unschedulable-runners' })
  @LogExecution('handle-unschedulable-runners')
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
    validateNetworkAllowList(networkAllowList)

    return networkAllowList
  }

  async createSshAccess(sandboxIdOrName: string, expiresInMinutes = 60, organizationId?: string): Promise<SshAccess> {
    //  check if sandbox exists
    const sandbox = await this.findOneByIdOrName(sandboxIdOrName, organizationId)

    // Revoke any existing SSH access for this sandbox
    await this.revokeSshAccess(sandbox.id)

    const sshAccess = new SshAccess()
    sshAccess.sandboxId = sandbox.id
    sshAccess.token = nanoid(32)
    sshAccess.expiresAt = new Date(Date.now() + expiresInMinutes * 60 * 1000)

    return await this.sshAccessRepository.save(sshAccess)
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
          runnerId: runner.id,
          runnerDomain: runner.domain,
        }
      }
    }

    return { valid: true, sandboxId: sshAccess.sandbox.id }
  }

  async getDistinctRegions(organizationId: string): Promise<string[]> {
    const result = await this.sandboxRepository
      .createQueryBuilder('sandbox')
      .select('DISTINCT sandbox.region', 'region')
      .where('sandbox.organizationId = :organizationId', { organizationId })
      .orderBy('sandbox.region', 'ASC')
      .getRawMany()

    return result.map((row) => row.region)
  }
}
