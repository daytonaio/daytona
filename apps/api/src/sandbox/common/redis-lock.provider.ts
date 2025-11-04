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

  async unlock(key: string): Promise<void> {
    await this.redis.del(key)
  }

  async waitForLock(key: string, ttl: number): Promise<void> {
    while (true) {
      const acquired = await this.lock(key, ttl)
      if (acquired) break
      await new Promise((resolve) => setTimeout(resolve, 50))
    }
  }
}
