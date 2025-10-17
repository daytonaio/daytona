/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, Inject, ExecutionContext } from '@nestjs/common'
import { ThrottlerGuard, ThrottlerRequest, ThrottlerModuleOptions, ThrottlerStorage } from '@nestjs/throttler'
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
    const user = req.user as any

    // Track by organization ID when available (shared quota per org)
    if (user?.organizationId) {
      return `auth:org:${user.organizationId}`
    }

    // Fallback to user ID for non-org routes (e.g., /users/me)
    if (user?.userId) {
      return `auth:user:${user.userId}`
    }

    // Ultimate fallback (shouldn't happen in normal flow)
    const ip = req.ips.length ? req.ips[0] : req.ip
    return `fallback:${ip}`
  }

  async handleRequest(requestProps: ThrottlerRequest): Promise<boolean> {
    const { context, throttler } = requestProps
    const request = context.switchToHttp().getRequest<Request>()
    const isAuthenticated = request.user && this.isValidAuthContext(request.user)

    // Skip rate limiting for M2M system roles (checked AFTER auth runs)
    if (this.isSystemRole(request.user)) {
      return true
    }

    // Check 'authenticated' throttler - applies to all authenticated routes
    // Routes can override with @Throttle({ authenticated: { limit, ttl } })
    if (throttler.name === 'authenticated') {
      if (isAuthenticated) {
        // Clear anonymous rate limit on successful authentication
        await this.clearAnonymousRateLimit(request, context)
        return super.handleRequest(requestProps)
      }
      return true
    }

    // Skip anonymous throttler (handled by AnonymousRateLimitGuard)
    if (throttler.name === 'anonymous') {
      return true
    }

    // For any other throttlers, defer to base ThrottlerGuard
    if (isAuthenticated) {
      return super.handleRequest(requestProps)
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

  private isValidAuthContext(user: any): boolean {
    return user && (user.userId || user.role)
  }

  private isSystemRole(user: any): boolean {
    // Skip rate limiting for M2M system roles (proxy, runner, ssh-gateway)
    return user?.role === 'ssh-gateway' || user?.role === 'proxy' || user?.role === 'runner'
  }
}
