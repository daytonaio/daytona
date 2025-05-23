/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ForbiddenException, Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Not, Repository, LessThan, In, JsonContains } from 'typeorm'
import { Workspace } from '../entities/workspace.entity'
import { CreateWorkspaceDto } from '../dto/create-workspace.dto'
import { WorkspaceState } from '../enums/workspace-state.enum'
import { WorkspaceClass } from '../enums/workspace-class.enum'
import { RunnerRegion } from '../enums/runner-region.enum'
import { WorkspaceDesiredState } from '../enums/workspace-desired-state.enum'
import { RunnerService } from './runner.service'
import { WorkspaceError } from '../../exceptions/workspace-error.exception'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { Cron, CronExpression } from '@nestjs/schedule'
import { RunnerState } from '../enums/runner-state.enum'
import { BackupState } from '../enums/backup-state.enum'
import { Image } from '../entities/image.entity'
import { ImageState } from '../enums/image-state.enum'
import { WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../constants/workspace.constants'
import { WorkspaceWarmPoolService } from './workspace-warm-pool.service'
import { EventEmitter2, OnEvent } from '@nestjs/event-emitter'
import { WarmPoolEvents } from '../constants/warmpool-events.constants'
import { WarmPoolTopUpRequested } from '../events/warmpool-topup-requested.event'
import { Runner } from '../entities/runner.entity'
import { PortPreviewUrlDto } from '../dto/port-preview-url.dto'
import { Organization } from '../../organization/entities/organization.entity'
import { WorkspaceEvents } from '../constants/workspace-events.constants'
import { WorkspaceStateUpdatedEvent } from '../events/workspace-state-updated.event'
import { BuildInfo } from '../entities/build-info.entity'
import { generateBuildInfoHash as generateBuildImageRef } from '../entities/build-info.entity'
import { WorkspaceBackupCreatedEvent } from '../events/workspace-backup-created.event'
import { WorkspaceDestroyedEvent } from '../events/workspace-destroyed.event'
import { WorkspaceStartedEvent } from '../events/workspace-started.event'
import { WorkspaceStoppedEvent } from '../events/workspace-stopped.event'
import { WorkspaceArchivedEvent } from '../events/workspace-archived.event'
import { OrganizationService } from '../../organization/services/organization.service'
import { OrganizationEvents } from '../../organization/constants/organization-events.constant'
import { OrganizationSuspendedWorkspaceStoppedEvent } from '../../organization/events/organization-suspended-workspace-stopped.event'
import { TypedConfigService } from '../../config/typed-config.service'

const DEFAULT_CPU = 1
const DEFAULT_MEMORY = 1
const DEFAULT_DISK = 3
const DEFAULT_GPU = 0

@Injectable()
export class WorkspaceService {
  private readonly logger = new Logger(WorkspaceService.name)

  constructor(
    @InjectRepository(Workspace)
    private readonly workspaceRepository: Repository<Workspace>,
    @InjectRepository(Image)
    private readonly imageRepository: Repository<Image>,
    @InjectRepository(Runner)
    private readonly runnerRepository: Repository<Runner>,
    @InjectRepository(BuildInfo)
    private readonly buildInfoRepository: Repository<BuildInfo>,
    private readonly runnerService: RunnerService,
    private readonly configService: TypedConfigService,
    private readonly warmPoolService: WorkspaceWarmPoolService,
    private readonly eventEmitter: EventEmitter2,
    private readonly organizationService: OrganizationService,
  ) { }

  private async validateOrganizationQuotas(
    organization: Organization,
    cpu: number,
    memory: number,
    disk: number,
    excludeWorkspaceId?: string,
  ): Promise<void> {
    await this.assertOrganizationIsNotSuspended(organization)

    // Check per-workspace resource limits
    if (cpu > organization.maxCpuPerWorkspace) {
      throw new ForbiddenException(
        `CPU request ${cpu} exceeds maximum allowed per workspace (${organization.maxCpuPerWorkspace})`,
      )
    }
    if (memory > organization.maxMemoryPerWorkspace) {
      throw new ForbiddenException(
        `Memory request ${memory}GB exceeds maximum allowed per workspace (${organization.maxMemoryPerWorkspace}GB)`,
      )
    }
    if (disk > organization.maxDiskPerWorkspace) {
      throw new ForbiddenException(
        `Disk request ${disk}GB exceeds maximum allowed per workspace (${organization.maxDiskPerWorkspace}GB)`,
      )
    }

    const ignoredStates = [WorkspaceState.DESTROYED, WorkspaceState.ARCHIVED, WorkspaceState.ERROR]

    const inactiveStates = [...ignoredStates, WorkspaceState.STOPPED, WorkspaceState.ARCHIVING]

    const resourceMetrics: {
      used_disk: number
      used_cpu: number
      used_mem: number
    } = await this.workspaceRepository
      .createQueryBuilder('workspace')
      .select([
        'SUM(CASE WHEN workspace.state NOT IN (:...ignoredStates) THEN workspace.disk ELSE 0 END) as used_disk',
        'SUM(CASE WHEN workspace.state NOT IN (:...inactiveStates) THEN workspace.cpu ELSE 0 END) as used_cpu',
        'SUM(CASE WHEN workspace.state NOT IN (:...inactiveStates) THEN workspace.mem ELSE 0 END) as used_mem',
      ])
      .where('workspace.organizationId = :organizationId', { organizationId: organization.id })
      .andWhere(
        excludeWorkspaceId ? 'workspace.id != :excludeWorkspaceId' : '1=1',
        excludeWorkspaceId ? { excludeWorkspaceId } : {},
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

  async archive(workspaceId: string): Promise<void> {
    const workspace = await this.workspaceRepository.findOne({
      where: {
        id: workspaceId,
      },
    })

    if (!workspace) {
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }

    if (String(workspace.state) !== String(workspace.desiredState)) {
      throw new WorkspaceError('State change in progress')
    }

    if (workspace.state !== WorkspaceState.STOPPED) {
      throw new WorkspaceError('Workspace is not stopped')
    }

    if (workspace.pending) {
      throw new WorkspaceError('Workspace state change in progress')
    }
    workspace.pending = true
    workspace.desiredState = WorkspaceDesiredState.ARCHIVED
    await this.workspaceRepository.save(workspace)

    this.eventEmitter.emit(WorkspaceEvents.ARCHIVED, new WorkspaceArchivedEvent(workspace))
  }

  async count(organizationId: string): Promise<number> {
    return this.workspaceRepository.count({
      where: {
        organizationId,
        state: Not(In([WorkspaceState.DESTROYED, WorkspaceState.ARCHIVED, WorkspaceState.ERROR])),
      },
    })
  }

  async create(
    organizationId: string,
    createWorkspaceDto: CreateWorkspaceDto,
    organization?: Organization,
  ): Promise<Workspace> {
    const cpu = createWorkspaceDto.cpu || DEFAULT_CPU
    const mem = createWorkspaceDto.memory || DEFAULT_MEMORY
    const disk = createWorkspaceDto.disk || DEFAULT_DISK
    const gpu = createWorkspaceDto.gpu || DEFAULT_GPU
    // Validate region and class
    const region = createWorkspaceDto.target || RunnerRegion.EU
    if (!this.isValidRegion(region)) {
      throw new BadRequestError('Invalid region')
    }
    const workspaceClass = createWorkspaceDto.class || WorkspaceClass.SMALL
    if (!this.isValidClass(workspaceClass)) {
      throw new BadRequestError('Invalid class')
    }

    // Validate organization quotas before creating workspace
    if (organizationId !== WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION) {
      if (!organization) {
        organization = await this.organizationService.findOne(organizationId)
        if (!organization) {
          throw new NotFoundException(`Organization with ID ${organizationId} not found`)
        }
      }

      await this.validateOrganizationQuotas(organization, cpu, mem, disk)
    }

    //  validate image
    let workspaceImage = createWorkspaceDto.image

    if ((!createWorkspaceDto.image || createWorkspaceDto.image.trim() === '') && !createWorkspaceDto.buildInfo) {
      workspaceImage = this.configService.getOrThrow('defaultImage')
    }

    const image = await this.imageRepository.findOne({
      where: [
        { organizationId, name: workspaceImage, state: ImageState.ACTIVE },
        { general: true, name: workspaceImage, state: ImageState.ACTIVE },
      ],
    })

    if (!createWorkspaceDto.buildInfo && (createWorkspaceDto.volumes || []).length === 0) {
      if (!image) {
        throw new BadRequestError(`Image ${workspaceImage} not found. Did you add it through the Daytona Dashboard?`)
      }

      if (organizationId !== WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION) {
        const warmPoolWorkspace = await this.warmPoolService.fetchWarmPoolWorkspace({
          organizationId: organizationId,
          image: workspaceImage,
          target: createWorkspaceDto.target,
          class: createWorkspaceDto.class,
          cpu,
          mem,
          disk,
          osUser: createWorkspaceDto.user,
          env: createWorkspaceDto.env,
          state: WorkspaceState.STARTED,
        })

        if (warmPoolWorkspace) {
          warmPoolWorkspace.public = createWorkspaceDto.public || false
          warmPoolWorkspace.labels = createWorkspaceDto.labels || {}
          if (createWorkspaceDto.autoStopInterval !== undefined) {
            warmPoolWorkspace.autoStopInterval = createWorkspaceDto.autoStopInterval
          }
          if (createWorkspaceDto.autoArchiveInterval !== undefined) {
            warmPoolWorkspace.autoArchiveInterval = this.resolveAutoArchiveInterval(
              createWorkspaceDto.autoArchiveInterval,
            )
          }
          warmPoolWorkspace.organizationId = organizationId
          warmPoolWorkspace.createdAt = new Date()
          const result = await this.workspaceRepository.save(warmPoolWorkspace)
          // Treat this as a newly started workspace
          this.eventEmitter.emit(
            WorkspaceEvents.STATE_UPDATED,
            new WorkspaceStateUpdatedEvent(warmPoolWorkspace, WorkspaceState.STARTED, WorkspaceState.STARTED),
          )
          return result
        }
      }
      //  [ end of warm pool logic ]
    }

    const workspace = new Workspace()

    workspace.organizationId = organizationId

    //  TODO: make configurable
    workspace.region = region
    workspace.class = workspaceClass
    workspace.image = workspaceImage
    //  TODO: default user should be configurable
    workspace.osUser = createWorkspaceDto.user || 'daytona'
    workspace.env = createWorkspaceDto.env || {}
    workspace.labels = createWorkspaceDto.labels || {}
    workspace.volumes = createWorkspaceDto.volumes || []

    workspace.cpu = cpu
    workspace.gpu = gpu
    workspace.mem = mem
    workspace.disk = disk

    workspace.public = createWorkspaceDto.public || false

    if (createWorkspaceDto.buildInfo) {
      const buildInfoImageRef = generateBuildImageRef(
        createWorkspaceDto.buildInfo.dockerfileContent,
        createWorkspaceDto.buildInfo.contextHashes,
      )

      // Check if buildInfo with the same imageRef already exists
      const existingBuildInfo = await this.buildInfoRepository.findOne({
        where: { imageRef: buildInfoImageRef },
      })

      if (existingBuildInfo) {
        workspace.buildInfo = existingBuildInfo
        await this.buildInfoRepository.update(workspace.buildInfo.imageRef, { lastUsedAt: new Date() })
      } else {
        const buildInfoEntity = this.buildInfoRepository.create({
          ...createWorkspaceDto.buildInfo,
        })
        await this.buildInfoRepository.save(buildInfoEntity)
        workspace.buildInfo = buildInfoEntity
      }
    }

    if (createWorkspaceDto.autoStopInterval !== undefined) {
      workspace.autoStopInterval = createWorkspaceDto.autoStopInterval
    }

    if (createWorkspaceDto.autoArchiveInterval !== undefined) {
      workspace.autoArchiveInterval = this.resolveAutoArchiveInterval(createWorkspaceDto.autoArchiveInterval)
    }

    const imageRef = workspace.buildInfo ? workspace.buildInfo.imageRef : image.internalName

    try {
      workspace.runnerId = await this.runnerService.getRandomAvailableRunner({
        region: workspace.region,
        workspaceClass: workspace.class,
        imageRef,
      })
    } catch (error) {
      if (
        error instanceof BadRequestError == false ||
        error.message !== 'No available runners' ||
        !workspace.buildInfo
      ) {
        throw error
      }
      workspace.state = WorkspaceState.PENDING_BUILD
    }

    await this.workspaceRepository.insert(workspace)
    return workspace
  }

  async createBackup(workspaceId: string): Promise<void> {
    const workspace = await this.workspaceRepository.findOne({
      where: {
        id: workspaceId,
      },
    })

    if (!workspace) {
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }

    if (![BackupState.COMPLETED, BackupState.NONE].includes(workspace.backupState)) {
      throw new WorkspaceError('Workspace backup is already in progress')
    }

    await this.workspaceRepository.update(workspaceId, {
      backupState: BackupState.PENDING,
    })

    this.eventEmitter.emit(WorkspaceEvents.BACKUP_CREATED, new WorkspaceBackupCreatedEvent(workspace))
  }

  async findAll(organizationId: string, labels?: { [key: string]: string }): Promise<Workspace[]> {
    return this.workspaceRepository.find({
      where: {
        organizationId,
        state: Not(WorkspaceState.DESTROYED),
        ...(labels ? { labels: JsonContains(labels) } : {}),
      },
    })
  }

  async findOne(workspaceId: string, returnDestroyed?: boolean): Promise<Workspace> {
    const workspace = await this.workspaceRepository.findOne({
      where: {
        id: workspaceId,
        ...(returnDestroyed ? {} : { state: Not(WorkspaceState.DESTROYED) }),
      },
    })

    if (!workspace) {
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }

    return workspace
  }

  async getPortPreviewUrl(workspaceId: string, port: number): Promise<PortPreviewUrlDto> {
    if (port < 1 || port > 65535) {
      throw new BadRequestError('Invalid port')
    }

    const workspace = await this.workspaceRepository.findOne({
      where: { id: workspaceId },
    })

    if (!workspace) {
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }

    // Validate workspace is in valid state
    if (workspace.state !== WorkspaceState.STARTED) {
      throw new WorkspaceError('Workspace must be started to get port preview URL')
    }

    // Get runner info
    const runner = await this.runnerService.findOne(workspace.runnerId)
    if (!runner) {
      throw new NotFoundException(`Runner not found for workspace ${workspaceId}`)
    }

    return {
      url: `https://${port}-${workspace.id}.${runner.domain}`,
      token: workspace.authToken,
    }
  }

  async destroy(workspaceId: string): Promise<void> {
    const workspace = await this.workspaceRepository.findOne({
      where: {
        id: workspaceId,
      },
    })

    if (!workspace) {
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }

    if (workspace.pending) {
      throw new WorkspaceError('Workspace state change in progress')
    }
    workspace.pending = true
    workspace.desiredState = WorkspaceDesiredState.DESTROYED
    await this.workspaceRepository.save(workspace)

    this.eventEmitter.emit(WorkspaceEvents.DESTROYED, new WorkspaceDestroyedEvent(workspace))
  }

  async start(workspaceId: string, organization: Organization): Promise<void> {
    const workspace = await this.workspaceRepository.findOne({
      where: {
        id: workspaceId,
      },
    })

    if (!workspace) {
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }

    if (String(workspace.state) !== String(workspace.desiredState)) {
      throw new WorkspaceError('State change in progress')
    }

    if (![WorkspaceState.STOPPED, WorkspaceState.ARCHIVED].includes(workspace.state)) {
      throw new WorkspaceError('Workspace is not in valid state')
    }

    await this.assertOrganizationIsNotSuspended(organization)

    if (workspace.runnerId) {
      // Add runner readiness check
      const runner = await this.runnerService.findOne(workspace.runnerId)
      if (runner.state !== RunnerState.READY) {
        throw new WorkspaceError('Runner is not ready')
      }

      if (runner.unschedulable && workspace.backupState !== BackupState.COMPLETED) {
        throw new WorkspaceError('Runner is unschedulable - can not start workspace until the backup is completed')
      }
    } else {
      //  restore operation
      //  like a new workspace creation, we need to validate quotas
      await this.validateOrganizationQuotas(organization, workspace.cpu, workspace.mem, workspace.disk, workspace.id)
    }

    if (workspace.pending) {
      throw new WorkspaceError('Workspace state change in progress')
    }

    workspace.pending = true
    workspace.desiredState = WorkspaceDesiredState.STARTED
    await this.workspaceRepository.save(workspace)

    this.eventEmitter.emit(WorkspaceEvents.STARTED, new WorkspaceStartedEvent(workspace))
  }

  async stop(workspaceId: string): Promise<void> {
    const workspace = await this.workspaceRepository.findOne({
      where: {
        id: workspaceId,
      },
    })

    if (!workspace) {
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }

    if (String(workspace.state) !== String(workspace.desiredState)) {
      throw new WorkspaceError('State change in progress')
    }

    if (workspace.state !== WorkspaceState.STARTED) {
      throw new WorkspaceError('Workspace is not started')
    }

    if (workspace.pending) {
      throw new WorkspaceError('Workspace state change in progress')
    }
    workspace.pending = true
    workspace.desiredState = WorkspaceDesiredState.STOPPED
    await this.workspaceRepository.save(workspace)

    this.eventEmitter.emit(WorkspaceEvents.STOPPED, new WorkspaceStoppedEvent(workspace))
  }

  async updatePublicStatus(workspaceId: string, isPublic: boolean): Promise<void> {
    const workspace = await this.workspaceRepository.findOne({
      where: { id: workspaceId },
    })

    if (!workspace) {
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }

    workspace.public = isPublic
    await this.workspaceRepository.save(workspace)
  }

  private isValidRegion(region: RunnerRegion): boolean {
    return Object.values(RunnerRegion).includes(region)
  }

  private isValidClass(workspaceClass: WorkspaceClass): boolean {
    return Object.values(WorkspaceClass).includes(workspaceClass)
  }

  async replaceLabels(workspaceId: string, labels: { [key: string]: string }): Promise<{ [key: string]: string }> {
    const workspace = await this.workspaceRepository.findOne({
      where: { id: workspaceId },
    })

    if (!workspace) {
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }

    // Replace all labels
    workspace.labels = labels
    await this.workspaceRepository.save(workspace)

    return workspace.labels
  }

  @Cron(CronExpression.EVERY_10_MINUTES)
  async cleanupDestroyedWorkspaces() {
    const twentyFourHoursAgo = new Date()
    twentyFourHoursAgo.setHours(twentyFourHoursAgo.getHours() - 24)

    const destroyedWorkspaces = await this.workspaceRepository.delete({
      state: WorkspaceState.DESTROYED,
      updatedAt: LessThan(twentyFourHoursAgo),
    })

    if (destroyedWorkspaces.affected > 0) {
      this.logger.debug(`Cleaned up ${destroyedWorkspaces.affected} destroyed workspaces`)
    }
  }

  async setAutostopInterval(workspaceId: string, interval: number): Promise<void> {
    const workspace = await this.workspaceRepository.findOne({
      where: { id: workspaceId },
    })

    if (!workspace) {
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }

    // Validate interval is non-negative
    if (interval < 0) {
      throw new BadRequestError('Auto-stop interval must be non-negative')
    }

    workspace.autoStopInterval = interval
    await this.workspaceRepository.save(workspace)
  }

  async setAutoArchiveInterval(workspaceId: string, interval: number): Promise<void> {
    const workspace = await this.workspaceRepository.findOne({
      where: { id: workspaceId },
    })

    if (!workspace) {
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }

    workspace.autoArchiveInterval = this.resolveAutoArchiveInterval(interval)
    await this.workspaceRepository.save(workspace)
  }

  @OnEvent(WarmPoolEvents.TOPUP_REQUESTED)
  private async createWarmPoolWorkspace(event: WarmPoolTopUpRequested) {
    const warmPoolItem = event.warmPool
    await this.create(WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION, {
      image: warmPoolItem.image,
      cpu: warmPoolItem.cpu,
      gpu: warmPoolItem.gpu,
      memory: warmPoolItem.mem,
      disk: warmPoolItem.disk,
      target: warmPoolItem.target,
      env: warmPoolItem.env,
      class: warmPoolItem.class,
    })
  }

  @Cron(CronExpression.EVERY_MINUTE)
  private async handleUnschedulableRunners() {
    const runners = await this.runnerRepository.find({ where: { unschedulable: true } })

    if (runners.length === 0) {
      return
    }

    //  find all workspaces that are using the unschedulable runners and have organizationId = '00000000-0000-0000-0000-000000000000'
    const workspaces = await this.workspaceRepository.find({
      where: {
        runnerId: In(runners.map((runner) => runner.id)),
        organizationId: '00000000-0000-0000-0000-000000000000',
        state: WorkspaceState.STARTED,
        desiredState: Not(WorkspaceDesiredState.DESTROYED),
      },
    })

    if (workspaces.length === 0) {
      return
    }

    const destroyPromises = workspaces.map((workspace) => this.destroy(workspace.id))
    const results = await Promise.allSettled(destroyPromises)

    // Log any failed workspace destructions
    results.forEach((result, index) => {
      if (result.status === 'rejected') {
        this.logger.error(`Failed to destroy workspace ${workspaces[index].id}: ${result.reason}`)
      }
    })
  }

  private async assertOrganizationIsNotSuspended(organization: Organization): Promise<void> {
    if (!organization.suspended) {
      return
    }

    if (organization.suspendedUntil ? organization.suspendedUntil > new Date() : true) {
      if (organization.suspensionReason) {
        throw new ForbiddenException(`Organization is suspended: ${organization.suspensionReason}`)
      } else {
        throw new ForbiddenException('Organization is suspended')
      }
    }
  }

  async isWorkspacePublic(workspaceId: string): Promise<boolean> {
    const workspace = await this.workspaceRepository.findOne({
      where: { id: workspaceId },
    })

    if (!workspace) {
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }

    return workspace.public
  }

  @OnEvent(OrganizationEvents.SUSPENDED_WORKSPACE_STOPPED)
  async handleSuspendedWorkspaceStopped(event: OrganizationSuspendedWorkspaceStoppedEvent) {
    await this.stop(event.workspaceId).catch((error) => {
      //  log the error for now, but don't throw it as it will be retried
      this.logger.error(
        `Error stopping workspace from suspended organization. WorkspaceId: ${event.workspaceId}: `,
        error,
      )
    })
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
