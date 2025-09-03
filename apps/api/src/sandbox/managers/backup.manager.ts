/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, OnApplicationShutdown } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { In, IsNull, LessThan, Not, Or, Repository } from 'typeorm'
import { Sandbox } from '../entities/sandbox.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { RunnerService } from '../services/runner.service'
import { RunnerState } from '../enums/runner-state.enum'
import { ResourceNotFoundError } from '../../exceptions/not-found.exception'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { BackupState } from '../enums/backup-state.enum'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../constants/sandbox.constants'
import { DockerProvider } from '../docker/docker-provider'
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

@Injectable()
export class BackupManager implements TrackableJobExecutions, OnApplicationShutdown {
  activeJobs = new Set<string>()

  private readonly logger = new Logger(BackupManager.name)

  constructor(
    @InjectRepository(Sandbox)
    private readonly sandboxRepository: Repository<Sandbox>,
    private readonly runnerService: RunnerService,
    private readonly runnerAdapterFactory: RunnerAdapterFactory,
    private readonly dockerRegistryService: DockerRegistryService,
    @InjectRedis() private readonly redis: Redis,
    private readonly dockerProvider: DockerProvider,
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
                await this.setBackupPending(sandbox.id).catch((error) => {
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
        .orderBy('sandbox.lastBackupAt', 'ASC')
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
                await this.updateSandboxBackupState(sandbox.id, BackupState.ERROR, fromAxiosError(error).message)
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

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'sync-stop-state-create-backups' })
  @TrackJobExecution()
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
              await this.setBackupPending(sandbox.id)
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

  async setBackupPending(sandboxId: string): Promise<void> {
    const sandbox = await this.sandboxRepository.findOneByOrFail({
      id: sandboxId,
    })

    if (!sandbox) {
      throw new ResourceNotFoundError('Sandbox not found')
    }

    if (sandbox.backupState === BackupState.COMPLETED) {
      return
    }

    // Allow backups for STARTED sandboxes or STOPPED sandboxes with runnerId
    if (
      !(
        sandbox.state === SandboxState.STARTED ||
        sandbox.state === SandboxState.ARCHIVING ||
        (sandbox.state === SandboxState.STOPPED && sandbox.runnerId)
      )
    ) {
      throw new BadRequestError('Sandbox must be started or stopped with assigned runner to create a backup')
    }

    if (sandbox.backupState === BackupState.IN_PROGRESS || sandbox.backupState === BackupState.PENDING) {
      return
    }

    // Get default registry
    const registry = await this.dockerRegistryService.getDefaultInternalRegistry()
    if (!registry) {
      throw new BadRequestError('No default registry configured')
    }

    // Generate backup snapshot name
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-')
    const backupSnapshot = `${registry.url.replace('https://', '').replace('http://', '')}/${registry.project}/backup-${sandbox.id}:${timestamp}`

    const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
      id: sandbox.id,
    })
    sandboxToUpdate.setBackupState(BackupState.PENDING, backupSnapshot, registry.id)
    await this.sandboxRepository.save(sandboxToUpdate)
  }

  private async checkBackupProgress(sandbox: Sandbox): Promise<void> {
    try {
      const runner = await this.runnerService.findOne(sandbox.runnerId)
      const runnerAdapter = await this.runnerAdapterFactory.create(runner)

      // Get sandbox info from runner
      const sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)

      switch (sandboxInfo.backupState) {
        case BackupState.COMPLETED: {
          sandbox.setBackupState(BackupState.COMPLETED)
          const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
            id: sandbox.id,
          })
          sandboxToUpdate.setBackupState(BackupState.COMPLETED)
          await this.sandboxRepository.save(sandboxToUpdate)
          break
        }
        case BackupState.ERROR: {
          await this.updateSandboxBackupState(sandbox.id, BackupState.ERROR, sandboxInfo.backupErrorReason)
          break
        }
        // If backup state is none, retry the backup process by setting the backup state to pending
        // This can happen if the runner is restarted or the operation is cancelled
        case BackupState.NONE: {
          await this.updateSandboxBackupState(sandbox.id, BackupState.PENDING)
          break
        }
        // If still in progress or any other state, do nothing and wait for next sync
      }
    } catch (error) {
      await this.updateSandboxBackupState(sandbox.id, BackupState.ERROR, fromAxiosError(error).message)
      throw error
    }
  }

  private async deleteSandboxBackupRepositoryFromRegistry(sandbox: Sandbox): Promise<void> {
    const registry = await this.dockerRegistryService.findOne(sandbox.backupRegistryId)

    try {
      await this.dockerProvider.deleteSandboxRepository(sandbox.id, registry)
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

      const runner = await this.runnerService.findOne(sandbox.runnerId)
      const runnerAdapter = await this.runnerAdapterFactory.create(runner)

      //  check if backup is already in progress on the runner
      const runnerSandbox = await runnerAdapter.sandboxInfo(sandbox.id)
      if (runnerSandbox.backupState?.toUpperCase() === 'IN_PROGRESS') {
        await this.updateSandboxBackupState(sandbox.id, BackupState.IN_PROGRESS)
        return
      }

      // Initiate backup on runner
      await runnerAdapter.createBackup(sandbox, sandbox.backupSnapshot, registry)

      await this.updateSandboxBackupState(sandbox.id, BackupState.IN_PROGRESS)
    } catch (error) {
      if (error.response?.status === 400 && error.response?.data?.message.includes('A backup is already in progress')) {
        await this.updateSandboxBackupState(sandbox.id, BackupState.IN_PROGRESS)
        return
      }
      await this.updateSandboxBackupState(sandbox.id, BackupState.ERROR, fromAxiosError(error).message)
      throw error
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  private async updateSandboxBackupState(
    sandboxId: string,
    backupState: BackupState,
    backupErrorReason?: string | null,
  ): Promise<void> {
    const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
      id: sandboxId,
    })
    sandboxToUpdate.setBackupState(backupState, undefined, undefined, backupErrorReason)
    await this.sandboxRepository.save(sandboxToUpdate)
  }

  @OnEvent(SandboxEvents.ARCHIVED)
  @TrackJobExecution()
  private async handleSandboxArchivedEvent(event: SandboxArchivedEvent) {
    this.setBackupPending(event.sandbox.id)
  }

  @OnEvent(SandboxEvents.DESTROYED)
  @TrackJobExecution()
  private async handleSandboxDestroyedEvent(event: SandboxDestroyedEvent) {
    this.deleteSandboxBackupRepositoryFromRegistry(event.sandbox)
  }

  @OnEvent(SandboxEvents.BACKUP_CREATED)
  @TrackJobExecution()
  private async handleSandboxBackupCreatedEvent(event: SandboxBackupCreatedEvent) {
    this.handlePendingBackup(event.sandbox)
  }
}
