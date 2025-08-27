/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { TypeOrmModule } from '@nestjs/typeorm'
import { WebhookService } from './services/webhook.service'
import { WebhookController } from './controllers/webhook.controller'
import { WebhookEventHandlerService } from './services/webhook-event-handler.service'
import { WebhookInitialization } from './entities/webhook-initialization.entity'
import { OrganizationModule } from '../organization/organization.module'
import { TypedConfigModule } from '../config/typed-config.module'
import { AuthModule } from '../auth/auth.module'

@Module({
  imports: [OrganizationModule, TypedConfigModule, TypeOrmModule.forFeature([WebhookInitialization]), AuthModule],
  controllers: [WebhookController],
  providers: [WebhookService, WebhookEventHandlerService],
  exports: [WebhookService],
})
export class WebhookModule {}
