/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { In, Not, Repository } from 'typeorm'
import { Workspace } from '../entities/workspace.entity'
import { WorkspaceState } from '../enums/workspace-state.enum'
import { NodeApiFactory } from '../runner-api/runnerApi'
import { NodeService } from '../services/node.service'
import { NodeState } from '../enums/node-state.enum'
import { ResourceNotFoundError } from '../../exceptions/not-found.exception'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../constants/workspace.constants'
import { DockerProvider } from '../docker/docker-provider'
import { fromAxiosError } from '../../common/utils/from-axios-error'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { OnEvent } from '@nestjs/event-emitter'
import { WorkspaceEvents } from '../constants/workspace-events.constants'
import { WorkspaceDestroyedEvent } from '../events/workspace-destroyed.event'
import { WorkspaceSnapshotCreatedEvent } from '../events/workspace-snapshot-created.event'
import { WorkspaceArchivedEvent } from '../events/workspace-archived.event'

@Injectable()
export class SnapshotManager {
  private readonly logger = new Logger(SnapshotManager.name)

  constructor(
    @InjectRepository(Workspace)
    private readonly workspaceRepository: Repository<Workspace>,
    private readonly nodeService: NodeService,
    private readonly nodeApiFactory: NodeApiFactory,
    private readonly dockerRegistryService: DockerRegistryService,
    @InjectRedis() private readonly redis: Redis,
    private readonly dockerProvider: DockerProvider,
    private readonly redisLockProvider: RedisLockProvider,
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
            state: In([WorkspaceState.STARTED, WorkspaceState.ARCHIVING]),
            snapshotState: In([SnapshotState.NONE, SnapshotState.COMPLETED]),
          },
          order: {
            lastSnapshotAt: 'ASC',
          },
          //  todo: increase this number when snapshot is stable
          take: 10,
        })

        await Promise.all(
          workspaces
            .filter(
              (workspace) =>
                !workspace.lastSnapshotAt || workspace.lastSnapshotAt < new Date(Date.now() - 1 * 60 * 60 * 1000),
            )
            .map(async (workspace) => {
              const lockKey = `workspace-snapshot-${workspace.id}`
              const hasLock = await this.redisLockProvider.lock(lockKey, 60)
              if (!hasLock) {
                return
              }

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

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'sync-snapshot-states' }) // Run every 10 seconds
  async syncSnapshotStates(): Promise<void> {
    //  lock the sync to only run one instance at a time
    const lockKey = 'sync-snapshot-states'
    const hasLock = await this.redisLockProvider.lock(lockKey, 10)
    if (!hasLock) {
      return
    }

    const workspaces = await this.workspaceRepository.find({
      where: {
        state: In([WorkspaceState.STARTED, WorkspaceState.STOPPED, WorkspaceState.ARCHIVING]),
        snapshotState: In([SnapshotState.PENDING, SnapshotState.IN_PROGRESS]),
      },
    })

    await Promise.all(
      workspaces.map(async (w) => {
        const lockKey = `workspace-snapshot-${w.id}`
        const hasLock = await this.redisLockProvider.lock(lockKey, 60)
        if (!hasLock) {
          return
        }

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
            await this.updateWorkspaceSnapshotState(workspace.id, SnapshotState.ERROR)
          } else {
            await this.redis.setex(errorRetryKey, 300, errorRetryCount + 1)
          }
        }
      }),
    ).catch((ex) => {
      this.logger.error(ex)
    })
  }

  async startSnapshotCreate(workspaceId: string): Promise<void> {
    const workspace = await this.workspaceRepository.findOneByOrFail({
      id: workspaceId,
    })

    if (!workspace) {
      throw new ResourceNotFoundError('Workspace not found')
    }

    // Allow snapshots for STARTED workspaces or STOPPED workspaces with nodeId
    if (
      !(
        workspace.state === WorkspaceState.STARTED ||
        workspace.state === WorkspaceState.ARCHIVING ||
        (workspace.state === WorkspaceState.STOPPED && workspace.nodeId)
      )
    ) {
      throw new BadRequestError('Workspace must be started or stopped with assigned node to create a snapshot')
    }

    if (workspace.snapshotState === SnapshotState.IN_PROGRESS || workspace.snapshotState === SnapshotState.PENDING) {
      return
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
    const existingSnapshotImages = workspace.existingSnapshotImages
    existingSnapshotImages.push({
      imageName: snapshotImage,
      createdAt: new Date(),
    })

    const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
      id: workspace.id,
    })
    workspaceToUpdate.existingSnapshotImages = existingSnapshotImages
    workspaceToUpdate.snapshotState = SnapshotState.PENDING
    workspaceToUpdate.snapshotRegistryId = registry.id
    workspaceToUpdate.snapshotImage = snapshotImage
    await this.workspaceRepository.save(workspaceToUpdate)
  }

  private async checkSnapshotProgress(workspace: Workspace): Promise<void> {
    try {
      const node = await this.nodeService.findOne(workspace.nodeId)
      const nodeWorkspaceApi = this.nodeApiFactory.createWorkspaceApi(node)

      // Get workspace info from node
      const workspaceInfo = await nodeWorkspaceApi.info(workspace.id)

      switch (workspaceInfo.data.snapshotState?.toUpperCase()) {
        case 'COMPLETED': {
          workspace.snapshotState = SnapshotState.COMPLETED
          workspace.lastSnapshotAt = new Date()
          const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
            id: workspace.id,
          })
          workspaceToUpdate.snapshotState = SnapshotState.COMPLETED
          workspaceToUpdate.lastSnapshotAt = new Date()
          await this.workspaceRepository.save(workspaceToUpdate)
          break
        }
        case 'FAILED':
        case 'ERROR': {
          await this.updateWorkspaceSnapshotState(workspace.id, SnapshotState.ERROR)
          break
        }

        // If still in progress or any other state, do nothing and wait for next sync
      }
    } catch (error) {
      await this.updateWorkspaceSnapshotState(workspace.id, SnapshotState.ERROR)
      throw error
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

      await this.updateWorkspaceSnapshotState(workspace.id, SnapshotState.IN_PROGRESS)
    } catch (error) {
      if (
        error.response?.status === 400 &&
        error.response?.data?.message.includes('A snapshot is already in progress')
      ) {
        await this.updateWorkspaceSnapshotState(workspace.id, SnapshotState.IN_PROGRESS)
        return
      }
      await this.updateWorkspaceSnapshotState(workspace.id, SnapshotState.ERROR)
      throw error
    }
  }

  @Cron(CronExpression.EVERY_30_SECONDS, { name: 'sync-stop-state-create-snapshots' }) // Run every 30 seconds
  async syncStopStateCreateSnapshots(): Promise<void> {
    const lockKey = 'sync-stop-state-create-snapshots'
    const hasLock = await this.redisLockProvider.lock(lockKey, 30)
    if (!hasLock) {
      return
    }

    const workspaces = await this.workspaceRepository.find({
      where: {
        state: In([WorkspaceState.STOPPED, WorkspaceState.ARCHIVING]),
        snapshotState: In([SnapshotState.NONE]),
      },
      //  todo: increase this number when auto-stop is stable
      take: 10,
    })

    await Promise.all(
      workspaces
        .filter((workspace) => workspace.nodeId !== null)
        .map(async (workspace) => {
          const lockKey = `workspace-snapshot-${workspace.id}`
          const thisLock = await this.redisLockProvider.lock(lockKey, 30)
          if (!thisLock) {
            return
          }

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

  private async updateWorkspaceSnapshotState(workspaceId: string, snapshotState: SnapshotState): Promise<void> {
    const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
      id: workspaceId,
    })
    workspaceToUpdate.snapshotState = snapshotState
    await this.workspaceRepository.save(workspaceToUpdate)
  }

  @OnEvent(WorkspaceEvents.ARCHIVED)
  private async handleWorkspaceArchivedEvent(event: WorkspaceArchivedEvent) {
    this.startSnapshotCreate(event.workspace.id)
  }

  @OnEvent(WorkspaceEvents.DESTROYED)
  private async handleWorkspaceDestroyedEvent(event: WorkspaceDestroyedEvent) {
    this.deleteSandboxSnapshotRepositoryFromRegistry(event.workspace)
  }

  @OnEvent(WorkspaceEvents.SNAPSHOT_CREATED)
  private async handleWorkspaceSnapshotCreatedEvent(event: WorkspaceSnapshotCreatedEvent) {
    this.handlePendingSnapshot(event.workspace)
  }
}
