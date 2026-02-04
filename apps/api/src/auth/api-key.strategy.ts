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
import { AuthContextType } from '../common/interfaces/auth-context.interface'
import { TypedConfigService } from '../config/typed-config.service'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { SystemRole } from '../user/enums/system-role.enum'
import { RunnerService } from '../sandbox/services/runner.service'
import { generateApiKeyHash } from '../common/utils/api-key'
import { RegionService } from '../region/services/region.service'
import { getApiKeyUserCacheKey, getApiKeyValidationCacheKey } from '../api-key/constants/api-key-cache-keys.constant'

type UserCache = {
  userId: string
  role: SystemRole
  email: string
}

@Injectable()
export class ApiKeyStrategy extends PassportStrategy(Strategy, 'api-key') implements OnModuleInit {
  private readonly logger = new Logger(ApiKeyStrategy.name)

  constructor(
    @InjectRedis() private readonly redis: Redis,
    private readonly apiKeyService: ApiKeyService,
    private readonly userService: UserService,
    private readonly configService: TypedConfigService,
    private readonly runnerService: RunnerService,
    private readonly regionService: RegionService,
  ) {
    super()
    this.logger.log('ApiKeyStrategy constructor called')
  }

  onModuleInit() {
    this.logger.log('ApiKeyStrategy initialized')
  }

  async validate(token: string): Promise<AuthContextType> {
    this.logger.debug('Validate method called')
    this.logger.debug(`Validating API key: ${token.substring(0, 8)}...`)

    const sshGatewayApiKey = this.configService.getOrThrow('sshGateway.apiKey')
    if (sshGatewayApiKey === token) {
      return {
        role: 'ssh-gateway',
      }
    }

    const proxyApiKey = this.configService.getOrThrow('proxy.apiKey')
    if (proxyApiKey === token) {
      return {
        role: 'proxy',
      }
    }

    const otelCollectorApiKey = this.configService.get('otelCollector.apiKey')
    if (otelCollectorApiKey && otelCollectorApiKey === token) {
      return {
        role: 'otel-collector',
      }
    }

    const healthCheckApiKey = this.configService.get('healthCheck.apiKey')
    if (healthCheckApiKey && healthCheckApiKey === token) {
      return {
        role: 'health-check',
      }
    }

    try {
      let apiKey = await this.getApiKeyCache(token)
      if (!apiKey) {
        // Cache miss - validate from database
        apiKey = await this.apiKeyService.getApiKeyByValue(token)
        this.logger.debug(`API key found for userId: ${apiKey.userId}`)

        // Check expiry BEFORE caching to prevent storing expired keys
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

      this.logger.debug(`Updating last used timestamp for API key: ${token.substring(0, 8)}...`)
      await this.apiKeyService.updateLastUsedAt(apiKey.organizationId, apiKey.userId, apiKey.name, new Date())

      let userCache = await this.getUserCache(apiKey.userId)
      if (!userCache) {
        const user = await this.userService.findOne(apiKey.userId)
        if (!user) {
          this.logger.error(`Api key has invalid user: ${apiKey.keySuffix} - ${apiKey.userId}`)
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

      const result = {
        userId: userCache.userId,
        role: userCache.role,
        email: userCache.email,
        apiKey,
        organizationId: apiKey.organizationId,
      }

      this.logger.debug('Authentication successful', result)
      return result
    } catch (error) {
      this.logger.debug('Error checking user API key:', error)
      // Continue to check runner API keys if user check fails
    }

    try {
      const runner = await this.runnerService.findByApiKey(token)
      if (runner) {
        this.logger.debug(`Runner API key found for runner: ${runner.id}`)
        return {
          role: 'runner',
          runnerId: runner.id,
          runner,
        }
      }
    } catch (error) {
      this.logger.debug('Error checking runner API key:', error)
    }

    try {
      const region = await this.regionService.findOneByProxyApiKey(token)
      if (region) {
        this.logger.debug(`Region proxy API key found for region: ${region.id}`)
        return {
          role: 'region-proxy',
          regionId: region.id,
        }
      }
    } catch (error) {
      this.logger.debug('Error checking region proxy API key:', error)
    }

    try {
      const region = await this.regionService.findOneBySshGatewayApiKey(token)
      if (region) {
        this.logger.debug(`Region SSH gateway API key found for region: ${region.id}`)
        return {
          role: 'region-ssh-gateway',
          regionId: region.id,
        }
      }
    } catch (error) {
      this.logger.debug('Error checking region SSH gateway API key:', error)
    }

    throw new UnauthorizedException('Invalid API key')
  }

  private async getUserCache(userId: string): Promise<UserCache | null> {
    try {
      const userCacheRaw = await this.redis.get(`api-key:user:${userId}`)
      if (userCacheRaw) {
        return JSON.parse(userCacheRaw)
      }
      return null
    } catch (error) {
      this.logger.error('Error getting user cache:', error)
      return null
    }
  }

  private async getApiKeyCache(token: string): Promise<ApiKey | null> {
    try {
      const cacheKey = this.generateValidationCacheKey(token)
      const cached = await this.redis.get(cacheKey)
      if (cached) {
        this.logger.debug('Using cached API key validation')
        const apiKey = JSON.parse(cached)
        // Parse Date fields from cached data
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
      }
      return null
    } catch (error) {
      this.logger.error('Error getting API key cache:', error)
      return null
    }
  }

  private generateValidationCacheKey(token: string): string {
    const keyHash = generateApiKeyHash(token)
    return getApiKeyValidationCacheKey(keyHash)
  }

  private generateUserCacheKey(userId: string): string {
    return getApiKeyUserCacheKey(userId)
  }
}
