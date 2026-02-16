/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { Sandbox } from '../../entities/sandbox.entity'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { DONT_SYNC_AGAIN, SandboxAction, SyncState, SYNC_AGAIN } from './sandbox.action'
import { BackupState } from '../../enums/backup-state.enum'
import { Repository } from 'typeorm'
import { LockCode, RedisLockProvider } from '../../common/redis-lock.provider'
import { RunnerService } from '../../services/runner.service'
import { InjectRepository } from '@nestjs/typeorm'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { RunnerAdapterFactory } from '../../runner-adapter/runnerAdapter'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { SandboxEvents } from '../../constants/sandbox-events.constants'
import { SandboxBackupCreatedEvent } from '../../events/sandbox-backup-created.event'

@Injectable()
export class SandboxArchiveAction extends SandboxAction {
  constructor(
    protected runnerService: RunnerService,
    protected runnerAdapterFactory: RunnerAdapterFactory,
    @InjectRepository(Sandbox)
    protected sandboxRepository: Repository<Sandbox>,
    protected readonly redisLockProvider: RedisLockProvider,
    @InjectRedis() private readonly redis: Redis,
    private readonly eventEmitter: EventEmitter2,
  ) {
    super(runnerService, runnerAdapterFactory, sandboxRepository, redisLockProvider)
  }

  async run(sandbox: Sandbox, lockCode: LockCode): Promise<SyncState> {
    const lockKey = 'archive-lock-' + sandbox.runnerId
    if (!(await this.redisLockProvider.lock(lockKey, 10))) {
      return DONT_SYNC_AGAIN
    }

    switch (sandbox.state) {
      case SandboxState.STOPPED:
      case SandboxState.ARCHIVING:
      case SandboxState.ERROR: {
        const isFromErrorState = sandbox.state === SandboxState.ERROR

        await this.redisLockProvider.unlock(lockKey)

        //  if the backup state is error, we need to retry the backup
        if (sandbox.backupState === BackupState.ERROR) {
          const archiveErrorRetryKey = 'archive-error-retry-' + sandbox.id
          const archiveErrorRetryCountRaw = await this.redis.get(archiveErrorRetryKey)
          const archiveErrorRetryCount = archiveErrorRetryCountRaw ? parseInt(archiveErrorRetryCountRaw) : 0
          //  if the archive error retry count is greater than 3, we need to mark the sandbox as error
          if (archiveErrorRetryCount > 3) {
            // Only transition to ERROR if not already in ERROR state
            if (!isFromErrorState) {
              await this.updateSandboxState(
                sandbox.id,
                SandboxState.ERROR,
                lockCode,
                undefined,
                'Failed to archive sandbox after 3 retries',
              )
            }
            await this.redis.del(archiveErrorRetryKey)
            return DONT_SYNC_AGAIN
          }
          await this.redis.setex('archive-error-retry-' + sandbox.id, 720, String(archiveErrorRetryCount + 1))

          //  recreate the backup to retry
          this.eventEmitter.emit(SandboxEvents.BACKUP_CREATED, new SandboxBackupCreatedEvent(sandbox))

          return DONT_SYNC_AGAIN
        }

        if (sandbox.backupState !== BackupState.COMPLETED) {
          return DONT_SYNC_AGAIN
        }

        //  when the backup is completed, destroy the sandbox on the runner
        //  and deassociate the sandbox from the runner
        const runner = await this.runnerService.findOneOrFail(sandbox.runnerId)
        const runnerAdapter = await this.runnerAdapterFactory.create(runner)

        try {
          const sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)
          switch (sandboxInfo.state) {
            case SandboxState.DESTROYING:
              //  wait until sandbox is destroyed on runner
              return SYNC_AGAIN
            case SandboxState.DESTROYED:
              if (isFromErrorState) {
                this.logger.warn(`Transitioning sandbox ${sandbox.id} from ERROR to ARCHIVED state (runner draining)`)
              }
              await this.updateSandboxState(sandbox.id, SandboxState.ARCHIVED, lockCode, null)
              return DONT_SYNC_AGAIN
            default:
              await runnerAdapter.destroySandbox(sandbox.id)
              return SYNC_AGAIN
          }
        } catch (error) {
          //  fail for errors other than sandbox not found or sandbox already destroyed
          if (
            !(
              (error.response?.data?.statusCode === 400 &&
                error.response?.data?.message.includes('Sandbox already destroyed')) ||
              error.response?.status === 404
            )
          ) {
            throw error
          }
          //  if the sandbox is already destroyed, do nothing
          if (isFromErrorState) {
            this.logger.warn(`Transitioning sandbox ${sandbox.id} from ERROR to ARCHIVED state (runner draining)`)
          }
          await this.updateSandboxState(sandbox.id, SandboxState.ARCHIVED, lockCode, null)
          return DONT_SYNC_AGAIN
        }
      }
    }

    return DONT_SYNC_AGAIN
  }
}
