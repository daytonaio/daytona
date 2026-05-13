/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { Sandbox } from '../../entities/sandbox.entity'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { DONT_SYNC_AGAIN, SandboxAction, SyncState, SYNC_AGAIN } from './sandbox.action'
import { RunnerState } from '../../enums/runner-state.enum'
import { RunnerService } from '../../services/runner.service'
import { RunnerAdapterFactory } from '../../runner-adapter/runnerAdapter'
import { SandboxRepository } from '../../repositories/sandbox.repository'
import { LockCode, RedisLockProvider } from '../../common/redis-lock.provider'
import { WithSpan } from '../../../common/decorators/otel.decorator'
import { SandboxVolumeMountService } from '../../services/sandbox-volume-mount.service'

@Injectable()
export class SandboxDestroyAction extends SandboxAction {
  private readonly destroyLogger = new Logger(SandboxDestroyAction.name)

  constructor(
    protected runnerService: RunnerService,
    protected runnerAdapterFactory: RunnerAdapterFactory,
    protected sandboxRepository: SandboxRepository,
    protected redisLockProvider: RedisLockProvider,
    private readonly sandboxVolumeMountService: SandboxVolumeMountService,
  ) {
    super(runnerService, runnerAdapterFactory, sandboxRepository, redisLockProvider)
  }

  // Revokes every per-sandbox layered mount token and drops the
  // `sandbox_volume` rows so subsequent reattach generates fresh tokens.
  // Best-effort: a failure here logs and continues — the rows are removed
  // either way and orphaned tokens are at worst a cleanup concern, not a
  // correctness one.
  private async detachVolumesBestEffort(sandboxId: string): Promise<void> {
    try {
      await this.sandboxVolumeMountService.detachVolumesFromSandbox(sandboxId)
    } catch (error) {
      this.destroyLogger.warn(`Failed to detach layered volumes for sandbox ${sandboxId}: ${error?.message ?? error}`)
    }
  }

  @WithSpan()
  async run(sandbox: Sandbox, lockCode: LockCode): Promise<SyncState> {
    if (sandbox.state === SandboxState.DESTROYED) {
      return DONT_SYNC_AGAIN
    }

    if (sandbox.state === SandboxState.ARCHIVED || sandbox.state === SandboxState.PENDING_BUILD) {
      await this.detachVolumesBestEffort(sandbox.id)
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

      if (sandboxInfo.state === SandboxState.DESTROYED) {
        await this.detachVolumesBestEffort(sandbox.id)
        await this.updateSandboxState(sandbox, SandboxState.DESTROYED, lockCode)
        return DONT_SYNC_AGAIN
      }

      if (sandbox.state !== SandboxState.DESTROYING) {
        await runnerAdapter.destroySandbox(sandbox.id)
        await this.updateSandboxState(sandbox, SandboxState.DESTROYING, lockCode)
      }

      return SYNC_AGAIN
    } catch (error) {
      //  if the sandbox is not found on runner, it is already destroyed
      if (error.response?.status === 404 || error.statusCode === 404) {
        await this.detachVolumesBestEffort(sandbox.id)
        await this.updateSandboxState(sandbox, SandboxState.DESTROYED, lockCode)
        return DONT_SYNC_AGAIN
      }

      throw error
    }
  }
}
