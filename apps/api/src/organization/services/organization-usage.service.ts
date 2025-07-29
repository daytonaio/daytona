/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BadRequestException, Injectable, Logger, NotFoundException } from '@nestjs/common'
import { OnEvent } from '@nestjs/event-emitter'
import { InjectRepository } from '@nestjs/typeorm'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { In, Not, Repository } from 'typeorm'
import { SANDBOX_USAGE_IGNORED_STATES } from '../constants/sandbox-usage-ignored-states.constant'
import { SANDBOX_USAGE_INACTIVE_STATES } from '../constants/sandbox-usage-inactive-states.constant'
import { SNAPSHOT_USAGE_IGNORED_STATES } from '../constants/snapshot-usage-ignored-states.constant'
import { VOLUME_USAGE_IGNORED_STATES } from '../constants/volume-usage-ignored-states.constant'
import { OrganizationUsageOverviewDto } from '../dto/organization-usage-overview.dto'
import { SandboxUsageOverviewInternalDto } from '../dto/sandbox-usage-overview-internal.dto'
import { SnapshotUsageOverviewInternalDto } from '../dto/snapshot-usage-overview-internal.dto'
import { VolumeUsageOverviewInternalDto } from '../dto/volume-usage-overview-internal.dto'
import { Organization } from '../entities/organization.entity'
import {
  getResourceTypeFromQuota,
  OrganizationUsageQuotaType,
  OrganizationUsageResourceType,
} from '../helpers/organization-usage.helper'
import { RedisLockProvider } from '../../sandbox/common/redis-lock.provider'
import { SandboxEvents } from '../../sandbox/constants/sandbox-events.constants'
import { SnapshotEvents } from '../../sandbox/constants/snapshot-events'
import { VolumeEvents } from '../../sandbox/constants/volume-events'
import { Sandbox } from '../../sandbox/entities/sandbox.entity'
import { Snapshot } from '../../sandbox/entities/snapshot.entity'
import { Volume } from '../../sandbox/entities/volume.entity'
import { SandboxCreatedEvent } from '../../sandbox/events/sandbox-create.event'
import { SandboxStateUpdatedEvent } from '../../sandbox/events/sandbox-state-updated.event'
import { SnapshotCreatedEvent } from '../../sandbox/events/snapshot-created.event'
import { SnapshotStateUpdatedEvent } from '../../sandbox/events/snapshot-state-updated.event'
import { VolumeCreatedEvent } from '../../sandbox/events/volume-created.event'
import { VolumeStateUpdatedEvent } from '../../sandbox/events/volume-state-updated.event'

@Injectable()
export class OrganizationUsageService {
  private readonly logger = new Logger(OrganizationUsageService.name)

  private readonly CACHE_TTL_SECONDS = 10

  // cache is considered stale if it was last populated from db more than CACHE_MAX_AGE_MS ago
  private readonly CACHE_MAX_AGE_MS = 60 * 60 * 1000

  constructor(
    @InjectRedis() private readonly redis: Redis,
    @InjectRepository(Organization)
    private readonly organizationRepository: Repository<Organization>,
    @InjectRepository(Sandbox)
    private readonly sandboxRepository: Repository<Sandbox>,
    @InjectRepository(Snapshot)
    private readonly snapshotRepository: Repository<Snapshot>,
    @InjectRepository(Volume)
    private readonly volumeRepository: Repository<Volume>,
    private readonly redisLockProvider: RedisLockProvider,
  ) {}

  // 1. public methods for total/sandbox/snapshot/volume usage overviews

  async getUsageOverview(organizationId: string, organization?: Organization): Promise<OrganizationUsageOverviewDto> {
    if (organization && organization.id !== organizationId) {
      throw new BadRequestException('Organization ID mismatch')
    }

    if (!organization) {
      organization = await this.organizationRepository.findOne({ where: { id: organizationId } })
    }

    if (!organization) {
      throw new NotFoundException(`Organization with ID ${organizationId} not found`)
    }

    const sandboxUsageOverview = await this.getSandboxUsageOverview(organizationId)
    const snapshotUsageOverview = await this.getSnapshotUsageOverview(organizationId)
    const volumeUsageOverview = await this.getVolumeUsageOverview(organizationId)

    return {
      totalCpuQuota: organization.totalCpuQuota,
      totalMemoryQuota: organization.totalMemoryQuota,
      totalDiskQuota: organization.totalDiskQuota,
      totalSnapshotQuota: organization.snapshotQuota,
      totalVolumeQuota: organization.volumeQuota,
      ...sandboxUsageOverview,
      ...snapshotUsageOverview,
      ...volumeUsageOverview,
    }
  }

