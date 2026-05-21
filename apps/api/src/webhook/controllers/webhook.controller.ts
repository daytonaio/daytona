/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Post, Get, Param, UseGuards, HttpCode, HttpStatus, NotFoundException } from '@nestjs/common'
import { ApiTags, ApiOperation, ApiResponse, ApiBearerAuth, ApiOAuth2, ApiHeader, ApiParam } from '@nestjs/swagger'
import { WebhookService } from '../services/webhook.service'
import { OrganizationAuthContextGuard } from '../../organization/guards/organization-auth-context.guard'
import { WebhookAppPortalAccessDto } from '../dto/webhook-app-portal-access.dto'
import { WebhookInitializationStatusDto } from '../dto/webhook-initialization-status.dto'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import { CustomHeaders } from '../../common/constants/header.constants'
import { IsOrganizationAuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { Audit } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'

@Controller('webhooks')
@ApiTags('webhooks')
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@AuthStrategy([AuthStrategyType.API_KEY, AuthStrategyType.JWT])
@UseGuards(AuthenticatedRateLimitGuard)
@UseGuards(OrganizationAuthContextGuard)
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

  @Post('organizations/:organizationId/initialize')
  @HttpCode(HttpStatus.CREATED)
  @ApiOperation({ summary: 'Initialize webhooks for an organization' })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiResponse({
    status: HttpStatus.CREATED,
    description: 'Webhooks initialized successfully',
    type: WebhookInitializationStatusDto,
  })
  @Audit({
    action: AuditAction.INITIALIZE_WEBHOOKS,
    targetType: AuditTarget.ORGANIZATION,
    targetIdFromRequest: (req) => req.params.organizationId,
  })
  async initializeWebhooks(
    @IsOrganizationAuthContext() authContext: OrganizationAuthContext,
  ): Promise<WebhookInitializationStatusDto> {
    await this.webhookService.createSvixApplication(authContext.organization)
    const status = await this.webhookService.getInitializationStatus(authContext.organization.id)
    if (!status) {
      throw new NotFoundException('Webhook initialization status not found')
    }
    return WebhookInitializationStatusDto.fromWebhookInitialization(status)
  }

  @Post('organizations/:organizationId/refresh-endpoints')
  @HttpCode(HttpStatus.NO_CONTENT)
  @ApiOperation({ summary: 'Refresh cached endpoint presence flag for an organization' })
  @ApiResponse({
    status: HttpStatus.NO_CONTENT,
    description: 'Endpoint flag refreshed',
  })
  @ApiResponse({
    status: HttpStatus.NOT_FOUND,
    description: 'Webhook initialization status not found',
  })
  async refreshEndpoints(@Param('organizationId') organizationId: string): Promise<void> {
    await this.webhookService.refreshEndpointFlagByOrg(organizationId)
  }
}
