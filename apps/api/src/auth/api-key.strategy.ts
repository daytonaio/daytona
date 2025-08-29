/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, UnauthorizedException, Logger, OnModuleInit } from '@nestjs/common'
import { PassportStrategy } from '@nestjs/passport'
import { Strategy } from 'passport-http-bearer'
import { ApiKeyService } from '../api-key/api-key.service'
import { UserService } from '../user/user.service'
import { AuthContext } from '../common/interfaces/auth-context.interface'
import { TypedConfigService } from '../config/typed-config.service'
import { ProxyContext } from '../common/interfaces/proxy-context.interface'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { SystemRole } from '../user/enums/system-role.enum'
import { SshGatewayContext } from '../common/interfaces/ssh-gateway-context.interface'

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
  ) {
    super()
    this.logger.log('ApiKeyStrategy constructor called')
  }

  onModuleInit() {
    this.logger.log('ApiKeyStrategy initialized')
  }

  async validate(token: string): Promise<AuthContext | ProxyContext | SshGatewayContext> {
    this.logger.debug('Validate method called')
    this.logger.debug(`Validating API key: ${token.substring(0, 8)}...`)

    const sshGatewayApiKey = this.configService.getOrThrow('sshGatewayApiKey')
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

    try {
      const apiKey = await this.apiKeyService.getApiKeyByValue(token)
      this.logger.debug(`API key found for userId: ${apiKey.userId}`)

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
        await this.redis.set(`api-key:user:${apiKey.userId}`, JSON.stringify(userCache), 'EX', 60)
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
      this.logger.debug('Error in validate method:', error)

      if (error instanceof UnauthorizedException) {
        throw error
      }

      throw new UnauthorizedException('Invalid API key')
    }
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
}
