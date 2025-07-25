/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { Sandbox } from '../../entities/sandbox.entity'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { DONT_SYNC_AGAIN, SandboxAction, SyncState, SYNC_AGAIN } from './sandbox.action'
import { BackupState } from '../../enums/backup-state.enum'
import { In, Repository } from 'typeorm'
import { RedisLockProvider } from '../../common/redis-lock.provider'
import { RunnerService } from '../../services/runner.service'
import { InjectRepository } from '@nestjs/typeorm'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { RunnerAdapterFactory } from '../../runner-adapter/runnerAdapter'
import { ToolboxService } from '../../services/toolbox.service'

@Injectable()
export class SandboxArchiveAction extends SandboxAction {
  constructor(
    protected runnerService: RunnerService,
    protected runnerAdapterFactory: RunnerAdapterFactory,
    @InjectRepository(Sandbox)
    protected sandboxRepository: Repository<Sandbox>,
    private readonly redisLockProvider: RedisLockProvider,
    @InjectRedis() private readonly redis: Redis,
    protected toolboxService: ToolboxService,
  ) {
    super(runnerService, runnerAdapterFactory, sandboxRepository, toolboxService)
  }

  async run(sandbox: Sandbox): Promise<SyncState> {
    const lockKey = 'archive-lock-' + sandbox.runnerId
    if (!(await this.redisLockProvider.lock(lockKey, 10))) {
      return DONT_SYNC_AGAIN
    }

    const inProgressOnRunner = await this.sandboxRepository.find({
      where: {
        runnerId: sandbox.runnerId,
        state: In([SandboxState.ARCHIVING]),
      },
      order: {
        lastActivityAt: 'DESC',
      },
      take: 100,
    })

    //  if the sandbox is already in progress, continue
    if (!inProgressOnRunner.find((s) => s.id === sandbox.id)) {
      //  max 3 sandboxes can be archived at the same time on the same runner
      //  this is to prevent the runner from being overloaded
      if (inProgressOnRunner.length > 2) {
        await this.redisLockProvider.unlock(lockKey)
        return DONT_SYNC_AGAIN
      }
    }

    switch (sandbox.state) {
      case SandboxState.STOPPED: {
        await this.updateSandboxState(sandbox.id, SandboxState.ARCHIVING)
        //  fallthrough to archiving state
      }
      case SandboxState.ARCHIVING: {
        await this.redisLockProvider.unlock(lockKey)

        //  if the backup state is error, we need to retry the backup
        if (sandbox.backupState === BackupState.ERROR) {
          const archiveErrorRetryKey = 'archive-error-retry-' + sandbox.id
          const archiveErrorRetryCountRaw = await this.redis.get(archiveErrorRetryKey)
          const archiveErrorRetryCount = archiveErrorRetryCountRaw ? parseInt(archiveErrorRetryCountRaw) : 0
          //  if the archive error retry count is greater than 3, we need to mark the sandbox as error
          if (archiveErrorRetryCount > 3) {
            await this.updateSandboxState(sandbox.id, SandboxState.ERROR, undefined, 'Failed to archive sandbox')
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

        // Check for timeout - if more than 120 minutes since last activity
        const timeout = new Date(Date.now() - 120 * 60 * 1000)
        if (sandbox.lastActivityAt < timeout) {
          await this.updateSandboxState(sandbox.id, SandboxState.ERROR, undefined, 'Archiving operation timed out')
          return DONT_SYNC_AGAIN
        }

        if (sandbox.backupState !== BackupState.COMPLETED) {
          return DONT_SYNC_AGAIN
        }

        //  when the backup is completed, destroy the sandbox on the runner
        //  and deassociate the sandbox from the runner
        const runner = await this.runnerService.findOne(sandbox.runnerId)
        const runnerAdapter = await this.runnerAdapterFactory.create(runner)

        try {
          const sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)
          switch (sandboxInfo.state) {
            case SandboxState.DESTROYING:
              //  wait until sandbox is destroyed on runner
              return SYNC_AGAIN
            case SandboxState.DESTROYED:
              await this.updateSandboxState(sandbox.id, SandboxState.ARCHIVED, null)
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
          await this.updateSandboxState(sandbox.id, SandboxState.ARCHIVED, null)
          return DONT_SYNC_AGAIN
        }
      }
    }

    return DONT_SYNC_AGAIN
  }
}
