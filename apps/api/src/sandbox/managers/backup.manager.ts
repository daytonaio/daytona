/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, OnApplicationShutdown } from '@nestjs/common'
import { Cron, CronExpression } from '@nestjs/schedule'
import { In, IsNull, LessThan, Not, Or } from 'typeorm'
import { Sandbox } from '../entities/sandbox.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { RunnerService } from '../services/runner.service'
import { RunnerState } from '../enums/runner-state.enum'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { BackupState } from '../enums/backup-state.enum'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../constants/sandbox.constants'
import { fromAxiosError } from '../../common/utils/from-axios-error'
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
              backupState: In([BackupState.NONE, BackupState.COMPLETED]),
              lastBackupAt: Or(IsNull(), LessThan(new Date(Date.now() - 1 * 60 * 60 * 1000))),
              autoDeleteInterval: Not(0),
            },
            order: {
              lastBackupAt: 'ASC',
            },
            //  todo: increase this number when backup is stable
            take: 10,
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

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'check-backup-states' })
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
      const sandboxes = await this.sandboxRepository
        .createQueryBuilder('sandbox')
        .innerJoin('runner', 'r', 'r.id = sandbox.runnerId')
        .where('sandbox.state IN (:...states)', {
          states: [SandboxState.ARCHIVING, SandboxState.STARTED, SandboxState.STOPPED],
        })
        .andWhere('sandbox.backupState IN (:...backupStates)', {
          backupStates: [BackupState.PENDING, BackupState.IN_PROGRESS],
        })
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
        .setParameters({
          archiving: SandboxState.ARCHIVING,
          stopped: SandboxState.STOPPED,
          started: SandboxState.STARTED,
        })
        .orderBy('state_priority', 'ASC')
        .addOrderBy('sandbox.lastBackupAt', 'ASC', 'NULLS FIRST') // Process sandboxes with no backups first
        .addOrderBy('sandbox.createdAt', 'ASC') // For equal lastBackupAt, process older sandboxes first
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
                await this.sandboxService.updateSandboxBackupState(
                  sandbox.id,
                  BackupState.ERROR,
                  undefined,
                  undefined,
                  fromAxiosError(error).message,
                )
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
                await this.sandboxService.updateSandboxBackupState(
                  sandbox.id,
                  BackupState.ERROR,
                  undefined,
                  undefined,
                  fromAxiosError(error).message,
                )
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

  private async checkBackupProgress(sandbox: Sandbox): Promise<void> {
    try {
      const runner = await this.runnerService.findOneOrFail(sandbox.runnerId)
      const runnerAdapter = await this.runnerAdapterFactory.create(runner)

      // Get sandbox info from runner
      const sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)

      switch (sandboxInfo.backupState) {
        case BackupState.COMPLETED: {
          await this.sandboxService.updateSandboxBackupState(sandbox.id, BackupState.COMPLETED)
          break
        }
        case BackupState.ERROR: {
          await this.sandboxService.updateSandboxBackupState(
            sandbox.id,
            BackupState.ERROR,
            undefined,
            undefined,
            sandboxInfo.backupErrorReason,
          )
          break
        }
        // If backup state is none, retry the backup process by setting the backup state to pending
        // This can happen if the runner is restarted or the operation is cancelled
        case BackupState.NONE: {
          await this.sandboxService.updateSandboxBackupState(sandbox.id, BackupState.PENDING)
          break
        }
        // If still in progress or any other state, do nothing and wait for next sync
      }
    } catch (error) {
      await this.sandboxService.updateSandboxBackupState(
        sandbox.id,
        BackupState.ERROR,
        undefined,
        undefined,
        fromAxiosError(error).message,
      )
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
      if (runnerSandbox.backupState?.toUpperCase() === 'IN_PROGRESS') {
        await this.sandboxService.updateSandboxBackupState(sandbox.id, BackupState.IN_PROGRESS)
        return
      }

      // Initiate backup on runner
      await runnerAdapter.createBackup(sandbox, sandbox.backupSnapshot, registry)

      await this.sandboxService.updateSandboxBackupState(sandbox.id, BackupState.IN_PROGRESS)
    } catch (error) {
      if (error.response?.status === 400 && error.response?.data?.message.includes('A backup is already in progress')) {
        await this.sandboxService.updateSandboxBackupState(sandbox.id, BackupState.IN_PROGRESS)
        return
      }
      await this.sandboxService.updateSandboxBackupState(
        sandbox.id,
        BackupState.ERROR,
        undefined,
        undefined,
        fromAxiosError(error).message,
      )
      throw error
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
