/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { RunnerService } from '../../services/runner.service'
import { RunnerAdapterFactory } from '../../runner-adapter/runnerAdapter'
import { Sandbox } from '../../entities/sandbox.entity'
import { SandboxRepository } from '../../repositories/sandbox.repository'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { BackupState } from '../../enums/backup-state.enum'
import { getStateChangeLockKey } from '../../utils/lock-key.util'
import { LockCode, RedisLockProvider } from '../../common/redis-lock.provider'

export const SYNC_AGAIN = 'sync-again'
export const DONT_SYNC_AGAIN = 'dont-sync-again'
export type SyncState = typeof SYNC_AGAIN | typeof DONT_SYNC_AGAIN

@Injectable()
export abstract class SandboxAction {
  protected readonly logger = new Logger(SandboxAction.name)

  constructor(
    protected readonly runnerService: RunnerService,
    protected runnerAdapterFactory: RunnerAdapterFactory,
    protected readonly sandboxRepository: SandboxRepository,
    protected readonly redisLockProvider: RedisLockProvider,
  ) {}

  abstract run(sandbox: Sandbox, lockCode: LockCode): Promise<SyncState>

  protected async updateSandboxState(
    sandbox: Sandbox,
    state: SandboxState,
    expectedLockCode: LockCode,
    runnerId?: string | null | undefined,
    errorReason?: string,
    daemonVersion?: string,
    backupState?: BackupState,
    recoverable?: boolean,
  ) {
    //  check if the lock code is still valid
    const lockKey = getStateChangeLockKey(sandbox.id)
    const currentLockCode = await this.redisLockProvider.getCode(lockKey)

    if (currentLockCode === null) {
      this.logger.warn(
        `no lock code found - state update action expired - skipping - sandboxId: ${sandbox.id} - state: ${state}`,
      )
      return
    }

    if (expectedLockCode.getCode() !== currentLockCode.getCode()) {
      this.logger.warn(
        `lock code mismatch - state update action expired - skipping - sandboxId: ${sandbox.id} - state: ${state}`,
      )
      return
    }

    if (state !== SandboxState.ARCHIVED && !sandbox.pending) {
      const err = new Error(`sandbox ${sandbox.id} is not in a pending state`)
      this.logger.error(err)
      return
    }

    if (sandbox.state === state && sandbox.runnerId === runnerId && sandbox.errorReason === errorReason) {
      return
    }

    const updateData: Partial<Sandbox> = {
      state,
    }

    if (runnerId !== undefined) {
      updateData.runnerId = runnerId
    }

    if (errorReason !== undefined) {
      updateData.errorReason = errorReason
      if (state === SandboxState.ERROR) {
        updateData.recoverable = recoverable ?? false
      }
    }

    if (sandbox.state === SandboxState.ERROR && !sandbox.errorReason) {
      updateData.errorReason = 'Sandbox is in error state during update'
      updateData.recoverable = false
    }

    if (daemonVersion !== undefined) {
      updateData.daemonVersion = daemonVersion
    }

    if (state == SandboxState.DESTROYED) {
      updateData.backupState = BackupState.NONE
    }

    if (backupState !== undefined) {
      Object.assign(updateData, Sandbox.getBackupStateUpdate(sandbox, backupState))
    }

    if (recoverable !== undefined) {
      updateData.recoverable = recoverable
    }

    await this.sandboxRepository.update(sandbox.id, { updateData, entity: sandbox })
  }
}
