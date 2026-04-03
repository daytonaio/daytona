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
import { RunnerAdapterFactory } from '../../runner-adapter/runnerAdapter'
import { SandboxRepository } from '../../repositories/sandbox.repository'
import { LockCode, RedisLockProvider } from '../../common/redis-lock.provider'
import { WithSpan } from '../../../common/decorators/otel.decorator'

@Injectable()
export class SandboxStopAction extends SandboxAction {
  constructor(
    protected runnerService: RunnerService,
    protected runnerAdapterFactory: RunnerAdapterFactory,
    protected sandboxRepository: SandboxRepository,
    protected redisLockProvider: RedisLockProvider,
  ) {
    super(runnerService, runnerAdapterFactory, sandboxRepository, redisLockProvider)
  }

  @WithSpan()
  async run(sandbox: Sandbox, lockCode: LockCode, force?: boolean): Promise<SyncState> {
    const runner = await this.runnerService.findOneOrFail(sandbox.sandboxState.runnerId)
    if (runner.state !== RunnerState.READY) {
      return DONT_SYNC_AGAIN
    }

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    if (sandbox.sandboxState.state === SandboxState.STARTED) {
      // stop sandbox
      await runnerAdapter.stopSandbox(sandbox.id, force)
      await this.updateSandboxState(sandbox, SandboxState.STOPPING, lockCode)

      //  sync states again immediately for sandbox
      return SYNC_AGAIN
    }

    if (sandbox.sandboxState.state !== SandboxState.STOPPING && sandbox.sandboxState.state !== SandboxState.ERROR) {
      return DONT_SYNC_AGAIN
    }

    const sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)

    if (sandboxInfo.state === SandboxState.STOPPED) {
      await this.updateSandboxState(
        sandbox,
        SandboxState.STOPPED,
        lockCode,
        undefined,
        undefined,
        undefined,
        BackupState.NONE,
      )
      return DONT_SYNC_AGAIN
    } else if (sandboxInfo.state === SandboxState.ERROR) {
      await this.updateSandboxState(
        sandbox,
        SandboxState.ERROR,
        lockCode,
        undefined,
        'Sandbox is in error state on runner',
      )
      return DONT_SYNC_AGAIN
    }

    return SYNC_AGAIN
  }
}
