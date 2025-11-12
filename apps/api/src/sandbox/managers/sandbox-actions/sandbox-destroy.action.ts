/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { Sandbox } from '../../entities/sandbox.entity'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { DONT_SYNC_AGAIN, SandboxAction, SyncState } from './sandbox.action'
import { RunnerState } from '../../enums/runner-state.enum'
import { RunnerService } from '../../services/runner.service'
import { RunnerAdapterFactory } from '../../runner-adapter/runnerAdapter'
import { Repository } from 'typeorm'
import { InjectRepository } from '@nestjs/typeorm'
import { LockCode, RedisLockProvider } from '../../common/redis-lock.provider'

@Injectable()
export class SandboxDestroyAction extends SandboxAction {
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
    // Only return early if already destroyed; allow ARCHIVED sandboxes to be destroyed
    if (sandbox.state === SandboxState.DESTROYED) {
      return DONT_SYNC_AGAIN
    }

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

    await this.updateSandboxState(sandbox.id, SandboxState.DESTROYED, lockCode)
    return DONT_SYNC_AGAIN
  }
}
