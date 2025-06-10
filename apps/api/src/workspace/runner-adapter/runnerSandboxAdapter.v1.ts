/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { Node } from '../entities/node.entity'
import { WorkspaceDesiredState } from '../enums/workspace-desired-state.enum'
import { WorkspaceState } from '../enums/workspace-state.enum'
import { Workspace } from '../entities/workspace.entity'
import { ImageNodeState } from '../enums/image-node-state.enum'
import { BuildInfo } from '../entities/build-info.entity'
import { In, Repository } from 'typeorm'
import { NodeState } from '../enums/node-state.enum'
import { CreateSandboxDTO, EnumsSandboxState as NodeWorkspaceState } from '@daytonaio/runner-api-client'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { COMPLETE_SYNC_TASK, RERUN_SYNC_TASK, RunnerSandboxAdapter, SyncTaskStatus } from './runnerSandboxAdapter'
import { ImageService } from '../services/image.service'
import { NodeService } from '../services/node.service'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { DockerProvider } from '../docker/docker-provider'
import Redis from 'ioredis'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { SandboxApi, DefaultApi, ImagesApi, Configuration } from '@daytonaio/runner-api-client'
import axios from 'axios'
import axiosDebug from 'axios-debug-log'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { fromAxiosError } from '../../common/utils/from-axios-error'

const isDebugEnabled = process.env.DEBUG === 'true'

if (isDebugEnabled) {
  axiosDebug({
    request: function (debug, config) {
      debug('Request with ' + JSON.stringify(config))
      return config
    },
    response: function (debug, response) {
      debug('Response with ' + response)
      return response
    },
    error: function (debug, error) {
      debug('Error with ' + error)
      return Promise.reject(error)
    },
  })
}

@Injectable()
export class RunnerSandboxAdapterV1 implements RunnerSandboxAdapter {
  private readonly logger = new Logger(RunnerSandboxAdapterV1.name)
  private apiClientNode: DefaultApi
  private apiClientSandbox: SandboxApi
  private apiClientImage: ImagesApi

  constructor(
    private readonly nodeService: NodeService,
    private readonly workspaceRepository: Repository<Workspace>,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly dockerRegistryService: DockerRegistryService,
    @InjectRedis() private readonly redis: Redis,
    private readonly imageService: ImageService,
    private readonly dockerProvider: DockerProvider,
  ) {}

  public async init(node: Node): Promise<void> {
    const axiosInstance = axios.create({
      baseURL: node.apiUrl,
      headers: {
        Authorization: `Bearer ${node.apiKey}`,
      },
      timeout: 1 * 60 * 60 * 1000, // 1 hour
    })

    axiosInstance.interceptors.response.use(
      (response) => {
        return response
      },
      (error) => {
        const errorMessage = error.response?.data?.message || error.response?.data || error.message || String(error)

        throw new Error(String(errorMessage))
      },
    )

    if (isDebugEnabled) {
      axiosDebug.addLogger(axiosInstance)
    }

    this.apiClientNode = new DefaultApi(new Configuration(), '', axiosInstance)
    this.apiClientSandbox = new SandboxApi(new Configuration(), '', axiosInstance)
    this.apiClientImage = new ImagesApi(new Configuration(), '', axiosInstance)
  }

  async syncInstanceState(workspace: Workspace): Promise<SyncTaskStatus> {
    switch (workspace.desiredState) {
      case WorkspaceDesiredState.STARTED: {
        return await this.handleWorkspaceDesiredStateStarted(workspace)
      }
      case WorkspaceDesiredState.STOPPED: {
        return await this.handleWorkspaceDesiredStateStopped(workspace)
      }
      case WorkspaceDesiredState.DESTROYED: {
        return await this.handleWorkspaceDesiredStateDestroyed(workspace)
      }
      case WorkspaceDesiredState.ARCHIVED: {
        return await this.handleWorkspaceDesiredStateArchived(workspace)
      }
    }
  }

