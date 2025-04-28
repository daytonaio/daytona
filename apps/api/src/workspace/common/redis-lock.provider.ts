/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { InjectRedis } from '@nestjs-modules/ioredis'
import { Injectable } from '@nestjs/common'
import { Redis } from 'ioredis'

@Injectable()
export class RedisLockProvider {
  constructor(@InjectRedis() private readonly redis: Redis) {}

  async lock(key: string, ttl: number): Promise<boolean> {
    if (await this.redis.get(key)) {
      return true
    }
    // //  sleep for 100ms to avoid race condition
    // await new Promise((resolve) => setTimeout(resolve, 100))
    // const hasLock2 = await this.redis.get(key)
    // if (hasLock2) {
    //   return true
    // }
    await this.redis.setex(key, ttl, '1')
    return false
  }

  async unlock(key: string): Promise<void> {
    await this.redis.del(key)
  }

  async waitForLock(key: string, ttl: number): Promise<void> {
    while (await this.redis.get(key)) {
      await new Promise((resolve) => setTimeout(resolve, 50))
    }

    await this.redis.setex(key, ttl, '1')
  }
}
