/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, Inject, ExecutionContext } from '@nestjs/common'
import { ThrottlerGuard, ThrottlerRequest, ThrottlerModuleOptions } from '@nestjs/throttler'
import { ThrottlerStorage } from '@nestjs/throttler/dist/throttler-storage.interface'
import { Reflector } from '@nestjs/core'
import { Request } from 'express'
import { getRedisConnectionToken } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'

@Injectable()
export class AuthenticatedRateLimitGuard extends ThrottlerGuard {
  private readonly logger = new Logger(AuthenticatedRateLimitGuard.name)

  constructor(
    options: ThrottlerModuleOptions,
    storageService: ThrottlerStorage,
    reflector: Reflector,
    @Inject(getRedisConnectionToken('throttler')) private readonly redis: Redis,
  ) {
    super(options, storageService, reflector)
  }

  protected async getTracker(req: Request): Promise<string> {
    const hasAuthHeader = req.headers.authorization?.startsWith('Bearer ')

    if (hasAuthHeader) {
      const token = req.headers.authorization
      return `auth:${this.hashToken(token)}`
    }

    // Fallback (shouldn't happen in normal flow)
    const ip = req.ips.length ? req.ips[0] : req.ip
    return `fallback:${ip}`
  }

  async handleRequest(requestProps: ThrottlerRequest): Promise<boolean> {
    const { context, throttler } = requestProps
    const request = context.switchToHttp().getRequest<Request>()
    const isAuthenticated = request.user && this.isValidAuthContext(request.user)

    if (throttler.name === 'authenticated') {
      if (isAuthenticated) {
        // Clear anonymous rate limit on successful authentication
        await this.clearAnonymousRateLimit(request, context)
        return super.handleRequest(requestProps)
      }
      return true
    }

    return true
  }

  private async clearAnonymousRateLimit(request: Request, context: ExecutionContext): Promise<void> {
    try {
      const ip = request.ips.length ? request.ips[0] : request.ip

      const anonymousTracker = `anonymous:${ip}`
      const anonymousThrottlerName = 'anonymous'

      // Generate the key using the same context and tracker as the anonymous guard
      const anonymousKey = this.generateKey(context, anonymousTracker, anonymousThrottlerName)

      // Construct the Redis keys using the same format as the throttler storage
      const keyPrefix = this.redis.options.keyPrefix || ''
      const hitKey = `${keyPrefix}{${anonymousKey}:${anonymousThrottlerName}}:hits`
      const blockKey = `${keyPrefix}{${anonymousKey}:${anonymousThrottlerName}}:blocked`

      // Delete the specific keys for this IP and context
      const deletedKeys = []
      if (await this.redis.exists(hitKey)) {
        await this.redis.del(hitKey)
        deletedKeys.push(hitKey)
      }
      if (await this.redis.exists(blockKey)) {
        await this.redis.del(blockKey)
        deletedKeys.push(blockKey)
      }
    } catch (error) {
      this.logger.warn('Failed to clear anonymous rate limit:', error)
      // Don't throw - rate limiting should not break authentication
    }
  }

  private hashToken(token: string): string {
    return Buffer.from(token).toString('base64').substring(0, 16)
  }

  private isValidAuthContext(user: any): boolean {
    return user && (user.userId || user.role)
  }
}
