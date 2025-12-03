/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, NestMiddleware, Inject } from '@nestjs/common'
import { Request, Response, NextFunction } from 'express'
import { ThrottlerException } from '@nestjs/throttler'
import { getRedisConnectionToken } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { TypedConfigService } from '../../config/typed-config.service'

/**
 * Middleware that checks if an IP is blocked due to too many failed auth attempts.
 * Runs BEFORE auth guards to block requests early and prevent wasting resources on auth.
 *
 * Flow:
 * 1. Request comes in
 * 2. This middleware checks Redis if IP has exceeded failed auth limit (isBlocked)
 * 3. If blocked: return 429 with rate limit headers immediately
 * 4. If not blocked: continue to auth guards
 */
@Injectable()
export class FailedAuthRateLimitMiddleware implements NestMiddleware {
  constructor(
    @Inject(getRedisConnectionToken('throttler')) private readonly redis: Redis,
    private readonly configService: TypedConfigService,
  ) {}

  async use(req: Request, res: Response, next: NextFunction) {
    const ip = req.ips.length ? req.ips[0] : req.ip
    const tracker = `failedauth:${ip}`
    const throttlerName = 'failed-auth'

    // Get failed-auth config from TypedConfigService
    const failedAuthConfig = this.configService.get('rateLimit.failedAuth')

    if (!failedAuthConfig || !failedAuthConfig.ttl || !failedAuthConfig.limit) {
      // If failed-auth throttler is not configured, skip
      return next()
    }

    try {
      // Build the Redis key (same format as ThrottlerStorageRedisService)
      const keyPrefix = this.redis.options.keyPrefix || ''
      const key = `${throttlerName}-${tracker}`
      const blockedKey = `${keyPrefix}{${key}:${throttlerName}}:blocked`

      // Check if IP is blocked
      const isBlocked = await this.redis.get(blockedKey)

      if (isBlocked) {
        // Get TTL for the blocked key
        const ttl = await this.redis.pttl(blockedKey)

        // Set rate limit headers to inform client (convert ttl from milliseconds to seconds)
        res.setHeader('X-RateLimit-Limit-failed-auth', failedAuthConfig.limit.toString())
        res.setHeader('X-RateLimit-Remaining-failed-auth', '0')
        res.setHeader('X-RateLimit-Reset-failed-auth', Math.ceil(ttl / 1000).toString())
        res.setHeader('Retry-After-failed-auth', Math.ceil(ttl / 1000).toString())

        throw new ThrottlerException()
      }

      // Not blocked, continue to auth guards
      next()
    } catch (error) {
      if (error instanceof ThrottlerException) {
        throw error
      }
      // If there's an error checking the rate limit, allow the request to continue
      // We don't want rate limiting failures to block legitimate requests
      next()
    }
  }
}
