/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RedisLockProvider } from '../../sandbox/common/redis-lock.provider'

/**
 * Base class for entity-specific locking.
 * Use when locks are scoped to individual entities (sandboxes, snapshots, etc).
 *
 * @example
 * ```typescript
 * export class SandboxService extends LockableEntity {
 *   protected getLockKey(id: string) { return `sandbox:${id}:state-change` }
 *
 *   async stop(id: string) {
 *     return this.withLock(id, 60, async () => {
 *       // Critical section - auto lock/unlock
 *     })
 *   }
 * }
 * ```
 */
export abstract class LockableEntity {
  constructor(protected readonly redisLockProvider: RedisLockProvider) {}

  /** Returns lock key for the given entity ID */
  protected abstract getLockKey(entityId: string): string

  /** Acquires lock, executes operation, releases lock (blocking - waits for lock) */
  protected async withLock<TResult>(
    entityId: string,
    timeoutSeconds: number,
    operation: () => Promise<TResult>,
  ): Promise<TResult> {
    const lockKey = this.getLockKey(entityId)
    await this.redisLockProvider.waitForLock(lockKey, timeoutSeconds)

    try {
      return await operation()
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  /** Tries to acquire lock, returns lock key or null (non-blocking - skips if locked) */
  protected async tryLock(entityId: string, ttlSeconds: number): Promise<string | null> {
    const lockKey = this.getLockKey(entityId)
    const acquired = await this.redisLockProvider.lock(lockKey, ttlSeconds)
    return acquired ? lockKey : null
  }

  /** Releases the lock */
  protected async unlock(entityId: string): Promise<void> {
    const lockKey = this.getLockKey(entityId)
    await this.redisLockProvider.unlock(lockKey)
  }
}
