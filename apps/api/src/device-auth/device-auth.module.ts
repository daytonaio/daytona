/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { TypeOrmModule } from '@nestjs/typeorm'
import { DeviceAuthorization } from './device-authorization.entity'
import { DeviceAuthService } from './device-auth.service'
import { DeviceAuthController } from './device-auth.controller'
import { ApiKeyModule } from '../api-key/api-key.module'
import { UserModule } from '../user/user.module'

@Module({
  imports: [TypeOrmModule.forFeature([DeviceAuthorization]), ApiKeyModule, UserModule],
  providers: [DeviceAuthService],
  controllers: [DeviceAuthController],
  exports: [DeviceAuthService],
})
export class DeviceAuthModule {}
