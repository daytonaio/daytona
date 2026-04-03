/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, UnauthorizedException, Logger, OnModuleInit } from '@nestjs/common'
import { PassportStrategy } from '@nestjs/passport'
import { Strategy } from 'passport-http-bearer'
import { ApiKeyService } from '../api-key/api-key.service'
import { ApiKey } from '../api-key/api-key.entity'
import { UserService } from '../user/user.service'
import { TypedConfigService } from '../config/typed-config.service'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { SystemRole } from '../user/enums/system-role.enum'
import { RunnerService } from '../sandbox/services/runner.service'
import { generateApiKeyHash } from '../common/utils/api-key'
import { RegionService } from '../region/services/region.service'
import { JWT_REGEX } from './constants/jwt-regex.constant'
import { AuthStrategyType } from './enums/auth-strategy-type.enum'
import { RequestWithAuthMetadata } from './interfaces/request-with-auth-metadata.interface'
import { UserAuthContext } from '../common/interfaces/user-auth-context.interface'
import { ProxyAuthContext } from '../common/interfaces/proxy-auth-context.interface'
import { RunnerAuthContext } from '../common/interfaces/runner-auth-context.interface'
import { SshGatewayAuthContext } from '../common/interfaces/ssh-gateway-auth-context.interface'
import { RegionProxyAuthContext } from '../common/interfaces/region-proxy-auth-context.interface'
import { RegionSSHGatewayAuthContext } from '../common/interfaces/region-ssh-gateway-auth-context.interface'
import { OtelCollectorAuthContext } from '../common/interfaces/otel-collector-auth-context.interface'
import { HealthCheckAuthContext } from '../common/interfaces/health-check-auth-context.interface'

type ApiKeyAuthContext =
  | UserAuthContext
  | ProxyAuthContext
  | RunnerAuthContext
  | SshGatewayAuthContext
  | RegionProxyAuthContext
  | RegionSSHGatewayAuthContext
  | OtelCollectorAuthContext
  | HealthCheckAuthContext

type UserCache = {
  userId: string
  role: SystemRole
  email: string
}

@Injectable()
export class ApiKeyStrategy extends PassportStrategy(Strategy, AuthStrategyType.API_KEY) implements OnModuleInit {
  private readonly logger = new Logger(ApiKeyStrategy.name)

  constructor(
    @InjectRedis() private readonly redis: Redis,
    private readonly apiKeyService: ApiKeyService,
    private readonly userService: UserService,
    private readonly configService: TypedConfigService,
    private readonly runnerService: RunnerService,
    private readonly regionService: RegionService,
  ) {
    super({ passReqToCallback: true })
  }

  onModuleInit() {
    this.logger.log('ApiKeyStrategy initialized')
  }

  async validate(request: RequestWithAuthMetadata, token: string): Promise<ApiKeyAuthContext | null> {
    if (!request.authMetadata?.isStrategyAllowed(AuthStrategyType.API_KEY)) {
      return null
    }

    return this.validateToken(token)
  }

