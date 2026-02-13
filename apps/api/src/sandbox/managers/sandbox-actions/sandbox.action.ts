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
import { Repository, FindOptionsWhere } from 'typeorm'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { SandboxDesiredState } from '../../enums/sandbox-desired-state.enum'
import { BackupState } from '../../enums/backup-state.enum'
import { getStateChangeLockKey } from '../../utils/lock-key.util'
import { LockCode, RedisLockProvider } from '../../common/redis-lock.provider'
import { SandboxLookupCacheInvalidationService } from '../../services/sandbox-lookup-cache-invalidation.service'
import { SandboxStateUpdatedEvent } from '../../events/sandbox-state-updated.event'
import { SandboxEvents } from '../../constants/sandbox-events.constants'

export const SYNC_AGAIN = 'sync-again'
export const DONT_SYNC_AGAIN = 'dont-sync-again'
export type SyncState = typeof SYNC_AGAIN | typeof DONT_SYNC_AGAIN

@Injectable()
export abstract class SandboxAction {
  protected readonly logger = new Logger(SandboxAction.name)

  @Inject(EventEmitter2)
  protected readonly eventEmitter: EventEmitter2

  @Inject(SandboxLookupCacheInvalidationService)
  protected readonly sandboxLookupCacheInvalidationService: SandboxLookupCacheInvalidationService

  constructor(
    protected readonly runnerService: RunnerService,
    protected runnerAdapterFactory: RunnerAdapterFactory,
    @InjectRepository(Sandbox)
    protected readonly sandboxRepository: Repository<Sandbox>,
    protected readonly redisLockProvider: RedisLockProvider,
  ) {}

  abstract run(sandbox: Sandbox, lockCode: LockCode): Promise<SyncState>

  /**
   * Updates a sandbox's state using a single targeted UPDATE query instead of
   * loading the full entity and saving all 35+ columns back.
   *
   * The caller must pass the already-loaded sandbox entity (from syncInstanceState)
   * so we can read desiredState, organizationId, name etc. for hook replication,
   * event emission, and cache invalidation — without an extra SELECT.
   *
   * @BeforeUpdate hooks (updateLastActivityAt, updatePendingFlag, handleDestroyedState,
   * validateDesiredState) are replicated inline.
   */
  protected async updateSandboxState(
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

    const oldState = sandbox.state

    // ── Build the update patch (same order as the old entity-mutation logic) ──

    const patch: Partial<Record<string, unknown>> = {}

    patch.state = state

    if (runnerId !== undefined) {
      patch.runnerId = runnerId
    }

    if (errorReason !== undefined) {
      patch.errorReason = errorReason
      if (state === SandboxState.ERROR) {
        patch.recoverable = recoverable ?? false
      }
    }

    //  If transitioning to ERROR without an error reason, set a default
    const effectiveErrorReason = errorReason !== undefined ? errorReason : sandbox.errorReason
    if (state === SandboxState.ERROR && !effectiveErrorReason) {
      patch.errorReason = 'Sandbox is in error state during update'
      patch.recoverable = false
    }

    if (daemonVersion !== undefined) {
      patch.daemonVersion = daemonVersion
    }

    //  Replicate handleDestroyedState @BeforeUpdate hook
    if (state === SandboxState.DESTROYED) {
      patch.runnerId = null
      patch.backupState = BackupState.NONE
    }

    //  Handle setBackupState (only BackupState.NONE is passed in practice)
    if (backupState !== undefined) {
      patch.backupState = backupState
      if (backupState === BackupState.NONE) {
        patch.backupSnapshot = null
      }
    }

    if (recoverable !== undefined) {
      patch.recoverable = recoverable
    }

    //  Replicate updateLastActivityAt @BeforeUpdate hook
    patch.lastActivityAt = new Date()

    //  Replicate updatePendingFlag @BeforeUpdate hook
    const effectiveDesiredState = sandbox.desiredState
    let pendingValue = sandbox.pending ?? false
    if (!pendingValue && String(state) !== String(effectiveDesiredState)) {
      pendingValue = true
    }
    if (pendingValue && String(state) === String(effectiveDesiredState)) {
      pendingValue = false
    }
    if (
      state === SandboxState.ERROR ||
      state === SandboxState.BUILD_FAILED ||
      effectiveDesiredState === SandboxDesiredState.ARCHIVED
    ) {
      pendingValue = false
    }
    patch.pending = pendingValue

    // ── Execute a single targeted UPDATE with WHERE guard ──

    const whereClause: FindOptionsWhere<Sandbox> = { id: sandbox.id }
    if (state !== SandboxState.ARCHIVED) {
      whereClause.pending = true
    }

    const result = await this.sandboxRepository.update(whereClause, patch)

    if (!result.affected || result.affected === 0) {
      //  Sandbox wasn't found or wasn't in the expected pending state.
      //  This indicates a concurrency conflict — log and return gracefully
      //  to avoid incorrectly setting the error state on an otherwise ready sandbox.
      const err = new Error(`sandbox ${sandbox.id} is not in a pending state`)
      this.logger.error(err)
      return undefined
    }

    // ── Apply patch to the in-memory entity so callers see the new values ──

    Object.assign(sandbox, patch)

    // ── Emit STATE_UPDATED event (replaces subscriber behaviour for save()) ──

    if (oldState !== state) {
      this.eventEmitter.emit(SandboxEvents.STATE_UPDATED, new SandboxStateUpdatedEvent(sandbox, oldState, state))
    }

    // ── Invalidate lookup cache (replaces subscriber behaviour for save()) ──

    try {
      this.sandboxLookupCacheInvalidationService.invalidate({
        sandboxId: sandbox.id,
        organizationId: sandbox.organizationId,
        name: sandbox.name,
      })
    } catch (error) {
      this.logger.warn(
        `Failed to invalidate sandbox lookup cache for ${sandbox.id}: ${error instanceof Error ? error.message : String(error)}`,
      )
    }

    return sandbox
  }
}
