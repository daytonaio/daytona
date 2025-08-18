/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ConflictException, Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { ApiKey } from './api-key.entity'
import * as crypto from 'crypto'
import { OrganizationUser } from '../organization/entities/organization-user.entity'
import { OrganizationMemberRole } from '../organization/enums/organization-member-role.enum'
import { OrganizationResourcePermission } from '../organization/enums/organization-resource-permission.enum'
import { OrganizationUserService } from '../organization/services/organization-user.service'
import { RedisLockProvider } from '../sandbox/common/redis-lock.provider'

@Injectable()
export class ApiKeyService {
  private readonly logger = new Logger(ApiKeyService.name)

  constructor(
    @InjectRepository(ApiKey)
    private apiKeyRepository: Repository<ApiKey>,
    private organizationUserService: OrganizationUserService,
    private readonly redisLockProvider: RedisLockProvider,
  ) {}

  private generateApiKeyValue(): string {
    return `dtn_${crypto.randomBytes(32).toString('hex')}`
  }

  private generateApiKeyHash(value: string): string {
    return crypto.createHash('sha256').update(value).digest('hex')
  }

  private getApiKeyPrefix(value: string): string {
    return value.substring(0, 3)
  }

  private getApiKeySuffix(value: string): string {
    return value.slice(-3)
  }

  async createApiKey(
    organizationId: string,
    userId: string,
    name: string,
    permissions: OrganizationResourcePermission[],
    expiresAt?: Date,
  ): Promise<{ apiKey: ApiKey; value: string }> {
    const existingKey = await this.apiKeyRepository.findOne({ where: { organizationId, userId, name } })
    if (existingKey) {
      throw new ConflictException('API key with this name already exists')
    }

    const value = this.generateApiKeyValue()

    const apiKey = await this.apiKeyRepository.save({
      organizationId,
      userId,
      name,
      keyHash: this.generateApiKeyHash(value),
      keyPrefix: this.getApiKeyPrefix(value),
      keySuffix: this.getApiKeySuffix(value),
      permissions,
      createdAt: new Date(),
      expiresAt,
    })

    return { apiKey, value }
  }

  async getApiKeys(organizationId: string, userId: string): Promise<ApiKey[]> {
    const apiKeys = await this.apiKeyRepository.find({
      where: { organizationId, userId },
      order: {
        lastUsedAt: {
          direction: 'DESC',
          nulls: 'LAST',
        },
        createdAt: 'DESC',
      },
    })

    const organizationUser = await this.organizationUserService.findOne(organizationId, userId)
    if (!organizationUser) {
      throw new NotFoundException('Organization user (API key owner) not found')
    }

    return apiKeys.map((apiKey) => {
      return {
        ...apiKey,
        permissions: this.getEffectivePermissions(apiKey, organizationUser),
      }
    })
  }

  async getApiKeyByName(organizationId: string, userId: string, name: string): Promise<ApiKey> {
    const apiKey = await this.apiKeyRepository.findOne({
      where: {
        organizationId,
        userId,
        name,
      },
    })

    if (!apiKey) {
      throw new NotFoundException('API key not found')
    }

    const organizationUser = await this.organizationUserService.findOne(organizationId, userId)
    if (!organizationUser) {
      throw new NotFoundException('Organization user (API key owner) not found')
    }

    apiKey.permissions = this.getEffectivePermissions(apiKey, organizationUser)
    return apiKey
  }

  async getApiKeyByValue(value: string): Promise<ApiKey> {
    const apiKey = await this.apiKeyRepository.findOne({
      where: {
        keyHash: this.generateApiKeyHash(value),
      },
    })

    if (!apiKey) {
      throw new NotFoundException('API key not found')
    }

    const organizationUser = await this.organizationUserService.findOne(apiKey.organizationId, apiKey.userId)
    if (!organizationUser) {
      throw new NotFoundException('Organization user (API key owner) not found')
    }

    apiKey.permissions = this.getEffectivePermissions(apiKey, organizationUser)
    return apiKey
  }

  async deleteApiKey(organizationId: string, userId: string, name: string): Promise<void> {
    const apiKey = await this.apiKeyRepository.findOne({ where: { organizationId, userId, name } })

    if (!apiKey) {
      throw new NotFoundException('API key not found')
    }

    await this.apiKeyRepository.remove(apiKey)
  }

  async updateLastUsedAt(organizationId: string, userId: string, name: string, lastUsedAt: Date): Promise<void> {
    const cooldownKey = `cooldown-${organizationId}-${userId}-${name}`

    // redis for cooldown period - 10 seconds
    // prevents database flooding when multiple requests are made at the same time
    const lastUpdateTime = await this.redisLockProvider.get(cooldownKey)
    if (lastUpdateTime) {
      const timestamp = parseInt(lastUpdateTime)

      if (isNaN(timestamp) || timestamp <= 0 || timestamp > Date.now()) {
        this.logger.warn(`Invalid timestamp in Redis for API key ${name}: ${lastUpdateTime}, proceeding with update`)
      } else {
        const timeSinceLastUpdate = Date.now() - timestamp
        const tenSecondsInMs = 10 * 1000

        if (timeSinceLastUpdate < tenSecondsInMs) {
          this.logger.debug(`Skipping update for API key ${name} - last update was ${timeSinceLastUpdate}ms ago`)
          return
        }
      }
    }

    await this.apiKeyRepository.update(
      {
        organizationId,
        userId,
        name,
      },
      { lastUsedAt },
    )

    // Set the cooldown key in Redis with current timestamp
    await this.redisLockProvider.set(cooldownKey, Date.now().toString(), 60) // 60 second TTL
  }

  private getEffectivePermissions(
    apiKey: ApiKey,
    organizationUser: OrganizationUser,
  ): OrganizationResourcePermission[] {
    if (organizationUser.role === OrganizationMemberRole.OWNER) {
      return apiKey.permissions
    }
    const organizationUserPermissions = new Set(organizationUser.assignedRoles.flatMap((role) => role.permissions))
    return apiKey.permissions.filter((permission) => organizationUserPermissions.has(permission))
  }
}
