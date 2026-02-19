/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, OnModuleDestroy } from '@nestjs/common'
import { createClient, ClickHouseClient } from '@clickhouse/client'
import { TypedConfigService } from '../config/typed-config.service'

@Injectable()
export class ClickHouseService implements OnModuleDestroy {
  private readonly logger = new Logger(ClickHouseService.name)
  private client: ClickHouseClient | null = null

  constructor(private readonly configService: TypedConfigService) {}

  private getClient(): ClickHouseClient | null {
    if (this.client) {
      return this.client
    }

    const config = this.configService.getClickHouseConfig()
    if (!config) {
      return null
    }

    this.client = createClient({
      url: config.url,
      username: config.username,
      password: config.password,
      database: config.database,
    })

    return this.client
  }

  async onModuleDestroy() {
    if (this.client) {
      await this.client.close()
    }
  }

  isConfigured(): boolean {
    return this.configService.getClickHouseConfig() !== null
  }

  async query<T>(query: string, params?: Record<string, unknown>): Promise<T[]> {
    const client = this.getClient()
    if (!client) {
      this.logger.warn('ClickHouse is not configured')
      return []
    }

    try {
      const result = await client.query({
        query,
        query_params: params,
        format: 'JSONEachRow',
        clickhouse_settings: {
          date_time_input_format: 'best_effort',
        },
      })

      return (await result.json()) as T[]
    } catch (error) {
      this.logger.error('ClickHouse query failed:', error)
      throw error
    }
  }

  async queryOne<T>(query: string, params?: Record<string, unknown>): Promise<T | null> {
    const results = await this.query<T>(query, params)
    return results.length > 0 ? results[0] : null
  }
}
