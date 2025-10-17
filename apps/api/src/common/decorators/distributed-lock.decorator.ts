/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RedisLockProvider } from '../../sandbox/common/redis-lock.provider'

type DistributedLockOptions = {
  lockKey?: string
  lockTtl?: number
}

/**
 * Redis lock decorator for exclusive execution. The lock is released automatically when the method completes.
 * redisLockProvider is required to be injected in the class.
 * @param options - The options for the Redis lock
 * @param options.lockKey - The key to use for the Redis lock
 * @param options.lockTtl - Time to live for the lock in seconds
 * @returns A decorator function that handles Redis locking
 */
export function DistributedLock(options?: DistributedLockOptions): MethodDecorator {
  return function (target: any, propertyKey: string, descriptor: PropertyDescriptor) {
    const originalMethod = descriptor.value

    descriptor.value = async function (...args: any[]) {
      if (!this.redisLockProvider) {
        throw new Error(`@DistributedLock requires 'redisLockProvider' property on ${target.constructor.name}`)
      }

      const redisLockProvider: RedisLockProvider = this.redisLockProvider

      // Generate lock key
      const lockKey = `lock:${options?.lockKey ?? target.constructor.name}.${propertyKey}`

      // Set default TTL if not provided
      const lockTtlMs = options?.lockTtl || 30 // 30 seconds default

      const hasLock = await redisLockProvider.lock(lockKey, lockTtlMs)
      if (!hasLock) {
        return
      }
      try {
        return await originalMethod.apply(this, args)
      } finally {
        await redisLockProvider.unlock(lockKey)
      }
    }
  }
}