  private async handleUnassignedBuildWorkspace(workspace: Workspace): Promise<SyncTaskStatus> {
    // Try to assign an available node with the image build
    let nodeId: string
    try {
      nodeId = await this.nodeService.getRandomAvailableNode({
        region: workspace.region,
        workspaceClass: workspace.class,
        imageRef: workspace.buildInfo.imageRef,
      })
    } catch (error) {
      // Continue to next assignment method
    }

    if (nodeId) {
      await this.updateWorkspaceState(workspace.id, WorkspaceState.UNKNOWN, nodeId)
      return RERUN_SYNC_TASK
    }

    // Try to assign an available node that is currently building the image
    const imageNodes = await this.nodeService.getImageNodes(workspace.buildInfo.imageRef)

    for (const imageNode of imageNodes) {
      const node = await this.nodeService.findOne(imageNode.nodeId)
      if (node.used < node.capacity) {
        if (imageNode.state === ImageNodeState.BUILDING_IMAGE) {
          await this.updateWorkspaceState(workspace.id, WorkspaceState.BUILDING_IMAGE, node.id)
          return RERUN_SYNC_TASK
        } else if (imageNode.state === ImageNodeState.ERROR) {
          await this.updateWorkspaceErrorState(workspace.id, imageNode.errorReason)
          return COMPLETE_SYNC_TASK
        }
      }
    }

    const excludedNodeIds = await this.nodeService.getNodesWithMultipleImagesBuilding()

    // Try to assign a new available node
    nodeId = await this.nodeService.getRandomAvailableNode({
      region: workspace.region,
      workspaceClass: workspace.class,
      excludedNodeIds: excludedNodeIds,
    })

    this.buildOnNode(workspace.buildInfo, nodeId, workspace.organizationId)

    await this.updateWorkspaceState(workspace.id, WorkspaceState.BUILDING_IMAGE, nodeId)
    await this.nodeService.recalculateNodeUsage(nodeId)
    return RERUN_SYNC_TASK
  }

