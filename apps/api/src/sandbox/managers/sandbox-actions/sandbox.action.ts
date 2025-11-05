/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { RunnerService } from '../../services/runner.service'
import { RunnerAdapterFactory } from '../../runner-adapter/runnerAdapter'
import { Sandbox } from '../../entities/sandbox.entity'
import { Repository } from 'typeorm'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { ToolboxService } from '../../services/toolbox.service'
import { BackupState } from '../../enums/backup-state.enum'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'

export const SYNC_AGAIN = 'sync-again'
export const DONT_SYNC_AGAIN = 'dont-sync-again'
export type SyncState = typeof SYNC_AGAIN | typeof DONT_SYNC_AGAIN

@Injectable()
export abstract class SandboxAction {
  protected readonly logger = new Logger(SandboxAction.name)

  constructor(
    protected readonly runnerService: RunnerService,
    protected runnerAdapterFactory: RunnerAdapterFactory,
    @InjectRepository(Sandbox)
    protected readonly sandboxRepository: Repository<Sandbox>,
    protected readonly toolboxService: ToolboxService,
    @InjectRedis() protected readonly redis: Redis,
  ) {}

  abstract run(sandbox: Sandbox): Promise<SyncState>

  protected async updateSandboxState(
    sandboxId: string,
    state: SandboxState,
    runnerId?: string | null | undefined,
    errorReason?: string,
    daemonVersion?: string,
  ) {
    const sandbox = await this.sandboxRepository.findOneBy({
      id: sandboxId,
      pending: true,
    })
    if (!sandbox) {
      //  this should never happen
      //  if it does, we need to log the error and return
      //  this indicates a concurrency error and should be investigated
      //  we don't to throw the error, just log it and return to avoid setting the error state
      //  on the otherwise ready sandbox
      const err = new Error(`sandbox ${sandboxId} is not in a pending state`)
      this.logger.error(err)
      return
    }

    if (sandbox.state === state && sandbox.runnerId === runnerId && sandbox.errorReason === errorReason) {
      return
    }

    sandbox.state = state

    if (runnerId !== undefined) {
      sandbox.runnerId = runnerId
    }

    if (errorReason !== undefined) {
      sandbox.errorReason = errorReason
    }

    if (daemonVersion !== undefined) {
      sandbox.daemonVersion = daemonVersion
    }

    if (sandbox.state == SandboxState.DESTROYED) {
      sandbox.backupState = BackupState.NONE
    }

    await this.sandboxRepository.save(sandbox)
  }

  /**
   * Retry wrapper for sandbox operations that should retry on specific error patterns
   * Uses Redis to count retries, allowing the cron job to recall the operations
   * @param sandboxId - The sandbox ID for generating a unique retry key
   * @param operationType - Type of operation (e.g., 'start', 'stop', 'destroy', 'archive') for key uniqueness
   * @param operation - The async operation to retry
   * @param retryErrorSubstrings - Array of error message substrings that should trigger retry
   * @param maxRetries - Maximum number of retry attempts (default: 3)
   * @param onError - Optional callback to handle specific error cases before retry check (returns SyncState if handled, null otherwise)
   * @returns SYNC_AGAIN on success, DONT_SYNC_AGAIN if retries exhausted or to let cron job retry, or throws error
   */
  protected async retryOperation<T>(
    sandboxId: string,
    operationType: string,
    operation: () => Promise<T>,
    retryErrorSubstrings: string[],
    maxRetries = 3,
    onError?: (error: any) => Promise<SyncState | null>,
  ): Promise<SyncState> {
    const retryKey = `retry:${operationType}:${sandboxId}`
    const retryCount = await this.redis.incr(retryKey)

    // Set TTL on first increment (720 seconds = 12 minutes)
    if (retryCount === 1) {
      await this.redis.expire(retryKey, 720)
    }

    try {
      await operation()
      // Success - clear retry counter
      await this.redis.del(retryKey)
      return SYNC_AGAIN
    } catch (error) {
      // Check for custom error handling first
      if (onError) {
        const handled = await onError(error)
        if (handled !== null) {
          // If handled, clear retry counter if it was successful
          if (handled === SYNC_AGAIN) {
            await this.redis.del(retryKey)
          }
          return handled
        }
      }

      // Check if error message contains a retryable error substring
      const isRetryableError =
        error?.message &&
        retryErrorSubstrings.some((substring) => error.message.toLowerCase().includes(substring.toLowerCase()))

      if (isRetryableError) {
        if (retryCount >= maxRetries) {
          // All retries exhausted - clear counter and let cron job pick it up
          await this.redis.del(retryKey)
          return DONT_SYNC_AGAIN
        }
        // Retryable error with retries remaining - let cron job retry
        return DONT_SYNC_AGAIN
      }

      // Not a retryable error - clear counter and rethrow
      await this.redis.del(retryKey)
      throw error
    }
  }
}
