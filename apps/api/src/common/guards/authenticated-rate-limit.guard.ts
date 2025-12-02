/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, Inject, ExecutionContext, Optional } from '@nestjs/common'
import { ThrottlerGuard, ThrottlerRequest, ThrottlerModuleOptions, ThrottlerStorage } from '@nestjs/throttler'
import { Reflector } from '@nestjs/core'
import { Request } from 'express'
import { getRedisConnectionToken } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { OrganizationService } from '../../organization/services/organization.service'
import { THROTTLER_SCOPE_KEY } from '../decorators/throttler-scope.decorator'
import { createHash } from 'crypto'

@Injectable()
export class AuthenticatedRateLimitGuard extends ThrottlerGuard {
  private readonly logger = new Logger(AuthenticatedRateLimitGuard.name)

  constructor(
    options: ThrottlerModuleOptions,
    storageService: ThrottlerStorage,
    reflector: Reflector,
    @Inject(getRedisConnectionToken('throttler')) private readonly redis: Redis,
    @Optional() private readonly organizationService?: OrganizationService,
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

  protected generateKey(context: ExecutionContext, suffix: string, name: string): string {
    // Override to make rate limiting per-rate-limit-type, not per-route
    // This ensures all routes share the same counter per rate limit type (authenticated, sandbox-create, sandbox-lifecycle)
    return createHash('sha256').update(`${name}-${suffix}`).digest('hex')
  }

  async handleRequest(requestProps: ThrottlerRequest): Promise<boolean> {
    const { context, throttler } = requestProps
    const request = context.switchToHttp().getRequest<Request>()
    const isAuthenticated = request.user && this.isValidAuthContext(request.user)

    // Skip rate limiting for M2M system roles (checked AFTER auth runs)
    if (this.isSystemRole(request.user)) {
      await this.clearAnonymousRateLimit(request, context)
      return true
    }

    // Skip anonymous throttler (handled by AnonymousRateLimitGuard)
    if (throttler.name === 'anonymous') {
      return true
    }

    // Check authenticated throttlers
    const authenticatedThrottlers = ['authenticated', 'sandbox-create', 'sandbox-lifecycle']
    if (authenticatedThrottlers.includes(throttler.name)) {
      if (isAuthenticated) {
        // Clear anonymous rate limit on successful authentication (once per request)
        // Do this BEFORE checking throttler scope so it happens for all authenticated routes
        await this.clearAnonymousRateLimit(request, context)

        // Only 'authenticated' applies to all routes by default
        // 'sandbox-create' and 'sandbox-lifecycle' only apply if explicitly configured via @SkipThrottle or @Throttle
        const isDefaultThrottler = throttler.name === 'authenticated'

        if (!isDefaultThrottler) {
          // Sandbox throttlers (sandbox-create, sandbox-lifecycle) are opt-in only
          // Check if this route declares this throttler scope via @ThrottlerScope() decorator
          const scopes = this.reflector.getAllAndOverride<string[]>(THROTTLER_SCOPE_KEY, [
            context.getHandler(),
            context.getClass(),
          ])

          // If the route hasn't declared this throttler in its scope, skip it
          if (!scopes || !scopes.includes(throttler.name)) {
            return true
          }
        }

        const user = request.user as any
        const orgId = user?.organizationId
        if (orgId) {
          const orgLimits = await this.getCachedOrganizationRateLimits(orgId)
          if (orgLimits) {
            const customLimit =
              throttler.name === 'authenticated'
                ? orgLimits.authenticated
                : throttler.name === 'sandbox-create'
                  ? orgLimits.sandboxCreate
                  : throttler.name === 'sandbox-lifecycle'
                    ? orgLimits.sandboxLifecycle
                    : undefined

            if (customLimit) {
              const modifiedProps = {
                ...requestProps,
                limit: customLimit,
              }
              return super.handleRequest(modifiedProps)
            }
          }
        }
        return super.handleRequest(requestProps)
      }
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
      await this.redis.del(hitKey)
      await this.redis.del(blockKey)
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

  private async getCachedOrganizationRateLimits(
    organizationId: string,
  ): Promise<{ authenticated: number | null; sandboxCreate: number | null; sandboxLifecycle: number | null } | null> {
    // If OrganizationService is not available (e.g., in UserModule), use default rate limits
    if (!this.organizationService) {
      return null
    }

    try {
      const cacheKey = `organization:rate-limits:${organizationId}`
      const cachedLimits = await this.redis.get(cacheKey)

      if (cachedLimits) {
        return JSON.parse(cachedLimits)
      }

      const organization = await this.organizationService.findOne(organizationId)
      if (organization) {
        const limits = {
          authenticated: organization.authenticatedRateLimit,
          sandboxCreate: organization.sandboxCreateRateLimit,
          sandboxLifecycle: organization.sandboxLifecycleRateLimit,
        }
        await this.redis.set(cacheKey, JSON.stringify(limits), 'EX', 60)
        return limits
      }

      return null
    } catch (error) {
      this.logger.error('Error getting cached organization rate limits:', error)
      return null
    }
  }
}
