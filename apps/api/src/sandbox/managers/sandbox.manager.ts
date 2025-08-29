/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, OnApplicationShutdown } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { In, MoreThanOrEqual, Not, Raw, Repository } from 'typeorm'
import { Sandbox } from '../entities/sandbox.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { RunnerService } from '../services/runner.service'
import { RunnerState } from '../enums/runner-state.enum'

import { RedisLockProvider } from '../common/redis-lock.provider'

import { SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../constants/sandbox.constants'

import { OnEvent } from '@nestjs/event-emitter'
import { SandboxEvents } from '../constants/sandbox-events.constants'
import { SandboxStoppedEvent } from '../events/sandbox-stopped.event'
import { SandboxStartedEvent } from '../events/sandbox-started.event'
import { SandboxArchivedEvent } from '../events/sandbox-archived.event'
import { SandboxDestroyedEvent } from '../events/sandbox-destroyed.event'
import { SandboxCreatedEvent } from '../events/sandbox-create.event'

import { OtelSpan } from '../../common/decorators/otel.decorator'

import { SandboxStartAction } from './sandbox-actions/sandbox-start.action'
import { SandboxStopAction } from './sandbox-actions/sandbox-stop.action'
import { SandboxDestroyAction } from './sandbox-actions/sandbox-destroy.action'
import { SandboxArchiveAction } from './sandbox-actions/sandbox-archive.action'
import { SYNC_AGAIN, DONT_SYNC_AGAIN } from './sandbox-actions/sandbox.action'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { TypedConfigService } from '../../config/typed-config.service'

import { TrackJobExecution } from '../../common/decorators/track-job-execution.decorator'
import { TrackableJobExecutions } from '../../common/interfaces/trackable-job-executions'
import { setTimeout } from 'timers/promises'

export const SYNC_INSTANCE_STATE_LOCK_KEY = 'sync-instance-state-'

@Injectable()
export class SandboxManager implements TrackableJobExecutions, OnApplicationShutdown {
  activeJobs = new Set<string>()

  private readonly logger = new Logger(SandboxManager.name)

  constructor(
    @InjectRepository(Sandbox)
    private readonly sandboxRepository: Repository<Sandbox>,
    private readonly runnerService: RunnerService,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly sandboxStartAction: SandboxStartAction,
    private readonly sandboxStopAction: SandboxStopAction,
    private readonly sandboxDestroyAction: SandboxDestroyAction,
    private readonly sandboxArchiveAction: SandboxArchiveAction,
    private readonly eventEmitter: EventEmitter2,
    private readonly configService: TypedConfigService,
  ) {}

  async onApplicationShutdown() {
    //  wait for all active jobs to finish
    while (this.activeJobs.size > 0) {
      this.logger.log(`Waiting for ${this.activeJobs.size} active jobs to finish`)
      await setTimeout(1000)
    }
  }

  @Cron(CronExpression.EVERY_MINUTE, { name: 'auto-stop-check' })
  @TrackJobExecution()
  @OtelSpan()
  async autostopCheck(): Promise<void> {
    const lockKey = 'auto-stop-check-worker-selected'
    //  lock the sync to only run one instance at a time
    if (!(await this.redisLockProvider.lock(lockKey, 60))) {
      return
    }

    try {
      // Get all ready runners
      const allRunners = await this.runnerService.findAll()
      const readyRunners = allRunners.filter((runner) => runner.state === RunnerState.READY)

      // Process all runners in parallel
      await Promise.all(
        readyRunners.map(async (runner) => {
          const sandboxes = await this.sandboxRepository.find({
            where: {
              runnerId: runner.id,
              organizationId: Not(SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION),
              state: SandboxState.STARTED,
              desiredState: SandboxDesiredState.STARTED,
              pending: Not(true),
              autoStopInterval: Not(0),
              lastActivityAt: Raw((alias) => `${alias} < NOW() - INTERVAL '1 minute' * "autoStopInterval"`),
            },
            order: {
              lastBackupAt: 'ASC',
            },
            //  todo: increase this number when auto-stop is stable
            take: 10,
          })

          await Promise.all(
            sandboxes.map(async (sandbox) => {
              const lockKey = SYNC_INSTANCE_STATE_LOCK_KEY + sandbox.id
              const acquired = await this.redisLockProvider.lock(lockKey, 30)
              if (!acquired) {
                return
              }

              try {
                sandbox.pending = true
                //  if auto-delete interval is 0, delete the sandbox immediately
                if (sandbox.autoDeleteInterval === 0) {
                  sandbox.desiredState = SandboxDesiredState.DESTROYED
                } else {
                  sandbox.desiredState = SandboxDesiredState.STOPPED
                }
                await this.sandboxRepository.save(sandbox)
                this.syncInstanceState(sandbox.id)
              } catch (error) {
                this.logger.error(`Error processing auto-stop state for sandbox ${sandbox.id}:`, error)
              } finally {
                await this.redisLockProvider.unlock(lockKey)
              }
            }),
          )
        }),
      )
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @Cron(CronExpression.EVERY_MINUTE, { name: 'auto-archive-check' })
  @TrackJobExecution()
  async autoArchiveCheck(): Promise<void> {
    const lockKey = 'auto-archive-check-worker-selected'
    //  lock the sync to only run one instance at a time
    if (!(await this.redisLockProvider.lock(lockKey, 60))) {
      return
    }

    try {
      const sandboxes = await this.sandboxRepository.find({
        where: {
          organizationId: Not(SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION),
          state: SandboxState.STOPPED,
          desiredState: SandboxDesiredState.STOPPED,
          pending: Not(true),
          lastActivityAt: Raw((alias) => `${alias} < NOW() - INTERVAL '1 minute' * "autoArchiveInterval"`),
        },
        order: {
          lastBackupAt: 'ASC',
        },
        take: 100,
      })

      await Promise.all(
        sandboxes.map(async (sandbox) => {
          const lockKey = SYNC_INSTANCE_STATE_LOCK_KEY + sandbox.id
          const acquired = await this.redisLockProvider.lock(lockKey, 30)
          if (!acquired) {
            return
          }

          try {
            sandbox.desiredState = SandboxDesiredState.ARCHIVED
            await this.sandboxRepository.save(sandbox)
            this.syncInstanceState(sandbox.id)
          } catch (error) {
            this.logger.error(`Error processing auto-archive state for sandbox ${sandbox.id}:`, error)
          } finally {
            await this.redisLockProvider.unlock(lockKey)
          }
        }),
      )
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @Cron(CronExpression.EVERY_MINUTE, { name: 'auto-delete-check' })
  @TrackJobExecution()
  async autoDeleteCheck(): Promise<void> {
    const lockKey = 'auto-delete-check-worker-selected'
    //  lock the sync to only run one instance at a time
    if (!(await this.redisLockProvider.lock(lockKey, 60))) {
      return
    }

    try {
      // Get all ready runners
      const allRunners = await this.runnerService.findAll()
      const readyRunners = allRunners.filter((runner) => runner.state === RunnerState.READY)

      // Process all runners in parallel
      await Promise.all(
        readyRunners.map(async (runner) => {
          const sandboxes = await this.sandboxRepository.find({
            where: {
              runnerId: runner.id,
              organizationId: Not(SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION),
              state: SandboxState.STOPPED,
              desiredState: SandboxDesiredState.STOPPED,
              pending: Not(true),
              autoDeleteInterval: MoreThanOrEqual(0),
              lastActivityAt: Raw((alias) => `${alias} < NOW() - INTERVAL '1 minute' * "autoDeleteInterval"`),
            },
            order: {
              lastActivityAt: 'ASC',
            },
            take: 100,
          })

          await Promise.all(
            sandboxes.map(async (sandbox) => {
              const lockKey = SYNC_INSTANCE_STATE_LOCK_KEY + sandbox.id
              const acquired = await this.redisLockProvider.lock(lockKey, 30)
              if (!acquired) {
                return
              }

              try {
                sandbox.pending = true
                sandbox.desiredState = SandboxDesiredState.DESTROYED
                await this.sandboxRepository.save(sandbox)
                this.syncInstanceState(sandbox.id)
              } catch (error) {
                this.logger.error(`Error processing auto-delete state for sandbox ${sandbox.id}:`, error)
              } finally {
                await this.redisLockProvider.unlock(lockKey)
              }
            }),
          )
        }),
      )
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'sync-states' })
  @TrackJobExecution()
  @OtelSpan()
  async syncStates(): Promise<void> {
    const globalLockKey = 'sync-states'
    if (!(await this.redisLockProvider.lock(globalLockKey, 30))) {
      return
    }

    try {
      const queryBuilder = this.sandboxRepository
        .createQueryBuilder('sandbox')
        .select(['sandbox.id'])
        .where('sandbox.state NOT IN (:...excludedStates)', {
          excludedStates: [SandboxState.DESTROYED, SandboxState.ERROR, SandboxState.BUILD_FAILED],
        })
        .andWhere('sandbox."desiredState"::text != sandbox.state::text')
        .andWhere('sandbox."desiredState"::text != :archived', { archived: SandboxDesiredState.ARCHIVED })
        .orderBy('sandbox."lastActivityAt"', 'ASC')

      const stream = await queryBuilder.stream()
      let processedCount = 0
      const maxProcessPerRun = 100
      const pendingProcesses: Promise<void>[] = []

      try {
        await new Promise<void>((resolve, reject) => {
          stream.on('data', (row: any) => {
            if (processedCount >= maxProcessPerRun) {
              resolve()
              return
            }

            // Process sandbox asynchronously but track the promise
            const processPromise = this.syncInstanceState(row.sandbox_id)
            pendingProcesses.push(processPromise)
            processedCount++

            // Limit concurrent processing to avoid overwhelming the system
            if (pendingProcesses.length >= 10) {
              stream.pause()
              Promise.all(pendingProcesses.splice(0, pendingProcesses.length))
                .then(() => stream.resume())
                .catch(reject)
            }
          })

          stream.on('end', () => {
            Promise.all(pendingProcesses)
              .then(() => {
                resolve()
              })
              .catch(reject)
          })

          stream.on('error', reject)
        })
      } finally {
        if (!stream.destroyed) {
          stream.destroy()
        }
      }
    } finally {
      await this.redisLockProvider.unlock(globalLockKey)
    }
  }

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'sync-archived-desired-states' })
  @TrackJobExecution()
  async syncArchivedDesiredStates(): Promise<void> {
    const lockKey = 'sync-archived-desired-states'
    if (!(await this.redisLockProvider.lock(lockKey, 30))) {
      return
    }

    const sandboxes = await this.sandboxRepository.find({
      where: {
        state: In([SandboxState.ARCHIVING, SandboxState.STOPPED]),
        desiredState: SandboxDesiredState.ARCHIVED,
      },
      take: 100,
      order: {
        lastActivityAt: 'ASC',
      },
    })

    await Promise.all(
      sandboxes.map(async (sandbox) => {
        this.syncInstanceState(sandbox.id)
      }),
    )
    await this.redisLockProvider.unlock(lockKey)
  }

  async syncInstanceState(sandboxId: string, startedAt = new Date()): Promise<void> {
    // If syncing for longer than 10 seconds, return
    // The sandbox will be continued in the next cron run
    // This prevents endless loops of syncing the same sandbox
    if (new Date().getTime() - startedAt.getTime() > 10000) {
      return
    }

    //  prevent syncState cron from running multiple instances of the same sandbox
    const lockKey = SYNC_INSTANCE_STATE_LOCK_KEY + sandboxId
    const acquired = await this.redisLockProvider.lock(lockKey, 360)
    if (!acquired) {
      return
    }

    const sandbox = await this.sandboxRepository.findOneByOrFail({
      id: sandboxId,
    })

    if ([SandboxState.DESTROYED, SandboxState.ERROR, SandboxState.BUILD_FAILED].includes(sandbox.state)) {
      await this.redisLockProvider.unlock(lockKey)
      return
    }

    let syncState = DONT_SYNC_AGAIN

    try {
      switch (sandbox.desiredState) {
        case SandboxDesiredState.STARTED: {
          syncState = await this.sandboxStartAction.run(sandbox)
          break
        }
        case SandboxDesiredState.STOPPED: {
          syncState = await this.sandboxStopAction.run(sandbox)
          break
        }
        case SandboxDesiredState.DESTROYED: {
          syncState = await this.sandboxDestroyAction.run(sandbox)
          break
        }
        case SandboxDesiredState.ARCHIVED: {
          syncState = await this.sandboxArchiveAction.run(sandbox)
          break
        }
      }
    } catch (error) {
      this.logger.error(`Error processing desired state for sandbox ${sandboxId}:`, error)

      const sandbox = await this.sandboxRepository.findOneBy({
        id: sandboxId,
      })
      if (!sandbox) {
        //  edge case where sandbox is deleted while desired state is being processed
        return
      }
      sandbox.state = SandboxState.ERROR
      sandbox.errorReason = error.message || String(error)
      await this.sandboxRepository.save(sandbox)
    }

    await this.redisLockProvider.unlock(lockKey)
    if (syncState === SYNC_AGAIN) {
      this.syncInstanceState(sandboxId, startedAt)
    }
  }

  @OnEvent(SandboxEvents.ARCHIVED)
  @TrackJobExecution()
  private async handleSandboxArchivedEvent(event: SandboxArchivedEvent) {
    this.syncInstanceState(event.sandbox.id).catch(this.logger.error)
  }

  @OnEvent(SandboxEvents.DESTROYED)
  @TrackJobExecution()
  private async handleSandboxDestroyedEvent(event: SandboxDestroyedEvent) {
    this.syncInstanceState(event.sandbox.id).catch(this.logger.error)
  }

  @OnEvent(SandboxEvents.STARTED)
  @TrackJobExecution()
  private async handleSandboxStartedEvent(event: SandboxStartedEvent) {
    this.syncInstanceState(event.sandbox.id).catch(this.logger.error)
  }

  @OnEvent(SandboxEvents.STOPPED)
  @TrackJobExecution()
  private async handleSandboxStoppedEvent(event: SandboxStoppedEvent) {
    this.syncInstanceState(event.sandbox.id).catch(this.logger.error)
  }

  @OnEvent(SandboxEvents.CREATED)
  @TrackJobExecution()
  private async handleSandboxCreatedEvent(event: SandboxCreatedEvent) {
    this.syncInstanceState(event.sandbox.id).catch(this.logger.error)
  }
}
