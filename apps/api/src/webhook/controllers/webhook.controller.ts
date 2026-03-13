/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Post, Get, Param, UseGuards, HttpStatus, NotFoundException } from '@nestjs/common'
import { ApiTags, ApiOperation, ApiResponse, ApiBearerAuth, ApiHeader, ApiOAuth2 } from '@nestjs/swagger'
import { WebhookService } from '../services/webhook.service'
import { CustomHeaders } from '../../common/constants/header.constants'
import { OrganizationActionGuard } from '../../organization/guards/organization-action.guard'
import { WebhookAppPortalAccessDto } from '../dto/webhook-app-portal-access.dto'
import { WebhookInitializationStatusDto } from '../dto/webhook-initialization-status.dto'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'

@ApiTags('webhooks')
@Controller('webhooks')
@AuthStrategy([AuthStrategyType.API_KEY, AuthStrategyType.JWT])
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@UseGuards(OrganizationActionGuard, AuthenticatedRateLimitGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class WebhookController {
  constructor(private readonly webhookService: WebhookService) {}

  @Post('organizations/:organizationId/app-portal-access')
  @ApiOperation({ summary: 'Get Svix Consumer App Portal access for an organization' })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'App Portal access generated successfully',
    type: WebhookAppPortalAccessDto,
  })
  async getAppPortalAccess(@Param('organizationId') organizationId: string): Promise<WebhookAppPortalAccessDto> {
    return this.webhookService.getAppPortalAccess(organizationId)
  }

  @Get('organizations/:organizationId/initialization-status')
  @ApiOperation({ summary: 'Get webhook initialization status for an organization' })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'Webhook initialization status',
    type: WebhookInitializationStatusDto,
  })
  @ApiResponse({
    status: HttpStatus.NOT_FOUND,
    description: 'Webhook initialization status not found',
  })
  async getInitializationStatus(
    @Param('organizationId') organizationId: string,
  ): Promise<WebhookInitializationStatusDto> {
    const status = await this.webhookService.getInitializationStatus(organizationId)
    if (!status) {
      throw new NotFoundException('Webhook initialization status not found')
    }
    return WebhookInitializationStatusDto.fromWebhookInitialization(status)
  }
}
