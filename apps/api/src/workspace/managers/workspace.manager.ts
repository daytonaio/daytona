/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { In, Not, Raw, Repository } from 'typeorm'
import { Workspace } from '../entities/workspace.entity'
import { WorkspaceState } from '../enums/workspace-state.enum'
import { WorkspaceDesiredState } from '../enums/workspace-desired-state.enum'
import { NodeApiFactory } from '../runner-api/runnerApi'
import { NodeService } from '../services/node.service'
import { EnumsSandboxState as NodeWorkspaceState } from '@daytonaio/runner-api-client'
import { NodeState } from '../enums/node-state.enum'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { ImageService } from '../services/image.service'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../constants/workspace.constants'
import { DockerProvider } from '../docker/docker-provider'
import { OrganizationService } from '../../organization/services/organization.service'
import { ImageNodeState } from '../enums/image-node-state.enum'
import { BuildInfo } from '../entities/build-info.entity'
import { CreateSandboxDTO } from '@daytonaio/runner-api-client'
import { fromAxiosError } from '../../common/utils/from-axios-error'
import { OnEvent } from '@nestjs/event-emitter'
import { WorkspaceEvents } from '../constants/workspace-events.constants'
import { WorkspaceStoppedEvent } from '../events/workspace-stopped.event'
import { WorkspaceStartedEvent } from '../events/workspace-started.event'
import { WorkspaceArchivedEvent } from '../events/workspace-archived.event'
import { WorkspaceDestroyedEvent } from '../events/workspace-destroyed.event'
import { WorkspaceCreatedEvent } from '../events/workspace-create.event'

type BreakFromSwitch = boolean
const SYNC_INSTANCE_STATE_LOCK_KEY = 'sync-instance-state-'

@Injectable()
export class WorkspaceManager {
  private readonly logger = new Logger(WorkspaceManager.name)

  constructor(
    @InjectRepository(Workspace)
    private readonly workspaceRepository: Repository<Workspace>,
    private readonly nodeService: NodeService,
    private readonly nodeApiFactory: NodeApiFactory,
    private readonly dockerRegistryService: DockerRegistryService,
    @InjectRedis() private readonly redis: Redis,
    private readonly imageService: ImageService,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly dockerProvider: DockerProvider,
    private readonly organizationService: OrganizationService,
  ) {}

