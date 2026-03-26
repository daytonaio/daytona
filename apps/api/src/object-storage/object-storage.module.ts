/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { ObjectStorageController } from './controllers/object-storage.controller'
import { ObjectStorageService } from './services/object-storage.service'
import { ConfigModule } from '@nestjs/config'
import { OrganizationModule } from '../organization/organization.module'

@Module({
  imports: [ConfigModule, OrganizationModule],
  controllers: [ObjectStorageController],
  providers: [ObjectStorageService],
  exports: [ObjectStorageService],
})
export class ObjectStorageModule {}
