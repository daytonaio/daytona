/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { In, Not, Repository } from 'typeorm'
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

@Injectable()
export class BackupManager {
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
  ) {}

  //  on init
  async onApplicationBootstrap() {
    await this.adHocBackupCheck()
  }

  //  todo: make frequency configurable or more efficient
  @Cron(CronExpression.EVERY_5_MINUTES, { name: 'ad-hoc-backup-check' })
  async adHocBackupCheck(): Promise<void> {
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
            state: In([SandboxState.STARTED, SandboxState.ARCHIVING]),
            backupState: In([BackupState.NONE, BackupState.COMPLETED]),
          },
          order: {
            lastBackupAt: 'ASC',
          },
          //  todo: increase this number when backup is stable
          take: 10,
        })

        await Promise.all(
          sandboxes
            .filter(
              (sandbox) => !sandbox.lastBackupAt || sandbox.lastBackupAt < new Date(Date.now() - 1 * 60 * 60 * 1000),
            )
            .map(async (sandbox) => {
              const lockKey = `sandbox-backup-${sandbox.id}`
              const hasLock = await this.redisLockProvider.lock(lockKey, 60)
              if (!hasLock) {
                return
              }

              try {
                //  todo: remove the catch handler asap
                await this.startBackupCreate(sandbox.id).catch((error) => {
                  if (error instanceof BadRequestError && error.message === 'A backup is already in progress') {
                    return
                  }
                  this.logger.error(`Failed to create backup for sandbox ${sandbox.id}:`, fromAxiosError(error))
                })
              } catch (error) {
                this.logger.error(`Error processing stop state for sandbox ${sandbox.id}:`, fromAxiosError(error))
              }
            }),
        )
      }),
    )
  }

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'sync-backup-states' }) // Run every 10 seconds
  async syncBackupStates(): Promise<void> {
    //  lock the sync to only run one instance at a time
    const lockKey = 'sync-backup-states'
    const hasLock = await this.redisLockProvider.lock(lockKey, 10)
    if (!hasLock) {
      return
    }

    const sandboxes = await this.sandboxRepository.find({
      where: {
        state: In([SandboxState.STARTED, SandboxState.STOPPED, SandboxState.ARCHIVING]),
        backupState: In([BackupState.PENDING, BackupState.IN_PROGRESS]),
      },
    })

    await Promise.all(
      sandboxes.map(async (s) => {
        const lockKey = `sandbox-backup-${s.id}`
        const hasLock = await this.redisLockProvider.lock(lockKey, 60)
        if (!hasLock) {
          return
        }

        const runner = await this.runnerService.findOne(s.runnerId)
        if (runner.state !== RunnerState.READY) {
          return
        }

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
            await this.updateWorkspacBackupState(sandbox.id, BackupState.ERROR)
          } else {
            await this.redis.setex(errorRetryKey, 300, errorRetryCount + 1)
          }
        }
      }),
    ).catch((ex) => {
      this.logger.error(ex)
    })
  }

  async startBackupCreate(sandboxId: string): Promise<void> {
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
    const backupSnapshot = `${registry.url}/${registry.project}/backup-${sandbox.id}:${timestamp}`

    //  if sandbox has a backup snapshot, add it to the existingBackupSnapshots array
    if (
      sandbox.lastBackupAt &&
      sandbox.backupSnapshot &&
      [BackupState.NONE, BackupState.COMPLETED].includes(sandbox.backupState)
    ) {
      sandbox.existingBackupSnapshots.push({
        snapshotName: sandbox.backupSnapshot,
        createdAt: sandbox.lastBackupAt,
      })
    }
    const existingBackupSnapshots = sandbox.existingBackupSnapshots
    existingBackupSnapshots.push({
      snapshotName: backupSnapshot,
      createdAt: new Date(),
    })

    const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
      id: sandbox.id,
    })
    sandboxToUpdate.existingBackupSnapshots = existingBackupSnapshots
    sandboxToUpdate.backupState = BackupState.PENDING
    sandboxToUpdate.backupRegistryId = registry.id
    sandboxToUpdate.backupSnapshot = backupSnapshot
    await this.sandboxRepository.save(sandboxToUpdate)
  }

  private async checkBackupProgress(sandbox: Sandbox): Promise<void> {
    try {
      const runner = await this.runnerService.findOne(sandbox.runnerId)
      const runnerAdapter = await this.runnerAdapterFactory.create(runner)

      // Get sandbox info from runner
      const sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)

      switch (sandboxInfo.backupState?.toUpperCase()) {
        case 'COMPLETED': {
          sandbox.backupState = BackupState.COMPLETED
          sandbox.lastBackupAt = new Date()
          const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
            id: sandbox.id,
          })
          sandboxToUpdate.backupState = BackupState.COMPLETED
          sandboxToUpdate.lastBackupAt = new Date()
          await this.sandboxRepository.save(sandboxToUpdate)
          break
        }
        case 'FAILED':
        case 'ERROR': {
          await this.updateWorkspacBackupState(sandbox.id, BackupState.ERROR)
          break
        }

        // If still in progress or any other state, do nothing and wait for next sync
      }
    } catch (error) {
      await this.updateWorkspacBackupState(sandbox.id, BackupState.ERROR)
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
    try {
      const registry = await this.dockerRegistryService.findOne(sandbox.backupRegistryId)
      if (!registry) {
        throw new Error('Registry not found')
      }

      const runner = await this.runnerService.findOne(sandbox.runnerId)
      const runnerAdapter = await this.runnerAdapterFactory.create(runner)

      //  check if backup is already in progress on the runner
      const runnerSandbox = await runnerAdapter.sandboxInfo(sandbox.id)
      if (runnerSandbox.backupState?.toUpperCase() === 'IN_PROGRESS') {
        return
      }

      // Initiate backup on runner
      await runnerAdapter.createBackup(sandbox, sandbox.backupSnapshot, registry)

      await this.updateWorkspacBackupState(sandbox.id, BackupState.IN_PROGRESS)
    } catch (error) {
      if (error.response?.status === 400 && error.response?.data?.message.includes('A backup is already in progress')) {
        await this.updateWorkspacBackupState(sandbox.id, BackupState.IN_PROGRESS)
        return
      }
      await this.updateWorkspacBackupState(sandbox.id, BackupState.ERROR)
      throw error
    }
  }

  @Cron(CronExpression.EVERY_30_SECONDS, { name: 'sync-stop-state-create-backups' }) // Run every 30 seconds
  async syncStopStateCreateBackups(): Promise<void> {
    const lockKey = 'sync-stop-state-create-backups'
    const hasLock = await this.redisLockProvider.lock(lockKey, 30)
    if (!hasLock) {
      return
    }

    const sandboxes = await this.sandboxRepository.find({
      where: {
        state: In([SandboxState.STOPPED, SandboxState.ARCHIVING]),
        backupState: In([BackupState.NONE]),
      },
      //  todo: increase this number when auto-stop is stable
      take: 10,
    })

    await Promise.all(
      sandboxes
        .filter((sandbox) => sandbox.runnerId !== null)
        .map(async (sandbox) => {
          const lockKey = `sandbox-backup-${sandbox.id}`
          const hasLock = await this.redisLockProvider.lock(lockKey, 30)
          if (!hasLock) {
            return
          }

          const runner = await this.runnerService.findOne(sandbox.runnerId)
          if (runner.state !== RunnerState.READY) {
            return
          }

          //  TODO: this should be revisited
          //  an error should be handled better and not just logged
          try {
            //  todo: remove the catch handler asap
            await this.startBackupCreate(sandbox.id).catch((error) => {
              if (error instanceof BadRequestError && error.message === 'A backup is already in progress') {
                return
              }
              this.logger.error(`Failed to create backup for sandbox ${sandbox.id}:`, fromAxiosError(error))
            })
          } catch (error) {
            this.logger.error(`Failed to create backup for sandbox ${sandbox.id}:`, fromAxiosError(error))
          }
        }),
    )
  }

  private async updateWorkspacBackupState(sandboxId: string, backupState: BackupState): Promise<void> {
    const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
      id: sandboxId,
    })
    sandboxToUpdate.backupState = backupState
    await this.sandboxRepository.save(sandboxToUpdate)
  }

  @OnEvent(SandboxEvents.ARCHIVED)
  private async handleSandboxArchivedEvent(event: SandboxArchivedEvent) {
    this.startBackupCreate(event.sandbox.id)
  }

  @OnEvent(SandboxEvents.DESTROYED)
  private async handleSandboxDestroyedEvent(event: SandboxDestroyedEvent) {
    this.deleteSandboxBackupRepositoryFromRegistry(event.sandbox)
  }

  @OnEvent(SandboxEvents.BACKUP_CREATED)
  private async handleSandboxBackupCreatedEvent(event: SandboxBackupCreatedEvent) {
    this.handlePendingBackup(event.sandbox)
  }
}
