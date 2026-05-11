/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { InjectRedis } from '@nestjs-modules/ioredis'
import { Injectable } from '@nestjs/common'
import { Redis } from 'ioredis'

type Acquired = boolean

export class LockCode {
  constructor(private readonly code: string) {}

  public getCode(): string {
    return this.code
  }
}

@Injectable()
export class RedisLockProvider {
  constructor(@InjectRedis() private readonly redis: Redis) {}

  async lock(key: string, ttl: number, code?: LockCode | null): Promise<Acquired> {
    const keyValue = code ? code.getCode() : '1'
    const acquired = await this.redis.set(key, keyValue, 'EX', ttl, 'NX')
    return !!acquired
  }

  async getCode(key: string): Promise<LockCode | null> {
    const keyValue = await this.redis.get(key)
    return keyValue ? new LockCode(keyValue) : null
  }

  /**
   * Release a lock. When `code` is provided the delete is ownership-aware: it only removes the key
   * if its stored value still matches the token from this acquisition (compare-and-delete via Lua),
   * so a lock that already expired and was re-acquired by another owner is never deleted. Callers
   * that don't pass a token keep the legacy unconditional delete.
   */
  async unlock(key: string, code?: LockCode | null): Promise<void> {
    if (code) {
      await this.redis.eval(
        "if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) else return 0 end",
        1,
        key,
        code.getCode(),
      )
      return
    }
    await this.redis.del(key)
  }

  async isLocked(key: string): Promise<boolean> {
    const exists = await this.redis.exists(key)
    return exists === 1
  }

  async waitForLock(key: string, ttl: number, timeoutMs?: number): Promise<void> {
    const deadline = timeoutMs !== undefined ? Date.now() + timeoutMs : null
    while (true) {
      if (deadline !== null && Date.now() >= deadline) {
        throw new Error(`Timed out after ${timeoutMs}ms waiting for lock '${key}'`)
      }
      const acquired = await this.lock(key, ttl)
      if (acquired) break
      await new Promise((resolve) => setTimeout(resolve, 50))
    }
  }
}
