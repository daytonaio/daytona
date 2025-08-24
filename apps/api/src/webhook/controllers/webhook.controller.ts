/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Controller,
  Post,
  Get,
  //  Delete,
  Body,
  Param,
  UseGuards,
  HttpStatus,
  //  HttpCode,
  NotFoundException,
} from '@nestjs/common'
import { ApiTags, ApiOperation, ApiResponse, ApiBearerAuth } from '@nestjs/swagger'
import { WebhookService } from '../services/webhook.service'
import { SendWebhookDto } from '../dto/send-webhook.dto'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { OrganizationService } from '../../organization/services/organization.service'

@ApiTags('webhooks')
@Controller('webhooks')
@UseGuards(CombinedAuthGuard)
@ApiBearerAuth()
export class WebhookController {
  constructor(
    private readonly webhookService: WebhookService,
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
  @ApiResponse({
    status: HttpStatus.FORBIDDEN,
    description: 'User does not have access to this organization',
  })
  @ApiResponse({
    status: HttpStatus.NOT_FOUND,
    description: 'Organization not found',
  })
  async getAppPortalAccess(@Param('organizationId') organizationId: string): Promise<{ url: string }> {
    // Check if user has access to this organization
    const organization = await this.organizationService.findOne(organizationId)
    if (!organization) {
      throw new NotFoundException('Organization not found')
    }

    // TODO: Add proper authorization check here
    // For now, we'll assume the user has access if they can see the organization

    const url = await this.webhookService.getAppPortalAccessUrl(organizationId)
    return { url }
  }

  @Post('organizations/:organizationId/send')
  @ApiOperation({ summary: 'Send a webhook message to an organization' })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'Webhook message sent successfully',
  })
  @ApiResponse({
    status: HttpStatus.FORBIDDEN,
    description: 'User does not have access to this organization',
  })
  @ApiResponse({
    status: HttpStatus.NOT_FOUND,
    description: 'Organization not found',
  })
  async sendWebhook(
    @Param('organizationId') organizationId: string,
    @Body() sendWebhookDto: SendWebhookDto,
  ): Promise<void> {
    // Check if user has access to this organization
    const organization = await this.organizationService.findOne(organizationId)
    if (!organization) {
      throw new NotFoundException('Organization not found')
    }

    // TODO: Add proper authorization check here
    // For now, we'll assume the user has access if they can see the organization

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
  @ApiResponse({
    status: HttpStatus.FORBIDDEN,
    description: 'User does not have access to this organization',
  })
  @ApiResponse({
    status: HttpStatus.NOT_FOUND,
    description: 'Organization not found',
  })
  async getMessageAttempts(
    @Param('organizationId') organizationId: string,
    @Param('messageId') messageId: string,
  ): Promise<any[]> {
    // Check if user has access to this organization
    const organization = await this.organizationService.findOne(organizationId)
    if (!organization) {
      throw new NotFoundException('Organization not found')
    }

    // TODO: Add proper authorization check here
    // For now, we'll assume the user has access if they can see the organization

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
  async getStatus(): Promise<{ enabled: boolean }> {
    return {
      enabled: this.webhookService.isEnabled(),
    }
  }
}