  async validateToken(token: string): Promise<ApiKeyAuthContext | null> {
    /**
     * Check configured API keys
     */
    const sshGatewayApiKey = this.configService.getOrThrow('sshGateway.apiKey')
    if (sshGatewayApiKey === token) {
      return { role: 'ssh-gateway' } satisfies SshGatewayAuthContext
    }

    const proxyApiKey = this.configService.getOrThrow('proxy.apiKey')
    if (proxyApiKey === token) {
      return { role: 'proxy' } satisfies ProxyAuthContext
    }

    const otelCollectorApiKey = this.configService.get('otelCollector.apiKey')
    if (otelCollectorApiKey && otelCollectorApiKey === token) {
      return { role: 'otel-collector' } satisfies OtelCollectorAuthContext
    }

    const healthCheckApiKey = this.configService.get('healthCheck.apiKey')
    if (healthCheckApiKey && healthCheckApiKey === token) {
      return { role: 'health-check' } satisfies HealthCheckAuthContext
    }

    /**
     * Tokens matching JWT structure are not API keys — skip DB lookups and delegate to the JWT strategy (if allowed)
     */
    if (JWT_REGEX.test(token)) {
      return null
    }

    /**
     * Check for valid user API key
     */
    try {
      let apiKey = await this.getApiKeyCache(token)
      if (!apiKey) {
        apiKey = await this.apiKeyService.getApiKeyByValue(token)

        // Check expiry before caching to prevent storing expired keys
        if (apiKey.expiresAt && apiKey.expiresAt < new Date()) {
          throw new UnauthorizedException('This API key has expired')
        }

        const validationCacheTtl = this.configService.get('apiKey.validationCacheTtlSeconds')
        const cacheKey = this.generateValidationCacheKey(token)
        await this.redis.setex(cacheKey, validationCacheTtl, JSON.stringify(apiKey))
      }

      if (apiKey.expiresAt && apiKey.expiresAt < new Date()) {
        throw new UnauthorizedException('This API key has expired')
      }

      await this.apiKeyService.updateLastUsedAt(apiKey.organizationId, apiKey.userId, apiKey.name, new Date())

      let userCache = await this.getUserCache(apiKey.userId)
      if (!userCache) {
        const user = await this.userService.findOne(apiKey.userId)

        if (!user) {
          throw new UnauthorizedException('User not found')
        }

        userCache = {
          userId: user.id,
          role: user.role,
          email: user.email,
        }
        const userCacheTtl = this.configService.get('apiKey.userCacheTtlSeconds')
        await this.redis.setex(this.generateUserCacheKey(apiKey.userId), userCacheTtl, JSON.stringify(userCache))
      }

      return {
        userId: userCache.userId,
        role: userCache.role,
        email: userCache.email,
        apiKey,
        organizationId: apiKey.organizationId,
      } satisfies UserAuthContext
    } catch (error) {
      this.logger.debug('User API key validation failed:', error)
    }

    /**
     * Check for valid runner API key
     */
    try {
      const runner = await this.runnerService.findByApiKey(token)
      if (runner) {
        return {
          role: 'runner',
          runnerId: runner.id,
          runner,
        } satisfies RunnerAuthContext
      }
    } catch (error) {
      this.logger.debug('Runner API key validation failed:', error)
    }

    /**
     * Check for valid region proxy API key
     */
    try {
      const region = await this.regionService.findOneByProxyApiKey(token)
      if (region) {
        return {
          role: 'region-proxy',
          regionId: region.id,
        } satisfies RegionProxyAuthContext
      }
    } catch (error) {
      this.logger.debug('Region proxy API key validation failed:', error)
    }

    /**
     * Check for valid region SSH gateway API key
     */
    try {
      const region = await this.regionService.findOneBySshGatewayApiKey(token)
      if (region) {
        return {
          role: 'region-ssh-gateway',
          regionId: region.id,
        } satisfies RegionSSHGatewayAuthContext
      }
    } catch (error) {
      this.logger.debug('Region SSH gateway API key validation failed:', error)
    }

    /**
     * No valid API key found
     */
    return null
  }

  private async getUserCache(userId: string): Promise<UserCache | null> {
    try {
      const cached = await this.redis.get(`api-key:user:${userId}`)
      if (!cached) {
        return null
      }
      return JSON.parse(cached)
    } catch (error) {
      this.logger.error('Error getting or parsing user cache:', error)
      return null
    }
  }

  private async getApiKeyCache(token: string): Promise<ApiKey | null> {
    try {
      const cacheKey = this.generateValidationCacheKey(token)
      const cached = await this.redis.get(cacheKey)

      if (!cached) {
        return null
      }

      const apiKey = JSON.parse(cached)

      // JSON.parse returns dates as strings — restore them to Date instances
      if (apiKey.createdAt) {
        apiKey.createdAt = new Date(apiKey.createdAt)
      }
      if (apiKey.lastUsedAt) {
        apiKey.lastUsedAt = new Date(apiKey.lastUsedAt)
      }
      if (apiKey.expiresAt) {
        apiKey.expiresAt = new Date(apiKey.expiresAt)
      }

      return apiKey
    } catch (error) {
      this.logger.error('Error getting or parsing API key cache:', error)
      return null
    }
  }

  private generateValidationCacheKey(token: string): string {
    return `api-key:validation:${generateApiKeyHash(token)}`
  }

  private generateUserCacheKey(userId: string): string {
    return `api-key:user:${userId}`
  }
}
