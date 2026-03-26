/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { HealthIndicatorService } from '@nestjs/terminus'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'

@Injectable()
export class RedisHealthIndicator {
  private readonly redis: Redis
  constructor(
    @InjectRedis() redis: Redis,
    private readonly healthIndicatorService: HealthIndicatorService,
  ) {
    this.redis = redis.duplicate({
      commandTimeout: 1000,
    })
  }

  async isHealthy(key: string) {
    // Start the health indicator check for the given key
    const indicator = this.healthIndicatorService.check(key)

    try {
      await this.redis.ping()
      return indicator.up()
    } catch (error) {
      return indicator.down(error)
    }
  }
}