  async getSandboxUsageOverview(
    organizationId: string,
    excludeSandboxId?: string,
  ): Promise<SandboxUsageOverviewInternalDto> {
    let cachedUsageOverview = await this.getCachedSandboxUsageOverview(organizationId)

    // cache hit
    if (cachedUsageOverview) {
      if (excludeSandboxId) {
        return await this.excludeSandboxFromUsageOverview(cachedUsageOverview, excludeSandboxId)
      }

      return cachedUsageOverview
    }

    // cache miss, wait for lock
    const lockKey = `org:${organizationId}:fetch-sandbox-usage-from-db`
    await this.redisLockProvider.waitForLock(lockKey, 60)

    try {
      // check if cache was updated while waiting for lock
      cachedUsageOverview = await this.getCachedSandboxUsageOverview(organizationId)

      // cache hit
      if (cachedUsageOverview) {
        if (excludeSandboxId) {
          return await this.excludeSandboxFromUsageOverview(cachedUsageOverview, excludeSandboxId)
        }

        return cachedUsageOverview
      }

      // cache miss, fetch from db
      const usageOverview = await this.fetchSandboxUsageFromDb(organizationId)

      if (excludeSandboxId) {
        return await this.excludeSandboxFromUsageOverview(usageOverview, excludeSandboxId)
      }

      return usageOverview
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  private async excludeSandboxFromUsageOverview(
    usageOverview: SandboxUsageOverviewInternalDto,
    excludeSandboxId: string,
  ): Promise<SandboxUsageOverviewInternalDto> {
    const excludedSandbox = await this.sandboxRepository.findOne({
      where: { id: excludeSandboxId },
    })

    if (!excludedSandbox) {
      return usageOverview
    }

    let cpuToSubtract = 0
    let memToSubtract = 0
    let diskToSubtract = 0

    if (!SANDBOX_USAGE_IGNORED_STATES.includes(excludedSandbox.state)) {
      diskToSubtract = excludedSandbox.disk
    }

    if (!SANDBOX_USAGE_INACTIVE_STATES.includes(excludedSandbox.state)) {
      cpuToSubtract = excludedSandbox.cpu
      memToSubtract = excludedSandbox.mem
    }

    return {
      ...usageOverview,
      currentCpuUsage: Math.max(0, usageOverview.currentCpuUsage - cpuToSubtract),
      currentMemoryUsage: Math.max(0, usageOverview.currentMemoryUsage - memToSubtract),
      currentDiskUsage: Math.max(0, usageOverview.currentDiskUsage - diskToSubtract),
    }
  }

  async getSnapshotUsageOverview(organizationId: string): Promise<SnapshotUsageOverviewInternalDto> {
    let cachedUsageOverview = await this.getCachedSnapshotUsageOverview(organizationId)

    // cache hit
    if (cachedUsageOverview) {
      return cachedUsageOverview
    }

    // cache miss, wait for lock
    const lockKey = `org:${organizationId}:fetch-snapshot-usage-from-db`
    await this.redisLockProvider.waitForLock(lockKey, 60)

    try {
      // check if cache was updated while waiting for lock
      cachedUsageOverview = await this.getCachedSnapshotUsageOverview(organizationId)

      // cache hit
      if (cachedUsageOverview) {
        return cachedUsageOverview
      }

      // cache miss, fetch from db
      return await this.fetchSnapshotUsageFromDb(organizationId)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  async getVolumeUsageOverview(organizationId: string): Promise<VolumeUsageOverviewInternalDto> {
    let cachedUsageOverview = await this.getCachedVolumeUsageOverview(organizationId)

    // cache hit
    if (cachedUsageOverview) {
      return cachedUsageOverview
    }

    // cache miss, wait for lock
    const lockKey = `org:${organizationId}:fetch-volume-usage-from-db`
    await this.redisLockProvider.waitForLock(lockKey, 60)

    try {
      // check if cache was updated while waiting for lock
      cachedUsageOverview = await this.getCachedVolumeUsageOverview(organizationId)

      // cache hit
      if (cachedUsageOverview) {
        return cachedUsageOverview
      }

      // cache miss, fetch from db
      return await this.fetchVolumeUsageFromDb(organizationId)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  // 2. helpers for getting sandbox/snapshot/volume usage overviews from cache

  private async getCachedSandboxUsageOverview(organizationId: string): Promise<SandboxUsageOverviewInternalDto | null> {
    const cpuUsage = await this.getQuotaUsageCachedValue(organizationId, 'cpu')
    const memoryUsage = await this.getQuotaUsageCachedValue(organizationId, 'memory')
    const diskUsage = await this.getQuotaUsageCachedValue(organizationId, 'disk')

    if (cpuUsage === null || memoryUsage === null || diskUsage === null) {
      return null
    }

    return {
      currentCpuUsage: cpuUsage,
      currentMemoryUsage: memoryUsage,
      currentDiskUsage: diskUsage,
    }
  }

  private async getCachedSnapshotUsageOverview(
    organizationId: string,
  ): Promise<SnapshotUsageOverviewInternalDto | null> {
    const snapshotUsage = await this.getQuotaUsageCachedValue(organizationId, 'snapshot_count')

    if (snapshotUsage === null) {
      return null
    }

    return {
      currentSnapshotUsage: snapshotUsage,
    }
  }

  private async getCachedVolumeUsageOverview(organizationId: string): Promise<VolumeUsageOverviewInternalDto | null> {
    const volumeUsage = await this.getQuotaUsageCachedValue(organizationId, 'volume_count')

    if (volumeUsage === null) {
      return null
    }

    return {
      currentVolumeUsage: volumeUsage,
    }
  }

  // 3. helpers for fetching sandbox/snapshot/volume usage overviews from db and caching them

  private async fetchSandboxUsageFromDb(organizationId: string): Promise<SandboxUsageOverviewInternalDto> {
    // fetch from db
    const sandboxUsageMetrics: {
      used_cpu: number
      used_mem: number
      used_disk: number
    } = await this.sandboxRepository
      .createQueryBuilder('sandbox')
      .select([
        'SUM(CASE WHEN sandbox.state NOT IN (:...inactiveStates) THEN sandbox.cpu ELSE 0 END) as used_cpu',
        'SUM(CASE WHEN sandbox.state NOT IN (:...inactiveStates) THEN sandbox.mem ELSE 0 END) as used_mem',
        'SUM(CASE WHEN sandbox.state NOT IN (:...ignoredStates) THEN sandbox.disk ELSE 0 END) as used_disk',
      ])
      .where('sandbox.organizationId = :organizationId', { organizationId })
      .setParameter('ignoredStates', SANDBOX_USAGE_IGNORED_STATES)
      .setParameter('inactiveStates', SANDBOX_USAGE_INACTIVE_STATES)
      .getRawOne()

    const cpuUsage = Number(sandboxUsageMetrics.used_cpu) || 0
    const memoryUsage = Number(sandboxUsageMetrics.used_mem) || 0
    const diskUsage = Number(sandboxUsageMetrics.used_disk) || 0

    // cache the results
    const cpuCacheKey = this.getQuotaUsageCacheKey(organizationId, 'cpu')
    const memoryCacheKey = this.getQuotaUsageCacheKey(organizationId, 'memory')
    const diskCacheKey = this.getQuotaUsageCacheKey(organizationId, 'disk')

    await this.redis
      .multi()
      .setex(cpuCacheKey, this.CACHE_TTL_SECONDS, cpuUsage)
      .setex(memoryCacheKey, this.CACHE_TTL_SECONDS, memoryUsage)
      .setex(diskCacheKey, this.CACHE_TTL_SECONDS, diskUsage)
      .exec()

    await this.resetCacheStaleness(organizationId, 'sandbox')

    return {
      currentCpuUsage: cpuUsage,
      currentMemoryUsage: memoryUsage,
      currentDiskUsage: diskUsage,
    }
  }

  private async fetchSnapshotUsageFromDb(organizationId: string): Promise<SnapshotUsageOverviewInternalDto> {
    // fetch from db
    const snapshotUsage = await this.snapshotRepository.count({
      where: {
        organizationId,
        state: Not(In(SNAPSHOT_USAGE_IGNORED_STATES)),
      },
    })

    // cache the result
    const cacheKey = this.getQuotaUsageCacheKey(organizationId, 'snapshot_count')
    await this.redis.setex(cacheKey, this.CACHE_TTL_SECONDS, snapshotUsage)

    await this.resetCacheStaleness(organizationId, 'snapshot')

    return {
      currentSnapshotUsage: snapshotUsage,
    }
  }

  private async fetchVolumeUsageFromDb(organizationId: string): Promise<VolumeUsageOverviewInternalDto> {
    // fetch from db
    const volumeUsage = await this.volumeRepository.count({
      where: {
        organizationId,
        state: Not(In(VOLUME_USAGE_IGNORED_STATES)),
      },
    })

    // cache the result
    const cacheKey = this.getQuotaUsageCacheKey(organizationId, 'volume_count')
    await this.redis.setex(cacheKey, this.CACHE_TTL_SECONDS, volumeUsage)

    await this.resetCacheStaleness(organizationId, 'volume')

    return {
      currentVolumeUsage: volumeUsage,
    }
  }

  // 4. helpers for cached quota usage values

  private getQuotaUsageCacheKey(organizationId: string, quotaType: OrganizationUsageQuotaType): string {
    return `org:${organizationId}:quota:${quotaType}:usage`
  }

  private async getQuotaUsageCachedValue(
    organizationId: string,
    quotaType: OrganizationUsageQuotaType,
  ): Promise<number | null> {
    const cacheKey = this.getQuotaUsageCacheKey(organizationId, quotaType)
    const cachedData = await this.redis.get(cacheKey)

    if (!cachedData) {
      return null
    }

    // must be a non-negative number
    const parsedValue = Number(cachedData)
    if (isNaN(parsedValue) || parsedValue < 0) {
      return null
    }

    const resourceType = getResourceTypeFromQuota(quotaType)
    const isStale = await this.isCacheStale(organizationId, resourceType)

    if (isStale) {
      return null
    }

    return parsedValue
  }

  private async updateQuotaUsage(
    organizationId: string,
    quotaType: OrganizationUsageQuotaType,
    delta: number,
  ): Promise<void> {
    // must be no-op if cache not present
    const script = `
    if redis.call("EXISTS", KEYS[1]) == 1 then
      redis.call("INCRBY", KEYS[1], ARGV[1])
      redis.call("EXPIRE", KEYS[1], ARGV[2])
    end
  `
    const cacheKey = this.getQuotaUsageCacheKey(organizationId, quotaType)
    await this.redis.eval(script, 1, cacheKey, delta.toString(), this.CACHE_TTL_SECONDS.toString())
  }

  // 5. helpers for cache staleness

  private getCacheStalenessKey(organizationId: string, resourceType: OrganizationUsageResourceType): string {
    return `org:${organizationId}:resource:${resourceType}:usage:fetched_at`
  }

  private async resetCacheStaleness(
    organizationId: string,
    resourceType: OrganizationUsageResourceType,
  ): Promise<void> {
    const cacheKey = this.getCacheStalenessKey(organizationId, resourceType)
    await this.redis.set(cacheKey, Date.now())
  }

  private async isCacheStale(organizationId: string, resourceType: OrganizationUsageResourceType): Promise<boolean> {
    const cacheKey = this.getCacheStalenessKey(organizationId, resourceType)
    const cachedData = await this.redis.get(cacheKey)

    if (!cachedData) {
      return true
    }

    const lastFetchedAtTimestamp = Number(cachedData)
    if (isNaN(lastFetchedAtTimestamp)) {
      return true
    }

    return Date.now() - lastFetchedAtTimestamp > this.CACHE_MAX_AGE_MS
  }

  // 6. event handlers for updating quota usage in cache

  @OnEvent(SandboxEvents.CREATED)
  async handleSandboxCreated(event: SandboxCreatedEvent) {
    const lockKey = `sandbox:${event.sandbox.id}:quota-usage-update`
    await this.redisLockProvider.waitForLock(lockKey, 60)

    try {
      await this.updateQuotaUsage(event.sandbox.organizationId, 'cpu', event.sandbox.cpu)
      await this.updateQuotaUsage(event.sandbox.organizationId, 'memory', event.sandbox.mem)
      await this.updateQuotaUsage(event.sandbox.organizationId, 'disk', event.sandbox.disk)
    } catch (error) {
      this.logger.warn(
        `Error updating cached sandbox quota usage for organization ${event.sandbox.organizationId}`,
        error,
      )
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @OnEvent(SandboxEvents.STATE_UPDATED)
  async handleSandboxStateUpdated(event: SandboxStateUpdatedEvent) {
    const lockKey = `sandbox:${event.sandbox.id}:quota-usage-update`
    await this.redisLockProvider.waitForLock(lockKey, 60)

    try {
      const cpuDelta = this.calculateQuotaUsageDelta(
        event.sandbox.cpu,
        event.oldState,
        event.newState,
        SANDBOX_USAGE_INACTIVE_STATES,
      )

      const memDelta = this.calculateQuotaUsageDelta(
        event.sandbox.mem,
        event.oldState,
        event.newState,
        SANDBOX_USAGE_INACTIVE_STATES,
      )

      const diskDelta = this.calculateQuotaUsageDelta(
        event.sandbox.disk,
        event.oldState,
        event.newState,
        SANDBOX_USAGE_IGNORED_STATES,
      )

      if (cpuDelta !== 0) {
        await this.updateQuotaUsage(event.sandbox.organizationId, 'cpu', cpuDelta)
      }

      if (memDelta !== 0) {
        await this.updateQuotaUsage(event.sandbox.organizationId, 'memory', memDelta)
      }

      if (diskDelta !== 0) {
        await this.updateQuotaUsage(event.sandbox.organizationId, 'disk', diskDelta)
      }
    } catch (error) {
      this.logger.warn(
        `Error updating cached sandbox quota usage for organization ${event.sandbox.organizationId}`,
        error,
      )
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @OnEvent(SnapshotEvents.CREATED)
  async handleSnapshotCreated(event: SnapshotCreatedEvent) {
    const lockKey = `snapshot:${event.snapshot.id}:quota-usage-update`
    await this.redisLockProvider.waitForLock(lockKey, 60)

    try {
      await this.updateQuotaUsage(event.snapshot.organizationId, 'snapshot_count', 1)
    } catch (error) {
      this.logger.warn(
        `Error updating cached snapshot quota usage for organization ${event.snapshot.organizationId}`,
        error,
      )
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @OnEvent(SnapshotEvents.STATE_UPDATED)
  async handleSnapshotStateUpdated(event: SnapshotStateUpdatedEvent) {
    const lockKey = `snapshot:${event.snapshot.id}:quota-usage-update`
    await this.redisLockProvider.waitForLock(lockKey, 60)

    try {
      const countDelta = this.calculateQuotaUsageDelta(1, event.oldState, event.newState, SNAPSHOT_USAGE_IGNORED_STATES)

      if (countDelta !== 0) {
        await this.updateQuotaUsage(event.snapshot.organizationId, 'snapshot_count', countDelta)
      }
    } catch (error) {
      this.logger.warn(
        `Error updating cached snapshot quota usage for organization ${event.snapshot.organizationId}`,
        error,
      )
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @OnEvent(VolumeEvents.CREATED)
  async handleVolumeCreated(event: VolumeCreatedEvent) {
    const lockKey = `volume:${event.volume.id}:quota-usage-update`
    await this.redisLockProvider.waitForLock(lockKey, 60)

    try {
      await this.updateQuotaUsage(event.volume.organizationId, 'volume_count', 1)
    } catch (error) {
      this.logger.warn(
        `Error updating cached volume quota usage for organization ${event.volume.organizationId}`,
        error,
      )
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @OnEvent(VolumeEvents.STATE_UPDATED)
  async handleVolumeStateUpdated(event: VolumeStateUpdatedEvent) {
    const lockKey = `volume:${event.volume.id}:quota-usage-update`
    await this.redisLockProvider.waitForLock(lockKey, 60)

    try {
      const countDelta = this.calculateQuotaUsageDelta(1, event.oldState, event.newState, VOLUME_USAGE_IGNORED_STATES)

      if (countDelta !== 0) {
        await this.updateQuotaUsage(event.volume.organizationId, 'volume_count', countDelta)
      }
    } catch (error) {
      this.logger.warn(
        `Error updating cached volume quota usage for organization ${event.volume.organizationId}`,
        error,
      )
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  private calculateQuotaUsageDelta<TState>(
    resourceAmount: number,
    oldState: TState,
    newState: TState,
    nonConsumingStates: TState[],
  ): number {
    const wasConsumingResources = !nonConsumingStates.includes(oldState)
    const isConsumingResources = !nonConsumingStates.includes(newState)

    if (!wasConsumingResources && isConsumingResources) {
      return resourceAmount
    }

    if (wasConsumingResources && !isConsumingResources) {
      return -resourceAmount
    }

    return 0
  }
}
