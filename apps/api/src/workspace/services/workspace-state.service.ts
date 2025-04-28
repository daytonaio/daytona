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
import { NodeService } from './node.service'
import { EnumsSandboxState as NodeWorkspaceState } from '@daytonaio/runner-api-client'
import { NodeState } from '../enums/node-state.enum'
import { ResourceNotFoundError } from '../../exceptions/not-found.exception'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { ImageService } from './image.service'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../constants/workspace.constants'
import { DockerProvider } from '../docker/docker-provider'
import { OrganizationService } from '../../organization/services/organization.service'
import { ImageNodeState } from '../enums/image-node-state.enum'
import { BuildInfo } from '../entities/build-info.entity'
import { CreateSandboxDTO } from '@daytonaio/runner-api-client'
import { fromAxiosError } from '../../common/utils/from-axios-error'

type BreakFromSwitch = boolean
const SYNC_INSTANCE_STATE_LOCK_KEY = 'sync-instance-state-'

@Injectable()
export class WorkspaceStateService {
  private readonly logger = new Logger(WorkspaceStateService.name)

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

  //  on init
  async onApplicationBootstrap() {
    await this.adHocSnapshotCheck()
  }

  //  todo: make frequency configurable or more efficient
  @Cron(CronExpression.EVERY_5_MINUTES, { name: 'ad-hoc-snapshot-check' })
  async adHocSnapshotCheck(): Promise<void> {
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
            snapshotState: In([SnapshotState.NONE, SnapshotState.COMPLETED]),
          },
          order: {
            lastSnapshotAt: 'ASC',
          },
          //  todo: increase this number when snapshot is stable
          take: 1,
        })

        await Promise.all(
          workspaces
            .filter(
              (workspace) =>
                !workspace.lastSnapshotAt || workspace.lastSnapshotAt < new Date(Date.now() - 1 * 60 * 60 * 1000),
            )
            .map(async (workspace) => {
              const lockKey = `workspace-snapshot-${workspace.id}`
              const hasLock = await this.redis.get(lockKey)
              if (hasLock) {
                return // Another instance is processing this workspace
              }
              //  sleep for 100ms to avoid race condition
              await new Promise((resolve) => setTimeout(resolve, 100))
              const hasLock2 = await this.redis.get(lockKey)
              if (hasLock2) {
                return
              }
              await this.redis.setex(lockKey, 30, '1')

              try {
                //  todo: remove the catch handler asap
                await this.startSnapshotCreate(workspace.id).catch((error) => {
                  if (error instanceof BadRequestError && error.message === 'A snapshot is already in progress') {
                    return
                  }
                  this.logger.error(`Failed to create snapshot for workspace ${workspace.id}:`, fromAxiosError(error))
                })
              } catch (error) {
                this.logger.error(`Error processing stop state for workspace ${workspace.id}:`, fromAxiosError(error))
              }
            }),
        )
      }),
    )
  }

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
          take: 1,
        })

        await Promise.all(
          workspaces.map(async (workspace) => {
            const lockKey = `workspace-autostop-${workspace.id}`
            const hasLock = await this.redis.get(lockKey)
            if (hasLock) {
              return // Another instance is processing this workspace
            }
            await this.redis.setex(lockKey, 30, '1')

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

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'sync-snapshot-states' }) // Run every 10 seconds
  async syncSnapshotStates(): Promise<void> {
    //  lock the sync to only run one instance at a time
    const lockKey = 'sync-snapshot-states'
    const hasLock = await this.redis.get(lockKey)
    if (hasLock) {
      return
    }
    await this.redis.setex(lockKey, 10, '1')

    const workspaces = await this.workspaceRepository.find({
      where: {
        state: In([WorkspaceState.STARTED, WorkspaceState.STOPPED]),
        snapshotState: In([SnapshotState.PENDING, SnapshotState.IN_PROGRESS]),
      },
    })

    await Promise.all(
      workspaces.map(async (w) => {
        const lockKey = `workspace-snapshot-${w.id}`
        const hasLock = await this.redis.get(lockKey)
        if (hasLock) {
          return
        }
        await new Promise((resolve) => setTimeout(resolve, 100))
        const hasLock2 = await this.redis.get(lockKey)
        if (hasLock2) {
          return
        }
        await this.redis.setex(lockKey, 60, '1')

        const node = await this.nodeService.findOne(w.nodeId)
        if (node.state !== NodeState.READY) {
          return
        }

        //  get the latest workspace state
        const workspace = await this.workspaceRepository.findOneByOrFail({
          id: w.id,
        })

        try {
          switch (workspace.snapshotState) {
            case SnapshotState.PENDING: {
              await this.handlePendingSnapshot(workspace)
              break
            }
            case SnapshotState.IN_PROGRESS: {
              await this.checkSnapshotProgress(workspace)
              break
            }
          }
        } catch (error) {
          this.logger.error(`Error processing snapshot for workspace ${workspace.id}:`, fromAxiosError(error))

          //  if error, retry 10 times
          const errorRetryKey = `${lockKey}-error-retry`
          const errorRetryCount = await this.redis.get(errorRetryKey)
          if (!errorRetryCount) {
            await this.redis.setex(errorRetryKey, 300, '1')
          } else if (parseInt(errorRetryCount) > 10) {
            workspace.snapshotState = SnapshotState.ERROR
            await this.workspaceRepository.save(workspace)
          } else {
            await this.redis.setex(errorRetryKey, 300, errorRetryCount + 1)
          }
        }
      }),
    ).catch((ex) => {
      this.logger.error(ex)
    })
  }

  @Cron(CronExpression.EVERY_30_SECONDS, { name: 'sync-stop-state-2' }) // Run every 30 seconds
  async syncStopState(): Promise<void> {
    const lockKey = 'sync-stop-state-2'
    const hasLock = await this.redis.get(lockKey)
    if (hasLock) {
      return
    }
    await this.redis.setex(lockKey, 30, '1')

    const workspaces = await this.workspaceRepository.find({
      where: {
        state: In([WorkspaceState.STOPPED, WorkspaceState.ARCHIVING]),
        snapshotState: In([SnapshotState.NONE]),
      },
      //  todo: increase this number when auto-stop is stable
      take: 5,
    })

    await Promise.all(
      workspaces
        .filter((workspace) => workspace.nodeId !== null)
        .map(async (workspace) => {
          const lockKey = `workspace-snapshot-${workspace.id}`
          const hasLock = await this.redis.get(lockKey)
          if (hasLock) {
            return // Another instance is processing this workspace
          }
          await new Promise((resolve) => setTimeout(resolve, 100))
          const hasLock2 = await this.redis.get(lockKey)
          if (hasLock2) {
            return
          }
          this.redis.setex(lockKey, 30, 1)

          const node = await this.nodeService.findOne(workspace.nodeId)
          if (node.state !== NodeState.READY) {
            return
          }

          //  TODO: this should be revisited
          //  an error should be handled better and not just logged
          try {
            //  todo: remove the catch handler asap
            await this.startSnapshotCreate(workspace.id).catch((error) => {
              if (error instanceof BadRequestError && error.message === 'A snapshot is already in progress') {
                return
              }
              this.logger.error(`Failed to create snapshot for workspace ${workspace.id}:`, fromAxiosError(error))
            })
          } catch (error) {
            this.logger.error(`Failed to create snapshot for workspace ${workspace.id}:`, fromAxiosError(error))
          }
        }),
    )
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
        const hasLock = await this.redis.get(lockKey)
        if (hasLock) {
          return
        }
        await this.redis.setex(lockKey, 30, '1')

        workspace.desiredState = WorkspaceDesiredState.STOPPED
        try {
          await this.workspaceRepository.save(workspace)
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
        case WorkspaceDesiredState.RESIZED: {
          await this.handleWorkspaceDesiredStateResized(workspace.id)
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
      workspace.state = WorkspaceState.ERROR
      workspace.errorReason = error.message || String(error)
      await this.workspaceRepository.save(workspace)
    }

    //  unlock the workspace after 10 seconds
    //  this will allow the syncState cron to run again, but will allow
    //  the syncInstanceState to complete any pending state changes
    await this.redis.setex(lockKey, 10, '1')
  }

  async startSnapshotCreate(workspaceId: string): Promise<Workspace> {
    const workspace = await this.workspaceRepository.findOneByOrFail({
      id: workspaceId,
    })

    if (!workspace) {
      throw new ResourceNotFoundError('Workspace not found')
    }

    // Allow snapshots for STARTED workspaces or STOPPED workspaces with nodeId
    if (
      !(workspace.state === WorkspaceState.STARTED || (workspace.state === WorkspaceState.STOPPED && workspace.nodeId))
    ) {
      throw new BadRequestError('Workspace must be started or stopped with assigned node to create a snapshot')
    }

    if (workspace.snapshotState === SnapshotState.IN_PROGRESS || workspace.snapshotState === SnapshotState.PENDING) {
      throw new BadRequestError('A snapshot is already in progress')
    }

    // Get default registry
    const registry = await this.dockerRegistryService.getDefaultInternalRegistry()
    if (!registry) {
      throw new BadRequestError('No default registry configured')
    }

    // Generate snapshot image name
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-')
    const snapshotImage = `${registry.url}/${registry.project}/snapshot-${workspace.id}:${timestamp}`

    //  if workspace has a snapshot image, add it to the existingSnapshotImages array
    if (
      workspace.lastSnapshotAt &&
      workspace.snapshotImage &&
      [SnapshotState.NONE, SnapshotState.COMPLETED].includes(workspace.snapshotState)
    ) {
      workspace.existingSnapshotImages.push({
        imageName: workspace.snapshotImage,
        createdAt: workspace.lastSnapshotAt,
      })
    }
    workspace.existingSnapshotImages.push({
      imageName: snapshotImage,
      createdAt: new Date(),
    })
    workspace.snapshotState = SnapshotState.PENDING
    workspace.snapshotRegistryId = registry.id
    workspace.snapshotImage = snapshotImage
    await this.workspaceRepository.save(workspace)

    return workspace
  }

  private async checkSnapshotProgress(workspace: Workspace): Promise<void> {
    try {
      const node = await this.nodeService.findOne(workspace.nodeId)
      const nodeWorkspaceApi = this.nodeApiFactory.createWorkspaceApi(node)

      // Get workspace info from node
      const workspaceInfo = await nodeWorkspaceApi.info(workspace.id)

      switch (workspaceInfo.data.snapshotState?.toUpperCase()) {
        case 'COMPLETED':
          workspace.snapshotState = SnapshotState.COMPLETED
          workspace.lastSnapshotAt = new Date()
          await this.workspaceRepository.save(workspace)
          break

        case 'FAILED':
        case 'ERROR':
          workspace.snapshotState = SnapshotState.ERROR
          await this.workspaceRepository.save(workspace)
          break

        // If still in progress or any other state, do nothing and wait for next sync
      }
    } catch (error) {
      workspace.snapshotState = SnapshotState.ERROR
      await this.workspaceRepository.save(workspace)
      throw error
    }
  }

  private async handleUnassignedBuildWorkspace(workspace: Workspace): Promise<void> {
    // Try to assign an available node with the image build
    let node: string
    try {
      node = await this.nodeService.getRandomAvailableNode(
        workspace.region,
        workspace.class,
        workspace.buildInfo.imageRef,
      )
    } catch (error) {
      // Continue to next assignment method
    }

    if (node) {
      workspace.nodeId = node
      workspace.state = WorkspaceState.UNKNOWN

      await this.workspaceRepository.save(workspace)
      this.syncInstanceState(workspace.id)
      return
    }

    // Try to assign an available node that is currently building the image
    const imageNodes = await this.nodeService.getImageNodes(workspace.buildInfo.imageRef)

    for (const imageNode of imageNodes) {
      const node = await this.nodeService.findOne(imageNode.nodeId)
      if (node.used < node.capacity) {
        if (imageNode.state === ImageNodeState.BUILDING_IMAGE) {
          workspace.nodeId = node.id
          workspace.state = WorkspaceState.BUILDING_IMAGE
          await this.workspaceRepository.save(workspace)
          return
        } else if (imageNode.state === ImageNodeState.ERROR) {
          workspace.state = WorkspaceState.ERROR
          workspace.errorReason = imageNode.errorReason
          await this.workspaceRepository.save(workspace)
          return
        }
      }
    }

    // Try to assign a new available node
    const nodeId = await this.nodeService.getRandomAvailableNode(workspace.region, workspace.class)

    workspace.nodeId = nodeId
    workspace.state = WorkspaceState.BUILDING_IMAGE

    this.buildOnNode(workspace.buildInfo, nodeId, workspace.organizationId)

    await this.workspaceRepository.save(workspace)
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
    switch (workspace.state) {
      case WorkspaceState.STOPPED: {
        //  if snapshot process hasn't started yet, start one
        if (workspace.snapshotState === SnapshotState.NONE) {
          //  TODO: this should be revisited.
          //  an error should be handled better and not just logged
          await this.startSnapshotCreate(workspace.id).catch((error) => {
            this.logger.error(`Failed to create snapshot for workspace ${workspace.id}:`, fromAxiosError(error))
          })
        }

        //  this should not happen
        if (workspace.snapshotState !== SnapshotState.COMPLETED) {
          const updateWorkspace = await this.workspaceRepository.findOneByOrFail({
            id: workspaceId,
          })
          updateWorkspace.state = WorkspaceState.ERROR
          updateWorkspace.errorReason = 'Can not archive sandbox if snapshot state is not completed'
          await this.workspaceRepository.save(updateWorkspace)
          break
        }

        //  check if the snapshot image exists in the snapshot registry
        const registry = await this.dockerRegistryService.findOne(workspace.snapshotRegistryId)
        if (!registry) {
          throw new Error('Registry not found')
        }

        let exists = false
        try {
          exists = await this.dockerProvider.checkImageExistsInRegistry(workspace.snapshotImage, registry)
        } catch (error) {
          this.logger.error(
            `Failed to check if snapshot image ${workspace.snapshotImage} exists in registry ${registry.id}:`,
            fromAxiosError(error),
          )
        }
        //  if the snapshot image does not exist in the registry, create a new snapshot
        if (!exists) {
          this.logger.error(`Snapshot image ${workspace.snapshotImage} does not exist in registry ${registry.id}`)

          const updateWorkspace = await this.workspaceRepository.findOneByOrFail({
            id: workspaceId,
          })
          //  revert workspace to stopped state and abort archive
          updateWorkspace.desiredState = WorkspaceDesiredState.STOPPED
          updateWorkspace.snapshotState = SnapshotState.NONE
          await this.workspaceRepository.save(updateWorkspace)

          await this.startSnapshotCreate(workspace.id).catch((error) => {
            this.logger.error(`Failed to create snapshot for workspace ${workspace.id}:`, fromAxiosError(error))
          })
          return
        }

        workspace.state = WorkspaceState.ARCHIVING
        await this.workspaceRepository.save(workspace)
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
        const workspaceInfoResponse = await nodeWorkspaceApi.info(workspace.id)
        const workspaceInfo = workspaceInfoResponse.data
        if (workspaceInfo.state === NodeWorkspaceState.SandboxStateDestroying) {
          //  wait until workspace is destroyed on node
          await this.redis.del(workspace.id)
          this.syncInstanceState(workspace.id)
          break
        }
        if (workspaceInfo.state !== NodeWorkspaceState.SandboxStateDestroyed) {
          try {
            await nodeWorkspaceApi.destroy(workspace.id)
          } catch (error) {
            //  if the workspace is already destroyed, do nothing
            if (
              !(
                (error.response?.data?.statusCode === 400 &&
                  error.response?.data?.message.includes('Workspace already destroyed')) ||
                error.response?.status === 404
              )
            ) {
              throw error
            }
          }
          //  wait until workspace is destroyed on node
          await this.redis.del(workspace.id)
          this.syncInstanceState(workspace.id)
          break
        }
        await nodeWorkspaceApi.removeDestroyed(workspace.id)

        //  unset the current nodeId
        workspace.nodeId = null
        workspace.state = WorkspaceState.ARCHIVED
        await this.workspaceRepository.save(workspace)
        //  sync states again immediately for workspace
        await this.redis.del(workspace.id)
        this.syncInstanceState(workspace.id)
        break
      }
    }
  }

  private async deleteSandboxSnapshotRepositoryFromRegistry(workspace: Workspace): Promise<void> {
    const registry = await this.dockerRegistryService.findOne(workspace.snapshotRegistryId)

    try {
      await this.dockerProvider.deleteSandboxRepository(workspace.id, registry)
    } catch (error) {
      this.logger.error(
        `Failed to delete snapshot repository ${workspace.id} from registry ${registry.id}:`,
        fromAxiosError(error),
      )
    }
  }

  private async handleWorkspaceDesiredStateDestroyed(workspaceId: string): Promise<void> {
    const workspace = await this.workspaceRepository.findOneByOrFail({
      id: workspaceId,
    })

    if (workspace.state === WorkspaceState.ARCHIVED) {
      await this.deleteSandboxSnapshotRepositoryFromRegistry(workspace)
      workspace.state = WorkspaceState.DESTROYED
      await this.workspaceRepository.save(workspace)
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

        //  delete snapshot images from registry
        await this.deleteSandboxSnapshotRepositoryFromRegistry(workspace)

        workspace.state = WorkspaceState.DESTROYED

        await this.workspaceRepository.save(workspace)
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
        workspace.state = WorkspaceState.DESTROYING
        await this.workspaceRepository.save(workspace)
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
          workspace.state = WorkspaceState.STARTED
          workspace.snapshotState = SnapshotState.NONE
          await this.workspaceRepository.save(workspace)
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
        workspace.state = WorkspaceState.STOPPING
        await this.workspaceRepository.save(workspace)
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
          case NodeWorkspaceState.SandboxStateStopped:
            workspace.state = WorkspaceState.STOPPED
            workspace.snapshotState = SnapshotState.NONE
            await this.workspaceRepository.save(workspace)
            break
          case NodeWorkspaceState.SandboxStateError:
            workspace.state = WorkspaceState.ERROR
            // workspace.errorReason = workspaceInfo.errorReason
            await this.workspaceRepository.save(workspace)
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
          workspace.state = WorkspaceState.STOPPED
          await this.workspaceRepository.save(workspace)
        }
        break
      }
    }
  }

  private async handlePendingSnapshot(workspace: Workspace): Promise<void> {
    try {
      const registry = await this.dockerRegistryService.findOne(workspace.snapshotRegistryId)
      if (!registry) {
        throw new Error('Registry not found')
      }

      const node = await this.nodeService.findOne(workspace.nodeId)
      const nodeWorkspaceApi = this.nodeApiFactory.createWorkspaceApi(node)

      //  check if snapshot is already in progress on the node
      const nodeWorkspaceResponse = await nodeWorkspaceApi.info(workspace.id)
      const nodeWorkspace = nodeWorkspaceResponse.data
      if (nodeWorkspace.snapshotState?.toUpperCase() === 'IN_PROGRESS') {
        return
      }

      // Initiate snapshot on node
      await nodeWorkspaceApi.createSnapshot(workspace.id, {
        registry: {
          url: registry.url,
          username: registry.username,
          password: registry.password,
        },
        image: workspace.snapshotImage,
      })

      workspace.snapshotState = SnapshotState.IN_PROGRESS
      await this.workspaceRepository.save(workspace)
    } catch (error) {
      if (
        error.response?.status === 400 &&
        error.response?.data?.message.includes('A snapshot is already in progress')
      ) {
        workspace.snapshotState = SnapshotState.IN_PROGRESS
        await this.workspaceRepository.save(workspace)
        return
      }
      workspace.snapshotState = SnapshotState.ERROR
      await this.workspaceRepository.save(workspace)
      throw error
    }
  }

  private async handleNodeWorkspaceBuildingImageStateOnDesiredStateStart(workspace: Workspace) {
    const imageNode = await this.nodeService.getImageNode(workspace.nodeId, workspace.buildInfo.imageRef)
    if (imageNode) {
      switch (imageNode.state) {
        case ImageNodeState.READY: {
          // TODO: "UNKNOWN" should probably be changed to something else
          workspace.state = WorkspaceState.UNKNOWN
          await this.workspaceRepository.save(workspace)
          this.syncInstanceState(workspace.id)
          return
        }
        case ImageNodeState.ERROR: {
          workspace.state = WorkspaceState.ERROR
          workspace.errorReason = imageNode.errorReason
          await this.workspaceRepository.save(workspace)
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
    workspace.state = WorkspaceState.CREATING
    await this.workspaceRepository.save(workspace)
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
          await this.workspaceRepository.save(workspace)
        }
      }
    }

    if (workspace.nodeId === null) {
      //  if workspace has no node, check if snapshot is completed
      //  if not, set workspace to error
      //  if snapshot is completed, get random available node and start workspace
      //  use the snapshot image to start the workspace

      if (workspace.snapshotState !== SnapshotState.COMPLETED) {
        workspace.state = WorkspaceState.ERROR
        workspace.errorReason = 'Workspace has no node and snapshot is not completed'
        await this.workspaceRepository.save(workspace)
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
        workspace.state = WorkspaceState.ERROR
        workspace.errorReason = 'No valid snapshot image found'
        await this.workspaceRepository.save(workspace)
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

      workspace.nodeId = nodeId
      workspace.state = WorkspaceState.RESTORING
      await this.workspaceRepository.save(workspace)
    } else {
      // if workspace has node, start workspace
      const node = await this.nodeService.findOne(workspace.nodeId)

      const nodeWorkspaceApi = this.nodeApiFactory.createWorkspaceApi(node)

      await nodeWorkspaceApi.start(workspace.id)

      workspace.state = WorkspaceState.STARTING
      await this.workspaceRepository.save(workspace)
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
      workspace.state = WorkspaceState.PULLING_IMAGE
      await this.workspaceRepository.save(workspace)

      await this.redis.del(workspace.id)
      this.syncInstanceState(workspace.id)
      return true
    }
    if (workspaceInfo.state === NodeWorkspaceState.SandboxStateError) {
      workspace.state = WorkspaceState.ERROR
      // workspace.errorReason = workspaceInfo.errorReason
      await this.workspaceRepository.save(workspace)
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
        workspace.state = WorkspaceState.STARTED
        //  if previous snapshot state is error or completed, set snapshot state to none
        if ([SnapshotState.ERROR, SnapshotState.COMPLETED].includes(workspace.snapshotState)) {
          workspace.snapshotState = SnapshotState.NONE
        }
        await this.workspaceRepository.save(workspace)

        //  if workspace was transferred to a new node, remove it from the old node
        if (workspace.prevNodeId) {
          const node = await this.nodeService.findOne(workspace.prevNodeId)
          if (!node) {
            this.logger.warn(`Previously assigned node ${workspace.prevNodeId} for workspace ${workspace.id} not found`)
            //  clear prevNodeId to avoid trying to cleanup on a non-existent node
            workspace.prevNodeId = null
            await this.workspaceRepository.save(workspace)
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
            await this.workspaceRepository.save(workspace)
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
        workspace.state = WorkspaceState.ERROR
        // workspace.errorReason = workspaceInfo.errorReason
        await this.workspaceRepository.save(workspace)
        break
      }
    }
    //  sync states again immediately for workspace
    await this.redis.del(workspace.id)
    this.syncInstanceState(workspace.id)
  }

  private async handleWorkspaceDesiredStateResized(workspaceId: string): Promise<void> {
    const workspace = await this.workspaceRepository.findOneByOrFail({
      id: workspaceId,
    })

    const node = await this.nodeService.findOne(workspace.nodeId)
    if (node.state !== NodeState.READY) {
      return
    }

    const nodeWorkspaceApi = this.nodeApiFactory.createWorkspaceApi(node)

    switch (workspace.state) {
      case WorkspaceState.RESIZING: {
        const workspaceInfoResponse = await nodeWorkspaceApi.info(workspace.id)
        const workspaceInfo = workspaceInfoResponse.data
        if (workspaceInfo.state === NodeWorkspaceState.SandboxStateStarted) {
          workspace.state = WorkspaceState.STARTED
          await this.workspaceRepository.save(workspace)
        }
        if (workspaceInfo.state === NodeWorkspaceState.SandboxStateStopped) {
          workspace.state = WorkspaceState.STOPPED
          await this.workspaceRepository.save(workspace)
        }
        if (workspaceInfo.state === NodeWorkspaceState.SandboxStateError) {
          workspace.state = WorkspaceState.ERROR
          // workspace.errorReason = workspaceInfo.errorReason
          await this.workspaceRepository.save(workspace)
        }
        break
      }
      case WorkspaceState.STOPPED:
      case WorkspaceState.STARTED: {
        try {
          // Update the workspace resources on the node
          await nodeWorkspaceApi.resize(workspace.id, {
            // TODO: Important - check
            cpu: workspace.cpu,
            gpu: workspace.gpu,
            memory: workspace.mem,
          })

          // Set the state back to the previous state since resize is complete
          workspace.state = WorkspaceState.RESIZING
          await this.workspaceRepository.save(workspace)
          this.syncInstanceState(workspace.id)
        } catch (error) {
          workspace.state = WorkspaceState.ERROR
          workspace.errorReason = `Failed to resize workspace: ${error.message}`
          await this.workspaceRepository.save(workspace)
          throw error
        }
        break
      }
    }
  }
}
