/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Inject, Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { RunnerService } from '../../services/runner.service'
import { RunnerAdapterFactory } from '../../runner-adapter/runnerAdapter'
import { Sandbox } from '../../entities/sandbox.entity'
import { Repository } from 'typeorm'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { BackupState } from '../../enums/backup-state.enum'
import { getStateChangeLockKey } from '../../utils/lock-key.util'
import { LockCode, RedisLockProvider } from '../../common/redis-lock.provider'
import { SandboxService } from '../../services/sandbox.service'

export const SYNC_AGAIN = 'sync-again'
export const DONT_SYNC_AGAIN = 'dont-sync-again'
export type SyncState = typeof SYNC_AGAIN | typeof DONT_SYNC_AGAIN

@Injectable()
export abstract class SandboxAction {
  protected readonly logger = new Logger(SandboxAction.name)

  @Inject(EventEmitter2)
  protected readonly eventEmitter: EventEmitter2

  @Inject(SandboxService)
  protected readonly sandboxService: SandboxService

  constructor(
    protected readonly runnerService: RunnerService,
    protected runnerAdapterFactory: RunnerAdapterFactory,
    @InjectRepository(Sandbox)
    protected readonly sandboxRepository: Repository<Sandbox>,
    protected readonly redisLockProvider: RedisLockProvider,
  ) {}

  abstract run(sandbox: Sandbox, lockCode: LockCode): Promise<SyncState>

  /**
   * Validates the Redis lock code and then delegates to SandboxService.updateSandboxState
   * for a single targeted UPDATE with hook replication, event emission, and cache invalidation.
   *
   * A WHERE guard of { pending: true } is added for non-ARCHIVED transitions to prevent
   * concurrent writes from overwriting each other.
   */
  protected async guardedUpdateState(
    sandbox: Sandbox,
    state: SandboxState,
    expectedLockCode: LockCode,
    runnerId?: string | null | undefined,
    errorReason?: string,
    daemonVersion?: string,
    backupState?: BackupState,
    recoverable?: boolean,
  ): Promise<Sandbox | undefined> {
    //  check if the lock code is still valid
    const lockKey = getStateChangeLockKey(sandbox.id)
    const currentLockCode = await this.redisLockProvider.getCode(lockKey)

    if (currentLockCode === null) {
      this.logger.warn(
        `no lock code found - state update action expired - skipping - sandboxId: ${sandbox.id} - state: ${state}`,
      )
      return undefined
    }

    if (expectedLockCode.getCode() !== currentLockCode.getCode()) {
      this.logger.warn(
        `lock code mismatch - state update action expired - skipping - sandboxId: ${sandbox.id} - state: ${state}`,
      )
      return undefined
    }

    //  If transitioning to ERROR without an error reason, set a default
    const effectiveErrorReason = errorReason !== undefined ? errorReason : sandbox.errorReason
    if (state === SandboxState.ERROR && !effectiveErrorReason) {
      errorReason = 'Sandbox is in error state during update'
      recoverable = false
    }

    //  WHERE guard: only update if the sandbox is still in 'pending' state
    //  (skip for ARCHIVED transitions since those aren't set to pending beforehand)
    const where = state !== SandboxState.ARCHIVED ? { pending: true } : undefined

    const updated = await this.sandboxService.updateSandboxState(
      sandbox,
      state,
      {
        runnerId,
        errorReason,
        recoverable,
        daemonVersion,
        backupState,
      },
      where,
    )

    if (!updated) {
      const err = new Error(`sandbox ${sandbox.id} is not in a pending state`)
      this.logger.error(err)
      return undefined
    }

    return sandbox
  }
}
