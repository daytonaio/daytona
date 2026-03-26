/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { TypeOrmModule } from '@nestjs/typeorm'
import { DataSource } from 'typeorm'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { SandboxUsagePeriod } from './entities/sandbox-usage-period.entity'
import { UsageService } from './services/usage.service'
import { RedisLockProvider } from '../sandbox/common/redis-lock.provider'
import { SandboxUsagePeriodArchive } from './entities/sandbox-usage-period-archive.entity'
import { SandboxRepository } from '../sandbox/repositories/sandbox.repository'
import { SandboxLookupCacheInvalidationService } from '../sandbox/services/sandbox-lookup-cache-invalidation.service'
import { Sandbox } from '../sandbox/entities/sandbox.entity'

@Module({
  imports: [TypeOrmModule.forFeature([SandboxUsagePeriod, Sandbox, SandboxUsagePeriodArchive])],
  providers: [
    UsageService,
    RedisLockProvider,
    SandboxLookupCacheInvalidationService,
    {
      provide: SandboxRepository,
      inject: [DataSource, EventEmitter2, SandboxLookupCacheInvalidationService],
      useFactory: (
        dataSource: DataSource,
        eventEmitter: EventEmitter2,
        sandboxLookupCacheInvalidationService: SandboxLookupCacheInvalidationService,
      ) => new SandboxRepository(dataSource, eventEmitter, sandboxLookupCacheInvalidationService),
    },
  ],
  exports: [UsageService],
})
export class UsageModule {}
