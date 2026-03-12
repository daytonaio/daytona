/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { NotificationService } from './services/notification.service'
import { NotificationGateway } from './gateways/notification.gateway'
import { NotificationRedisEmitter } from './emitters/notification-redis.emitter'
import { NotificationEmitter } from './gateways/notification-emitter.abstract'
import { OrganizationModule } from '../organization/organization.module'
import { SandboxModule } from '../sandbox/sandbox.module'
import { RedisModule } from '@nestjs-modules/ioredis'
import { AuthModule } from '../auth/auth.module'
import { RegionModule } from '../region/region.module'
import { isApiEnabled } from '../common/utils/app-mode'

const gatewayEnabled = isApiEnabled() && process.env.NOTIFICATION_GATEWAY_DISABLED !== 'true'

@Module({
  imports: [OrganizationModule, SandboxModule, RedisModule, AuthModule, RegionModule],
  providers: [
    NotificationService,
    ...(gatewayEnabled
      ? [NotificationGateway, { provide: NotificationEmitter, useExisting: NotificationGateway }]
      : [{ provide: NotificationEmitter, useClass: NotificationRedisEmitter }]),
  ],
  exports: [NotificationService],
})
export class NotificationModule {}
