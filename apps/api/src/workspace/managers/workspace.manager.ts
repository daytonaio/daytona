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
import { RERUN_SYNC_TASK, RunnerSandboxAdapterFactory } from '../runner-adapter/runnerSandboxAdapter'
import { NodeService } from '../services/node.service'
import { NodeState } from '../enums/node-state.enum'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../constants/workspace.constants'
import { fromAxiosError } from '../../common/utils/from-axios-error'
import { OnEvent } from '@nestjs/event-emitter'
import { WorkspaceEvents } from '../constants/workspace-events.constants'
import { WorkspaceStoppedEvent } from '../events/workspace-stopped.event'
import { WorkspaceStartedEvent } from '../events/workspace-started.event'
import { WorkspaceArchivedEvent } from '../events/workspace-archived.event'
import { WorkspaceDestroyedEvent } from '../events/workspace-destroyed.event'
import { WorkspaceCreatedEvent } from '../events/workspace-create.event'
import { ImageNode } from '../entities/image-node.entity'

const SYNC_INSTANCE_STATE_LOCK_KEY = 'sync-instance-state-'

@Injectable()
export class WorkspaceManager {
  private readonly logger = new Logger(WorkspaceManager.name)

  constructor(
    @InjectRepository(Workspace)
    private readonly workspaceRepository: Repository<Workspace>,
    @InjectRepository(ImageNode)
    private readonly nodeService: NodeService,
    private readonly runnerSandboxAdapterFactory: RunnerSandboxAdapterFactory,
    @InjectRedis() private readonly redis: Redis,
    private readonly redisLockProvider: RedisLockProvider,
  ) {}

  private async syncInstanceState(workspaceId: string) {
    const lockKey = SYNC_INSTANCE_STATE_LOCK_KEY + workspaceId
    const acquired = await this.redisLockProvider.lock(lockKey, 360)
    if (!acquired) {
      return
    }

    const workspace = await this.workspaceRepository.findOneByOrFail({
      id: workspaceId,
    })

    //  NOTE: this should be revisited if it's not needed
    if (workspace.state === WorkspaceState.ERROR) {
      return
    }

    const node = await this.nodeService.findOne(workspace.nodeId)
    const runnerSandboxAdapter = await this.runnerSandboxAdapterFactory.create(node)

    try {
      const result = await runnerSandboxAdapter.syncInstanceState(workspace)
      await this.redisLockProvider.unlock(lockKey)
      if (result === RERUN_SYNC_TASK) {
        await this.syncInstanceState(workspaceId)
      }
    } catch (error) {
      //  TODO: legacy error handling / retry logic
      //        should be revisited

      if (error.code === 'ECONNRESET') {
        await this.redisLockProvider.unlock(lockKey)
        await this.syncInstanceState(workspaceId)
      } else {
        this.logger.error(`Error processing desired state for workspace ${workspaceId}:`, String(error))
        const errorReason = error.message || String(error)

        const workspace = await this.workspaceRepository.findOneBy({
          id: workspaceId,
        })
        if (!workspace) {
          //  edge case where workspace is deleted while desired state is being processed
          await this.redisLockProvider.unlock(lockKey)
          return
        }
        workspace.state = WorkspaceState.ERROR
        if (errorReason !== undefined) {
          workspace.errorReason = errorReason
        }
        await this.workspaceRepository.save(workspace)
        await this.redisLockProvider.unlock(lockKey)
      }
    }
  }

  @Cron(CronExpression.EVERY_MINUTE, { name: 'auto-stop-check' })
  private async autostopCheck(): Promise<void> {
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
            const acquired = await this.redisLockProvider.lock(lockKey, 30)
            if (!acquired) {
              return
            }

            try {
              workspace.desiredState = WorkspaceDesiredState.STOPPED
              await this.workspaceRepository.save(workspace)
              await this.redisLockProvider.unlock(lockKey)
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
  private async syncStates(): Promise<void> {
    const lockKey = 'sync-states'
    if (!(await this.redisLockProvider.lock(lockKey, 30))) {
      return
    }

    const workspaces = await this.workspaceRepository.find({
      where: {
        state: Not(In([WorkspaceState.DESTROYED, WorkspaceState.ERROR])),
        desiredState: Raw(
          () =>
            `"Workspace"."desiredState"::text != "Workspace"."state"::text AND "Workspace"."desiredState"::text != 'archived'`,
        ),
      },
      take: 100,
      order: {
        lastActivityAt: 'DESC',
      },
    })

    await Promise.all(
      workspaces.map(async (workspace) => {
        this.syncInstanceState(workspace.id)
      }),
    )
    await this.redisLockProvider.unlock(lockKey)
  }

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'sync-archived-desired-states' })
  private async syncArchivedDesiredStates(): Promise<void> {
    const lockKey = 'sync-archived-desired-states'
    if (!(await this.redisLockProvider.lock(lockKey, 30))) {
      return
    }

    const nodesWith3InProgress = await this.workspaceRepository
      .createQueryBuilder('workspace')
      .select('"nodeId"')
      .where('"workspace"."state" = :state', { state: WorkspaceState.ARCHIVING })
      .groupBy('"nodeId"')
      .having('COUNT(*) >= 3')
      .getRawMany()

    const workspaces = await this.workspaceRepository.find({
      where: [
        {
          state: WorkspaceState.ARCHIVING,
          desiredState: WorkspaceDesiredState.ARCHIVED,
        },
        {
          state: Not(In([WorkspaceState.ARCHIVED, WorkspaceState.DESTROYED, WorkspaceState.ERROR])),
          desiredState: WorkspaceDesiredState.ARCHIVED,
          nodeId: Not(In(nodesWith3InProgress.map((node) => node.nodeId))),
        },
      ],
      take: 100,
      order: {
        lastActivityAt: 'DESC',
      },
    })

    await Promise.all(
      workspaces.map(async (workspace) => {
        this.syncInstanceState(workspace.id)
      }),
    )
    await this.redisLockProvider.unlock(lockKey)
  }

  @OnEvent(WorkspaceEvents.ARCHIVED)
  private async handleWorkspaceArchivedEvent(event: WorkspaceArchivedEvent) {
    this.syncInstanceState(event.workspace.id).catch(this.logger.error)
  }

  @OnEvent(WorkspaceEvents.DESTROYED)
  private async handleWorkspaceDestroyedEvent(event: WorkspaceDestroyedEvent) {
    this.syncInstanceState(event.workspace.id).catch(this.logger.error)
  }

  @OnEvent(WorkspaceEvents.STARTED)
  private async handleWorkspaceStartedEvent(event: WorkspaceStartedEvent) {
    this.syncInstanceState(event.workspace.id).catch(this.logger.error)
  }

  @OnEvent(WorkspaceEvents.STOPPED)
  private async handleWorkspaceStoppedEvent(event: WorkspaceStoppedEvent) {
    this.syncInstanceState(event.workspace.id).catch(this.logger.error)
  }

  @OnEvent(WorkspaceEvents.CREATED)
  private async handleWorkspaceCreatedEvent(event: WorkspaceCreatedEvent) {
    this.syncInstanceState(event.workspace.id).catch(this.logger.error)
  }
}
