/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { TypeOrmModule } from '@nestjs/typeorm'
import { DeviceAuthController } from './device-auth.controller'
import { DeviceAuthService } from './device-auth.service'
import { DeviceAuthorizationRequest } from './device-auth.entity'
import { ApiKeyModule } from '../api-key/api-key.module'
import { OrganizationModule } from '../organization/organization.module'

@Module({
  imports: [TypeOrmModule.forFeature([DeviceAuthorizationRequest]), ApiKeyModule, OrganizationModule],
  controllers: [DeviceAuthController],
  providers: [DeviceAuthService],
  exports: [DeviceAuthService],
})
export class DeviceAuthModule {}
