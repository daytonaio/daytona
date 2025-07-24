/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ForbiddenException, Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Not, Repository, LessThan, In, JsonContains, FindOptionsWhere, Or } from 'typeorm'
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
import { PortPreviewUrlDto } from '../dto/port-preview-url.dto'
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

const DEFAULT_CPU = 1
const DEFAULT_MEMORY = 1
const DEFAULT_DISK = 3
const DEFAULT_GPU = 0

@Injectable()
export class SandboxService {
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
    private readonly runnerService: RunnerService,
    private readonly configService: TypedConfigService,
    private readonly warmPoolService: SandboxWarmPoolService,
    private readonly eventEmitter: EventEmitter2,
    private readonly organizationService: OrganizationService,
  ) {}

  private async validateOrganizationQuotas(
    organization: Organization,
    cpu: number,
    memory: number,
    disk: number,
    excludeSandboxId?: string,
  ): Promise<void> {
    this.organizationService.assertOrganizationIsNotSuspended(organization)

    // Check per-sandbox resource limits
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

    const ignoredStates = [SandboxState.DESTROYED, SandboxState.ARCHIVED, SandboxState.ERROR, SandboxState.BUILD_FAILED]

    const inactiveStates = [...ignoredStates, SandboxState.STOPPED, SandboxState.ARCHIVING]

    const resourceMetrics: {
      used_disk: number
      used_cpu: number
      used_mem: number
    } = await this.sandboxRepository
      .createQueryBuilder('sandbox')
      .select([
        'SUM(CASE WHEN sandbox.state NOT IN (:...ignoredStates) THEN sandbox.disk ELSE 0 END) as used_disk',
        'SUM(CASE WHEN sandbox.state NOT IN (:...inactiveStates) THEN sandbox.cpu ELSE 0 END) as used_cpu',
        'SUM(CASE WHEN sandbox.state NOT IN (:...inactiveStates) THEN sandbox.mem ELSE 0 END) as used_mem',
      ])
      .where('sandbox.organizationId = :organizationId', { organizationId: organization.id })
      .andWhere(
        excludeSandboxId ? 'sandbox.id != :excludeSandboxId' : '1=1',
        excludeSandboxId ? { excludeSandboxId } : {},
      )
      .setParameter('ignoredStates', ignoredStates)
      .setParameter('inactiveStates', inactiveStates)
      .getRawOne()

    const usedDisk = Number(resourceMetrics.used_disk) || 0
    const usedCpu = Number(resourceMetrics.used_cpu) || 0
    const usedMem = Number(resourceMetrics.used_mem) || 0

    if (usedDisk + disk > organization.totalDiskQuota) {
      throw new ForbiddenException(
        `Total disk quota exceeded (${usedDisk + disk}GB > ${organization.totalDiskQuota}GB)`,
      )
    }

    // Check total resource quotas
    if (usedCpu + cpu > organization.totalCpuQuota) {
      throw new ForbiddenException(`Total CPU quota exceeded (${usedCpu + cpu} > ${organization.totalCpuQuota})`)
    }

    if (usedMem + memory > organization.totalMemoryQuota) {
      throw new ForbiddenException(
        `Total memory quota exceeded (${usedMem + memory}GB > ${organization.totalMemoryQuota}GB)`,
      )
    }
  }

  async archive(sandboxId: string): Promise<void> {
    const sandbox = await this.sandboxRepository.findOne({
      where: {
        id: sandboxId,
      },
    })

    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    if (String(sandbox.state) !== String(sandbox.desiredState)) {
      throw new SandboxError('State change in progress')
    }

    if (sandbox.state !== SandboxState.STOPPED) {
      throw new SandboxError('Sandbox is not stopped')
    }

    if (sandbox.pending) {
      throw new SandboxError('Sandbox state change in progress')
    }
    sandbox.pending = true
    sandbox.desiredState = SandboxDesiredState.ARCHIVED
    await this.sandboxRepository.save(sandbox)

    this.eventEmitter.emit(SandboxEvents.ARCHIVED, new SandboxArchivedEvent(sandbox))
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
      throw new BadRequestError(`Snapshot ${snapshotIdOrName} not found. Did you add it through the Daytona Dashboard?`)
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

    await this.validateOrganizationQuotas(organization, cpu, mem, disk)

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
        return await this.assignWarmPoolSandbox(warmPoolSandbox, createSandboxDto, organization.id)
      }
    }

    const runner = await this.runnerService.getRandomAvailableRunner({
      region,
      sandboxClass,
      snapshotRef: snapshot.internalName,
    })

    const sandbox = new Sandbox()

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
  }

  private async assignWarmPoolSandbox(
    warmPoolSandbox: Sandbox,
    createSandboxDto: CreateSandboxDto,
    organizationId: string,
  ): Promise<SandboxDto> {
    warmPoolSandbox.public = createSandboxDto.public || false
    warmPoolSandbox.labels = createSandboxDto.labels || {}
    warmPoolSandbox.organizationId = organizationId
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

    const runner = await this.runnerService.findOne(warmPoolSandbox.runnerId)

    const result = await this.sandboxRepository.save(warmPoolSandbox)

    // Treat this as a newly started sandbox
    this.eventEmitter.emit(
      SandboxEvents.STATE_UPDATED,
      new SandboxStateUpdatedEvent(warmPoolSandbox, SandboxState.STARTED, SandboxState.STARTED),
    )
    return SandboxDto.fromSandbox(result, runner.domain)
  }

  async createFromBuildInfo(createSandboxDto: CreateSandboxDto, organization: Organization): Promise<SandboxDto> {
    const region = this.getValidatedOrDefaultRegion(createSandboxDto.target)
    const sandboxClass = this.getValidatedOrDefaultClass(createSandboxDto.class)

    const cpu = createSandboxDto.cpu || DEFAULT_CPU
    const mem = createSandboxDto.memory || DEFAULT_MEMORY
    const disk = createSandboxDto.disk || DEFAULT_DISK
    const gpu = createSandboxDto.gpu || DEFAULT_GPU

    await this.validateOrganizationQuotas(organization, cpu, mem, disk)

    const sandbox = new Sandbox()

    // sandbox = from

    sandbox.organizationId = organization.id

    //  TODO: make configurable
    sandbox.region = region
    sandbox.class = sandboxClass
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
      if (error instanceof BadRequestError == false || error.message !== 'No available runners' || !sandbox.buildInfo) {
        throw error
      }
      sandbox.state = SandboxState.PENDING_BUILD
    }

    await this.sandboxRepository.insert(sandbox)
    return SandboxDto.fromSandbox(sandbox, runner?.domain)
  }

  async createBackup(sandboxId: string): Promise<void> {
    const sandbox = await this.sandboxRepository.findOne({
      where: {
        id: sandboxId,
      },
    })

    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    if (![BackupState.COMPLETED, BackupState.NONE].includes(sandbox.backupState)) {
      throw new SandboxError('Sandbox backup is already in progress')
    }

    await this.sandboxRepository.update(sandboxId, {
      backupState: BackupState.PENDING,
    })

    this.eventEmitter.emit(SandboxEvents.BACKUP_CREATED, new SandboxBackupCreatedEvent(sandbox))
  }

  async findAll(
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
        sandbox.desiredState !== SandboxDesiredState.DESTROYED)
    ) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    return sandbox
  }

  async getPortPreviewUrl(sandboxId: string, port: number): Promise<PortPreviewUrlDto> {
    if (port < 1 || port > 65535) {
      throw new BadRequestError('Invalid port')
    }

    const sandbox = await this.sandboxRepository.findOne({
      where: { id: sandboxId },
    })

    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    // Validate sandbox is in valid state
    if (sandbox.state !== SandboxState.STARTED) {
      throw new SandboxError('Sandbox must be started to get port preview URL')
    }

    // Get runner info
    const runner = await this.runnerService.findOne(sandbox.runnerId)
    if (!runner) {
      throw new NotFoundException(`Runner not found for sandbox ${sandboxId}`)
    }

    return {
      url: `https://${port}-${sandbox.id}.${runner.domain}`,
      token: sandbox.authToken,
    }
  }

  async destroy(sandboxId: string): Promise<void> {
    const sandbox = await this.sandboxRepository.findOne({
      where: {
        id: sandboxId,
      },
    })

    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    if (sandbox.pending) {
      throw new SandboxError('Sandbox state change in progress')
    }
    sandbox.pending = true
    sandbox.desiredState = SandboxDesiredState.DESTROYED
    await this.sandboxRepository.save(sandbox)

    this.eventEmitter.emit(SandboxEvents.DESTROYED, new SandboxDestroyedEvent(sandbox))
  }

  async start(sandboxId: string, organization: Organization): Promise<void> {
    const sandbox = await this.sandboxRepository.findOne({
      where: {
        id: sandboxId,
      },
    })

    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    if (String(sandbox.state) !== String(sandbox.desiredState)) {
      throw new SandboxError('State change in progress')
    }

    if (![SandboxState.STOPPED, SandboxState.ARCHIVED].includes(sandbox.state)) {
      throw new SandboxError('Sandbox is not in valid state')
    }

    this.organizationService.assertOrganizationIsNotSuspended(organization)

    if (sandbox.runnerId) {
      // Add runner readiness check
      const runner = await this.runnerService.findOne(sandbox.runnerId)
      if (runner.state !== RunnerState.READY) {
        throw new SandboxError('Runner is not ready')
      }
    } else {
      //  restore operation
      //  like a new sandbox creation, we need to validate quotas
      await this.validateOrganizationQuotas(organization, sandbox.cpu, sandbox.mem, sandbox.disk, sandbox.id)
    }

    if (sandbox.pending) {
      throw new SandboxError('Sandbox state change in progress')
    }

    sandbox.pending = true
    sandbox.desiredState = SandboxDesiredState.STARTED
    await this.sandboxRepository.save(sandbox)

    this.eventEmitter.emit(SandboxEvents.STARTED, new SandboxStartedEvent(sandbox))
  }

  async stop(sandboxId: string): Promise<void> {
    const sandbox = await this.sandboxRepository.findOne({
      where: {
        id: sandboxId,
      },
    })

    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

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
  }

  async updatePublicStatus(sandboxId: string, isPublic: boolean): Promise<void> {
    const sandbox = await this.sandboxRepository.findOne({
      where: { id: sandboxId },
    })

    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    sandbox.public = isPublic
    await this.sandboxRepository.save(sandbox)
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

  async replaceLabels(sandboxId: string, labels: { [key: string]: string }): Promise<{ [key: string]: string }> {
    const sandbox = await this.sandboxRepository.findOne({
      where: { id: sandboxId },
    })

    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    // Replace all labels
    sandbox.labels = labels
    await this.sandboxRepository.save(sandbox)

    return sandbox.labels
  }

  @Cron(CronExpression.EVERY_10_MINUTES)
  async cleanupDestroyedSandboxs() {
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

  async setAutostopInterval(sandboxId: string, interval: number): Promise<void> {
    const sandbox = await this.sandboxRepository.findOne({
      where: { id: sandboxId },
    })

    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    sandbox.autoStopInterval = this.resolveAutoStopInterval(interval)
    await this.sandboxRepository.save(sandbox)
  }

  async setAutoArchiveInterval(sandboxId: string, interval: number): Promise<void> {
    const sandbox = await this.sandboxRepository.findOne({
      where: { id: sandboxId },
    })

    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    sandbox.autoArchiveInterval = this.resolveAutoArchiveInterval(interval)
    await this.sandboxRepository.save(sandbox)
  }

  async setAutoDeleteInterval(sandboxId: string, interval: number): Promise<void> {
    const sandbox = await this.sandboxRepository.findOne({
      where: { id: sandboxId },
    })

    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }

    sandbox.autoDeleteInterval = interval
    await this.sandboxRepository.save(sandbox)
  }

  @OnEvent(WarmPoolEvents.TOPUP_REQUESTED)
  private async createWarmPoolSandbox(event: WarmPoolTopUpRequested) {
    await this.createForWarmPool(event.warmPool)
  }

  @Cron(CronExpression.EVERY_MINUTE)
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
}
