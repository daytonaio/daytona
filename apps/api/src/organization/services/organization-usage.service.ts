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
import { SANDBOX_STATES_CONSUMING_COMPUTE } from '../constants/sandbox-states-consuming-compute.constant'
import { SANDBOX_STATES_CONSUMING_DISK } from '../constants/sandbox-states-consuming-disk.constant'
import { SNAPSHOT_USAGE_IGNORED_STATES } from '../constants/snapshot-usage-ignored-states.constant'
import { VOLUME_USAGE_IGNORED_STATES } from '../constants/volume-usage-ignored-states.constant'
import { OrganizationUsageOverviewDto } from '../dto/organization-usage-overview.dto'
import {
  PendingSandboxUsageOverviewInternalDto,
  SandboxUsageOverviewInternalDto,
  SandboxUsageOverviewWithPendingInternalDto,
} from '../dto/sandbox-usage-overview-internal.dto'
import {
  PendingSnapshotUsageOverviewInternalDto,
  SnapshotUsageOverviewInternalDto,
  SnapshotUsageOverviewWithPendingInternalDto,
} from '../dto/snapshot-usage-overview-internal.dto'
import {
  PendingVolumeUsageOverviewInternalDto,
  VolumeUsageOverviewInternalDto,
  VolumeUsageOverviewWithPendingInternalDto,
} from '../dto/volume-usage-overview-internal.dto'
import { Organization } from '../entities/organization.entity'
import { OrganizationUsageQuotaType, OrganizationUsageResourceType } from '../helpers/organization-usage.helper'
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

  /**
   * Time-to-live for cached quota usage values
   */
  private readonly CACHE_TTL_SECONDS = 60

  /**
   * Cache is considered stale if it was last populated from db more than `CACHE_MAX_AGE_MS` ago
   */
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

  /**
   * Get the current usage overview for all organization quotas.
   *
   * @param organizationId
   * @param organization - Provide the organization entity to avoid fetching it from the database (optional)
   */
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
      currentCpuUsage: sandboxUsageOverview.currentCpuUsage,
      currentMemoryUsage: sandboxUsageOverview.currentMemoryUsage,
      currentDiskUsage: sandboxUsageOverview.currentDiskUsage,
      currentSnapshotUsage: snapshotUsageOverview.currentSnapshotUsage,
      currentVolumeUsage: volumeUsageOverview.currentVolumeUsage,
    }
  }

  /**
   * Get the current and pending usage overview for sandbox-related organization quotas.
   *
   * @param organizationId
   * @param excludeSandboxId - If provided, the usage overview will exclude the current usage of the sandbox with the given ID
   */
  async getSandboxUsageOverview(
    organizationId: string,
    excludeSandboxId?: string,
  ): Promise<SandboxUsageOverviewWithPendingInternalDto> {
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

      // get pending usage separately since it's not stored in DB
      const pendingUsageOverview = await this.getCachedPendingSandboxUsageOverview(organizationId)

      const combinedUsageOverview: SandboxUsageOverviewWithPendingInternalDto = {
        ...usageOverview,
        ...pendingUsageOverview,
      }

      if (excludeSandboxId) {
        return await this.excludeSandboxFromUsageOverview(combinedUsageOverview, excludeSandboxId)
      }

      return combinedUsageOverview
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  /**
   * Get the current and pending usage overview for snapshot-related organization quotas.
   *
   * @param organizationId
   */
  async getSnapshotUsageOverview(organizationId: string): Promise<SnapshotUsageOverviewWithPendingInternalDto> {
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
      const usageOverview = await this.fetchSnapshotUsageFromDb(organizationId)

      // get pending usage separately since it's not stored in DB
      const pendingUsageOverview = await this.getCachedPendingSnapshotUsageOverview(organizationId)

      return {
        ...usageOverview,
        ...pendingUsageOverview,
      }
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  /**
   * Get the current and pending usage overview for volume-related organization quotas.
   *
   * @param organizationId
   */
  async getVolumeUsageOverview(organizationId: string): Promise<VolumeUsageOverviewWithPendingInternalDto> {
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
      const usageOverview = await this.fetchVolumeUsageFromDb(organizationId)

      // get pending usage separately since it's not stored in DB
      const pendingUsageOverview = await this.getCachedPendingVolumeUsageOverview(organizationId)

      return {
        ...usageOverview,
        ...pendingUsageOverview,
      }
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  /**
   * Exclude the current usage of a specific sandbox from the usage overview.
   *
   * @param usageOverview
   * @param excludeSandboxId
   */
  private async excludeSandboxFromUsageOverview<T extends SandboxUsageOverviewInternalDto>(
    usageOverview: T,
    excludeSandboxId: string,
  ): Promise<T> {
    const excludedSandbox = await this.sandboxRepository.findOne({
      where: { id: excludeSandboxId },
    })

    if (!excludedSandbox) {
      return usageOverview
    }

    let cpuToSubtract = 0
    let memToSubtract = 0
    let diskToSubtract = 0

    if (SANDBOX_STATES_CONSUMING_COMPUTE.includes(excludedSandbox.state)) {
      cpuToSubtract = excludedSandbox.cpu
      memToSubtract = excludedSandbox.mem
    }

    if (SANDBOX_STATES_CONSUMING_DISK.includes(excludedSandbox.state)) {
      diskToSubtract = excludedSandbox.disk
    }

    return {
      ...usageOverview,
      currentCpuUsage: Math.max(0, usageOverview.currentCpuUsage - cpuToSubtract),
      currentMemoryUsage: Math.max(0, usageOverview.currentMemoryUsage - memToSubtract),
      currentDiskUsage: Math.max(0, usageOverview.currentDiskUsage - diskToSubtract),
    }
  }

  /**
   * Get the cached current and pending usage overview for sandbox-related organization quotas.
   *
   * @param organizationId
   */
  private async getCachedSandboxUsageOverview(
    organizationId: string,
  ): Promise<SandboxUsageOverviewWithPendingInternalDto | null> {
    const script = `
      return {
        redis.call("GET", KEYS[1]),
        redis.call("GET", KEYS[2]),
        redis.call("GET", KEYS[3]),
        redis.call("GET", KEYS[4]),
        redis.call("GET", KEYS[5]),
        redis.call("GET", KEYS[6])
      }
    `

    const result = (await this.redis.eval(
      script,
      6,
      this.getCurrentQuotaUsageCacheKey(organizationId, 'cpu'),
      this.getCurrentQuotaUsageCacheKey(organizationId, 'memory'),
      this.getCurrentQuotaUsageCacheKey(organizationId, 'disk'),
      this.getPendingQuotaUsageCacheKey(organizationId, 'cpu'),
      this.getPendingQuotaUsageCacheKey(organizationId, 'memory'),
      this.getPendingQuotaUsageCacheKey(organizationId, 'disk'),
    )) as (string | null)[]

    const [cpuUsage, memoryUsage, diskUsage, pendingCpuUsage, pendingMemoryUsage, pendingDiskUsage] = result

    // Cache miss
    if (cpuUsage === null || memoryUsage === null || diskUsage === null) {
      return null
    }

    // Check cache staleness for current usage
    const isStale = await this.isCacheStale(organizationId, 'sandbox')

    if (isStale) {
      return null
    }

    // Validate current usage values are non-negative numbers
    const parsedCpuUsage = this.parseNonNegativeCachedValue(cpuUsage)
    const parsedMemoryUsage = this.parseNonNegativeCachedValue(memoryUsage)
    const parsedDiskUsage = this.parseNonNegativeCachedValue(diskUsage)

    if (parsedCpuUsage === null || parsedMemoryUsage === null || parsedDiskUsage === null) {
      return null
    }

    // Parse pending usage values (null is acceptable)
    const parsedPendingCpuUsage = this.parseNonNegativeCachedValue(pendingCpuUsage)
    const parsedPendingMemoryUsage = this.parseNonNegativeCachedValue(pendingMemoryUsage)
    const parsedPendingDiskUsage = this.parseNonNegativeCachedValue(pendingDiskUsage)

    return {
      currentCpuUsage: parsedCpuUsage,
      currentMemoryUsage: parsedMemoryUsage,
      currentDiskUsage: parsedDiskUsage,
      pendingCpuUsage: parsedPendingCpuUsage,
      pendingMemoryUsage: parsedPendingMemoryUsage,
      pendingDiskUsage: parsedPendingDiskUsage,
    }
  }

  /**
   * Get the cached pending usage overview for sandbox-related organization quotas.
   *
   * @param organizationId
   */
  private async getCachedPendingSandboxUsageOverview(
    organizationId: string,
  ): Promise<PendingSandboxUsageOverviewInternalDto> {
    const script = `
      return {
        redis.call("GET", KEYS[1]),
        redis.call("GET", KEYS[2]),
        redis.call("GET", KEYS[3])
      }
    `
    const result = (await this.redis.eval(
      script,
      3,
      this.getPendingQuotaUsageCacheKey(organizationId, 'cpu'),
      this.getPendingQuotaUsageCacheKey(organizationId, 'memory'),
      this.getPendingQuotaUsageCacheKey(organizationId, 'disk'),
    )) as (string | null)[]

    const [pendingCpuUsage, pendingMemoryUsage, pendingDiskUsage] = result

    const parsedPendingCpuUsage = this.parseNonNegativeCachedValue(pendingCpuUsage)
    const parsedPendingMemoryUsage = this.parseNonNegativeCachedValue(pendingMemoryUsage)
    const parsedPendingDiskUsage = this.parseNonNegativeCachedValue(pendingDiskUsage)

    return {
      pendingCpuUsage: parsedPendingCpuUsage,
      pendingMemoryUsage: parsedPendingMemoryUsage,
      pendingDiskUsage: parsedPendingDiskUsage,
    }
  }

  /**
   * Get the cached overview for current and pending usage for snapshot-related organization quotas.
   *
   * @param organizationId
   */
  private async getCachedSnapshotUsageOverview(
    organizationId: string,
  ): Promise<SnapshotUsageOverviewWithPendingInternalDto | null> {
    const script = `
      return {
        redis.call("GET", KEYS[1]),
        redis.call("GET", KEYS[2])
      }
    `
    const result = (await this.redis.eval(
      script,
      2,
      this.getCurrentQuotaUsageCacheKey(organizationId, 'snapshot_count'),
      this.getPendingQuotaUsageCacheKey(organizationId, 'snapshot_count'),
    )) as (string | null)[]

    const [currentSnapshotUsage, pendingSnapshotUsage] = result

    // Cache miss
    if (currentSnapshotUsage === null) {
      return null
    }

    // Check cache staleness for current usage
    const isStale = await this.isCacheStale(organizationId, 'snapshot')

    if (isStale) {
      return null
    }

    // Validate current usage values are non-negative numbers
    const parsedCurrentSnapshotUsage = this.parseNonNegativeCachedValue(currentSnapshotUsage)

    if (parsedCurrentSnapshotUsage === null) {
      return null
    }

    // Parse pending usage values (null is acceptable)
    const parsedPendingSnapshotUsage = this.parseNonNegativeCachedValue(pendingSnapshotUsage)

    return {
      currentSnapshotUsage: parsedCurrentSnapshotUsage,
      pendingSnapshotUsage: parsedPendingSnapshotUsage,
    }
  }

  /**
   * Get the cached pending usage overview for snapshot-related organization quotas.
   *
   * @param organizationId
   */
  private async getCachedPendingSnapshotUsageOverview(
    organizationId: string,
  ): Promise<PendingSnapshotUsageOverviewInternalDto> {
    const script = `
      return {
        redis.call("GET", KEYS[1])
      }
    `
    const result = (await this.redis.eval(
      script,
      1,
      this.getPendingQuotaUsageCacheKey(organizationId, 'snapshot_count'),
    )) as (string | null)[]

    const [pendingSnapshotUsage] = result

    // Parse pending usage values (null is acceptable)
    const parsedPendingSnapshotUsage = this.parseNonNegativeCachedValue(pendingSnapshotUsage)

    return {
      pendingSnapshotUsage: parsedPendingSnapshotUsage,
    }
  }

  /**
   * Get the cached overview for current and pending usage for volume-related organization quotas.
   *
   * @param organizationId
   */
  private async getCachedVolumeUsageOverview(
    organizationId: string,
  ): Promise<VolumeUsageOverviewWithPendingInternalDto | null> {
    const script = `
    return {
      redis.call("GET", KEYS[1]),
      redis.call("GET", KEYS[2])
    }
  `

    const result = (await this.redis.eval(
      script,
      2,
      this.getCurrentQuotaUsageCacheKey(organizationId, 'volume_count'),
      this.getPendingQuotaUsageCacheKey(organizationId, 'volume_count'),
    )) as (string | null)[]

    const [currentVolumeUsage, pendingVolumeUsage] = result

    if (currentVolumeUsage === null) {
      return null
    }

    // Check cache staleness for current usage
    const isStale = await this.isCacheStale(organizationId, 'volume')

    if (isStale) {
      return null
    }

    // Validate current usage values are non-negative numbers
    const parsedCurrentVolumeUsage = this.parseNonNegativeCachedValue(currentVolumeUsage)

    if (parsedCurrentVolumeUsage === null) {
      return null
    }

    // Parse pending usage values (null is acceptable)
    const parsedPendingVolumeUsage = this.parseNonNegativeCachedValue(pendingVolumeUsage)

    return {
      currentVolumeUsage: parsedCurrentVolumeUsage,
      pendingVolumeUsage: parsedPendingVolumeUsage,
    }
  }

  /**
   * Get the cached pending usage overview for volume-related organization quotas.
   *
   * @param organizationId
   */
  private async getCachedPendingVolumeUsageOverview(
    organizationId: string,
  ): Promise<PendingVolumeUsageOverviewInternalDto> {
    const script = `
      return {
        redis.call("GET", KEYS[1])
      }
    `

    const result = (await this.redis.eval(
      script,
      1,
      this.getPendingQuotaUsageCacheKey(organizationId, 'volume_count'),
    )) as (string | null)[]

    const [pendingVolumeUsage] = result

    // Parse pending usage values (null is acceptable)
    const parsedPendingVolumeUsage = this.parseNonNegativeCachedValue(pendingVolumeUsage)

    return {
      pendingVolumeUsage: parsedPendingVolumeUsage,
    }
  }

  /**
   * Attempts to parse a given value to a non-negative number.
   *
   * @param value - The value to parse.
   * @returns The parsed non-negative number or `null` if the given value is null or not a non-negative number.
   */
  private parseNonNegativeCachedValue(value: string | null): number | null {
    if (value === null) {
      return null
    }

    const parsedValue = Number(value)

    if (isNaN(parsedValue) || parsedValue < 0) {
      return null
    }

    return parsedValue
  }

  /**
   * Fetch the current usage overview for sandbox-related organization quotas from the database and cache the results.
   *
   * @param organizationId
   */
  async fetchSandboxUsageFromDb(organizationId: string): Promise<SandboxUsageOverviewInternalDto> {
    // fetch from db
    const sandboxUsageMetrics: {
      used_cpu: number
      used_mem: number
      used_disk: number
    } = await this.sandboxRepository
      .createQueryBuilder('sandbox')
      .select([
        'SUM(CASE WHEN sandbox.state IN (:...statesConsumingCompute) THEN sandbox.cpu ELSE 0 END) as used_cpu',
        'SUM(CASE WHEN sandbox.state IN (:...statesConsumingCompute) THEN sandbox.mem ELSE 0 END) as used_mem',
        'SUM(CASE WHEN sandbox.state IN (:...statesConsumingDisk) THEN sandbox.disk ELSE 0 END) as used_disk',
      ])
      .where('sandbox.organizationId = :organizationId', { organizationId })
      .setParameter('statesConsumingCompute', SANDBOX_STATES_CONSUMING_COMPUTE)
      .setParameter('statesConsumingDisk', SANDBOX_STATES_CONSUMING_DISK)
      .getRawOne()

    const cpuUsage = Number(sandboxUsageMetrics.used_cpu) || 0
    const memoryUsage = Number(sandboxUsageMetrics.used_mem) || 0
    const diskUsage = Number(sandboxUsageMetrics.used_disk) || 0

    // cache the results
    const cpuCacheKey = this.getCurrentQuotaUsageCacheKey(organizationId, 'cpu')
    const memoryCacheKey = this.getCurrentQuotaUsageCacheKey(organizationId, 'memory')
    const diskCacheKey = this.getCurrentQuotaUsageCacheKey(organizationId, 'disk')

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

  /**
   * Fetch the current usage overview for snapshot-related organization quotas from the database and cache the results.
   *
   * @param organizationId
   */
  private async fetchSnapshotUsageFromDb(organizationId: string): Promise<SnapshotUsageOverviewInternalDto> {
    // fetch from db
    const snapshotUsage = await this.snapshotRepository.count({
      where: {
        organizationId,
        state: Not(In(SNAPSHOT_USAGE_IGNORED_STATES)),
      },
    })

    // cache the result
    const cacheKey = this.getCurrentQuotaUsageCacheKey(organizationId, 'snapshot_count')
    await this.redis.setex(cacheKey, this.CACHE_TTL_SECONDS, snapshotUsage)

    await this.resetCacheStaleness(organizationId, 'snapshot')

    return {
      currentSnapshotUsage: snapshotUsage,
    }
  }

  /**
   * Fetch the current usage overview for volume-related organization quotas from the database and cache the results.
   *
   * @param organizationId
   */
  private async fetchVolumeUsageFromDb(organizationId: string): Promise<VolumeUsageOverviewInternalDto> {
    // fetch from db
    const volumeUsage = await this.volumeRepository.count({
      where: {
        organizationId,
        state: Not(In(VOLUME_USAGE_IGNORED_STATES)),
      },
    })

    // cache the result
    const cacheKey = this.getCurrentQuotaUsageCacheKey(organizationId, 'volume_count')
    await this.redis.setex(cacheKey, this.CACHE_TTL_SECONDS, volumeUsage)

    await this.resetCacheStaleness(organizationId, 'volume')

    return {
      currentVolumeUsage: volumeUsage,
    }
  }

  /**
   * Get the cache key for the current usage of a given organization quota.
   *
   * @param organizationId
   * @param quotaType
   */
  private getCurrentQuotaUsageCacheKey(organizationId: string, quotaType: OrganizationUsageQuotaType): string {
    return `org:${organizationId}:quota:${quotaType}:usage`
  }

  /**
   * Get the cache key for the pending usage of a given organization quota.
   *
   * @param organizationId
   * @param quotaType
   */
  private getPendingQuotaUsageCacheKey(organizationId: string, quotaType: OrganizationUsageQuotaType): string {
    return `org:${organizationId}:pending-${quotaType}`
  }

  /**
   * Updates the current usage of a given organization quota in the cache. If cache is not present, this method is a no-op.
   *
   * If the corresponding quota type has pending usage in the cache and the delta is positive, the pending usage is decremented accordingly.
   *
   * @param organizationId
   * @param quotaType
   * @param delta
   */
  private async updateCurrentQuotaUsage(
    organizationId: string,
    quotaType: OrganizationUsageQuotaType,
    delta: number,
  ): Promise<void> {
    const script = `
      local cacheKey = KEYS[1]
      local pendingCacheKey = KEYS[2]
      local delta = tonumber(ARGV[1])
      local ttl = tonumber(ARGV[2])

      if redis.call("EXISTS", cacheKey) == 1 then
        redis.call("INCRBY", cacheKey, delta)
        redis.call("EXPIRE", cacheKey, ttl)
      end
      
      local pending = tonumber(redis.call("GET", pendingCacheKey))
      if pending and pending > 0 and delta > 0 then
        redis.call("DECRBY", pendingCacheKey, delta)
      end
    `

    await this.redis.eval(
      script,
      2,
      this.getCurrentQuotaUsageCacheKey(organizationId, quotaType),
      this.getPendingQuotaUsageCacheKey(organizationId, quotaType),
      delta.toString(),
      this.CACHE_TTL_SECONDS.toString(),
    )
  }

  /**
   * Increments the pending usage for sandbox-related organization quotas.
   *
   * Pending usage is used to protect against race conditions to prevent quota abuse.
   *
   * If a user action will result in increased quota usage, we will first increment the pending usage.
   *
   * When the user action is complete, this pending usage will be transfered to the actual usage.
   *
   * As a safeguard, an expiration time is set on the pending usage cache to prevent lockout for new operations.
   *
   * @param organizationId
   * @param cpu - The amount of CPU to increment.
   * @param memory - The amount of memory to increment.
   * @param disk - The amount of disk to increment.
   * @param excludeSandboxId - If provided, pending usage will be incremented only for quotas that are not consumed by the sandbox in its current state.
   * @returns an object with the boolean values indicating if the pending usage was incremented for each quota type
   */
  async incrementPendingSandboxUsage(
    organizationId: string,
    cpu: number,
    memory: number,
    disk: number,
    excludeSandboxId?: string,
  ): Promise<{
    cpuIncremented: boolean
    memoryIncremented: boolean
    diskIncremented: boolean
  }> {
    // determine for which quota types we should increment the pending usage
    let shouldIncrementCpu = true
    let shouldIncrementMemory = true
    let shouldIncrementDisk = true

    if (excludeSandboxId) {
      const excludedSandbox = await this.sandboxRepository.findOne({
        where: { id: excludeSandboxId },
      })

      if (excludedSandbox) {
        if (SANDBOX_STATES_CONSUMING_COMPUTE.includes(excludedSandbox.state)) {
          shouldIncrementCpu = false
          shouldIncrementMemory = false
        }

        if (SANDBOX_STATES_CONSUMING_DISK.includes(excludedSandbox.state)) {
          shouldIncrementDisk = false
        }
      }
    }

    // increment the pending usage for necessary quota types
    const script = `
      local cpuKey = KEYS[1]
      local memoryKey = KEYS[2]
      local diskKey = KEYS[3]

      local shouldIncrementCpu = ARGV[1] == "true"
      local shouldIncrementMemory = ARGV[2] == "true"
      local shouldIncrementDisk = ARGV[3] == "true"

      local cpuIncrement = tonumber(ARGV[4])
      local memoryIncrement = tonumber(ARGV[5])
      local diskIncrement = tonumber(ARGV[6])

      local ttl = tonumber(ARGV[7])
    
      if shouldIncrementCpu then
        redis.call("INCRBY", cpuKey, cpuIncrement)
        redis.call("EXPIRE", cpuKey, ttl)
      end

      if shouldIncrementMemory then
        redis.call("INCRBY", memoryKey, memoryIncrement)
        redis.call("EXPIRE", memoryKey, ttl)
      end

      if shouldIncrementDisk then
        redis.call("INCRBY", diskKey, diskIncrement)
        redis.call("EXPIRE", diskKey, ttl)
      end
    `

    await this.redis.eval(
      script,
      3,
      this.getPendingQuotaUsageCacheKey(organizationId, 'cpu'),
      this.getPendingQuotaUsageCacheKey(organizationId, 'memory'),
      this.getPendingQuotaUsageCacheKey(organizationId, 'disk'),
      shouldIncrementCpu.toString(),
      shouldIncrementMemory.toString(),
      shouldIncrementDisk.toString(),
      cpu.toString(),
      memory.toString(),
      disk.toString(),
      this.CACHE_TTL_SECONDS.toString(),
    )

    return {
      cpuIncremented: shouldIncrementCpu,
      memoryIncremented: shouldIncrementMemory,
      diskIncremented: shouldIncrementDisk,
    }
  }

  /**
   * Decrements the pending usage for sandbox-related organization quotas.
   *
   * Use this method to roll back pending usage after incrementing it for an action that was subsequently rejected.
   *
   * Pending usage is used to protect against race conditions to prevent quota abuse.
   *
   * If a user action will result in increased quota usage, we will first increment the pending usage.
   *
   * When the user action is complete, this pending usage will be transfered to the actual usage.
   *
   * @param organizationId
   * @param cpu - If provided, the amount of CPU to decrement.
   * @param memory - If provided, the amount of memory to decrement.
   * @param disk - If provided, the amount of disk to decrement.
   */
  async decrementPendingSandboxUsage(
    organizationId: string,
    cpu?: number,
    memory?: number,
    disk?: number,
  ): Promise<void> {
    // decrement the pending usage for necessary quota types
    const script = `
      local cpuKey = KEYS[1]
      local memoryKey = KEYS[2] 
      local diskKey = KEYS[3]

      local cpuDecrement = tonumber(ARGV[1])
      local memoryDecrement = tonumber(ARGV[2])
      local diskDecrement = tonumber(ARGV[3])
      
      if cpuDecrement then
        redis.call("DECRBY", cpuKey, cpuDecrement)
      end

      if memoryDecrement then
        redis.call("DECRBY", memoryKey, memoryDecrement)
      end

      if diskDecrement then
        redis.call("DECRBY", diskKey, diskDecrement)
      end
    `

    await this.redis.eval(
      script,
      3,
      this.getPendingQuotaUsageCacheKey(organizationId, 'cpu'),
      this.getPendingQuotaUsageCacheKey(organizationId, 'memory'),
      this.getPendingQuotaUsageCacheKey(organizationId, 'disk'),
      cpu?.toString() ?? '0',
      memory?.toString() ?? '0',
      disk?.toString() ?? '0',
    )
  }

  /**
   * Increments the pending usage for snapshot-related organization quotas.
   *
   * Pending usage is used to protect against race conditions to prevent quota abuse.
   *
   * If a user action will result in increased quota usage, we will first increment the pending usage.
   *
   * When the user action is complete, this pending usage will be transfered to the actual usage.
   *
   * As a safeguard, an expiration time is set on the pending usage cache to prevent lockout for new operations.
   *
   * @param organizationId
   * @param snapshotCount - The count of snapshots to increment.
   */
  async incrementPendingSnapshotUsage(organizationId: string, snapshotCount: number): Promise<void> {
    const script = `
      local snapshotCountKey = KEYS[1]

      local snapshotCountIncrement = tonumber(ARGV[1])
      local ttl = tonumber(ARGV[2])
    
      redis.call("INCRBY", snapshotCountKey, snapshotCountIncrement)
      redis.call("EXPIRE", snapshotCountKey, ttl)
    `

    await this.redis.eval(
      script,
      1,
      this.getPendingQuotaUsageCacheKey(organizationId, 'snapshot_count'),
      snapshotCount.toString(),
      this.CACHE_TTL_SECONDS.toString(),
    )
  }

  /**
   * Decrements the pending usage for snapshot-related organization quotas.
   *
   * Use this method to roll back pending usage after incrementing it for an action that was subsequently rejected.
   *
   * Pending usage is used to protect against race conditions to prevent quota abuse.
   *
   * If a user action will result in increased quota usage, we will first increment the pending usage.
   *
   * When the user action is complete, this pending usage will be transfered to the actual usage.
   *
   * @param organizationId
   * @param snapshotCount - If provided, the count of snapshots to decrement.
   */
  async decrementPendingSnapshotUsage(organizationId: string, snapshotCount?: number): Promise<void> {
    // decrement the pending usage for necessary quota types
    const script = `
      local snapshotCountKey = KEYS[1]

      local snapshotCountDecrement = tonumber(ARGV[1])
      
      if snapshotCountDecrement then
        redis.call("DECRBY", snapshotCountKey, snapshotCountDecrement)
      end
    `

    await this.redis.eval(
      script,
      1,
      this.getPendingQuotaUsageCacheKey(organizationId, 'snapshot_count'),
      snapshotCount?.toString() ?? '0',
    )
  }

  /**
   * Increments the pending usage for volume-related organization quotas.
   *
   * Pending usage is used to protect against race conditions to prevent quota abuse.
   *
   * If a user action will result in increased quota usage, we will first increment the pending usage.
   *
   * When the user action is complete, this pending usage will be transfered to the actual usage.
   *
   * As a safeguard, an expiration time is set on the pending usage cache to prevent lockout for new operations.
   *
   * @param organizationId
   * @param volumeCount - The count of volumes to increment.
   */
  async incrementPendingVolumeUsage(organizationId: string, volumeCount: number): Promise<void> {
    const script = `
      local volumeCountKey = KEYS[1]

      local volumeCountIncrement = tonumber(ARGV[1])
      local ttl = tonumber(ARGV[2])
    
      redis.call("INCRBY", volumeCountKey, volumeCountIncrement)
      redis.call("EXPIRE", volumeCountKey, ttl)
    `

    await this.redis.eval(
      script,
      1,
      this.getPendingQuotaUsageCacheKey(organizationId, 'volume_count'),
      volumeCount.toString(),
      this.CACHE_TTL_SECONDS.toString(),
    )
  }

  /**
   * Decrements the pending usage for volume-related organization quotas.
   *
   * Use this method to roll back pending usage after incrementing it for an action that was subsequently rejected.
   *
   * Pending usage is used to protect against race conditions to prevent quota abuse.
   *
   * If a user action will result in increased quota usage, we will first increment the pending usage.
   *
   * When the user action is complete, this pending usage will be transfered to the actual usage.
   *
   * @param organizationId
   * @param volumeCount - If provided, the count of volumes to decrement.
   */
  async decrementPendingVolumeUsage(organizationId: string, volumeCount?: number): Promise<void> {
    // decrement the pending usage for necessary quota types
    const script = `
      local volumeCountKey = KEYS[1]

      local volumeCountDecrement = tonumber(ARGV[1])
      
      if volumeCountDecrement then
        redis.call("DECRBY", volumeCountKey, volumeCountDecrement)
      end
    `

    await this.redis.eval(
      script,
      1,
      this.getPendingQuotaUsageCacheKey(organizationId, 'volume_count'),
      volumeCount?.toString() ?? '0',
    )
  }

  /**
   * Get the cache key for the timestamp of the last time the cached usage of organization quotas for a given resource type was populated from the database.
   *
   * @param organizationId
   * @param resourceType
   */
  private getCacheStalenessKey(organizationId: string, resourceType: OrganizationUsageResourceType): string {
    return `org:${organizationId}:resource:${resourceType}:usage:fetched_at`
  }

  /**
   * Reset the timestamp of the last time the cached usage of organization quotas for a given resource type was populated from the database.
   *
   * @param organizationId
   * @param resourceType
   */
  private async resetCacheStaleness(
    organizationId: string,
    resourceType: OrganizationUsageResourceType,
  ): Promise<void> {
    const cacheKey = this.getCacheStalenessKey(organizationId, resourceType)
    await this.redis.set(cacheKey, Date.now())
  }

  /**
   * Check if the cached usage of organization quotas for a given resource type was last populated from the database more than CACHE_MAX_AGE_MS ago.
   *
   * @param organizationId
   * @param resourceType
   * @returns `true` if the cached usage is stale, `false` otherwise
   */
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

  @OnEvent(SandboxEvents.CREATED)
  async handleSandboxCreated(event: SandboxCreatedEvent) {
    const lockKey = `sandbox:${event.sandbox.id}:quota-usage-update`
    await this.redisLockProvider.waitForLock(lockKey, 60)

    try {
      await this.updateCurrentQuotaUsage(event.sandbox.organizationId, 'cpu', event.sandbox.cpu)
      await this.updateCurrentQuotaUsage(event.sandbox.organizationId, 'memory', event.sandbox.mem)
      await this.updateCurrentQuotaUsage(event.sandbox.organizationId, 'disk', event.sandbox.disk)
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
        SANDBOX_STATES_CONSUMING_COMPUTE,
      )

      const memDelta = this.calculateQuotaUsageDelta(
        event.sandbox.mem,
        event.oldState,
        event.newState,
        SANDBOX_STATES_CONSUMING_COMPUTE,
      )

      const diskDelta = this.calculateQuotaUsageDelta(
        event.sandbox.disk,
        event.oldState,
        event.newState,
        SANDBOX_STATES_CONSUMING_DISK,
      )

      if (cpuDelta !== 0) {
        await this.updateCurrentQuotaUsage(event.sandbox.organizationId, 'cpu', cpuDelta)
      }

      if (memDelta !== 0) {
        await this.updateCurrentQuotaUsage(event.sandbox.organizationId, 'memory', memDelta)
      }

      if (diskDelta !== 0) {
        await this.updateCurrentQuotaUsage(event.sandbox.organizationId, 'disk', diskDelta)
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
      await this.updateCurrentQuotaUsage(event.snapshot.organizationId, 'snapshot_count', 1)
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
        await this.updateCurrentQuotaUsage(event.snapshot.organizationId, 'snapshot_count', countDelta)
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
      await this.updateCurrentQuotaUsage(event.volume.organizationId, 'volume_count', 1)
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
        await this.updateCurrentQuotaUsage(event.volume.organizationId, 'volume_count', countDelta)
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
    statesConsumingResource: TState[],
  ): number {
    const wasConsumingResource = statesConsumingResource.includes(oldState)
    const isConsumingResource = statesConsumingResource.includes(newState)

    if (!wasConsumingResource && isConsumingResource) {
      return resourceAmount
    }

    if (wasConsumingResource && !isConsumingResource) {
      return -resourceAmount
    }

    return 0
  }
}
