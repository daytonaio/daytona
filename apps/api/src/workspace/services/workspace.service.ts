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
import { NodeRegion } from '../enums/node-region.enum'
import { WorkspaceDesiredState } from '../enums/workspace-desired-state.enum'
import { NodeApiFactory } from '../runner-api/runnerApi'
import { NodeService } from './node.service'
import { WorkspaceError } from '../../exceptions/workspace-error.exception'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { WorkspaceStateService } from './workspace-state.service'
import { Cron, CronExpression } from '@nestjs/schedule'
import { NodeState } from '../enums/node-state.enum'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { Image } from '../entities/image.entity'
import { ImageState } from '../enums/image-state.enum'
import { WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../constants/workspace.constants'
import { ConfigService } from '@nestjs/config'
import { OrganizationService } from '../../organization/services/organization.service'
import { ResizeDto } from '../../workspace/dto/resize.dto'
import { WorkspaceWarmPoolService } from './workspace-warm-pool.service'
import { EventEmitter2, OnEvent } from '@nestjs/event-emitter'
import { WarmPoolEvents } from '../constants/warmpool-events.constants'
import { WarmPoolTopUpRequested } from '../events/warmpool-topup-requested.event'
import { Node } from '../entities/node.entity'
import { PortPreviewUrlDto } from '../dto/port-preview-url.dto'
import { Organization } from '../../organization/entities/organization.entity'
import { WorkspaceEvents } from '../constants/workspace-events.constants'
import { WorkspaceStateUpdatedEvent } from '../events/workspace-state-updated.event'
import { BuildInfo } from '../entities/build-info.entity'
import { generateBuildInfoHash as generateBuildImageRef } from '../entities/build-info.entity'

@Injectable()
export class WorkspaceService {
  private readonly logger = new Logger(WorkspaceService.name)

  constructor(
    @InjectRepository(Workspace)
    private readonly workspaceRepository: Repository<Workspace>,
    @InjectRepository(Image)
    private readonly imageRepository: Repository<Image>,
    @InjectRepository(Node)
    private readonly nodeRepository: Repository<Node>,
    @InjectRepository(BuildInfo)
    private readonly buildInfoRepository: Repository<BuildInfo>,
    private readonly nodeService: NodeService,
    private readonly workspaceStateService: WorkspaceStateService,
    private readonly nodeApiFactory: NodeApiFactory,
    private readonly configService: ConfigService,
    private readonly warmPoolService: WorkspaceWarmPoolService,
    private readonly organizationService: OrganizationService,
    private readonly eventEmitter: EventEmitter2,
  ) {}

  private async validateOrganizationQuotas(
    organizationId: string,
    cpu: number,
    memory: number,
    disk: number,
    excludeWorkspaceId?: string,
  ): Promise<void> {
    const organization = await this.organizationService.findOne(organizationId)

    if (!organization) {
      throw new NotFoundException(`Organization with ID ${organizationId} not found`)
    }

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

    // Get total disk usage from all hot workspaces
    const hotWorkspaces = await this.workspaceRepository.find({
      where: {
        organizationId: organization.id,
        state: Not(In([WorkspaceState.DESTROYED, WorkspaceState.ARCHIVED, WorkspaceState.ERROR])),
        id: excludeWorkspaceId ? Not(excludeWorkspaceId) : undefined,
      },
    })

    const currentDisk = hotWorkspaces.reduce((sum, ws) => sum + ws.disk, 0)
    if (currentDisk + disk > organization.totalDiskQuota) {
      throw new ForbiddenException(
        `Total disk quota exceeded (${currentDisk + disk}GB > ${organization.totalDiskQuota}GB)`,
      )
    }

    // Get current resource usage from active workspaces
    const activeWorkspaces = await this.workspaceRepository.find({
      where: {
        organizationId,
        state: In([
          WorkspaceState.STARTED,
          WorkspaceState.STARTING,
          WorkspaceState.RESTORING,
          WorkspaceState.PULLING_IMAGE,
          WorkspaceState.CREATING,
        ]),
        id: excludeWorkspaceId ? Not(excludeWorkspaceId) : undefined,
      },
    })

    const currentCpu = activeWorkspaces.reduce((sum, ws) => sum + ws.cpu, 0)
    const currentMemory = activeWorkspaces.reduce((sum, ws) => sum + ws.mem, 0)

    // Check total resource quotas
    if (currentCpu + cpu > organization.totalCpuQuota) {
      throw new ForbiddenException(`Total CPU quota exceeded (${currentCpu + cpu} > ${organization.totalCpuQuota})`)
    }
    if (currentMemory + memory > organization.totalMemoryQuota) {
      throw new ForbiddenException(
        `Total memory quota exceeded (${currentMemory + memory}GB > ${organization.totalMemoryQuota}GB)`,
      )
    }

    // Check concurrent workspace limit
    const startedWorkspaces = activeWorkspaces.filter((ws) => ws.state === WorkspaceState.STARTED).length

    if (startedWorkspaces >= organization.maxConcurrentWorkspaces) {
      throw new ForbiddenException(
        `Maximum number of concurrent workspaces (${organization.maxConcurrentWorkspaces}) reached`,
      )
    }

    // Check total workspace quota if set
    if (organization.workspaceQuota > 0 && activeWorkspaces.length >= organization.workspaceQuota) {
      throw new ForbiddenException(`Workspace quota limit (${organization.workspaceQuota}) reached`)
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

    if (workspace.snapshotState !== SnapshotState.COMPLETED) {
      throw new WorkspaceError('Workspace snapshot is not completed')
    }

    if (workspace.pending) {
      throw new WorkspaceError('Workspace state change in progress')
    }
    workspace.pending = true
    workspace.desiredState = WorkspaceDesiredState.ARCHIVED
    await this.workspaceRepository.save(workspace)
    this.workspaceStateService.syncInstanceState(workspace.id).catch((err) => this.logger.error(err))
  }

  async count(organizationId: string): Promise<number> {
    return this.workspaceRepository.count({
      where: {
        organizationId,
        state: Not(WorkspaceState.DESTROYED),
      },
    })
  }

  async create(organizationId: string | null, createWorkspaceDto: CreateWorkspaceDto): Promise<Workspace> {
    // Validate region and class
    const region = createWorkspaceDto.target || NodeRegion.EU
    if (!this.isValidRegion(region)) {
      throw new BadRequestError('Invalid region')
    }
    const workspaceClass = createWorkspaceDto.class || WorkspaceClass.SMALL
    if (!this.isValidClass(workspaceClass)) {
      throw new BadRequestError('Invalid class')
    }

    // Validate organization quotas before creating workspace
    if (organizationId !== WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION) {
      await this.validateOrganizationQuotas(
        organizationId,
        createWorkspaceDto.cpu || 2,
        createWorkspaceDto.memory || 4,
        createWorkspaceDto.disk || 10,
      )
    }

    //  validate image
    let workspaceImage = createWorkspaceDto.image

    if ((!createWorkspaceDto.image || createWorkspaceDto.image.trim() === '') && !createWorkspaceDto.buildInfo) {
      workspaceImage = this.configService.get<string>('DEFAULT_IMAGE')
    }

    const image = await this.imageRepository.findOne({
      where: [
        { organizationId, name: workspaceImage, state: ImageState.ACTIVE },
        { general: true, name: workspaceImage, state: ImageState.ACTIVE },
      ],
    })

    if (!createWorkspaceDto.buildInfo) {
      if (!image) {
        throw new BadRequestError(`Image ${workspaceImage} not found or not accessible`)
      }

      if (organizationId !== WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION) {
        const warmPoolWorkspace = await this.warmPoolService.fetchWarmPoolWorkspace({
          organizationId: organizationId,
          image: workspaceImage,
          target: createWorkspaceDto.target,
          class: createWorkspaceDto.class,
          cpu: createWorkspaceDto.cpu,
          mem: createWorkspaceDto.memory,
          disk: createWorkspaceDto.disk,
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

    workspace.cpu = createWorkspaceDto.cpu || 2
    workspace.gpu = createWorkspaceDto.gpu || 0
    workspace.mem = createWorkspaceDto.memory || 4
    workspace.disk = createWorkspaceDto.disk || 10

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

    const imageRef = workspace.buildInfo ? workspace.buildInfo.imageRef : image.internalName

    try {
      workspace.nodeId = await this.nodeService.getRandomAvailableNode(workspace.region, workspace.class, imageRef)
    } catch (error) {
      if (error instanceof BadRequestError == false || error.message !== 'No available nodes' || !workspace.buildInfo) {
        throw error
      }
      workspace.state = WorkspaceState.PENDING_BUILD
    }

    const response = await this.workspaceRepository.save(workspace)
    this.workspaceStateService.syncInstanceState(response.id).catch((err) => this.logger.error(err))
    return response
  }

  async createSnapshot(workspaceId: string): Promise<void> {
    const workspace = await this.workspaceRepository.findOne({
      where: {
        id: workspaceId,
      },
    })

    if (!workspace) {
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }

    await this.workspaceStateService.startSnapshotCreate(workspace.id)
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

    // Get node info
    const node = await this.nodeService.findOne(workspace.nodeId)
    if (!node) {
      throw new NotFoundException(`Node not found for workspace ${workspaceId}`)
    }

    return {
      url: `https://${port}-${workspace.id}.${node.domain}`,
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

    if ([WorkspaceState.DESTROYED, WorkspaceState.UNKNOWN, WorkspaceState.CREATING].includes(workspace.state)) {
      throw new WorkspaceError('Workspace can not be destroyed at this time')
    }

    if (workspace.pending) {
      throw new WorkspaceError('Workspace state change in progress')
    }
    workspace.pending = true
    workspace.desiredState = WorkspaceDesiredState.DESTROYED
    await this.workspaceRepository.save(workspace)
    this.workspaceStateService.syncInstanceState(workspace.id).catch((err) => this.logger.error(err))
  }

  async start(workspaceId: string): Promise<void> {
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

    // Check concurrent workspace limit before starting
    const organization = await this.organizationService.findOne(workspace.organizationId)
    if (!organization) {
      throw new NotFoundException(`Organization with ID ${workspace.organizationId} not found`)
    }

    await this.assertOrganizationIsNotSuspended(organization)

    const startedWorkspaces = await this.workspaceRepository.count({
      where: {
        organizationId: workspace.organizationId,
        state: WorkspaceState.STARTED,
      },
    })

    if (startedWorkspaces >= organization.maxConcurrentWorkspaces) {
      throw new ForbiddenException(
        `Maximum number of concurrent workspaces (${organization.maxConcurrentWorkspaces}) reached`,
      )
    }

    if (workspace.nodeId) {
      // Add node readiness check
      const node = await this.nodeService.findOne(workspace.nodeId)
      if (node.state !== NodeState.READY) {
        throw new WorkspaceError('Node is not ready')
      }

      if (node.unschedulable && workspace.snapshotState !== SnapshotState.COMPLETED) {
        throw new WorkspaceError('Node is unschedulable - can not start workspace until the snapshot is completed')
      }
    } else {
      //  restore operation
      //  like a new workspace creation, we need to validate quotas
      await this.validateOrganizationQuotas(
        workspace.organizationId,
        workspace.cpu,
        workspace.mem,
        workspace.disk,
        workspace.id,
      )
    }

    if (workspace.pending) {
      throw new WorkspaceError('Workspace state change in progress')
    }

    workspace.pending = true
    workspace.desiredState = WorkspaceDesiredState.STARTED
    await this.workspaceRepository.save(workspace)
    this.workspaceStateService.syncInstanceState(workspace.id).catch((err) => this.logger.error(err))
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
    this.workspaceStateService.syncInstanceState(workspace.id).catch((err) => this.logger.error(err))
  }

  async resize(workspaceId: string, resizeDto: ResizeDto): Promise<void> {
    const workspace = await this.workspaceRepository.findOne({
      where: { id: workspaceId },
    })

    if (!workspace) {
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }

    if (resizeDto.gpu != 0) {
      throw new ForbiddenException('GPU resize is not supported')
    }

    if (String(workspace.state) !== String(workspace.desiredState)) {
      throw new WorkspaceError('State change in progress')
    }

    if (![WorkspaceState.STARTED, WorkspaceState.STOPPED].includes(workspace.state)) {
      throw new WorkspaceError('Workspace must be in started or stopped state')
    }

    if (workspace.pending) {
      throw new WorkspaceError('Workspace state change in progress')
    }

    //  check for quotas
    await this.validateOrganizationQuotas(
      workspace.organizationId,
      resizeDto.cpu,
      resizeDto.memory,
      workspace.disk,
      workspaceId,
    )

    workspace.cpu = resizeDto.cpu
    workspace.gpu = resizeDto.gpu
    workspace.mem = resizeDto.memory

    workspace.pending = true
    workspace.desiredState = WorkspaceDesiredState.RESIZED
    await this.workspaceRepository.save(workspace)
    this.workspaceStateService.syncInstanceState(workspace.id).catch((err) => this.logger.error(err))
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

  private isValidRegion(region: NodeRegion): boolean {
    return Object.values(NodeRegion).includes(region)
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
  private async handleUnschedulableNodes() {
    const nodes = await this.nodeRepository.find({ where: { unschedulable: true } })

    if (nodes.length === 0) {
      return
    }

    //  find all workspaces that are using the unschedulable nodes and have organizationId = '00000000-0000-0000-0000-000000000000'
    const workspaces = await this.workspaceRepository.find({
      where: {
        nodeId: In(nodes.map((node) => node.id)),
        organizationId: '00000000-0000-0000-0000-000000000000',
        state: WorkspaceState.STARTED,
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
}
