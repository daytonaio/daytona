/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Post, Get, Body, Param, UseGuards, HttpStatus, NotFoundException } from '@nestjs/common'
import { ApiTags, ApiOperation, ApiResponse, ApiBearerAuth, ApiHeader } from '@nestjs/swagger'
import { WebhookService } from '../services/webhook.service'
import { SendWebhookDto } from '../dto/send-webhook.dto'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { CustomHeaders } from '../../common/constants/header.constants'
import { SystemActionGuard } from '../../auth/system-action.guard'
import { OrganizationAccessGuard } from '../../organization/guards/organization-access.guard'
import { RequiredSystemRole } from '../../common/decorators/required-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { OrganizationService } from '../../organization/services/organization.service'
import { WebhookInitializationCheckerService } from '../services/webhook-initialization-checker.service'

@ApiTags('webhooks')
@Controller('webhooks')
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@UseGuards(CombinedAuthGuard, SystemActionGuard, OrganizationAccessGuard)
@ApiBearerAuth()
export class WebhookController {
  constructor(
    private readonly webhookService: WebhookService,
    private readonly webhookInitializationChecker: WebhookInitializationCheckerService,
    private readonly organizationService: OrganizationService,
  ) {}

  @Post('organizations/:organizationId/app-portal-access')
  @ApiOperation({ summary: 'Get Svix Consumer App Portal access URL for an organization' })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'App Portal access URL generated successfully',
    schema: {
      type: 'object',
      properties: {
        url: { type: 'string', description: 'App Portal access URL' },
      },
    },
  })
  @Audit({
    action: AuditAction.GET_WEBHOOK_APP_PORTAL_ACCESS,
    targetType: AuditTarget.ORGANIZATION,
    targetIdFromRequest: (req) => req.params.organizationId,
  })
  async getAppPortalAccess(@Param('organizationId') organizationId: string): Promise<{ url: string }> {
    const url = await this.webhookService.getAppPortalAccessUrl(organizationId)
    return { url }
  }

  @Post('organizations/:organizationId/send')
  @ApiOperation({ summary: 'Send a webhook message to an organization' })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'Webhook message sent successfully',
  })
  @RequiredSystemRole(SystemRole.ADMIN)
  @Audit({
    action: AuditAction.SEND_WEBHOOK_MESSAGE,
    targetType: AuditTarget.ORGANIZATION,
    targetIdFromRequest: (req) => req.params.organizationId,
    requestMetadata: {
      body: (req: TypedRequest<SendWebhookDto>) => ({
        eventType: req.body?.eventType,
        payload: req.body?.payload,
        eventId: req.body?.eventId,
      }),
    },
  })
  async sendWebhook(
    @Param('organizationId') organizationId: string,
    @Body() sendWebhookDto: SendWebhookDto,
  ): Promise<void> {
    await this.webhookService.sendWebhook(
      organizationId,
      sendWebhookDto.eventType,
      sendWebhookDto.payload,
      sendWebhookDto.eventId,
    )
  }

  @Get('organizations/:organizationId/messages/:messageId/attempts')
  @ApiOperation({ summary: 'Get delivery attempts for a webhook message' })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'List of delivery attempts',
    type: [Object],
  })
  @RequiredSystemRole(SystemRole.ADMIN)
  async getMessageAttempts(
    @Param('organizationId') organizationId: string,
    @Param('messageId') messageId: string,
  ): Promise<any[]> {
    return this.webhookService.getMessageAttempts(organizationId, messageId)
  }

  @Get('status')
  @ApiOperation({ summary: 'Get webhook service status' })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'Webhook service status',
    schema: {
      type: 'object',
      properties: {
        enabled: { type: 'boolean' },
      },
    },
  })
  @RequiredSystemRole(SystemRole.ADMIN)
  async getStatus(): Promise<{ enabled: boolean }> {
    return {
      enabled: this.webhookService.isEnabled(),
    }
  }

  @Get('organizations/:organizationId/initialization-status')
  @ApiOperation({ summary: 'Get webhook initialization status for an organization' })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'Webhook initialization status',
    schema: {
      type: 'object',
      properties: {
        organizationId: { type: 'string' },
        endpointsCreated: { type: 'boolean' },
        svixApplicationCreated: { type: 'boolean' },
        lastError: { type: 'string', nullable: true },
        retryCount: { type: 'number' },
        createdAt: { type: 'string' },
        updatedAt: { type: 'string' },
      },
    },
  })
  @ApiResponse({
    status: HttpStatus.NOT_FOUND,
    description: 'Organization not found',
  })
  async getInitializationStatus(@Param('organizationId') organizationId: string): Promise<any> {
    // Check if user has access to this organization
    const organization = await this.organizationService.findOne(organizationId)
    if (!organization) {
      throw new NotFoundException('Organization not found')
    }

    // TODO: Add proper authorization check here
    // For now, we'll assume the user has access if they can see the organization

    return this.webhookInitializationChecker.getInitializationStatus(organizationId)
  }
}
