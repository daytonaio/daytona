/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { Sandbox } from '../../entities/sandbox.entity'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { DONT_SYNC_AGAIN, SandboxAction, SyncState, SYNC_AGAIN } from './sandbox.action'
import { BackupState } from '../../enums/backup-state.enum'
import { RunnerState } from '../../enums/runner-state.enum'
import { RunnerService } from '../../services/runner.service'
import { RunnerAdapterFactory, RunnerSandboxInfo } from '../../runner-adapter/runnerAdapter'
import { Repository } from 'typeorm'
import { InjectRepository } from '@nestjs/typeorm'
import { LockCode, RedisLockProvider } from '../../common/redis-lock.provider'

@Injectable()
export class SandboxStopAction extends SandboxAction {
  constructor(
    protected runnerService: RunnerService,
    protected runnerAdapterFactory: RunnerAdapterFactory,
    @InjectRepository(Sandbox)
    protected sandboxRepository: Repository<Sandbox>,
    protected redisLockProvider: RedisLockProvider,
  ) {
    super(runnerService, runnerAdapterFactory, sandboxRepository, redisLockProvider)
  }

  async run(sandbox: Sandbox, lockCode: LockCode): Promise<SyncState> {
    const runner = await this.runnerService.findOne(sandbox.runnerId)
    if (runner.state !== RunnerState.READY) {
      return DONT_SYNC_AGAIN
    }

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    switch (sandbox.state) {
      case SandboxState.STARTED: {
        // stop sandbox
        await runnerAdapter.stopSandbox(sandbox.id)
        await this.updateSandboxState(sandbox.id, SandboxState.STOPPING, lockCode)
        //  sync states again immediately for sandbox
        return SYNC_AGAIN
      }
      case SandboxState.STOPPING: {
        // check if sandbox is stopped
        let sandboxInfo: RunnerSandboxInfo
        try {
          sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)
        } catch (error) {
          if (error.response?.status === 404) {
            await this.updateSandboxState(sandbox.id, SandboxState.ERROR, lockCode, undefined, error)
            return DONT_SYNC_AGAIN
          }
          throw error
        }

        switch (sandboxInfo.state) {
          case SandboxState.STOPPED: {
            await this.updateSandboxState(
              sandbox.id,
              SandboxState.STOPPED,
              lockCode,
              undefined,
              undefined,
              undefined,
              BackupState.NONE,
            )
            return DONT_SYNC_AGAIN
          }
          case SandboxState.ERROR: {
            await this.updateSandboxState(
              sandbox.id,
              SandboxState.ERROR,
              lockCode,
              undefined,
              'Sandbox is in error state on runner',
            )
            return DONT_SYNC_AGAIN
          }
        }
        return SYNC_AGAIN
      }
      case SandboxState.ERROR: {
        let sandboxInfo: RunnerSandboxInfo
        try {
          sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)
        } catch (error) {
          if (error.response?.status === 404) {
            await this.updateSandboxState(sandbox.id, SandboxState.ERROR, lockCode, undefined, error)
            return DONT_SYNC_AGAIN
          }
          throw error
        }

        if (sandboxInfo.state === SandboxState.STOPPED) {
          await this.updateSandboxState(sandbox.id, SandboxState.STOPPED, lockCode)
        }
      }
    }

    return DONT_SYNC_AGAIN
  }
}
