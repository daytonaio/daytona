/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { InjectDataSource } from '@nestjs/typeorm'
import { DataSource, IsNull, Raw } from 'typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { SandboxLastActivity } from '../entities/sandbox-last-activity.entity'
import { LogExecution } from '../../common/decorators/log-execution.decorator'
import { WithInstrumentation } from '../../common/decorators/otel.decorator'
import { TypedConfigService } from '../../config/typed-config.service'

const REDIS_ACTIVITY_KEY = 'sandbox:activity'

interface SandboxActivityUpdate {
  sandboxId: string
  lastActivityAt: Date
}

@Injectable()
export class SandboxActivityService {
  private readonly logger = new Logger(SandboxActivityService.name)

  constructor(
    @InjectRedis() private readonly redis: Redis,
    @InjectDataSource() private readonly dataSource: DataSource,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly configService: TypedConfigService,
  ) {}

  /**
   * Buffers a last activity timestamp in Redis (throttled to once per configured throttle TTL).
   *
   * Relies on the periodic flush to the database.
   */
  async updateLastActivityAt(sandboxId: string, lastActivityAt: Date): Promise<void> {
    const lockKey = `sandbox:update-last-activity:${sandboxId}`
    const acquired = await this.redisLockProvider.lock(
      lockKey,
      this.configService.getOrThrow('sandboxActivity.throttleTtlSeconds'),
    )
    if (!acquired) {
      return
    }
    await this.redis.zadd(REDIS_ACTIVITY_KEY, lastActivityAt.getTime(), sandboxId)
  }

  /**
   * Read the last activity timestamp for a sandbox.
   *
   * Checks Redis buffer first, falls back to the database.
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
   * Flush buffered activity timestamps from Redis to the database in bulk.
   * Processes entries in batches to avoid oversized transactions.
   *
   * Frequency must be < 1min to prevent unintended auto-lifecycle actions.
   */
  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'flush-activity-to-db' })
  @LogExecution('flush-activity-to-db')
  @WithInstrumentation()
  async flushActivityToDb(): Promise<void> {
    const lockKey = 'flush-activity-to-db-lock'
    const lockTtl = 30
    const acquired = await this.redisLockProvider.lock(lockKey, lockTtl)
    if (!acquired) {
      return
    }

    try {
      let totalFlushed = 0

      const batchSize = this.configService.getOrThrow('sandboxActivity.flushBatchSize')

      while (true) {
        const entries = await this.redis.zpopmin(REDIS_ACTIVITY_KEY, batchSize)

        if (entries.length === 0) {
          break
        }

        const updates: SandboxActivityUpdate[] = []
        for (let i = 0; i < entries.length; i += 2) {
          updates.push({
            sandboxId: entries[i],
            lastActivityAt: new Date(Number(entries[i + 1])),
          })
        }

        await this.bulkUpsertActivity(updates)
        totalFlushed += updates.length

        if (updates.length < batchSize) {
          break
        }
      }

      if (totalFlushed > 0) {
        this.logger.debug(`Flushed ${totalFlushed} activity timestamps to the database`)
      }
    } catch (error) {
      this.logger.error('Error flushing activity timestamps to the database:', error)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  /**
   * Builds a query to upsert activity timestamps into the database.
   *
   * Uses a conditional upsert that only updates when the incoming timestamp is newer, preventing updates to stale buffered values.
   */
  private buildUpsertQuery(values: SandboxActivityUpdate | SandboxActivityUpdate[]) {
    return this.dataSource
      .createQueryBuilder()
      .insert()
      .into(SandboxLastActivity)
      .values(values)
      .orUpdate(['lastActivityAt'], ['sandboxId'], {
        overwriteCondition: {
          where: [
            { lastActivityAt: IsNull() },
            { lastActivityAt: Raw((alias) => `${alias} < EXCLUDED."lastActivityAt"`) },
          ],
        },
      })
  }

  /**
   * Bulk upserts activity timestamps into the database.
   *
   * In case of FK violations, falls back to individual upserts to skip deleted sandbox(es).
   */
  private async bulkUpsertActivity(updates: SandboxActivityUpdate[]): Promise<void> {
    if (updates.length === 0) {
      this.logger.debug('No activity updates to flush')
      return
    }

    try {
      await this.buildUpsertQuery(updates).execute()
    } catch (bulkUpsertError) {
      if (bulkUpsertError.code === '23503') {
        this.logger.warn(
          'Bulk upsert for activity timestamps failed with FK violation, falling back to individual upserts',
        )
        for (const update of updates) {
          try {
            await this.buildUpsertQuery(update).execute()
          } catch (error) {
            if (error.code === '23503') {
              this.logger.warn(`Skipping activity flush for sandbox ${update.sandboxId} (deleted)`)
            } else {
              throw error
            }
          }
        }
      } else {
        throw bulkUpsertError
      }
    }
  }
}
