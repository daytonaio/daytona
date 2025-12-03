/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Inject, ExecutionContext } from '@nestjs/common'
import { getRedisConnectionToken } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { Request, Response } from 'express'
import { ThrottlerException } from '@nestjs/throttler'
import { TypedConfigService } from '../config/typed-config.service'

/**
 * Service to track failed authentication attempts across all auth guards.
 * Shared logic for both JWT and API key authentication failures.
 */
@Injectable()
export class FailedAuthTrackerService {
  constructor(
    @Inject(getRedisConnectionToken('throttler')) private readonly redis: Redis,
    private readonly configService: TypedConfigService,
  ) {}

  async incrementFailedAuth(context: ExecutionContext): Promise<void> {
    try {
      const request = context.switchToHttp().getRequest<Request>()
      const response = context.switchToHttp().getResponse<Response>()
      const ip = request.ips.length ? request.ips[0] : request.ip
      const tracker = `failedauth:${ip}`
      const throttlerName = 'failed-auth'

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

      // Set headers (convert ttlRemaining from milliseconds to seconds to match other rate limiters)
      response.setHeader(`X-RateLimit-Limit-${throttlerName}`, limit.toString())
      response.setHeader(`X-RateLimit-Remaining-${throttlerName}`, Math.max(0, limit - hits).toString())
      response.setHeader(`X-RateLimit-Reset-${throttlerName}`, Math.ceil(ttlRemaining / 1000).toString())

      // Check if blocked
      if (hits > limit) {
        await this.redis.set(blockedKey, '1', 'PX', ttl)
        response.setHeader(`Retry-After-${throttlerName}`, Math.ceil(ttl / 1000).toString())
        throw new ThrottlerException()
      }
    } catch (error) {
      if (error instanceof ThrottlerException) {
        throw error
      }
      // Silently fail - don't block auth if rate limiting has issues
    }
  }
}
