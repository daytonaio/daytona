/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, OnApplicationShutdown } from '@nestjs/common'
import { Cron, CronExpression } from '@nestjs/schedule'
import { Brackets, In, IsNull, LessThan, Not, Or, Repository } from 'typeorm'
import { InjectRepository } from '@nestjs/typeorm'
import { Sandbox } from '../entities/sandbox.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { RunnerService } from '../services/runner.service'
import { RunnerState } from '../enums/runner-state.enum'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { BackupState } from '../enums/backup-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../constants/sandbox.constants'
import { BACKUP_RETRY_ERROR_SUBSTRINGS } from '../constants/errors-for-backup-retry'
import { fromAxiosError } from '../../common/utils/from-axios-error'
import { sanitizeSandboxError } from '../utils/sanitize-error.util'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { OnEvent } from '@nestjs/event-emitter'
import { SandboxEvents } from '../constants/sandbox-events.constants'
import { SandboxDestroyedEvent } from '../events/sandbox-destroyed.event'
import { SandboxBackupCreatedEvent } from '../events/sandbox-backup-created.event'
import { SandboxArchivedEvent } from '../events/sandbox-archived.event'
import { RunnerAdapterFactory } from '../runner-adapter/runnerAdapter'
import { TypedConfigService } from '../../config/typed-config.service'

import { TrackJobExecution } from '../../common/decorators/track-job-execution.decorator'
import { TrackableJobExecutions } from '../../common/interfaces/trackable-job-executions'
import { setTimeout } from 'timers/promises'
import { LogExecution } from '../../common/decorators/log-execution.decorator'
import { WithInstrumentation } from '../../common/decorators/otel.decorator'
import { DockerRegistry } from '../../docker-registry/entities/docker-registry.entity'
import { SandboxService } from '../services/sandbox.service'
import { SandboxRepository } from '../repositories/sandbox.repository'
import { Job } from '../entities/job.entity'
import { JobStatus } from '../enums/job-status.enum'
import { JobType } from '../enums/job-type.enum'
import { ResourceType } from '../enums/resource-type.enum'
import { JobStateHandlerService } from '../services/job-state-handler.service'
import { SandboxConflictError } from '../errors/sandbox-conflict.error'

@Injectable()
export class BackupManager implements TrackableJobExecutions, OnApplicationShutdown {
  activeJobs = new Set<string>()

  private readonly logger = new Logger(BackupManager.name)

  constructor(
    private readonly sandboxRepository: SandboxRepository,
    private readonly sandboxService: SandboxService,
    private readonly runnerService: RunnerService,
    private readonly runnerAdapterFactory: RunnerAdapterFactory,
    private readonly dockerRegistryService: DockerRegistryService,
    @InjectRedis() private readonly redis: Redis,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly configService: TypedConfigService,
    @InjectRepository(Job)
    private readonly jobRepository: Repository<Job>,
    private readonly jobStateHandlerService: JobStateHandlerService,
  ) {}

  //  on init
  async onApplicationBootstrap() {
    await this.adHocBackupCheck()
  }

  async onApplicationShutdown() {
    //  wait for all active jobs to finish
    while (this.activeJobs.size > 0) {
      this.logger.log(`Waiting for ${this.activeJobs.size} active jobs to finish`)
      await setTimeout(1000)
    }
  }

  /**
   * Increments a retry counter in Redis and returns whether the operation should be retried.
   * @param key - The Redis key for the retry counter
   * @param maxRetries - Maximum number of retries allowed (default: 3)
   * @param ttlSeconds - TTL for the retry counter in seconds (default: 300)
   * @returns true if should retry, false if max retries exceeded
   */
  private async runnerIsDraining(sandbox: Sandbox): Promise<boolean> {
    if (!sandbox.runnerId) return false
    try {
      const runner = await this.runnerService.findOne(sandbox.runnerId)
      return runner?.draining === true
    } catch {
      return false
    }
  }

  /** Route recoverable backup failures on draining runners into state=ERROR so /recover and
   *  drain auto-recover (both gate on state=ERROR + recoverable=true) catch them. */
  private async markErroredIfDraining(
    sandbox: Sandbox,
    errorReason: string | null,
    recoverable: boolean,
    isOnDrainingRunner: boolean,
  ): Promise<void> {
    if (!isOnDrainingRunner || !recoverable || sandbox.state === SandboxState.ERROR) return
    // Mirror errorReason to backupErrorReason so the runner's DeduceRecoveryType matches either.
    await this.sandboxRepository.updateWhere(sandbox.id, {
      updateData: {
        state: SandboxState.ERROR,
        errorReason,
      },
      whereCondition: { state: sandbox.state },
    })
  }

