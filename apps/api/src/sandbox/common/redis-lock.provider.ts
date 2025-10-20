/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { InjectRedis } from '@nestjs-modules/ioredis'
import { Injectable } from '@nestjs/common'
import { Redis } from 'ioredis'

type Acquired = boolean

@Injectable()
export class RedisLockProvider {
  constructor(@InjectRedis() private readonly redis: Redis) {}

  async lock(key: string, ttl: number): Promise<Acquired> {
    const acquired = await this.redis.set(key, '1', 'EX', ttl, 'NX')
    return !!acquired
  }

  async unlock(key: string): Promise<void> {
    await this.redis.del(key)
  }

  async renewLock(key: string, ttl: number): Promise<boolean> {
    // Check if the lock exists and renew it
    const exists = await this.redis.exists(key)
    if (exists) {
      await this.redis.expire(key, ttl)
      return true
    }
    return false
  }

  async waitForLock(key: string, ttl: number): Promise<void> {
    while (true) {
      const acquired = await this.lock(key, ttl)
      if (acquired) break
      await new Promise((resolve) => setTimeout(resolve, 50))
    }
  }
}
