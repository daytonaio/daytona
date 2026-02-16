/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, OnApplicationShutdown } from '@nestjs/common'
import { Cron, CronExpression } from '@nestjs/schedule'
import { In, IsNull, MoreThanOrEqual, Not, Raw } from 'typeorm'
import { randomUUID } from 'crypto'

import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { RunnerService } from '../services/runner.service'

import { RedisLockProvider, LockCode } from '../common/redis-lock.provider'

import { SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../constants/sandbox.constants'

import { SandboxEvents } from '../constants/sandbox-events.constants'
import { SandboxStoppedEvent } from '../events/sandbox-stopped.event'
import { SandboxStartedEvent } from '../events/sandbox-started.event'
import { SandboxArchivedEvent } from '../events/sandbox-archived.event'
import { SandboxDestroyedEvent } from '../events/sandbox-destroyed.event'
import { SandboxCreatedEvent } from '../events/sandbox-create.event'

import { WithInstrumentation } from '../../common/decorators/otel.decorator'

import { SandboxStartAction } from './sandbox-actions/sandbox-start.action'
import { SandboxStopAction } from './sandbox-actions/sandbox-stop.action'
import { SandboxDestroyAction } from './sandbox-actions/sandbox-destroy.action'
import { SandboxArchiveAction } from './sandbox-actions/sandbox-archive.action'
import { SYNC_AGAIN, DONT_SYNC_AGAIN } from './sandbox-actions/sandbox.action'

import { TrackJobExecution } from '../../common/decorators/track-job-execution.decorator'
import { TrackableJobExecutions } from '../../common/interfaces/trackable-job-executions'
import { setTimeout } from 'timers/promises'
import { LogExecution } from '../../common/decorators/log-execution.decorator'
import { SandboxRepository } from '../repositories/sandbox.repository'
import { getStateChangeLockKey } from '../utils/lock-key.util'
import { BackupState } from '../enums/backup-state.enum'
import { OnAsyncEvent } from '../../common/decorators/on-async-event.decorator'
import { sanitizeSandboxError } from '../utils/sanitize-error.util'
import { Sandbox } from '../entities/sandbox.entity'
import { RunnerAdapterFactory } from '../runner-adapter/runnerAdapter'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { OrganizationService } from '../../organization/services/organization.service'
import { TypedConfigService } from '../../config/typed-config.service'
import { BackupManager } from './backup.manager'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'

@Injectable()
export class SandboxManager implements TrackableJobExecutions, OnApplicationShutdown {
  activeJobs = new Set<string>()

  private readonly logger = new Logger(SandboxManager.name)

  constructor(
    private readonly sandboxRepository: SandboxRepository,
    private readonly runnerService: RunnerService,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly sandboxStartAction: SandboxStartAction,
    private readonly sandboxStopAction: SandboxStopAction,
    private readonly sandboxDestroyAction: SandboxDestroyAction,
    private readonly sandboxArchiveAction: SandboxArchiveAction,
    private readonly configService: TypedConfigService,
    private readonly dockerRegistryService: DockerRegistryService,
    private readonly organizationService: OrganizationService,
    private readonly runnerAdapterFactory: RunnerAdapterFactory,
    private readonly backupManager: BackupManager,
    @InjectRedis() private readonly redis: Redis,
  ) {}

  async onApplicationShutdown() {
    //  wait for all active jobs to finish
    while (this.activeJobs.size > 0) {
      this.logger.log(`Waiting for ${this.activeJobs.size} active jobs to finish`)
      await setTimeout(1000)
    }
  }

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'auto-stop-check' })
  @TrackJobExecution()
  @WithInstrumentation()
  @LogExecution('auto-stop-check')
  @WithInstrumentation()
  async autostopCheck(): Promise<void> {
    const lockKey = 'auto-stop-check-worker-selected'
    //  lock the sync to only run one instance at a time
    if (!(await this.redisLockProvider.lock(lockKey, 60))) {
      return
    }

    try {
      const readyRunners = await this.runnerService.findAllReady()

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
            take: 100,
          })

          await Promise.all(
            sandboxes.map(async (sandbox) => {
              const lockKey = getStateChangeLockKey(sandbox.id)
              const acquired = await this.redisLockProvider.lock(lockKey, 30)
              if (!acquired) {
                return
              }

              try {
                //  if auto-delete interval is 0, delete the sandbox immediately
                if (sandbox.autoDeleteInterval === 0) {
                  sandbox.applyDesiredDestroyedState()
                } else {
                  sandbox.pending = true
                  sandbox.desiredState = SandboxDesiredState.STOPPED
                }
                await this.sandboxRepository.saveWhere(sandbox, { pending: false, state: sandbox.state })

                this.syncInstanceState(sandbox.id).catch(this.logger.error)
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

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'auto-archive-check' })
  @TrackJobExecution()
  @LogExecution('auto-archive-check')
  @WithInstrumentation()
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
          const lockKey = getStateChangeLockKey(sandbox.id)
          const acquired = await this.redisLockProvider.lock(lockKey, 30)
          if (!acquired) {
            return
          }

          try {
            sandbox.desiredState = SandboxDesiredState.ARCHIVED
            await this.sandboxRepository.saveWhere(sandbox, { pending: false, state: sandbox.state })
            this.syncInstanceState(sandbox.id).catch(this.logger.error)
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

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'auto-delete-check' })
  @TrackJobExecution()
  @LogExecution('auto-delete-check')
  @WithInstrumentation()
  async autoDeleteCheck(): Promise<void> {
    const lockKey = 'auto-delete-check-worker-selected'
    //  lock the sync to only run one instance at a time
    if (!(await this.redisLockProvider.lock(lockKey, 60))) {
      return
    }

    try {
      const readyRunners = await this.runnerService.findAllReady()

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
              const lockKey = getStateChangeLockKey(sandbox.id)
              const acquired = await this.redisLockProvider.lock(lockKey, 30)
              if (!acquired) {
                return
              }

              try {
                sandbox.applyDesiredDestroyedState()
                await this.sandboxRepository.saveWhere(sandbox, { pending: false, state: sandbox.state })

                this.syncInstanceState(sandbox.id).catch(this.logger.error)
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

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'draining-runner-sandboxes-check' })
  @TrackJobExecution()
  @LogExecution('draining-runner-sandboxes-check')
  @WithInstrumentation()
  async drainingRunnerSandboxesCheck(): Promise<void> {
    const lockKey = 'draining-runner-sandboxes-check'
    const lockTtl = 10 * 60 // seconds (10 min)
    if (!(await this.redisLockProvider.lock(lockKey, lockTtl))) {
      return
    }

    try {
      const skip = (await this.redis.get('draining-runner-sandboxes-skip')) || 0

      const drainingRunners = await this.runnerService.findDrainingPaginated(Number(skip), 10)

      this.logger.debug(`Checking ${drainingRunners.length} draining runners for sandbox migration (offset: ${skip})`)

      if (drainingRunners.length === 0) {
        await this.redis.set('draining-runner-sandboxes-skip', 0)
        return
      }

      await this.redis.set('draining-runner-sandboxes-skip', Number(skip) + drainingRunners.length)

      await Promise.allSettled(
        drainingRunners.map(async (runner) => {
          try {
            const sandboxes = await this.sandboxRepository.find({
              where: {
                runnerId: runner.id,
                state: SandboxState.STOPPED,
                desiredState: SandboxDesiredState.STOPPED,
                backupState: BackupState.COMPLETED,
                backupSnapshot: Not(IsNull()),
              },
              take: 100,
            })

            this.logger.debug(
              `Found ${sandboxes.length} eligible sandboxes on draining runner ${runner.id} for migration`,
            )

            await Promise.allSettled(
              sandboxes.map(async (sandbox) => {
                const sandboxLockKey = getStateChangeLockKey(sandbox.id)
                const hasSandboxLock = await this.redisLockProvider.lock(sandboxLockKey, 60)
                if (!hasSandboxLock) {
                  return
                }

                try {
                  const startScoreThreshold = this.configService.get('runnerScore.thresholds.start') || 0
                  const targetRunner = await this.runnerService.getRandomAvailableRunner({
                    snapshotRef: sandbox.backupSnapshot,
                    excludedRunnerIds: [runner.id],
                    availabilityScoreThreshold: startScoreThreshold,
                  })

                  await this.reassignSandbox(sandbox, runner.id, targetRunner.id)
                } catch (e) {
                  this.logger.error(`Error migrating sandbox ${sandbox.id} from draining runner ${runner.id}`, e)
                } finally {
                  await this.redisLockProvider.unlock(sandboxLockKey)
                }
              }),
            )

            // Archive ERROR sandboxes that have completed backups on this draining runner
            await this.archiveErroredSandboxesOnDrainingRunner(runner.id)

            // Retry backups for non-started sandboxes with errored backup state
            await this.retryErroredBackupsOnDrainingRunner(runner.id)
          } catch (e) {
            this.logger.error(`Error processing draining runner ${runner.id} for sandbox migration`, e)
          }
        }),
      )
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  private async archiveErroredSandboxesOnDrainingRunner(runnerId: string): Promise<void> {
    const erroredSandboxes = await this.sandboxRepository.find({
      where: {
        runnerId,
        state: SandboxState.ERROR,
        desiredState: Not(In([SandboxDesiredState.DESTROYED, SandboxDesiredState.ARCHIVED])),
        backupState: BackupState.COMPLETED,
        backupSnapshot: Not(IsNull()),
      },
      take: 100,
    })

    if (erroredSandboxes.length === 0) {
      return
    }

    this.logger.debug(
      `Found ${erroredSandboxes.length} errored sandboxes with completed backups on draining runner ${runnerId}`,
    )

    await Promise.allSettled(
      erroredSandboxes.map(async (sandbox) => {
        const sandboxLockKey = getStateChangeLockKey(sandbox.id)
        const acquired = await this.redisLockProvider.lock(sandboxLockKey, 30)
        if (!acquired) {
          return
        }

        try {
          this.logger.warn(
            `Setting desired state to ARCHIVED for errored sandbox ${sandbox.id} on draining runner ${runnerId} (previous desired state: ${sandbox.desiredState})`,
          )
          sandbox.desiredState = SandboxDesiredState.ARCHIVED
          await this.sandboxRepository.saveWhere(sandbox, { state: SandboxState.ERROR })
        } catch (e) {
          this.logger.error(
            `Failed to set desired state to ARCHIVED for errored sandbox ${sandbox.id} on draining runner ${runnerId}`,
            e,
          )
        } finally {
          await this.redisLockProvider.unlock(sandboxLockKey)
        }
      }),
    )
  }

  private static readonly DRAINING_BACKUP_RETRY_TTL_SECONDS = 12 * 60 * 60 // 12 hours

  private async retryErroredBackupsOnDrainingRunner(runnerId: string): Promise<void> {
    const erroredSandboxes = await this.sandboxRepository.find({
      where: [
        {
          runnerId,
          state: SandboxState.STOPPED,
          desiredState: SandboxDesiredState.STOPPED,
          backupState: BackupState.ERROR,
        },
        {
          runnerId,
          state: SandboxState.ERROR,
          backupState: In([BackupState.ERROR, BackupState.NONE]),
          desiredState: Not(SandboxDesiredState.DESTROYED),
        },
      ],
      take: 100,
    })

    if (erroredSandboxes.length === 0) {
      return
    }

    this.logger.debug(`Found ${erroredSandboxes.length} sandboxes with errored backups on draining runner ${runnerId}`)

    await Promise.allSettled(
      erroredSandboxes.map(async (sandbox) => {
        const redisKey = `draining:backup-retry:${sandbox.id}`

        // Check if we've already retried within the last 12 hours
        const alreadyRetried = await this.redis.exists(redisKey)
        if (alreadyRetried) {
          this.logger.debug(
            `Skipping backup retry for sandbox ${sandbox.id} on draining runner ${runnerId} — already retried within 12 hours`,
          )
          return
        }

        try {
          await this.backupManager.setBackupPending(sandbox)
          await this.redis.set(redisKey, '1', 'EX', SandboxManager.DRAINING_BACKUP_RETRY_TTL_SECONDS)
          this.logger.log(`Retried backup for sandbox ${sandbox.id} on draining runner ${runnerId}`)
        } catch (e) {
          this.logger.error(`Failed to retry backup for sandbox ${sandbox.id} on draining runner ${runnerId}`, e)
        }
      }),
    )
  }

  private async reassignSandbox(sandbox: Sandbox, oldRunnerId: string, newRunnerId: string): Promise<void> {
    this.logger.debug(
      `Starting sandbox reassignment for ${sandbox.id} from runner ${oldRunnerId} to runner ${newRunnerId}`,
    )

    // Safety check: ensure sandbox is not pending
    if (sandbox.pending) {
      this.logger.warn(
        `Sandbox ${sandbox.id} is pending, skipping reassignment from runner ${oldRunnerId} to runner ${newRunnerId}`,
      )
      return
    }

    if (!sandbox.backupRegistryId) {
      throw new Error(`Sandbox ${sandbox.id} has no backup registry`)
    }

    const registry = await this.dockerRegistryService.findOne(sandbox.backupRegistryId)
    if (!registry) {
      throw new Error(`Registry ${sandbox.backupRegistryId} not found for sandbox ${sandbox.id}`)
    }

    const organization = await this.organizationService.findOne(sandbox.organizationId)

    const metadata = {
      ...organization?.sandboxMetadata,
      sandboxName: sandbox.name,
    }

    const newRunner = await this.runnerService.findOneOrFail(newRunnerId)
    const newRunnerAdapter = await this.runnerAdapterFactory.create(newRunner)

    const originalSnapshot = sandbox.snapshot
    sandbox.snapshot = sandbox.backupSnapshot

    try {
      // Pass undefined for entrypoint as the backup snapshot already has it baked in and use skipStart
      await newRunnerAdapter.createSandbox(sandbox, registry, undefined, metadata, undefined, true)
      this.logger.debug(`Created sandbox ${sandbox.id} on new runner ${newRunnerId} with skipStart`)
    } catch (e) {
      // Restore original snapshot on failure
      sandbox.snapshot = originalSnapshot
      this.logger.error(`Failed to create sandbox ${sandbox.id} on new runner ${newRunnerId}`, e)
      throw e
    }

    // Re-fetch sandbox from DB to get fresh state (the in-memory entity may be stale)
    const freshSandbox = await this.sandboxRepository.findOne({ where: { id: sandbox.id } })
    if (!freshSandbox || freshSandbox.pending) {
      this.logger.warn(
        `Sandbox ${sandbox.id} is pending or missing, aborting reassignment from runner ${oldRunnerId} to runner ${newRunnerId}`,
      )

      // Roll back: remove the sandbox from the new runner since we won't complete the migration
      try {
        await newRunnerAdapter.destroySandbox(sandbox.id)
        this.logger.debug(`Rolled back sandbox ${sandbox.id} creation on new runner ${newRunnerId}`)
      } catch (rollbackErr) {
        this.logger.error(
          `Failed to roll back sandbox ${sandbox.id} on new runner ${newRunnerId} after pending check`,
          rollbackErr,
        )
      }
      return
    }

    // Update the sandbox to use the new runner; roll back on failure
    try {
      await this.sandboxRepository.update(sandbox.id, {
        prevRunnerId: sandbox.runnerId,
        runnerId: newRunnerId,
      })
    } catch (e) {
      this.logger.error(`Failed to update sandbox ${sandbox.id} runnerId to ${newRunnerId}, rolling back`, e)

      // Roll back: remove the sandbox from the new runner
      try {
        await newRunnerAdapter.destroySandbox(sandbox.id)
        this.logger.debug(`Rolled back sandbox ${sandbox.id} creation on new runner ${newRunnerId}`)
      } catch (rollbackErr) {
        this.logger.error(
          `Failed to roll back sandbox ${sandbox.id} on new runner ${newRunnerId} after DB update failure`,
          rollbackErr,
        )
      }
      throw e
    }

    this.logger.log(`Migrated sandbox ${sandbox.id} from draining runner ${oldRunnerId} to runner ${newRunnerId}`)

    // Best effort deletion of the sandbox on the old runner
    try {
      const oldRunner = await this.runnerService.findOne(oldRunnerId)
      if (oldRunner) {
        const oldRunnerAdapter = await this.runnerAdapterFactory.create(oldRunner)
        await oldRunnerAdapter.destroySandbox(sandbox.id)
        this.logger.debug(`Deleted sandbox ${sandbox.id} from old runner ${oldRunnerId}`)
      }
    } catch (e) {
      this.logger.warn(`Best effort deletion failed for sandbox ${sandbox.id} on old runner ${oldRunnerId}`, e)
    }
  }

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'sync-states' })
  @TrackJobExecution()
  @WithInstrumentation()
  @LogExecution('sync-states')
  async syncStates(): Promise<void> {
    const globalLockKey = 'sync-states'
    const lockTtl = 10 * 60 // seconds (10 min)
    if (!(await this.redisLockProvider.lock(globalLockKey, lockTtl))) {
      return
    }

    try {
      const queryBuilder = this.sandboxRepository
        .createQueryBuilder('sandbox')
        .select(['sandbox.id'])
        .where('sandbox.state NOT IN (:...excludedStates)', {
          excludedStates: [
            SandboxState.DESTROYED,
            SandboxState.ERROR,
            SandboxState.BUILD_FAILED,
            SandboxState.RESIZING,
          ],
        })
        .andWhere('sandbox."desiredState"::text != sandbox.state::text')
        .andWhere('sandbox."desiredState"::text != :archived', { archived: SandboxDesiredState.ARCHIVED })
        .orderBy('sandbox."lastActivityAt"', 'DESC')

      const stream = await queryBuilder.stream()
      let processedCount = 0
      const maxProcessPerRun = 1000
      const pendingProcesses: Promise<void>[] = []

      try {
        await new Promise<void>((resolve, reject) => {
          stream.on('data', async (row: any) => {
            if (processedCount >= maxProcessPerRun) {
              resolve()
              return
            }

            const lockKey = getStateChangeLockKey(row.sandbox_id)
            if (await this.redisLockProvider.isLocked(lockKey)) {
              // Sandbox is already being processed, skip it
              return
            }

            // Process sandbox asynchronously but track the promise
            const processPromise = this.syncInstanceState(row.sandbox_id)
            pendingProcesses.push(processPromise)
            processedCount++

            // Limit concurrent processing to avoid overwhelming the system
            if (pendingProcesses.length >= 10) {
              stream.pause()
              Promise.allSettled(pendingProcesses.splice(0, pendingProcesses.length))
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
  @LogExecution('sync-archived-desired-states')
  @WithInstrumentation()
  async syncArchivedDesiredStates(): Promise<void> {
    const lockKey = 'sync-archived-desired-states'
    if (!(await this.redisLockProvider.lock(lockKey, 30))) {
      return
    }

    const sandboxes = await this.sandboxRepository.find({
      where: {
        state: In([SandboxState.ARCHIVING, SandboxState.STOPPED, SandboxState.ERROR]),
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

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'sync-archived-completed-states' })
  @TrackJobExecution()
  @LogExecution('sync-archived-completed-states')
  async syncArchivedCompletedStates(): Promise<void> {
    const lockKey = 'sync-archived-completed-states'
    if (!(await this.redisLockProvider.lock(lockKey, 30))) {
      return
    }

    const sandboxes = await this.sandboxRepository.find({
      where: {
        state: In([SandboxState.ARCHIVING, SandboxState.STOPPED, SandboxState.ERROR]),
        desiredState: SandboxDesiredState.ARCHIVED,
        backupState: BackupState.COMPLETED,
      },
      take: 100,
      order: {
        updatedAt: 'ASC',
      },
    })

    await Promise.allSettled(
      sandboxes.map(async (sandbox) => {
        await this.syncInstanceState(sandbox.id)
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

    //  generate a random lock code to prevent race condition if sandbox action continues
    //  after the lock expires
    const lockCode = new LockCode(randomUUID())

    //  prevent syncState cron from running multiple instances of the same sandbox
    const lockKey = getStateChangeLockKey(sandboxId)
    const acquired = await this.redisLockProvider.lock(lockKey, 30, lockCode)
    if (!acquired) {
      return
    }

    const sandbox = await this.sandboxRepository.findOneOrFail({
      where: { id: sandboxId },
    })

    if (
      [SandboxState.DESTROYED, SandboxState.ERROR, SandboxState.BUILD_FAILED, SandboxState.RESIZING].includes(
        sandbox.state,
      )
    ) {
      // Allow ERROR → ARCHIVED transition (e.g., during runner draining)
      if (!(sandbox.state === SandboxState.ERROR && sandbox.desiredState === SandboxDesiredState.ARCHIVED)) {
        await this.redisLockProvider.unlock(lockKey)
        return
      }
    }

    //  prevent potential race condition, or SYNC_AGAIN loop bug
    //  this should never happen
    if (String(sandbox.state) === String(sandbox.desiredState)) {
      this.logger.warn(`Sandbox ${sandboxId} is already in the desired state ${sandbox.desiredState}, skipping sync`)
      await this.redisLockProvider.unlock(lockKey)
      return
    }

    let syncState = DONT_SYNC_AGAIN

    try {
      switch (sandbox.desiredState) {
        case SandboxDesiredState.STARTED: {
          syncState = await this.sandboxStartAction.run(sandbox, lockCode)
          break
        }
        case SandboxDesiredState.STOPPED: {
          syncState = await this.sandboxStopAction.run(sandbox, lockCode)
          break
        }
        case SandboxDesiredState.DESTROYED: {
          syncState = await this.sandboxDestroyAction.run(sandbox, lockCode)
          break
        }
        case SandboxDesiredState.ARCHIVED: {
          syncState = await this.sandboxArchiveAction.run(sandbox, lockCode)
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

      const { recoverable, errorReason } = sanitizeSandboxError(error)
      sandbox.errorReason = errorReason
      sandbox.recoverable = recoverable

      await this.sandboxRepository.save(sandbox)
    }

    await this.redisLockProvider.unlock(lockKey)
    if (syncState === SYNC_AGAIN) {
      this.syncInstanceState(sandboxId, startedAt)
    }
  }

  @OnAsyncEvent({
    event: SandboxEvents.ARCHIVED,
  })
  @TrackJobExecution()
  private async handleSandboxArchivedEvent(event: SandboxArchivedEvent) {
    await this.syncInstanceState(event.sandbox.id)
  }

  @OnAsyncEvent({
    event: SandboxEvents.DESTROYED,
  })
  @TrackJobExecution()
  private async handleSandboxDestroyedEvent(event: SandboxDestroyedEvent) {
    await this.syncInstanceState(event.sandbox.id)
  }

  @OnAsyncEvent({
    event: SandboxEvents.STARTED,
  })
  @TrackJobExecution()
  private async handleSandboxStartedEvent(event: SandboxStartedEvent) {
    await this.syncInstanceState(event.sandbox.id)
  }

  @OnAsyncEvent({
    event: SandboxEvents.STOPPED,
  })
  @TrackJobExecution()
  private async handleSandboxStoppedEvent(event: SandboxStoppedEvent) {
    await this.syncInstanceState(event.sandbox.id)
  }

  @OnAsyncEvent({
    event: SandboxEvents.CREATED,
  })
  @TrackJobExecution()
  private async handleSandboxCreatedEvent(event: SandboxCreatedEvent) {
    await this.syncInstanceState(event.sandbox.id)
  }
}