  private async shouldRetry(key: string, maxRetries = 3, ttlSeconds = 300): Promise<boolean> {
    const retryCount = await this.redis.get(key)
    const currentCount = retryCount ? parseInt(retryCount) : 0

    if (currentCount >= maxRetries) {
      await this.redis.del(key)
      return false
    }

    await this.redis.setex(key, ttlSeconds, String(currentCount + 1))
    return true
  }

  //  todo: make frequency configurable or more efficient
  @Cron(CronExpression.EVERY_5_MINUTES, { name: 'ad-hoc-backup-check' })
  @TrackJobExecution()
  @LogExecution('ad-hoc-backup-check')
  @WithInstrumentation()
  async adHocBackupCheck(): Promise<void> {
    const lockKey = 'ad-hoc-backup-check'
    const hasLock = await this.redisLockProvider.lock(lockKey, 5 * 60)
    if (!hasLock) {
      return
    }

    // Get all ready runners
    const readyRunners = await this.runnerService.findAllReady()

    try {
      // Process all runners in parallel
      await Promise.all(
        readyRunners.map(async (runner) => {
          const sandboxes = await this.sandboxRepository.find({
            where: {
              runnerId: runner.id,
              organizationId: Not(SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION),
              state: SandboxState.STARTED,
              desiredState: Not(SandboxDesiredState.DESTROYED),
              backupState: In([BackupState.NONE, BackupState.COMPLETED]),
              lastBackupAt: Or(IsNull(), LessThan(new Date(Date.now() - 1 * 60 * 60 * 1000))),
              autoDeleteInterval: Not(0),
            },
            order: {
              lastBackupAt: 'ASC',
            },
            take: 25,
          })

          await Promise.all(
            sandboxes.map(async (sandbox) => {
              const lockKey = `sandbox-backup-${sandbox.id}`
              const hasLock = await this.redisLockProvider.lock(lockKey, 60)
              if (!hasLock) {
                return
              }

              try {
                //  todo: remove the catch handler asap
                await this.setBackupPending(sandbox).catch((error) => {
                  if (error instanceof BadRequestError && error.message === 'A backup is already in progress') {
                    return
                  }
                  this.logger.error(`Failed to create backup for sandbox ${sandbox.id}:`, fromAxiosError(error))
                })
              } catch (error) {
                this.logger.error(`Error processing stop state for sandbox ${sandbox.id}:`, fromAxiosError(error))
              } finally {
                await this.redisLockProvider.unlock(lockKey)
              }
            }),
          )
        }),
      )
    } catch (error) {
      this.logger.error(`Error processing backups: `, error)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'check-backup-states', waitForCompletion: true })
  @TrackJobExecution()
  @LogExecution('check-backup-states')
  @WithInstrumentation()
  async checkBackupStates(): Promise<void> {
    //  lock the sync to only run one instance at a time
    const lockKey = 'check-backup-states'
    const hasLock = await this.redisLockProvider.lock(lockKey, 10)
    if (!hasLock) {
      return
    }

    try {
      // PENDING only — IN_PROGRESS is handled by checkBackupStatesInProgress so a slow
      // in-progress poll can't block fresh PENDING backups. distinctOn(runnerId) caps
      // each runner to one PENDING sandbox per tick to spread load across the fleet.
      const sandboxes = await this.sandboxRepository
        .createQueryBuilder('sandbox')
        .innerJoin('runner', 'r', 'r.id = sandbox.runnerId')
        .where('sandbox.state IN (:...states)', {
          states: [SandboxState.ARCHIVING, SandboxState.STARTED, SandboxState.STOPPED],
        })
        .andWhere('sandbox.backupState IN (:...backupStates)', {
          backupStates: [BackupState.PENDING],
        })
        .andWhere('sandbox.desiredState != :destroyed', { destroyed: SandboxDesiredState.DESTROYED })
        .andWhere('r.state = :ready', { ready: RunnerState.READY })
        // Prioritize manual archival action, then auto-archive poller, then ad-hoc backup poller
        .addSelect(
          `
          CASE sandbox.state
            WHEN :archiving THEN 1
            WHEN :stopped   THEN 2
            WHEN :started   THEN 3
            ELSE 999
          END
          `,
          'state_priority',
        )
        .addSelect(
          `
          CASE sandbox."backupState"
            WHEN :in_progress THEN 1
            ELSE 999
          END
          `,
          'backup_state_priority',
        )
        .setParameters({
          in_progress: BackupState.IN_PROGRESS,
          archiving: SandboxState.ARCHIVING,
          stopped: SandboxState.STOPPED,
          started: SandboxState.STARTED,
        })
        .distinctOn(['sandbox.runnerId'])
        .orderBy('sandbox.runnerId', 'ASC')
        .addOrderBy('backup_state_priority', 'ASC')
        .addOrderBy('state_priority', 'ASC')
        .addOrderBy('sandbox.lastBackupAt', 'ASC', 'NULLS FIRST') // Process sandboxes with no backups first
        .take(250)
        .getMany()

      await Promise.allSettled(
        sandboxes.map(async (s) => {
          const lockKey = `sandbox-backup-${s.id}`
          const hasLock = await this.redisLockProvider.lock(lockKey, 60)
          if (!hasLock) {
            return
          }

          try {
            //  get the latest sandbox state
            const sandbox = await this.sandboxRepository.findOneByOrFail({
              id: s.id,
            })

            try {
              switch (sandbox.backupState) {
                case BackupState.PENDING: {
                  await this.handlePendingBackup(sandbox)
                  break
                }
              }
            } catch (error) {
              //  if error, retry 10 times
              const errorRetryKey = `${lockKey}-error-retry`
              const errorRetryCount = await this.redis.get(errorRetryKey)
              if (!errorRetryCount) {
                await this.redis.setex(errorRetryKey, 300, '1')
              } else if (parseInt(errorRetryCount) > 10) {
                this.logger.error(`Error processing backup for sandbox ${sandbox.id}:`, fromAxiosError(error))
                const { recoverable, errorReason } = sanitizeSandboxError(error)
                const isArchiveFlow = sandbox.desiredState === SandboxDesiredState.ARCHIVED
                const isOnDrainingRunner = await this.runnerIsDraining(sandbox)
                await this.sandboxService.updateSandboxBackupState(
                  sandbox.id,
                  BackupState.ERROR,
                  undefined,
                  undefined,
                  errorReason,
                  recoverable && (isArchiveFlow || isOnDrainingRunner),
                )
                await this.markErroredIfDraining(sandbox, errorReason, recoverable, isOnDrainingRunner)
              } else {
                await this.redis.setex(errorRetryKey, 300, errorRetryCount + 1)
              }
            }
          } catch (error) {
            this.logger.error(`Error processing backup for sandbox ${s.id}:`, error)
          } finally {
            await this.redisLockProvider.unlock(lockKey)
          }
        }),
      )
    } catch (error) {
      this.logger.error(`Error processing backups: `, error)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'check-backup-states-in-progress', waitForCompletion: true })
  @TrackJobExecution()
  @WithInstrumentation()
  @LogExecution('check-backup-states-in-progress')
  async checkBackupStatesInProgress(): Promise<void> {
    //  lock the sync to only run one instance at a time
    const lockKey = 'check-backup-states-in-progress'
    const hasLock = await this.redisLockProvider.lock(lockKey, 10)
    if (!hasLock) {
      return
    }

    try {
      // Random ordering avoids head-of-line blocking on a slow runner repeatedly
      // returning the same in-progress sandboxes.
      const sandboxes = await this.sandboxRepository
        .createQueryBuilder('sandbox')
        .addSelect('RANDOM()', 'rand')
        .innerJoin('runner', 'r', 'r.id = sandbox.runnerId')
        .where('sandbox.state IN (:...states)', {
          states: [SandboxState.ARCHIVING, SandboxState.STARTED, SandboxState.STOPPED],
        })
        .andWhere('sandbox.backupState IN (:...backupStates)', {
          backupStates: [BackupState.IN_PROGRESS],
        })
        .andWhere('sandbox.desiredState != :destroyed', { destroyed: SandboxDesiredState.DESTROYED })
        .andWhere('r.state = :ready', { ready: RunnerState.READY })
        .orderBy('rand')
        .take(200)
        .getMany()

      await Promise.allSettled(
        sandboxes.map(async (s) => {
          const lockKey = `sandbox-backup-${s.id}`
          const hasLock = await this.redisLockProvider.lock(lockKey, 60)
          if (!hasLock) {
            return
          }

          try {
            //  get the latest sandbox state
            const sandbox = await this.sandboxRepository.findOneByOrFail({
              id: s.id,
            })

            try {
              switch (sandbox.backupState) {
                case BackupState.IN_PROGRESS: {
                  await this.checkBackupProgress(sandbox)
                  break
                }
              }
            } catch (error) {
              //  if error, retry 10 times
              const errorRetryKey = `${lockKey}-error-retry`
              const errorRetryCount = await this.redis.get(errorRetryKey)
              if (!errorRetryCount) {
                await this.redis.setex(errorRetryKey, 300, '1')
              } else if (parseInt(errorRetryCount) > 10) {
                this.logger.error(`Error processing backup for sandbox ${sandbox.id}:`, fromAxiosError(error))
                const { recoverable, errorReason } = sanitizeSandboxError(error)
                const isArchiveFlow = sandbox.desiredState === SandboxDesiredState.ARCHIVED
                const isOnDrainingRunner = await this.runnerIsDraining(sandbox)
                await this.sandboxService.updateSandboxBackupState(
                  sandbox.id,
                  BackupState.ERROR,
                  undefined,
                  undefined,
                  errorReason,
                  recoverable && (isArchiveFlow || isOnDrainingRunner),
                )
                await this.markErroredIfDraining(sandbox, errorReason, recoverable, isOnDrainingRunner)
              } else {
                await this.redis.setex(errorRetryKey, 300, errorRetryCount + 1)
              }
            }
          } catch (error) {
            this.logger.error(`Error processing backup for sandbox ${s.id}:`, error)
          } finally {
            await this.redisLockProvider.unlock(lockKey)
          }
        }),
      )
    } catch (error) {
      this.logger.error(`Error processing backups: `, error)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'check-backup-states-errored-draining' })
  @TrackJobExecution()
  @LogExecution('check-backup-states-errored-draining')
  @WithInstrumentation()
  async checkBackupStatesForErroredDraining(): Promise<void> {
    const lockKey = 'check-backup-states-errored-draining'
    const hasLock = await this.redisLockProvider.lock(lockKey, 10)
    if (!hasLock) {
      return
    }

    try {
      const sandboxes = await this.sandboxRepository
        .createQueryBuilder('sandbox')
        .innerJoin('runner', 'r', 'r.id = sandbox.runnerId')
        .where('sandbox.state = :error', { error: SandboxState.ERROR })
        .andWhere('sandbox.backupState IN (:...backupStates)', {
          backupStates: [BackupState.PENDING, BackupState.IN_PROGRESS],
        })
        .andWhere('r.state = :ready', { ready: RunnerState.READY })
        .andWhere('r."draining" = true')
        .addOrderBy('sandbox.lastBackupAt', 'ASC', 'NULLS FIRST')
        .addOrderBy('sandbox.createdAt', 'ASC')
        .take(100)
        .getMany()

      await Promise.allSettled(
        sandboxes.map(async (s) => {
          const lockKey = `sandbox-backup-${s.id}`
          const hasLock = await this.redisLockProvider.lock(lockKey, 60)
          if (!hasLock) {
            return
          }

          try {
            const sandbox = await this.sandboxRepository.findOneByOrFail({
              id: s.id,
            })

            try {
              switch (sandbox.backupState) {
                case BackupState.PENDING: {
                  await this.handlePendingBackup(sandbox)
                  break
                }
                case BackupState.IN_PROGRESS: {
                  await this.checkBackupProgress(sandbox)
                  break
                }
              }
            } catch (error) {
              const errorRetryKey = `${lockKey}-error-retry`
              const errorRetryCount = await this.redis.get(errorRetryKey)
              if (!errorRetryCount) {
                await this.redis.setex(errorRetryKey, 300, '1')
              } else if (parseInt(errorRetryCount) > 10) {
                this.logger.error(
                  `Error processing backup for errored sandbox ${sandbox.id} on draining runner:`,
                  fromAxiosError(error),
                )
                const { recoverable, errorReason } = sanitizeSandboxError(error)
                const isArchiveFlow = sandbox.desiredState === SandboxDesiredState.ARCHIVED
                const isOnDrainingRunner = await this.runnerIsDraining(sandbox)
                await this.sandboxService.updateSandboxBackupState(
                  sandbox.id,
                  BackupState.ERROR,
                  undefined,
                  undefined,
                  errorReason,
                  recoverable && (isArchiveFlow || isOnDrainingRunner),
                )
                await this.markErroredIfDraining(sandbox, errorReason, recoverable, isOnDrainingRunner)
              } else {
                await this.redis.setex(errorRetryKey, 300, errorRetryCount + 1)
              }
            }
          } catch (error) {
            this.logger.error(`Error processing backup for errored sandbox ${s.id} on draining runner:`, error)
          } finally {
            await this.redisLockProvider.unlock(lockKey)
          }
        }),
      )
    } catch (error) {
      this.logger.error(`Error processing backups for errored sandboxes on draining runners: `, error)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'sync-stop-state-create-backups' })
  @TrackJobExecution()
  @LogExecution('sync-stop-state-create-backups')
  @WithInstrumentation()
  async syncStopStateCreateBackups(): Promise<void> {
    const lockKey = 'sync-stop-state-create-backups'
    const hasLock = await this.redisLockProvider.lock(lockKey, 10)
    if (!hasLock) {
      return
    }

    try {
      const sandboxes = await this.sandboxRepository
        .createQueryBuilder('sandbox')
        .innerJoin('runner', 'r', 'r.id = sandbox.runnerId')
        .where('sandbox.state IN (:...states)', { states: [SandboxState.ARCHIVING, SandboxState.STOPPED] })
        .andWhere('sandbox.backupState = :none', { none: BackupState.NONE })
        .andWhere('sandbox.desiredState != :destroyed', { destroyed: SandboxDesiredState.DESTROYED })
        .andWhere('r.state = :ready', { ready: RunnerState.READY })
        .take(100)
        .getMany()

      await Promise.allSettled(
        sandboxes
          .filter((sandbox) => sandbox.runnerId !== null)
          .map(async (sandbox) => {
            const lockKey = `sandbox-backup-${sandbox.id}`
            const hasLock = await this.redisLockProvider.lock(lockKey, 30)
            if (!hasLock) {
              return
            }

            try {
              await this.setBackupPending(sandbox)
            } catch (error) {
              this.logger.error(`Error processing backup for sandbox ${sandbox.id}:`, error)
            } finally {
              await this.redisLockProvider.unlock(lockKey)
            }
          }),
      )
    } catch (error) {
      this.logger.error(`Error processing backups: `, error)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @Cron(CronExpression.EVERY_10_MINUTES, { name: 'check-stale-in-progress-backups' })
  @TrackJobExecution()
  @LogExecution('check-stale-in-progress-backups')
  @WithInstrumentation()
  async checkStaleInProgressBackups(): Promise<void> {
    const lockKey = 'check-stale-in-progress-backups'
    const lockTtlSeconds = 5 * 60
    const hasLock = await this.redisLockProvider.lock(lockKey, lockTtlSeconds)
    if (!hasLock) {
      return
    }

    try {
      const twoHoursAgo = new Date(Date.now() - 2 * 60 * 60 * 1000)

      const staleSandboxes = await this.sandboxRepository.find({
        where: {
          backupState: BackupState.IN_PROGRESS,
          desiredState: Not(SandboxDesiredState.DESTROYED),
          updatedAt: LessThan(twoHoursAgo),
        },
        order: {
          updatedAt: 'ASC',
        },
        take: 100,
      })

      for (const sandbox of staleSandboxes) {
        try {
          await this.sandboxRepository.updateWhere(sandbox.id, {
            updateData: {
              backupState: BackupState.ERROR,
              backupErrorReason: 'Backup timed out after 2 hours',
            },
            whereCondition: {
              backupState: BackupState.IN_PROGRESS,
              desiredState: Not(SandboxDesiredState.DESTROYED),
              updatedAt: LessThan(twoHoursAgo),
            },
          })
          this.logger.warn(`Backup for sandbox ${sandbox.id} timed out after 2 hours`)
        } catch (error) {
          this.logger.error(`Failed to mark stale backup as errored for sandbox ${sandbox.id}:`, error)
        }
      }
    } catch (error) {
      this.logger.error('Error checking for stale in-progress backups:', error)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  async setBackupPending(sandbox: Sandbox): Promise<void> {
    if (sandbox.backupState === BackupState.COMPLETED) {
      return
    }

    // Allow backups for STARTED sandboxes, STOPPED/ERROR sandboxes with runnerId, or ARCHIVING sandboxes
    if (
      !(
        sandbox.state === SandboxState.STARTED ||
        sandbox.state === SandboxState.ARCHIVING ||
        (sandbox.state === SandboxState.STOPPED && sandbox.runnerId) ||
        (sandbox.state === SandboxState.ERROR && sandbox.runnerId)
      )
    ) {
      throw new BadRequestError('Sandbox must be started, stopped, or errored with assigned runner to create a backup')
    }

    if (sandbox.backupState === BackupState.IN_PROGRESS || sandbox.backupState === BackupState.PENDING) {
      return
    }

    let registry: DockerRegistry | null = null

    if (sandbox.backupRegistryId) {
      registry = await this.dockerRegistryService.findOne(sandbox.backupRegistryId)
    } else {
      registry = await this.dockerRegistryService.getAvailableBackupRegistry(sandbox.region)
    }

    if (!registry) {
      throw new BadRequestError('No backup registry configured')
    }
    // Generate backup snapshot name
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-')
    const backupSnapshot = `${registry.url.replace('https://', '').replace('http://', '')}/${registry.project || 'daytona'}/backup-${sandbox.id}:${timestamp}`

    await this.sandboxService.updateSandboxBackupState(sandbox.id, BackupState.PENDING, backupSnapshot, registry.id)
  }

  /**
   * Reconcile an in-progress backup on a v2 runner from its CREATE_BACKUP job.
   *
   * On v2 runners backup completion is driven by the job-state handler, which is invoked
   * once (fire-and-forget) when the runner reports the job terminal. If that invocation is
   * lost (e.g. a transient error, or it raced with the PENDING->IN_PROGRESS write), the
   * sandbox is stranded in InProgress while its job is already Completed/Failed - the
   * stale-job sweep ignores terminal jobs and the runner has nothing left to report.
   *
   * Re-running the authoritative completion handler here is idempotent and resolves the
   * sandbox to Completed/Error. The v2 RunnerAdapter's `sandboxInfo` only echoes the DB
   * backupState, so the runner-polling path below would otherwise be a no-op for v2.
   */
  private async reconcileV2BackupFromJob(sandbox: Sandbox): Promise<void> {
    const job = await this.jobRepository.findOne({
      where: {
        resourceType: ResourceType.SANDBOX,
        resourceId: sandbox.id,
        type: JobType.CREATE_BACKUP,
      },
      order: { createdAt: 'DESC' },
    })

    if (!job) {
      return
    }

    if (job.status === JobStatus.COMPLETED || job.status === JobStatus.FAILED) {
      // handleJobCompletion -> handleCreateBackupJobCompletion swallows its own errors,
      // so this won't throw and won't trip the caller's error-retry/ERROR path.
      await this.jobStateHandlerService.handleJobCompletion(job)
    }
  }

  private async checkBackupProgress(sandbox: Sandbox): Promise<void> {
    try {
      const runner = await this.runnerService.findOneOrFail(sandbox.runnerId)

      // v2+ runners don't expose live backup state via the runner API - completion is
      // tracked through the CREATE_BACKUP job. Reconcile from the job instead of polling.
      if (runner.apiVersion !== '0') {
        await this.reconcileV2BackupFromJob(sandbox)
        return
      }

      const runnerAdapter = await this.runnerAdapterFactory.create(runner)

      // Get sandbox info from runner
      const sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)

      switch (sandboxInfo.backupState) {
        case BackupState.COMPLETED: {
          // Only accept completion if the runner-reported snapshot matches the DB snapshot
          if (sandboxInfo.backupSnapshot && sandboxInfo.backupSnapshot !== sandbox.backupSnapshot) {
            this.logger.warn(
              `Ignoring stale backup completion for sandbox ${sandbox.id}: runner snapshot ${sandboxInfo.backupSnapshot} does not match DB snapshot ${sandbox.backupSnapshot}`,
            )
            break
          }
          await this.sandboxService.updateSandboxBackupState(sandbox.id, BackupState.COMPLETED)
          break
        }
        case BackupState.ERROR: {
          // Only accept failure if the runner-reported snapshot matches the DB snapshot
          if (sandboxInfo.backupSnapshot && sandboxInfo.backupSnapshot !== sandbox.backupSnapshot) {
            this.logger.warn(
              `Ignoring stale backup failure for sandbox ${sandbox.id}: runner snapshot ${sandboxInfo.backupSnapshot} does not match DB snapshot ${sandbox.backupSnapshot}`,
            )
            break
          }
          // Surface recoverable=true for archive flows or any backup error on a draining runner.
          const isArchiveFlow = sandbox.desiredState === SandboxDesiredState.ARCHIVED
          const isOnDrainingRunner = runner.draining === true
          const recoverable = sandboxInfo.recoverable ?? false
          await this.sandboxService.updateSandboxBackupState(
            sandbox.id,
            BackupState.ERROR,
            undefined,
            undefined,
            sandboxInfo.backupErrorReason,
            recoverable && (isArchiveFlow || isOnDrainingRunner),
          )
          await this.markErroredIfDraining(
            sandbox,
            sandboxInfo.backupErrorReason ?? null,
            recoverable,
            isOnDrainingRunner,
          )
          break
        }
        // If backup state is none, retry the backup process by setting the backup state to pending
        // This can happen if the runner is restarted or the operation is cancelled.
        // Bound the retries so a runner stuck reporting NONE doesn't loop forever.
        case BackupState.NONE: {
          const noneRetryKey = `sandbox-backup-${sandbox.id}-none-retry`
          if (await this.shouldRetry(noneRetryKey)) {
            await this.sandboxService.updateSandboxBackupState(sandbox.id, BackupState.PENDING)
          } else {
            this.logger.error(`Backup for sandbox ${sandbox.id} failed: runner repeatedly reports no backup state`)
            await this.sandboxService.updateSandboxBackupState(
              sandbox.id,
              BackupState.ERROR,
              undefined,
              undefined,
              'Backup failed: runner repeatedly reports no backup state',
            )
          }
          break
        }
        // If still in progress or any other state, do nothing and wait for next sync
      }
    } catch (error) {
      const { recoverable, errorReason } = sanitizeSandboxError(error)
      const isArchiveFlow = sandbox.desiredState === SandboxDesiredState.ARCHIVED
      const isOnDrainingRunner = await this.runnerIsDraining(sandbox)
      await this.sandboxService.updateSandboxBackupState(
        sandbox.id,
        BackupState.ERROR,
        undefined,
        undefined,
        errorReason,
        recoverable && (isArchiveFlow || isOnDrainingRunner),
      )
      await this.markErroredIfDraining(sandbox, errorReason, recoverable, isOnDrainingRunner)
      throw error
    }
  }

  private async deleteSandboxBackupRepositoryFromRegistry(sandbox: Sandbox): Promise<void> {
    const registry = await this.dockerRegistryService.findOne(sandbox.backupRegistryId)

    try {
      await this.dockerRegistryService.deleteSandboxRepository(sandbox.id, registry)
    } catch (error) {
      this.logger.error(
        `Failed to delete backup repository ${sandbox.id} from registry ${registry.id}:`,
        fromAxiosError(error),
      )
    }
  }

  /**
   * Transition a backup from PENDING to IN_PROGRESS without clobbering a terminal state
   * that a concurrent completion may have already written.
   *
   * On v2 runners a CREATE_BACKUP job can be claimed and completed (or fail fast, e.g.
   * "No such container") within milliseconds of being created. The fire-and-forget job
   * completion handler then writes Completed/Error. An unconditional write here would
   * overwrite that terminal state back to InProgress and strand the sandbox forever
   * (the job is already terminal, so neither the stale-job sweep nor the in-progress
   * poller would ever resolve it). The guarded update is a no-op once the sandbox has
   * left PENDING.
   */
  private async markBackupInProgressIfPending(sandboxId: string): Promise<void> {
    try {
      await this.sandboxRepository.updateWhere(sandboxId, {
        updateData: { backupState: BackupState.IN_PROGRESS },
        whereCondition: { backupState: BackupState.PENDING },
      })
    } catch (error) {
      if (error instanceof SandboxConflictError) {
        // Backup already left PENDING (e.g. completion handler set Completed/Error). Don't clobber it.
        this.logger.debug(`Skipping InProgress transition for sandbox ${sandboxId}: backup no longer PENDING`)
        return
      }
      throw error
    }
  }

  private async handlePendingBackup(sandbox: Sandbox): Promise<void> {
    const lockKey = `runner-${sandbox.runnerId}-backup-lock`
    try {
      await this.redisLockProvider.waitForLock(lockKey, 10)

      const backupsInProgress = await this.sandboxRepository.count({
        where: {
          runnerId: sandbox.runnerId,
          backupState: BackupState.IN_PROGRESS,
        },
      })
      if (backupsInProgress >= this.configService.getOrThrow('maxConcurrentBackupsPerRunner')) {
        return
      }

      const registry = await this.dockerRegistryService.findOne(sandbox.backupRegistryId)
      if (!registry) {
        throw new Error('Registry not found')
      }

      const runner = await this.runnerService.findOneOrFail(sandbox.runnerId)
      const runnerAdapter = await this.runnerAdapterFactory.create(runner)

      //  check if backup is already in progress on the runner
      const runnerSandbox = await runnerAdapter.sandboxInfo(sandbox.id)
      if (runnerSandbox.backupState === BackupState.IN_PROGRESS) {
        await this.markBackupInProgressIfPending(sandbox.id)
        return
      }

      // Initiate backup on runner
      await runnerAdapter.createBackup(sandbox, sandbox.backupSnapshot, registry)

      await this.markBackupInProgressIfPending(sandbox.id)
    } catch (error) {
      if (error.response?.status === 400 && error.response?.data?.message.includes('A backup is already in progress')) {
        await this.markBackupInProgressIfPending(sandbox.id)
        return
      }
      const { recoverable, errorReason } = sanitizeSandboxError(error)
      const isArchiveFlow = sandbox.desiredState === SandboxDesiredState.ARCHIVED
      const isOnDrainingRunner = await this.runnerIsDraining(sandbox)
      await this.sandboxService.updateSandboxBackupState(
        sandbox.id,
        BackupState.ERROR,
        undefined,
        undefined,
        errorReason,
        recoverable && (isArchiveFlow || isOnDrainingRunner),
      )
      await this.markErroredIfDraining(sandbox, errorReason, recoverable, isOnDrainingRunner)
      throw error
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @Cron(CronExpression.EVERY_10_MINUTES, { name: 'retry-errored-backups', waitForCompletion: true })
  @TrackJobExecution()
  @LogExecution('retry-errored-backups')
  @WithInstrumentation()
  async retryErroredBackups(): Promise<void> {
    const retryIntervalHours = this.configService.getOrThrow('backupRetryIntervalHours')
    if (retryIntervalHours <= 0) {
      return
    }

    const lockKey = 'retry-errored-backups'
    const hasLock = await this.redisLockProvider.lock(lockKey, 5 * 60)
    if (!hasLock) {
      return
    }

    try {
      const cutoff = new Date(Date.now() - retryIntervalHours * 60 * 60 * 1000)

      // Filter eligible (transient) errors in SQL so each run only selects retryable sandboxes
      // instead of pulling a random batch that might be mostly non-retryable.
      const sandboxes = await this.sandboxRepository
        .createQueryBuilder('sandbox')
        .addSelect('RANDOM()', 'rand')
        .where('sandbox.backupState = :error', { error: BackupState.ERROR })
        .andWhere('sandbox.desiredState != :destroyed', { destroyed: SandboxDesiredState.DESTROYED })
        .andWhere('sandbox.updatedAt < :cutoff', { cutoff })
        .andWhere(
          new Brackets((qb) => {
            BACKUP_RETRY_ERROR_SUBSTRINGS.forEach((substring, index) => {
              const param = `retrySubstring${index}`
              qb.orWhere(`sandbox.backupErrorReason ILIKE :${param}`, { [param]: `%${substring}%` })
            })
          }),
        )
        // Random ordering avoids head-of-line blocking on the same sandboxes that keep failing.
        .orderBy('rand')
        .take(100)
        .getMany()

      const results = await Promise.allSettled(
        sandboxes.map(async (s) => {
          // Take the same per-sandbox lock used by the rest of the backup workflows to avoid
          // racing with in-flight backup processing and causing state flapping.
          const sandboxLockKey = `sandbox-backup-${s.id}`
          const hasSandboxLock = await this.redisLockProvider.lock(sandboxLockKey, 60)
          if (!hasSandboxLock) {
            return
          }

          try {
            // Re-fetch the latest state inside the lock; only reset if it is still errored.
            const sandbox = await this.sandboxRepository.findOneByOrFail({ id: s.id })
            if (sandbox.backupState !== BackupState.ERROR) {
              return
            }

            this.logger.log(`Retrying backup for sandbox ${sandbox.id} (error: ${sandbox.backupErrorReason})`)
            await this.sandboxService.updateSandboxBackupState(sandbox.id, BackupState.NONE)
          } finally {
            await this.redisLockProvider.unlock(sandboxLockKey)
          }
        }),
      )

      results.forEach((result, index) => {
        if (result.status === 'rejected') {
          this.logger.error(
            `Failed to reset backup state for sandbox ${sandboxes[index].id}:`,
            fromAxiosError(result.reason),
          )
        }
      })
    } catch (error) {
      this.logger.error('Error retrying errored backups:', error)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @OnEvent(SandboxEvents.ARCHIVED)
  @TrackJobExecution()
  private async handleSandboxArchivedEvent(event: SandboxArchivedEvent) {
    this.setBackupPending(event.sandbox)
  }

  @OnEvent(SandboxEvents.DESTROYED)
  @TrackJobExecution()
  private async handleSandboxDestroyedEvent(event: SandboxDestroyedEvent) {
    this.deleteSandboxBackupRepositoryFromRegistry(event.sandbox)
  }

  @OnEvent(SandboxEvents.BACKUP_CREATED)
  @TrackJobExecution()
  private async handleSandboxBackupCreatedEvent(event: SandboxBackupCreatedEvent) {
    this.setBackupPending(event.sandbox)
  }
}
