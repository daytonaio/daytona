/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Inject, Logger } from '@nestjs/common'
import { getRedisConnectionToken } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { Request, Response } from 'express'
import { ThrottlerException } from '@nestjs/throttler'
import { TypedConfigService } from '../config/typed-config.service'
import { setRateLimitHeaders } from '../common/utils/rate-limit-headers.util'

/**
 * Service to track failed authentication attempts across all auth guards.
 * Shared logic for both JWT and API key authentication failures.
 */
@Injectable()
export class FailedAuthTrackerService {
  private readonly logger = new Logger(FailedAuthTrackerService.name)

  constructor(
    @Inject(getRedisConnectionToken('throttler')) private readonly redis: Redis,
    private readonly configService: TypedConfigService,
  ) {}

  async incrementFailedAuth(request: Request, response: Response): Promise<void> {
    try {
      const ip = request.ips.length ? request.ips[0] : request.ip
      const throttlerName = 'failed-auth'
      const tracker = `${throttlerName}:${ip}`

      // Get failed-auth config from TypedConfigService
      const failedAuthConfig = this.configService.get('rateLimit.failedAuth')
      if (!failedAuthConfig || !failedAuthConfig.ttl || !failedAuthConfig.limit) {
        // If failed-auth throttler is not configured, skip tracking
        return
      }

      const limit = failedAuthConfig.limit
      const ttl = failedAuthConfig.ttl * 1000 // Convert seconds to milliseconds

      const keyPrefix = this.redis.options.keyPrefix || ''
      const key = `${throttlerName}-${tracker}`
      const hitKey = `${keyPrefix}{${key}:${throttlerName}}:hits`
      const blockedKey = `${keyPrefix}{${key}:${throttlerName}}:blocked`

      // Increment hits
      const hits = await this.redis.incr(hitKey)
      if (hits === 1) {
        await this.redis.pexpire(hitKey, ttl)
      }
      const ttlRemaining = await this.redis.pttl(hitKey)

      // Set rate limit headers
      setRateLimitHeaders(response, {
        throttlerName,
        limit,
        remaining: Math.max(0, limit - hits),
        resetSeconds: Math.ceil(ttlRemaining / 1000),
      })

      // Check if blocked
      if (hits >= limit) {
        await this.redis.set(blockedKey, '1', 'PX', ttl)
        setRateLimitHeaders(response, {
          throttlerName,
          limit,
          remaining: 0,
          resetSeconds: Math.ceil(ttl / 1000),
          retryAfterSeconds: Math.ceil(ttl / 1000),
        })
        throw new ThrottlerException()
      }
    } catch (error) {
      if (error instanceof ThrottlerException) {
        throw error
      }
      // Log error but don't block auth if rate limiting has issues
      this.logger.error('Failed to track authentication failure:', error)
    }
  }
}
