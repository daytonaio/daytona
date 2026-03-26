/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { ApiKeyController } from './api-key.controller'
import { ApiKeyService } from './api-key.service'
import { ApiKey } from './api-key.entity'
import { TypeOrmModule } from '@nestjs/typeorm'
import { OrganizationModule } from '../organization/organization.module'
import { RedisLockProvider } from '../sandbox/common/redis-lock.provider'

@Module({
  imports: [OrganizationModule, TypeOrmModule.forFeature([ApiKey])],
  controllers: [ApiKeyController],
  providers: [ApiKeyService, RedisLockProvider],
  exports: [ApiKeyService],
})
export class ApiKeyModule {}
