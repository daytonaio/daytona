/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Get, Logger, ServiceUnavailableException, UseGuards } from '@nestjs/common'
import { HealthCheckService, HealthCheck, TypeOrmHealthIndicator } from '@nestjs/terminus'
import { RedisHealthIndicator } from './redis.health'
import { AnonymousRateLimitGuard } from '../common/guards/anonymous-rate-limit.guard'
import { AuthenticatedRateLimitGuard } from '../common/guards/authenticated-rate-limit.guard'
import { CombinedAuthGuard } from '../auth/combined-auth.guard'
import { HealthCheckGuard } from '../auth/health-check.guard'

@Controller('health')
export class HealthController {
  private readonly logger = new Logger(HealthController.name)

  constructor(
    private health: HealthCheckService,
    private db: TypeOrmHealthIndicator,
    private redis: RedisHealthIndicator,
  ) {}

  @Get()
  @UseGuards(AnonymousRateLimitGuard)
  live() {
    return { status: 'ok' }
  }

  @Get('ready')
  @UseGuards(CombinedAuthGuard, HealthCheckGuard, AuthenticatedRateLimitGuard)
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
