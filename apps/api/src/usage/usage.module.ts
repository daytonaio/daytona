/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { TypeOrmModule } from '@nestjs/typeorm'
import { WorkspaceUsagePeriod } from './entities/workspace-usage-period.entity'
import { UsageService } from './services/usage.service'
import { RedisLockProvider } from '../workspace/common/redis-lock.provider'

@Module({
  imports: [TypeOrmModule.forFeature([WorkspaceUsagePeriod])],
  providers: [UsageService, RedisLockProvider],
  exports: [UsageService],
})
export class UsageModule {}
