/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { TypeOrmModule } from '@nestjs/typeorm'
import { AuditController } from './controllers/audit.controller'
import { AuditLog } from './entities/audit-log.entity'
import { AuditInterceptor } from './interceptors/audit.interceptor'
import { AuditService } from './services/audit.service'
import { AuditLogSubscriber } from './subscribers/audit-log.subscriber'
import { RedisLockProvider } from '../sandbox/common/redis-lock.provider'

@Module({
  imports: [TypeOrmModule.forFeature([AuditLog])],
  controllers: [AuditController],
  providers: [AuditService, AuditInterceptor, AuditLogSubscriber, RedisLockProvider],
  exports: [AuditService, AuditInterceptor],
})
export class AuditModule {}