  // Initiates the image build on the runner and creates an ImageNode depending on the result
  private async buildOnNode(buildInfo: BuildInfo, nodeId: string, organizationId: string) {
    let retries = 0

    while (retries < 10) {
      try {
        await this.apiClientImage.buildImage({
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

    const response = (await this.apiClientImage.imageExists(buildInfo.imageRef)).data
    let state = ImageNodeState.BUILDING_IMAGE
    if (response && response.exists) {
      state = ImageNodeState.READY
    }

    await this.nodeService.createImageNode(nodeId, buildInfo.imageRef, state)
  }

  private async handleWorkspaceDesiredStateArchived(workspace: Workspace): Promise<SyncTaskStatus> {
    const lockKey = 'archive-lock-' + workspace.nodeId
    if (!(await this.redisLockProvider.lock(lockKey, 10))) {
      return COMPLETE_SYNC_TASK
    }

    const inProgressOnNode = await this.workspaceRepository.find({
      where: {
        nodeId: workspace.nodeId,
        state: In([WorkspaceState.ARCHIVING]),
      },
      order: {
        lastActivityAt: 'DESC',
      },
      take: 100,
    })

    //  if the workspace is already in progress, continue
    if (!inProgressOnNode.find((w) => w.id === workspace.id)) {
      //  max 3 workspaces can be archived at the same time on the same node
      //  this is to prevent the node from being overloaded
      if (inProgressOnNode.length > 2) {
        await this.redisLockProvider.unlock(lockKey)
        return
      }
    }

    switch (workspace.state) {
      case WorkspaceState.STOPPED: {
        await this.updateWorkspaceState(workspace.id, WorkspaceState.ARCHIVING)
        //  fallthrough to archiving state
      }
      case WorkspaceState.ARCHIVING: {
        await this.redisLockProvider.unlock(lockKey)

        //  if the snapshot state is error, we need to retry the snapshot
        if (workspace.snapshotState === SnapshotState.ERROR) {
          const archiveErrorRetryKey = 'archive-error-retry-' + workspace.id
          const archiveErrorRetryCountRaw = await this.redis.get(archiveErrorRetryKey)
          const archiveErrorRetryCount = archiveErrorRetryCountRaw ? parseInt(archiveErrorRetryCountRaw) : 0
          //  if the archive error retry count is greater than 3, we need to mark the workspace as error
          if (archiveErrorRetryCount > 3) {
            await this.updateWorkspaceErrorState(workspace.id, 'Failed to archive workspace')
            await this.redis.del(archiveErrorRetryKey)
            return COMPLETE_SYNC_TASK
          }
          await this.redis.setex('archive-error-retry-' + workspace.id, 720, String(archiveErrorRetryCount + 1))

          //  reset the snapshot state to pending to retry the snapshot
          await this.workspaceRepository.update(workspace.id, {
            snapshotState: SnapshotState.PENDING,
          })

          return COMPLETE_SYNC_TASK
        }

        // Check for timeout - if more than 30 minutes since last activity
        const thirtyMinutesAgo = new Date(Date.now() - 30 * 60 * 1000)
        if (workspace.lastActivityAt < thirtyMinutesAgo) {
          await this.updateWorkspaceErrorState(workspace.id, 'Archiving operation timed out')
          return COMPLETE_SYNC_TASK
        }

        if (workspace.snapshotState !== SnapshotState.COMPLETED) {
          return COMPLETE_SYNC_TASK
        }

        try {
          const workspaceInfoResponse = await this.apiClientSandbox.info(workspace.id)
          const workspaceInfo = workspaceInfoResponse.data
          switch (workspaceInfo.state) {
            case NodeWorkspaceState.SandboxStateDestroying:
              //  wait until workspace is destroyed on node
              return RERUN_SYNC_TASK
            case NodeWorkspaceState.SandboxStateDestroyed:
              await this.updateWorkspaceState(workspace.id, WorkspaceState.ARCHIVED, null)
              return COMPLETE_SYNC_TASK
            default:
              await this.apiClientSandbox.destroy(workspace.id)
              return RERUN_SYNC_TASK
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
          await this.updateWorkspaceState(workspace.id, WorkspaceState.ARCHIVED, null)
          return COMPLETE_SYNC_TASK
        }
      }
    }

    return COMPLETE_SYNC_TASK
  }

  private async handleWorkspaceDesiredStateDestroyed(workspace: Workspace): Promise<SyncTaskStatus> {
    if (workspace.state === WorkspaceState.ARCHIVED) {
      await this.updateWorkspaceState(workspace.id, WorkspaceState.DESTROYED)
      return COMPLETE_SYNC_TASK
    }

    const node = await this.nodeService.findOne(workspace.nodeId)
    if (node.state !== NodeState.READY) {
      //  console.debug(`Node ${node.id} is not ready`);
      return COMPLETE_SYNC_TASK
    }

    switch (workspace.state) {
      case WorkspaceState.DESTROYED:
        return COMPLETE_SYNC_TASK
      case WorkspaceState.DESTROYING: {
        try {
          const workspaceInfoResponse = await this.apiClientSandbox.info(workspace.id)
          const workspaceInfo = workspaceInfoResponse.data
          if (
            workspaceInfo.state === NodeWorkspaceState.SandboxStateDestroyed ||
            workspaceInfo.state === NodeWorkspaceState.SandboxStateError
          ) {
            await this.apiClientSandbox.removeDestroyed(workspace.id)
          }
        } catch (e) {
          //  if the workspace is not found on node, it is already destroyed
          if (!e.response || e.response.status !== 404) {
            throw e
          }
        }

        await this.updateWorkspaceState(workspace.id, WorkspaceState.DESTROYED)
        return RERUN_SYNC_TASK
      }
      default: {
        // destroy workspace
        try {
          const workspaceInfoResponse = await this.apiClientSandbox.info(workspace.id)
          const workspaceInfo = workspaceInfoResponse.data
          if (workspaceInfo?.state === NodeWorkspaceState.SandboxStateDestroyed) {
            return COMPLETE_SYNC_TASK
          }
          await this.apiClientSandbox.destroy(workspace.id)
        } catch (e) {
          //  if the workspace is not found on node, it is already destroyed
          if (e.response.status !== 404) {
            throw e
          }
        }
        await this.updateWorkspaceState(workspace.id, WorkspaceState.DESTROYING)
        return RERUN_SYNC_TASK
      }
    }
  }

  private async handleWorkspaceDesiredStateStarted(workspace: Workspace): Promise<SyncTaskStatus> {
    switch (workspace.state) {
      case WorkspaceState.PENDING_BUILD: {
        return this.handleUnassignedBuildWorkspace(workspace)
      }
      case WorkspaceState.BUILDING_IMAGE: {
        return this.handleNodeWorkspaceBuildingImageStateOnDesiredStateStart(workspace)
      }
      case WorkspaceState.UNKNOWN: {
        return this.handleNodeWorkspaceUnknownStateOnDesiredStateStart(workspace)
      }
      case WorkspaceState.ARCHIVED:
      case WorkspaceState.STOPPED: {
        return this.handleNodeWorkspaceStoppedOrArchivedStateOnDesiredStateStart(workspace)
      }
      case WorkspaceState.RESTORING:
      case WorkspaceState.CREATING: {
        return this.handleNodeWorkspacePullingImageStateCheck(workspace)
      }
      case WorkspaceState.PULLING_IMAGE:
      case WorkspaceState.STARTING: {
        return this.handleNodeWorkspaceStartedStateCheck(workspace)
      }
      default: {
        throw new Error(`Unsupported workspace state: ${workspace.state}`)
      }
    }
  }

  private async handleWorkspaceDesiredStateStopped(workspace: Workspace): Promise<SyncTaskStatus> {
    const node = await this.nodeService.findOne(workspace.nodeId)
    if (node.state !== NodeState.READY) {
      //  console.debug(`Node ${node.id} is not ready`);
      return COMPLETE_SYNC_TASK
    }

    switch (workspace.state) {
      case WorkspaceState.STARTED: {
        // stop workspace
        await this.apiClientSandbox.stop(workspace.id)
        await this.updateWorkspaceState(workspace.id, WorkspaceState.STOPPING)
        //  sync states again immediately for workspace
        return RERUN_SYNC_TASK
      }
      case WorkspaceState.STOPPING: {
        // check if workspace is stopped
        const workspaceInfoResponse = await this.apiClientSandbox.info(workspace.id)
        const workspaceInfo = workspaceInfoResponse.data
        switch (workspaceInfo.state) {
          case NodeWorkspaceState.SandboxStateStopped: {
            const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
              id: workspace.id,
            })
            workspaceToUpdate.state = WorkspaceState.STOPPED
            workspaceToUpdate.snapshotState = SnapshotState.NONE
            await this.workspaceRepository.save(workspaceToUpdate)
            return RERUN_SYNC_TASK
          }
          case NodeWorkspaceState.SandboxStateError: {
            await this.updateWorkspaceErrorState(workspace.id, 'Sandbox is in error state on runner')
            return COMPLETE_SYNC_TASK
          }
        }
        return RERUN_SYNC_TASK
      }
      default: {
        throw new Error(`Unsupported workspace state: ${workspace.state}`)
      }
    }
  }

  private async handleNodeWorkspaceBuildingImageStateOnDesiredStateStart(
    workspace: Workspace,
  ): Promise<SyncTaskStatus> {
    const imageNode = await this.nodeService.getImageNode(workspace.nodeId, workspace.buildInfo.imageRef)
    if (imageNode) {
      switch (imageNode.state) {
        case ImageNodeState.READY: {
          // TODO: "UNKNOWN" should probably be changed to something else
          await this.updateWorkspaceState(workspace.id, WorkspaceState.UNKNOWN)
          return RERUN_SYNC_TASK
        }
        case ImageNodeState.ERROR: {
          await this.updateWorkspaceErrorState(workspace.id, imageNode.errorReason)
          return COMPLETE_SYNC_TASK
        }
      }
    }
    if (!imageNode || imageNode.state === ImageNodeState.BUILDING_IMAGE) {
      // Sleep for a second and go back to syncing instance state
      await new Promise((resolve) => setTimeout(resolve, 1000))
      return RERUN_SYNC_TASK
    }

    return COMPLETE_SYNC_TASK
  }

  private async handleNodeWorkspaceUnknownStateOnDesiredStateStart(workspace: Workspace): Promise<SyncTaskStatus> {
    const node = await this.nodeService.findOne(workspace.nodeId)
    if (node.state !== NodeState.READY) {
      //  console.debug(`Node ${node.id} is not ready`);
      return COMPLETE_SYNC_TASK
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

    await this.apiClientSandbox.create(createWorkspaceDto)
    await this.updateWorkspaceState(workspace.id, WorkspaceState.CREATING)
    //  sync states again immediately for workspace
    return RERUN_SYNC_TASK
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
  ): Promise<SyncTaskStatus> {
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

      if (workspace.snapshotState === SnapshotState.COMPLETED) {
        const usageThreshold = 35
        const runningWorkspacesCount = await this.workspaceRepository.count({
          where: {
            nodeId: workspace.nodeId,
            state: WorkspaceState.STARTED,
          },
        })
        if (runningWorkspacesCount > usageThreshold) {
          //  TODO: usage should be based on compute usage

          const image = await this.imageService.getImageByName(workspace.image, workspace.organizationId)
          const availableNodes = await this.nodeService.findAvailableNodes({
            region: workspace.region,
            workspaceClass: workspace.class,
            imageRef: image.internalName,
          })
          const lessUsedNodes = availableNodes.filter((node) => node.id !== workspace.nodeId)

          //  temp workaround to move workspaces to less used node
          if (lessUsedNodes.length > 0) {
            await this.workspaceRepository.update(workspace.id, {
              nodeId: null,
              prevNodeId: workspace.nodeId,
            })
            try {
              await this.apiClientSandbox.removeDestroyed(workspace.id)
            } catch (e) {
              this.logger.error(
                `Failed to cleanup workspace ${workspace.id} on previous node ${node.id}:`,
                fromAxiosError(e),
              )
            }
            workspace.prevNodeId = workspace.nodeId
            workspace.nodeId = null
          }
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
        return COMPLETE_SYNC_TASK
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
        return RERUN_SYNC_TASK
      }

      const image = await this.imageService.getImageByName(workspace.image, workspace.organizationId)

      //  exclude the node that the last node workspace was on
      const availableNodes = (
        await this.nodeService.findAvailableNodes({
          region: workspace.region,
          workspaceClass: workspace.class,
          imageRef: image.internalName,
        })
      ).filter((node) => node.id != workspace.prevNodeId)

      //  get random node from available nodes
      const randomNodeIndex = (min: number, max: number) => Math.floor(Math.random() * (max - min + 1) + min)
      const nodeId = availableNodes[randomNodeIndex(0, availableNodes.length - 1)].id

      await this.apiClientSandbox.create({
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
      await this.apiClientSandbox.start(workspace.id)

      await this.updateWorkspaceState(workspace.id, WorkspaceState.STARTING)
      return RERUN_SYNC_TASK
    }

    return RERUN_SYNC_TASK
  }

  //  used to check if workspace is pulling image on node and update workspace state accordingly
  private async handleNodeWorkspacePullingImageStateCheck(workspace: Workspace): Promise<SyncTaskStatus> {
    //  edge case when workspace is being transferred to a new node
    if (!workspace.nodeId) {
      return RERUN_SYNC_TASK
    }

    const workspaceInfoResponse = await this.apiClientSandbox.info(workspace.id)
    const workspaceInfo = workspaceInfoResponse.data

    if (workspaceInfo.state === NodeWorkspaceState.SandboxStatePullingImage) {
      await this.updateWorkspaceState(workspace.id, WorkspaceState.PULLING_IMAGE)
    } else if (workspaceInfo.state === NodeWorkspaceState.SandboxStateError) {
      await this.updateWorkspaceErrorState(workspace.id)
    } else {
      await this.updateWorkspaceState(workspace.id, WorkspaceState.STARTING)
    }

    return RERUN_SYNC_TASK
  }

  //  used to check if workspace is started on node and update workspace state accordingly
  //  also used to handle the case where a workspace is started on a node and then transferred to a new node
  private async handleNodeWorkspaceStartedStateCheck(workspace: Workspace): Promise<SyncTaskStatus> {
    const workspaceInfoResponse = await this.apiClientSandbox.info(workspace.id)
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
          try {
            // First try to destroy the workspace
            await this.apiClientSandbox.destroy(workspace.id)

            // Wait for workspace to be destroyed before removing
            let retries = 0
            while (retries < 10) {
              try {
                const workspaceInfo = await this.apiClientSandbox.info(workspace.id)
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
            await this.apiClientSandbox.removeDestroyed(workspace.id)
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

    return RERUN_SYNC_TASK
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
}
