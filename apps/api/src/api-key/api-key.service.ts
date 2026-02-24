/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ConflictException, Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { EntityManager, Repository, ArrayOverlap } from 'typeorm'
import { ApiKey } from './api-key.entity'
import { OrganizationResourcePermission } from '../organization/enums/organization-resource-permission.enum'
import { RedisLockProvider } from '../sandbox/common/redis-lock.provider'
import { OnAsyncEvent } from '../common/decorators/on-async-event.decorator'
import { OrganizationEvents } from '../organization/constants/organization-events.constant'
import { OrganizationResourcePermissionsUnassignedEvent } from '../organization/events/organization-resource-permissions-unassigned.event'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { generateApiKeyHash, generateApiKeyValue } from '../common/utils/api-key'
import { OrganizationDeletedEvent } from '../organization/events/organization-deleted.event'
import { getApiKeyValidationCacheKey, getApiKeyUserCacheKey } from './constants/api-key-cache-keys.constant'

@Injectable()
export class ApiKeyService {
  private readonly logger = new Logger(ApiKeyService.name)

  constructor(
    @InjectRepository(ApiKey)
    private apiKeyRepository: Repository<ApiKey>,
    private readonly redisLockProvider: RedisLockProvider,
    @InjectRedis() private readonly redis: Redis,
  ) {}

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
    apiKeyValue?: string,
  ): Promise<{ apiKey: ApiKey; value: string }> {
    const existingKey = await this.apiKeyRepository.findOne({ where: { organizationId, userId, name } })
    if (existingKey) {
      throw new ConflictException('API key with this name already exists')
    }

    const value = apiKeyValue || generateApiKeyValue()

    const apiKey = await this.apiKeyRepository.save({
      organizationId,
      userId,
      name,
      keyHash: generateApiKeyHash(value),
      keyPrefix: this.getApiKeyPrefix(value),
      keySuffix: this.getApiKeySuffix(value),
      permissions,
      createdAt: new Date(),
      expiresAt,
    })

    return { apiKey, value }
  }

  async getApiKeys(organizationId: string, userId?: string): Promise<ApiKey[]> {
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

    return apiKeys
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

    return apiKey
  }

  async getApiKeyByValue(value: string): Promise<ApiKey> {
    const apiKey = await this.apiKeyRepository.findOne({
      where: {
        keyHash: generateApiKeyHash(value),
      },
    })

    if (!apiKey) {
      throw new NotFoundException('API key not found')
    }

    return apiKey
  }

  async deleteApiKey(organizationId: string, userId: string, name: string): Promise<void> {
    const apiKey = await this.apiKeyRepository.findOne({ where: { organizationId, userId, name } })

    if (!apiKey) {
      throw new NotFoundException('API key not found')
    }

    await this.deleteWithEntityManager(this.apiKeyRepository.manager, apiKey)
  }

  async updateLastUsedAt(organizationId: string, userId: string, name: string, lastUsedAt: Date): Promise<void> {
    const cooldownKey = `api-key-last-used-update-${organizationId}-${userId}-${name}`

    const aquired = await this.redisLockProvider.lock(cooldownKey, 10)

    // redis for cooldown period - 10 seconds
    // prevents database flooding when multiple requests are made at the same time
    if (!aquired) {
      return
    }

    await this.apiKeyRepository.update(
      {
        organizationId,
        userId,
        name,
      },
      { lastUsedAt },
    )
  }

  private async deleteWithEntityManager(entityManager: EntityManager, apiKey: ApiKey): Promise<void> {
    await entityManager.remove(apiKey)
    // Invalidate cache when API key is deleted
    await this.invalidateApiKeyCache(apiKey.keyHash)
  }

  private async invalidateApiKeyCache(keyHash: string): Promise<void> {
    try {
      const cacheKey = getApiKeyValidationCacheKey(keyHash)
      await this.redis.del(cacheKey)
      this.logger.debug(`Invalidated cache for API key: ${cacheKey}`)
    } catch (error) {
      this.logger.error('Error invalidating API key cache:', error)
    }
  }

  @OnAsyncEvent({
    event: OrganizationEvents.PERMISSIONS_UNASSIGNED,
  })
  async handleOrganizationResourcePermissionsUnassignedEvent(
    payload: OrganizationResourcePermissionsUnassignedEvent,
  ): Promise<void> {
    const apiKeysToRevoke = await this.apiKeyRepository.find({
      where: {
        organizationId: payload.organizationId,
        userId: payload.userId,
        permissions: ArrayOverlap(payload.unassignedPermissions),
      },
    })

    await Promise.all(apiKeysToRevoke.map((apiKey) => this.deleteWithEntityManager(payload.entityManager, apiKey)))
  }

  @OnAsyncEvent({
    event: OrganizationEvents.DELETED,
  })
  async handleOrganizationDeletedEvent(payload: OrganizationDeletedEvent): Promise<void> {
    const { entityManager, organizationId } = payload

    // Get API keys before deletion to invalidate caches
    const apiKeys = await entityManager.find(ApiKey, {
      where: { organizationId },
      select: ['keyHash', 'userId'],
    })

    await entityManager.delete(ApiKey, { organizationId })

    // Invalidate caches
    try {
      const validationCacheKeys = apiKeys.map((apiKey) => getApiKeyValidationCacheKey(apiKey.keyHash))
      const userCacheKeys = apiKeys.map((apiKey) => getApiKeyUserCacheKey(apiKey.userId))
      const allKeys = [...validationCacheKeys, ...userCacheKeys]
      if (allKeys.length > 0) {
        const BATCH_SIZE = 500
        for (let i = 0; i < allKeys.length; i += BATCH_SIZE) {
          const batch = allKeys.slice(i, i + BATCH_SIZE)
          await this.redis.del(...batch)
        }
      }
    } catch (error) {
      this.logger.error(`Failed to invalidate API key caches for organization ${organizationId}:`, error)
    }
  }
}
