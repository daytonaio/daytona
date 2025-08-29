/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Get, Logger, ServiceUnavailableException } from '@nestjs/common'
import { HealthCheckService, HealthCheck, TypeOrmHealthIndicator } from '@nestjs/terminus'
import { RedisHealthIndicator } from './redis.health'

@Controller('health')
export class HealthController {
  private readonly logger = new Logger(HealthController.name)

  constructor(
    private health: HealthCheckService,
    private db: TypeOrmHealthIndicator,
    private redis: RedisHealthIndicator,
  ) {}

  @Get()
  @HealthCheck()
  async check() {
    try {
      const result = await this.health.check([() => this.db.pingCheck('database'), () => this.redis.isHealthy('redis')])
      return { status: result.status }
    } catch (error) {
      this.logger.error(error)
      throw new ServiceUnavailableException()
    }
  }
}
