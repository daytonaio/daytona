/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { Sandbox } from '../../entities/sandbox.entity'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { DONT_SYNC_AGAIN, SandboxAction, SyncState } from './sandbox.action'
import { BackupState } from '../../enums/backup-state.enum'
import { Repository } from 'typeorm'
import { LockCode, RedisLockProvider } from '../../common/redis-lock.provider'
import { RunnerService } from '../../services/runner.service'
import { InjectRepository } from '@nestjs/typeorm'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { RunnerAdapterFactory } from '../../runner-adapter/runnerAdapter'
import { RunnerState } from '../../enums/runner-state.enum'

@Injectable()
export class SandboxArchiveAction extends SandboxAction {
  constructor(
    protected runnerService: RunnerService,
    protected runnerAdapterFactory: RunnerAdapterFactory,
    @InjectRepository(Sandbox)
    protected sandboxRepository: Repository<Sandbox>,
    protected readonly redisLockProvider: RedisLockProvider,
    @InjectRedis() private readonly redis: Redis,
  ) {
    super(runnerService, runnerAdapterFactory, sandboxRepository, redisLockProvider)
  }

  async run(sandbox: Sandbox, lockCode: LockCode): Promise<SyncState> {
    const lockKey = 'archive-lock-' + sandbox.runnerId
    if (!(await this.redisLockProvider.lock(lockKey, 10))) {
      return DONT_SYNC_AGAIN
    }

    // Only proceed with archiving if the sandbox is in STOPPED or ARCHIVING state.
    // For all other states, do not proceed with archiving.
    if (sandbox.state !== SandboxState.STOPPED && sandbox.state !== SandboxState.ARCHIVING) {
      return DONT_SYNC_AGAIN
    }

    await this.redisLockProvider.unlock(lockKey)

    //  if the backup state is error, we need to retry the backup
    if (sandbox.backupState === BackupState.ERROR) {
      const archiveErrorRetryKey = 'archive-error-retry-' + sandbox.id
      const archiveErrorRetryCountRaw = await this.redis.get(archiveErrorRetryKey)
      const archiveErrorRetryCount = archiveErrorRetryCountRaw ? parseInt(archiveErrorRetryCountRaw) : 0
      //  if the archive error retry count is greater than 3, we need to mark the sandbox as error
      if (archiveErrorRetryCount > 3) {
        await this.updateSandboxState(
          sandbox.id,
          SandboxState.ERROR,
          lockCode,
          undefined,
          'Failed to archive sandbox after 3 retries',
        )
        await this.redis.del(archiveErrorRetryKey)
        return DONT_SYNC_AGAIN
      }
      await this.redis.setex('archive-error-retry-' + sandbox.id, 720, String(archiveErrorRetryCount + 1))

      //  reset the backup state to pending to retry the backup
      await this.sandboxRepository.update(sandbox.id, {
        backupState: BackupState.PENDING,
      })

      return DONT_SYNC_AGAIN
    }

    if (sandbox.backupState !== BackupState.COMPLETED) {
      return DONT_SYNC_AGAIN
    }

    //  when the backup is completed, destroy the sandbox on the runner
    //  and deassociate the sandbox from the runner
    const runner = await this.runnerService.findOne(sandbox.runnerId)
    if (runner.state !== RunnerState.READY) {
      return DONT_SYNC_AGAIN
    }

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    try {
      await runnerAdapter.destroySandbox(sandbox.id)
    } catch (error) {
      //  log errors other than sandbox not found
      if (error.response?.status !== 404) {
        this.logger.error(`Failed to destroy sandbox ${sandbox.id} on runner ${runner.id}:`, error)
        throw error
      }
    }

    await this.updateSandboxState(sandbox.id, SandboxState.ARCHIVED, lockCode, null)
    return DONT_SYNC_AGAIN
  }
}
