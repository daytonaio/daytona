/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { InjectDataSource } from '@nestjs/typeorm'
import { DataSource } from 'typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { SandboxLastActivity } from '../entities/sandbox-last-activity.entity'
import { LogExecution } from '../../common/decorators/log-execution.decorator'
import { WithInstrumentation } from '../../common/decorators/otel.decorator'
import { OnEvent } from '@nestjs/event-emitter'
import { SandboxEvents } from '../constants/sandbox-events.constants'
import { SandboxCreatedEvent } from '../events/sandbox-create.event'
import { SandboxStateUpdatedEvent } from '../events/sandbox-state-updated.event'

const REDIS_ACTIVITY_KEY = 'sandbox:activity'
const ACTIVITY_THROTTLE_TTL = 45
const FLUSH_BATCH_SIZE = 1000

@Injectable()
export class SandboxActivityService {
  private readonly logger = new Logger(SandboxActivityService.name)

  constructor(
    @InjectRedis() private readonly redis: Redis,
    @InjectDataSource() private readonly dataSource: DataSource,
    private readonly redisLockProvider: RedisLockProvider,
  ) {}

  /**
   * Update last-activity for a sandbox.
   * By default, buffers in Redis (throttled to once per 45s) and relies on the periodic flush to PG.
   * When `immediate` is true, writes directly to PG as well, bypassing the throttle.
   * Use immediate for state transitions where stale PG data could cause premature autostop/autoarchive.
   */
  async updateLastActivityAt(sandboxId: string, lastActivityAt: Date, immediate = false): Promise<void> {
    if (immediate) {
      await this.dataSource.getRepository(SandboxLastActivity).upsert({ sandboxId, lastActivityAt }, ['sandboxId'])
    } else {
      const lockKey = `sandbox:update-last-activity:${sandboxId}`
      const acquired = await this.redisLockProvider.lock(lockKey, ACTIVITY_THROTTLE_TTL)
      if (!acquired) {
        return
      }
    }

    await this.redis.zadd(REDIS_ACTIVITY_KEY, lastActivityAt.getTime(), sandboxId)
  }

  /**
   * Read the last activity timestamp for a sandbox.
   * Checks Redis buffer first, falls back to PG.
   */
  async getLastActivityAt(sandboxId: string): Promise<Date | null> {
    const score = await this.redis.zscore(REDIS_ACTIVITY_KEY, sandboxId)
    if (score !== null) {
      return new Date(Number(score))
    }

    const row = await this.dataSource.getRepository(SandboxLastActivity).findOne({ where: { sandboxId } })

    return row?.lastActivityAt ?? null
  }

  /**
   * Insert a row into sandbox_last_activity for a newly created sandbox.
   * Called during sandbox creation to seed the initial activity timestamp.
   */
  async initializeActivity(sandboxId: string, timestamp: Date): Promise<void> {
    await this.dataSource
      .getRepository(SandboxLastActivity)
      .upsert({ sandboxId, lastActivityAt: timestamp }, ['sandboxId'])
    await this.redis.zadd(REDIS_ACTIVITY_KEY, timestamp.getTime(), sandboxId)
  }

  /**
   * Remove activity tracking for a destroyed sandbox.
   */
  async removeActivity(sandboxId: string): Promise<void> {
    await this.redis.zrem(REDIS_ACTIVITY_KEY, sandboxId)
    // PG row is cascade-deleted when sandbox is deleted
  }

  /**
   * Flush buffered activity timestamps from Redis to PG in bulk.
   * Runs every 60 seconds. Processes entries in batches to avoid oversized transactions.
   */
  @Cron(CronExpression.EVERY_MINUTE, { name: 'flush-activity-to-pg' })
  @LogExecution('flush-activity-to-pg')
  @WithInstrumentation()
  async flushActivityToPg(): Promise<void> {
    const lockKey = 'flush-activity-to-pg-lock'
    const acquired = await this.redisLockProvider.lock(lockKey, 55)
    if (!acquired) {
      return
    }

    try {
      let cursor = 0
      let totalFlushed = 0

      // Process in batches using ZPOPMIN to atomically read-and-remove
      // We use ZRANGEBYSCORE + ZREM instead of ZPOPMIN for compatibility
      while (true) {
        // Read a batch of entries
        const entries = await this.redis.zrangebyscore(
          REDIS_ACTIVITY_KEY,
          '-inf',
          '+inf',
          'WITHSCORES',
          'LIMIT',
          cursor,
          FLUSH_BATCH_SIZE,
        )

        if (entries.length === 0) {
          break
        }

        // Parse entries: [member, score, member, score, ...]
        const updates: { sandboxId: string; lastActivityAt: Date }[] = []
        for (let i = 0; i < entries.length; i += 2) {
          updates.push({
            sandboxId: entries[i],
            lastActivityAt: new Date(Number(entries[i + 1])),
          })
        }

        if (updates.length > 0) {
          await this.bulkUpsertActivity(updates)
          totalFlushed += updates.length
        }

        if (updates.length < FLUSH_BATCH_SIZE) {
          break
        }

        cursor += FLUSH_BATCH_SIZE
      }

      if (totalFlushed > 0) {
        this.logger.debug(`Flushed ${totalFlushed} activity timestamps to PG`)
      }
    } catch (error) {
      this.logger.error('Error flushing activity timestamps to PG:', error)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  /**
   * Bulk upsert activity timestamps into PG.
   * Uses TypeORM's upsert which generates ON CONFLICT internally.
   */
  private async bulkUpsertActivity(updates: { sandboxId: string; lastActivityAt: Date }[]): Promise<void> {
    if (updates.length === 0) return

    await this.dataSource.getRepository(SandboxLastActivity).upsert(updates, ['sandboxId'])
  }

  @OnEvent(SandboxEvents.CREATED)
  private async handleSandboxCreatedEvent(event: SandboxCreatedEvent): Promise<void> {
    this.initializeActivity(event.sandbox.id, event.sandbox.createdAt).catch((error) => {
      this.logger.error(`Failed to initialize activity for sandbox ${event.sandbox.id}: ${error}`)
    })
  }

  @OnEvent(SandboxEvents.STATE_UPDATED)
  private async handleSandboxStateUpdatedEvent(event: SandboxStateUpdatedEvent): Promise<void> {
    this.updateLastActivityAt(event.sandbox.id, new Date(), true).catch((error) => {
      this.logger.error(`Failed to update activity for sandbox ${event.sandbox.id}: ${error}`)
    })
  }
}
