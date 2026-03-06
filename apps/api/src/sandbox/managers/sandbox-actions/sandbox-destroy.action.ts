/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { Sandbox } from '../../entities/sandbox.entity'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { DONT_SYNC_AGAIN, SandboxAction, SyncState, SYNC_AGAIN } from './sandbox.action'
import { RunnerState } from '../../enums/runner-state.enum'
import { RunnerService } from '../../services/runner.service'
import { RunnerAdapterFactory } from '../../runner-adapter/runnerAdapter'
import { SandboxRepository } from '../../repositories/sandbox.repository'
import { LockCode, RedisLockProvider } from '../../common/redis-lock.provider'
import { WithSpan } from '../../../common/decorators/otel.decorator'

@Injectable()
export class SandboxDestroyAction extends SandboxAction {
  constructor(
    protected runnerService: RunnerService,
    protected runnerAdapterFactory: RunnerAdapterFactory,
    protected sandboxRepository: SandboxRepository,
    protected redisLockProvider: RedisLockProvider,
  ) {
    super(runnerService, runnerAdapterFactory, sandboxRepository, redisLockProvider)
  }

  @WithSpan()
  async run(sandbox: Sandbox, lockCode: LockCode): Promise<SyncState> {
    if (sandbox.state === SandboxState.DESTROYED) {
      return DONT_SYNC_AGAIN
    }

    if (sandbox.state === SandboxState.ARCHIVED || sandbox.state === SandboxState.PENDING_BUILD) {
      await this.updateSandboxState(sandbox, SandboxState.DESTROYED, lockCode)
      return DONT_SYNC_AGAIN
    }

    const runner = await this.runnerService.findOneOrFail(sandbox.runnerId)
    if (runner.state !== RunnerState.READY) {
      return DONT_SYNC_AGAIN
    }

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    try {
      const sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)
      switch (sandboxInfo.state) {
        case SandboxState.DESTROYING:
          return SYNC_AGAIN
        case SandboxState.DESTROYED: {
          await this.updateSandboxState(sandbox, SandboxState.DESTROYED, lockCode)
          return DONT_SYNC_AGAIN
        }
        default: {
          // destroy sandbox
          await runnerAdapter.destroySandbox(sandbox.id)
          await this.updateSandboxState(sandbox, SandboxState.DESTROYING, lockCode)
          return SYNC_AGAIN
        }
      }
    } catch (e) {
      //  if the sandbox is not found on runner, it is already destroyed
      if (e.response?.status !== 404 && e.statusCode !== 404) {
        throw e
      }

      await this.updateSandboxState(sandbox, SandboxState.DESTROYED, lockCode)
      return DONT_SYNC_AGAIN
    }
  }
}
