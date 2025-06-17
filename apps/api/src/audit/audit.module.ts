/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { TypeOrmModule } from '@nestjs/typeorm'
import { AuditLog } from './entities/audit-log.entity'
import { AuditService } from './services/audit.service'
import { AuditInterceptor } from './interceptors/audit.interceptor'
import { AuditLogSubscriber } from './subscribers/audit-log.subscriber'
import { AuditController } from './controllers/audit.controller'

@Module({
  imports: [TypeOrmModule.forFeature([AuditLog])],
  controllers: [AuditController],
  providers: [AuditService, AuditInterceptor, AuditLogSubscriber],
  exports: [AuditService, AuditInterceptor],
})
export class AuditModule {}