  @Cron(CronExpression.EVERY_MINUTE, { name: 'auto-stop-check' })
  async autostopCheck(): Promise<void> {
    //  lock the sync to only run one instance at a time
    const snapshotCheckWorkerSelected = await this.redis.get('auto-stop-check-worker-selected')
    if (snapshotCheckWorkerSelected) {
      return
    }
    //  keep the worker selected for 1 minute
    await this.redis.setex('auto-stop-check-worker-selected', 60, '1')

    // Get all ready nodes
    const allNodes = await this.nodeService.findAll()
    const readyNodes = allNodes.filter((node) => node.state === NodeState.READY)

    // Process all nodes in parallel
    await Promise.all(
      readyNodes.map(async (node) => {
        const workspaces = await this.workspaceRepository.find({
          where: {
            nodeId: node.id,
            organizationId: Not(WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION),
            state: WorkspaceState.STARTED,
            autoStopInterval: Not(0),
            lastActivityAt: Raw((alias) => `${alias} < NOW() - INTERVAL '1 minute' * "autoStopInterval"`),
          },
          order: {
            lastSnapshotAt: 'ASC',
          },
          //  todo: increase this number when auto-stop is stable
          take: 10,
        })

        await Promise.all(
          workspaces.map(async (workspace) => {
            const lockKey = SYNC_INSTANCE_STATE_LOCK_KEY + workspace.id
            const locked = await this.redisLockProvider.lock(lockKey, 30)
            if (locked) {
              return
            }

            try {
              workspace.desiredState = WorkspaceDesiredState.STOPPED
              await this.workspaceRepository.save(workspace)
              await this.redis.del(lockKey)
              this.syncInstanceState(workspace.id)
            } catch (error) {
              this.logger.error(
                `Error processing auto-stop state for workspace ${workspace.id}:`,
                fromAxiosError(error),
              )
            }
          }),
        )
      }),
    )
  }

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'sync-states' })
  async syncStates(): Promise<void> {
    const lockKey = 'sync-states'
    if (await this.redisLockProvider.lock(lockKey, 30)) {
      return
    }

    const workspaces = await this.workspaceRepository.find({
      where: {
        state: Not(In([WorkspaceState.DESTROYED, WorkspaceState.ERROR])),
        desiredState: Raw(() => '"Workspace"."desiredState"::text != "Workspace"."state"::text'),
      },
      take: 100,
      order: {
        lastActivityAt: 'DESC',
      },
    })

    await Promise.all(
      workspaces.map(async (workspace) => {
        //  if the workspace is already being processed, skip it
        const lockKey = SYNC_INSTANCE_STATE_LOCK_KEY + workspace.id
        const hasLock = await this.redis.get(lockKey)
        if (hasLock) {
          return
        }
        this.syncInstanceState(workspace.id)
      }),
    )
    await this.redisLockProvider.unlock(lockKey)
  }

  @Cron(CronExpression.EVERY_10_MINUTES, { name: 'stop-suspended-organization-workspaces' })
  async stopSuspendedOrganizationWorkspaces(): Promise<void> {
    //  lock the sync to only run one instance at a time
    const lockKey = 'stop-suspended-organization-workspaces'
    const hasLock = await this.redis.get(lockKey)
    if (hasLock) {
      return
    }
    //  keep the worker selected for 1 minute
    await this.redis.setex(lockKey, 60, '1')

    const suspendedOrganizations = await this.organizationService.findSuspended(
      // Find organization suspended more than 24 hours ago
      new Date(Date.now() - 1 * 1000 * 60 * 60 * 24),
    )

    const suspendedOrganizationIds = suspendedOrganizations.map((organization) => organization.id)

    const workspaces = await this.workspaceRepository.find({
      where: {
        organizationId: In(suspendedOrganizationIds),
        state: WorkspaceState.STARTED,
      },
    })

    await Promise.allSettled(
      workspaces.map(async (workspace) => {
        //  if the workspace is already being processed, skip it
        const lockKey = SYNC_INSTANCE_STATE_LOCK_KEY + workspace.id
        const hasLock = await this.redisLockProvider.lock(lockKey, 30)
        if (!hasLock) {
          return
        }

        try {
          await this.updateWorkspaceState(workspace.id, WorkspaceState.STOPPING)
          await this.handleWorkspaceDesiredStateStopped(workspace.id)
        } catch (error) {
          this.logger.error(
            `Error stopping workspace from suspended organization. WorkspaceId: ${workspace.id}: `,
            fromAxiosError(error),
          )
        } finally {
          await this.redis.del(lockKey)
        }
      }),
    )

    await this.redis.del(lockKey)
  }

  async syncInstanceState(workspaceId: string): Promise<void> {
    //  prevent syncState cron from running multiple instances of the same workspace
    const lockKey = SYNC_INSTANCE_STATE_LOCK_KEY + workspaceId
    await this.redis.setex(lockKey, 360, '1')

    const workspace = await this.workspaceRepository.findOneByOrFail({
      id: workspaceId,
    })

    try {
      switch (workspace.desiredState) {
        case WorkspaceDesiredState.STARTED: {
          await this.handleWorkspaceDesiredStateStarted(workspace.id)
          break
        }
        case WorkspaceDesiredState.STOPPED: {
          await this.handleWorkspaceDesiredStateStopped(workspace.id)
          break
        }
        case WorkspaceDesiredState.DESTROYED: {
          await this.handleWorkspaceDesiredStateDestroyed(workspace.id)
          break
        }
        case WorkspaceDesiredState.ARCHIVED: {
          await this.handleWorkspaceDesiredStateArchived(workspace.id)
          break
        }
      }
    } catch (error) {
      if (error.code === 'ECONNRESET') {
        await this.redis.del(lockKey)
        this.syncInstanceState(workspaceId)
        return
      }

      this.logger.error(`Error processing desired state for workspace ${workspaceId}:`, fromAxiosError(error))

      const workspace = await this.workspaceRepository.findOneBy({
        id: workspaceId,
      })
      if (!workspace) {
        //  edge case where workspace is deleted while desired state is being processed
        return
      }
      await this.updateWorkspaceErrorState(workspace.id, error.message || String(error))
    }

    //  unlock the workspace after 10 seconds
    //  this will allow the syncState cron to run again, but will allow
    //  the syncInstanceState to complete any pending state changes
    await this.redis.setex(lockKey, 10, '1')
  }

  private async handleUnassignedBuildWorkspace(workspace: Workspace): Promise<void> {
    // Try to assign an available node with the image build
    let nodeId: string
    try {
      nodeId = await this.nodeService.getRandomAvailableNode(
        workspace.region,
        workspace.class,
        workspace.buildInfo.imageRef,
      )
    } catch (error) {
      // Continue to next assignment method
    }

    if (nodeId) {
      await this.updateWorkspaceState(workspace.id, WorkspaceState.UNKNOWN, nodeId)
      this.syncInstanceState(workspace.id)
      return
    }

    // Try to assign an available node that is currently building the image
    const imageNodes = await this.nodeService.getImageNodes(workspace.buildInfo.imageRef)

    for (const imageNode of imageNodes) {
      const node = await this.nodeService.findOne(imageNode.nodeId)
      if (node.used < node.capacity) {
        if (imageNode.state === ImageNodeState.BUILDING_IMAGE) {
          const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
            id: workspace.id,
          })
          workspaceToUpdate.nodeId = node.id
          workspaceToUpdate.state = WorkspaceState.BUILDING_IMAGE
          await this.workspaceRepository.save(workspaceToUpdate)
          return
        } else if (imageNode.state === ImageNodeState.ERROR) {
          await this.updateWorkspaceErrorState(workspace.id, imageNode.errorReason)
          return
        }
      }
    }

    // Try to assign a new available node
    nodeId = await this.nodeService.getRandomAvailableNode(workspace.region, workspace.class)

    this.buildOnNode(workspace.buildInfo, nodeId, workspace.organizationId)

    await this.updateWorkspaceState(workspace.id, WorkspaceState.BUILDING_IMAGE, nodeId)
    await this.nodeService.recalculateNodeUsage(nodeId)
    this.syncInstanceState(workspace.id)
  }

  // Initiates the image build on the runner and creates an ImageNode depending on the result
  async buildOnNode(buildInfo: BuildInfo, nodeId: string, organizationId: string) {
    const node = await this.nodeService.findOne(nodeId)
    const nodeImageApi = this.nodeApiFactory.createImageApi(node)

    let retries = 0

    while (retries < 10) {
      try {
        await nodeImageApi.buildImage({
          image: buildInfo.imageRef,
          organizationId: organizationId,
          dockerfile: buildInfo.dockerfileContent,
          context: buildInfo.contextHashes,
        })
        break
      } catch (err) {
        if (err.code !== 'ECONNRESET') {
          await this.nodeService.createImageNode(nodeId, buildInfo.imageRef, ImageNodeState.ERROR, err.message)
          return
        }
      }
      retries++
      await new Promise((resolve) => setTimeout(resolve, retries * 1000))
    }

    if (retries === 10) {
      await this.nodeService.createImageNode(nodeId, buildInfo.imageRef, ImageNodeState.ERROR, 'Timeout while building')
      return
    }

    await this.nodeService.createImageNode(nodeId, buildInfo.imageRef, ImageNodeState.BUILDING_IMAGE)
  }

  private async handleWorkspaceDesiredStateArchived(workspaceId: string): Promise<void> {
    const workspace = await this.workspaceRepository.findOneByOrFail({
      id: workspaceId,
    })

    const inProgressOnNode = await this.workspaceRepository.count({
      where: {
        nodeId: workspace.nodeId,
        state: In([WorkspaceState.ARCHIVING]),
        id: Not(workspaceId),
      },
    })

    //  max 3 workspaces can be archived at the same time on the same node
    //  this is to prevent the node from being overloaded
    if (inProgressOnNode > 2) {
      return
    }

    switch (workspace.state) {
      case WorkspaceState.STOPPED: {
        await this.updateWorkspaceState(workspaceId, WorkspaceState.ARCHIVING)
        //  fallthrough to archiving state
      }
      case WorkspaceState.ARCHIVING: {
        //  TODO: timeout logic
        if (workspace.snapshotState !== SnapshotState.COMPLETED) {
          await this.redis.del(workspace.id)
          break
        }

        //  when the snapshot is completed, destroy the workspace on the node
        //  and deassociate the workspace from the node
        const node = await this.nodeService.findOne(workspace.nodeId)
        const nodeWorkspaceApi = this.nodeApiFactory.createWorkspaceApi(node)

        try {
          const workspaceInfoResponse = await nodeWorkspaceApi.info(workspace.id)
          const workspaceInfo = workspaceInfoResponse.data
          switch (workspaceInfo.state) {
            case NodeWorkspaceState.SandboxStateDestroying:
              //  wait until workspace is destroyed on node
              await this.redis.del(workspace.id)
              this.syncInstanceState(workspace.id)
              break
            case NodeWorkspaceState.SandboxStateDestroyed:
              await this.updateWorkspaceState(workspaceId, WorkspaceState.ARCHIVED, null)
              break
            default:
              await nodeWorkspaceApi.destroy(workspace.id)
              break
          }
        } catch (error) {
          //  fail for errors other than workspace not found or workspace already destroyed
          if (
            !(
              (error.response?.data?.statusCode === 400 &&
                error.response?.data?.message.includes('Workspace already destroyed')) ||
              error.response?.status === 404
            )
          ) {
            throw error
          }
          //  if the workspace is already destroyed, do nothing
          await this.updateWorkspaceState(workspaceId, WorkspaceState.ARCHIVED, null)
        }
        break
      }
    }
  }

  private async handleWorkspaceDesiredStateDestroyed(workspaceId: string): Promise<void> {
    const workspace = await this.workspaceRepository.findOneByOrFail({
      id: workspaceId,
    })

    if (workspace.state === WorkspaceState.ARCHIVED) {
      await this.updateWorkspaceState(workspace.id, WorkspaceState.DESTROYED)
      return
    }

    const node = await this.nodeService.findOne(workspace.nodeId)
    if (node.state !== NodeState.READY) {
      //  console.debug(`Node ${node.id} is not ready`);
      return
    }

    switch (workspace.state) {
      case WorkspaceState.DESTROYED:
        break
      case WorkspaceState.DESTROYING: {
        // check if workspace is destroyed
        const nodeWorkspaceApi = this.nodeApiFactory.createWorkspaceApi(node)

        try {
          const workspaceInfoResponse = await nodeWorkspaceApi.info(workspaceId)
          const workspaceInfo = workspaceInfoResponse.data
          if (
            workspaceInfo.state === NodeWorkspaceState.SandboxStateDestroyed ||
            workspaceInfo.state === NodeWorkspaceState.SandboxStateError
          ) {
            await nodeWorkspaceApi.removeDestroyed(workspaceId)
          }
        } catch (e) {
          //  if the workspace is not found on node, it is already destroyed
          if (!e.response || e.response.status !== 404) {
            throw e
          }
        }

        await this.updateWorkspaceState(workspace.id, WorkspaceState.DESTROYED)
        //  sync states again immediately for workspace
        this.syncInstanceState(workspace.id)
        break
      }
      default: {
        // destroy workspace
        try {
          const nodeWorkspaceApi = this.nodeApiFactory.createWorkspaceApi(node)
          const workspaceInfoResponse = await nodeWorkspaceApi.info(workspaceId)
          const workspaceInfo = workspaceInfoResponse.data
          if (workspaceInfo?.state === NodeWorkspaceState.SandboxStateDestroyed) {
            break
          }
          await nodeWorkspaceApi.destroy(workspace.id)
        } catch (e) {
          //  if the workspace is not found on node, it is already destroyed
          if (e.response.status !== 404) {
            throw e
          }
        }
        await this.updateWorkspaceState(workspace.id, WorkspaceState.DESTROYING)
        this.syncInstanceState(workspace.id)
        break
      }
    }
  }

  private async handleWorkspaceDesiredStateStarted(workspaceId: string): Promise<void> {
    const workspace = await this.workspaceRepository.findOneByOrFail({
      id: workspaceId,
    })

    switch (workspace.state) {
      case WorkspaceState.PENDING_BUILD: {
        await this.handleUnassignedBuildWorkspace(workspace)
        break
      }
      case WorkspaceState.BUILDING_IMAGE: {
        await this.handleNodeWorkspaceBuildingImageStateOnDesiredStateStart(workspace)
        break
      }
      case WorkspaceState.UNKNOWN: {
        await this.handleNodeWorkspaceUnknownStateOnDesiredStateStart(workspace)
        break
      }
      case WorkspaceState.ARCHIVED:
      case WorkspaceState.STOPPED: {
        if (await this.handleNodeWorkspaceStoppedOrArchivedStateOnDesiredStateStart(workspace)) {
          break
        }
      }
      // eslint-disable-next-line no-fallthrough
      case WorkspaceState.RESTORING:
      case WorkspaceState.CREATING:
        if (await this.handleNodeWorkspacePullingImageStateCheck(workspace)) {
          break
        }
      //  fallthrough to check if workspace is already started
      case WorkspaceState.PULLING_IMAGE:
      case WorkspaceState.STARTING: {
        await this.handleNodeWorkspaceStartedStateCheck(workspace)
        break
      }
      //  TODO: remove this case
      case WorkspaceState.ERROR: {
        //  TODO: remove this asap
        //  this was a temporary solution to recover from the false positive error state
        if (workspace.id.startsWith('err_')) {
          return
        }
        const node = await this.nodeService.findOne(workspace.nodeId)
        const nodeWorkspaceApi = this.nodeApiFactory.createWorkspaceApi(node)
        const workspaceInfoResponse = await nodeWorkspaceApi.info(workspace.id)
        const workspaceInfo = workspaceInfoResponse.data
        if (workspaceInfo.state === NodeWorkspaceState.SandboxStateStarted) {
          const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
            id: workspace.id,
          })
          workspaceToUpdate.state = WorkspaceState.STARTED
          workspaceToUpdate.snapshotState = SnapshotState.NONE
          await this.workspaceRepository.save(workspaceToUpdate)
        }
        break
      }
    }
  }

  private async handleWorkspaceDesiredStateStopped(workspaceId: string): Promise<void> {
    const workspace = await this.workspaceRepository.findOneByOrFail({
      id: workspaceId,
    })
    const node = await this.nodeService.findOne(workspace.nodeId)
    if (node.state !== NodeState.READY) {
      //  console.debug(`Node ${node.id} is not ready`);
      return
    }

    switch (workspace.state) {
      case WorkspaceState.STARTED: {
        // stop workspace
        const nodeWorkspaceApi = this.nodeApiFactory.createWorkspaceApi(node)
        await nodeWorkspaceApi.stop(workspace.id)
        await this.updateWorkspaceState(workspace.id, WorkspaceState.STOPPING)
        //  sync states again immediately for workspace
        await this.redis.del(workspace.id)
        this.syncInstanceState(workspace.id)
        break
      }
      case WorkspaceState.STOPPING: {
        // check if workspace is stopped
        const node = await this.nodeService.findOne(workspace.nodeId)
        const nodeWorkspaceApi = this.nodeApiFactory.createWorkspaceApi(node)
        const workspaceInfoResponse = await nodeWorkspaceApi.info(workspace.id)
        const workspaceInfo = workspaceInfoResponse.data
        switch (workspaceInfo.state) {
          case NodeWorkspaceState.SandboxStateStopped: {
            const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
              id: workspace.id,
            })
            workspaceToUpdate.state = WorkspaceState.STOPPED
            workspaceToUpdate.snapshotState = SnapshotState.NONE
            await this.workspaceRepository.save(workspaceToUpdate)
            break
          }
          case NodeWorkspaceState.SandboxStateError:
            {
              await this.updateWorkspaceErrorState(workspace.id, 'Sandbox is in error state on runner')
              break
            }
            break
        }
        //  sync states again immediately for workspace
        await this.redis.del(workspace.id)
        this.syncInstanceState(workspace.id)
        break
      }
      case WorkspaceState.ERROR: {
        if (workspace.id.startsWith('err_')) {
          return
        }
        const node = await this.nodeService.findOne(workspace.nodeId)
        const nodeWorkspaceApi = this.nodeApiFactory.createWorkspaceApi(node)
        const workspaceInfoResponse = await nodeWorkspaceApi.info(workspace.id)
        const workspaceInfo = workspaceInfoResponse.data
        if (workspaceInfo.state === NodeWorkspaceState.SandboxStateStopped) {
          await this.updateWorkspaceState(workspace.id, WorkspaceState.STOPPED)
        }
        break
      }
    }
  }

  private async handleNodeWorkspaceBuildingImageStateOnDesiredStateStart(workspace: Workspace) {
    const imageNode = await this.nodeService.getImageNode(workspace.nodeId, workspace.buildInfo.imageRef)
    if (imageNode) {
      switch (imageNode.state) {
        case ImageNodeState.READY: {
          // TODO: "UNKNOWN" should probably be changed to something else
          await this.workspaceRepository.update(workspace.id, {
            state: WorkspaceState.UNKNOWN,
          })
          this.syncInstanceState(workspace.id)
          return
        }
        case ImageNodeState.ERROR: {
          await this.workspaceRepository.update(workspace.id, {
            state: WorkspaceState.ERROR,
            errorReason: imageNode.errorReason,
          })
          return
        }
      }
    }
    if (!imageNode || imageNode.state === ImageNodeState.BUILDING_IMAGE) {
      // Sleep for a second and go back to syncing instance state
      await new Promise((resolve) => setTimeout(resolve, 1000))
      this.syncInstanceState(workspace.id)
      return
    }
  }

  private async handleNodeWorkspaceUnknownStateOnDesiredStateStart(workspace: Workspace) {
    const node = await this.nodeService.findOne(workspace.nodeId)
    if (node.state !== NodeState.READY) {
      //  console.debug(`Node ${node.id} is not ready`);
      return
    }

    let createWorkspaceDto: CreateSandboxDTO = {
      id: workspace.id,
      osUser: workspace.osUser,
      image: '',
      // TODO: organizationId: workspace.organizationId,
      userId: workspace.organizationId,
      storageQuota: workspace.disk,
      memoryQuota: workspace.mem,
      cpuQuota: workspace.cpu,
      // gpuQuota: workspace.gpu,
      env: workspace.env,
      // public: workspace.public,
      volumes: workspace.volumes,
    }

    if (!workspace.buildInfo) {
      //  get internal image name
      const image = await this.imageService.getImageByName(workspace.image, workspace.organizationId)
      const internalImageName = image.internalName

      const registry = await this.dockerRegistryService.findOneByImageName(internalImageName, workspace.organizationId)
      if (!registry) {
        throw new Error('No registry found for image')
      }

      createWorkspaceDto = {
        ...createWorkspaceDto,
        image: internalImageName,
        entrypoint: image.entrypoint,
        registry: {
          url: registry.url,
          username: registry.username,
          password: registry.password,
        },
      }
    } else {
      createWorkspaceDto = {
        ...createWorkspaceDto,
        image: workspace.buildInfo.imageRef,
        entrypoint: this.getEntrypointFromDockerfile(workspace.buildInfo.dockerfileContent),
      }
    }

    const nodeWorkspaceApi = this.nodeApiFactory.createWorkspaceApi(node)
    await nodeWorkspaceApi.create(createWorkspaceDto)
    await this.updateWorkspaceState(workspace.id, WorkspaceState.CREATING)
    //  sync states again immediately for workspace
    await this.redis.del(workspace.id)
    this.syncInstanceState(workspace.id)
  }

  // TODO: revise/cleanup
  private getEntrypointFromDockerfile(dockerfileContent: string): string[] {
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

    // Match CMD with either a string or JSON array
    const cmdMatch = dockerfileContent.match(/CMD\s+(.*)/)
    if (cmdMatch) {
      const rawCmd = cmdMatch[1].trim()
      try {
        const parsed = JSON.parse(rawCmd)
        if (Array.isArray(parsed)) {
          return parsed
        }
      } catch {
        return [rawCmd.replace(/["']/g, '')]
      }
    }

    return ['sleep', 'infinity']
  }

  private async handleNodeWorkspaceStoppedOrArchivedStateOnDesiredStateStart(
    workspace: Workspace,
  ): Promise<BreakFromSwitch> {
    //  check if workspace is assigned to a node and if that node is unschedulable
    //  if it is, move workspace to prevNodeId, and set nodeId to null
    //  this will assign a new node to the workspace and restore the workspace from the latest snapshot
    if (workspace.nodeId) {
      const node = await this.nodeService.findOne(workspace.nodeId)
      if (node.unschedulable) {
        //  check if workspace has a valid snapshot
        if (workspace.snapshotState !== SnapshotState.COMPLETED) {
          //  if not, keep workspace on the same node
        } else {
          workspace.prevNodeId = workspace.nodeId
          workspace.nodeId = null

          const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
            id: workspace.id,
          })
          workspaceToUpdate.prevNodeId = workspace.nodeId
          workspaceToUpdate.nodeId = null
          await this.workspaceRepository.save(workspaceToUpdate)
        }
      }
    }

    if (workspace.nodeId === null) {
      //  if workspace has no node, check if snapshot is completed
      //  if not, set workspace to error
      //  if snapshot is completed, get random available node and start workspace
      //  use the snapshot image to start the workspace

      if (workspace.snapshotState !== SnapshotState.COMPLETED) {
        await this.updateWorkspaceErrorState(workspace.id, 'Workspace has no node and snapshot is not completed')
        return true
      }

      const registry = await this.dockerRegistryService.findOne(workspace.snapshotRegistryId)
      if (!registry) {
        throw new Error('No registry found for image')
      }

      const existingImages = workspace.existingSnapshotImages.map((existingImage) => existingImage.imageName)
      let validSnapshotImage
      let exists = false

      while (existingImages.length > 0) {
        try {
          if (!validSnapshotImage) {
            //  last image is the current image, so we don't need to check it
            //  just in case, we'll use the value from the snapshotImage property
            validSnapshotImage = workspace.snapshotImage
            existingImages.pop()
          } else {
            validSnapshotImage = existingImages.pop()
          }
          if (await this.dockerProvider.checkImageExistsInRegistry(validSnapshotImage, registry)) {
            exists = true
            break
          }
        } catch (error) {
          this.logger.error(
            `Failed to check if snapshot image ${workspace.snapshotImage} exists in registry ${registry.id}:`,
            fromAxiosError(error),
          )
        }
      }

      if (!exists) {
        await this.updateWorkspaceErrorState(workspace.id, 'No valid snapshot image found')
        return true
      }

      const image = await this.imageService.getImageByName(workspace.image, workspace.organizationId)

      const availableNodes = await this.nodeService.findAvailableNodes(
        workspace.region,
        workspace.class,
        image.internalName,
      )

      //  if there are available nodes with the workspace base image,
      //  search for available nodes with the base image cached on the node
      //  otherwise, search all available nodes
      const includeImage = availableNodes.length > 0 ? image.internalName : undefined

      const nodeId = await this.nodeService.getRandomAvailableNode(workspace.region, workspace.class, includeImage)
      const node = await this.nodeService.findOne(nodeId)
      if (node.state !== NodeState.READY) {
        //  console.debug(`Node ${node.id} is not ready`);
        return
      }

      const nodeWorkspaceApi = this.nodeApiFactory.createWorkspaceApi(node)

      await nodeWorkspaceApi.create({
        id: workspace.id,
        image: validSnapshotImage,
        osUser: workspace.osUser,
        // TODO: organizationId: workspace.organizationId,
        userId: workspace.organizationId,
        storageQuota: workspace.disk,
        memoryQuota: workspace.mem,
        cpuQuota: workspace.cpu,
        // gpuQuota: workspace.gpu,
        env: workspace.env,
        // public: workspace.public,
        registry: {
          url: registry.url,
          username: registry.username,
          password: registry.password,
        },
      })

      await this.updateWorkspaceState(workspace.id, WorkspaceState.RESTORING, nodeId)
    } else {
      // if workspace has node, start workspace
      const node = await this.nodeService.findOne(workspace.nodeId)

      const nodeWorkspaceApi = this.nodeApiFactory.createWorkspaceApi(node)

      await nodeWorkspaceApi.start(workspace.id)

      await this.updateWorkspaceState(workspace.id, WorkspaceState.STARTING)
      //  sync states again immediately for workspace
      await this.redis.del(workspace.id)
      this.syncInstanceState(workspace.id)
      return true
    }
    return false
  }

  //  used to check if workspace is pulling image on node and update workspace state accordingly
  private async handleNodeWorkspacePullingImageStateCheck(workspace: Workspace): Promise<BreakFromSwitch> {
    const node = await this.nodeService.findOne(workspace.nodeId)
    const nodeWorkspaceApi = this.nodeApiFactory.createWorkspaceApi(node)
    const workspaceInfoResponse = await nodeWorkspaceApi.info(workspace.id)
    const workspaceInfo = workspaceInfoResponse.data

    if (workspaceInfo.state === NodeWorkspaceState.SandboxStatePullingImage) {
      await this.updateWorkspaceState(workspace.id, WorkspaceState.PULLING_IMAGE)

      await this.redis.del(workspace.id)
      this.syncInstanceState(workspace.id)
      return true
    }
    if (workspaceInfo.state === NodeWorkspaceState.SandboxStateError) {
      await this.updateWorkspaceErrorState(workspace.id)
      return true
    }
    return false
  }

  //  used to check if workspace is started on node and update workspace state accordingly
  //  also used to handle the case where a workspace is started on a node and then transferred to a new node
  private async handleNodeWorkspaceStartedStateCheck(workspace: Workspace) {
    const node = await this.nodeService.findOne(workspace.nodeId)
    const nodeWorkspaceApi = this.nodeApiFactory.createWorkspaceApi(node)
    const workspaceInfoResponse = await nodeWorkspaceApi.info(workspace.id)
    const workspaceInfo = workspaceInfoResponse.data

    switch (workspaceInfo.state) {
      case NodeWorkspaceState.SandboxStateStarted: {
        //  if previous snapshot state is error or completed, set snapshot state to none
        if ([SnapshotState.ERROR, SnapshotState.COMPLETED].includes(workspace.snapshotState)) {
          workspace.snapshotState = SnapshotState.NONE

          const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
            id: workspace.id,
          })
          workspaceToUpdate.state = WorkspaceState.STARTED
          workspaceToUpdate.snapshotState = SnapshotState.NONE
          await this.workspaceRepository.save(workspaceToUpdate)
        } else {
          await this.updateWorkspaceState(workspace.id, WorkspaceState.STARTED)
        }

        //  if workspace was transferred to a new node, remove it from the old node
        if (workspace.prevNodeId) {
          const node = await this.nodeService.findOne(workspace.prevNodeId)
          if (!node) {
            this.logger.warn(`Previously assigned node ${workspace.prevNodeId} for workspace ${workspace.id} not found`)
            //  clear prevNodeId to avoid trying to cleanup on a non-existent node
            workspace.prevNodeId = null

            const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
              id: workspace.id,
            })
            workspaceToUpdate.prevNodeId = null
            await this.workspaceRepository.save(workspaceToUpdate)
            break
          }
          const nodeWorkspaceApi = this.nodeApiFactory.createWorkspaceApi(node)
          try {
            // First try to destroy the workspace
            await nodeWorkspaceApi.destroy(workspace.id)

            // Wait for workspace to be destroyed before removing
            let retries = 0
            while (retries < 10) {
              try {
                const workspaceInfo = await nodeWorkspaceApi.info(workspace.id)
                if (workspaceInfo.data.state === NodeWorkspaceState.SandboxStateDestroyed) {
                  break
                }
              } catch (e) {
                if (e.response?.status === 404) {
                  break // Workspace already gone
                }
                throw e
              }
              await new Promise((resolve) => setTimeout(resolve, 1000 * retries))
              retries++
            }

            // Finally remove the destroyed workspace
            await nodeWorkspaceApi.removeDestroyed(workspace.id)
            workspace.prevNodeId = null

            const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
              id: workspace.id,
            })
            workspaceToUpdate.prevNodeId = null
            await this.workspaceRepository.save(workspaceToUpdate)
          } catch (e) {
            this.logger.error(
              `Failed to cleanup workspace ${workspace.id} on previous node ${node.id}:`,
              fromAxiosError(e),
            )
          }
        }
        break
      }
      case NodeWorkspaceState.SandboxStateError: {
        await this.updateWorkspaceErrorState(workspace.id)
        break
      }
    }
    //  sync states again immediately for workspace
    await this.redis.del(workspace.id)
    this.syncInstanceState(workspace.id)
  }

  private async updateWorkspaceState(workspaceId: string, state: WorkspaceState, nodeId?: string | null | undefined) {
    const workspace = await this.workspaceRepository.findOneByOrFail({
      id: workspaceId,
    })
    if (workspace.state === state) {
      return
    }
    workspace.state = state
    if (nodeId !== undefined) {
      workspace.nodeId = nodeId
    }

    await this.workspaceRepository.save(workspace)
  }

  private async updateWorkspaceErrorState(workspaceId: string, errorReason?: string) {
    const workspace = await this.workspaceRepository.findOneByOrFail({
      id: workspaceId,
    })
    workspace.state = WorkspaceState.ERROR
    if (errorReason !== undefined) {
      workspace.errorReason = errorReason
    }
    await this.workspaceRepository.save(workspace)
  }

  @OnEvent(WorkspaceEvents.ARCHIVED)
  private async handleWorkspaceArchivedEvent(event: WorkspaceArchivedEvent) {
    this.handleWorkspaceDesiredStateArchived(event.workspace.id)
  }

  @OnEvent(WorkspaceEvents.DESTROYED)
  private async handleWorkspaceDestroyedEvent(event: WorkspaceDestroyedEvent) {
    this.handleWorkspaceDesiredStateDestroyed(event.workspace.id)
  }

  @OnEvent(WorkspaceEvents.STARTED)
  private async handleWorkspaceStartedEvent(event: WorkspaceStartedEvent) {
    this.handleWorkspaceDesiredStateStarted(event.workspace.id)
  }

  @OnEvent(WorkspaceEvents.STOPPED)
  private async handleWorkspaceStoppedEvent(event: WorkspaceStoppedEvent) {
    this.handleWorkspaceDesiredStateStopped(event.workspace.id)
  }

  @OnEvent(WorkspaceEvents.CREATED)
  private async handleWorkspaceCreatedEvent(event: WorkspaceCreatedEvent) {
    this.handleWorkspaceDesiredStateStarted(event.workspace.id)
  }
}
