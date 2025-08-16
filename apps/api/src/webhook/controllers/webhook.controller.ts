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
import { WebhookEndpointInitializerService } from '../services/webhook-endpoint-initializer.service'
import { WebhookInitializationManagerService } from '../services/webhook-initialization-manager.service'
// import { CreateWebhookEndpointDto } from '../dto/create-webhook-endpoint.dto'
import { WebhookEndpointDto } from '../dto/webhook-endpoint.dto'
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
    private readonly webhookEndpointInitializerService: WebhookEndpointInitializerService,
    private readonly webhookInitializationManagerService: WebhookInitializationManagerService,
    private readonly organizationService: OrganizationService,
  ) {}

  // @Post('organizations/:organizationId/endpoints')
  // @ApiOperation({ summary: 'Create a new webhook endpoint for an organization' })
  // @ApiResponse({
  //   status: HttpStatus.CREATED,
  //   description: 'Webhook endpoint created successfully',
  //   type: WebhookEndpointDto,
  // })
  // @ApiResponse({
  //   status: HttpStatus.FORBIDDEN,
  //   description: 'User does not have access to this organization',
  // })
  // @ApiResponse({
  //   status: HttpStatus.NOT_FOUND,
  //   description: 'Organization not found',
  // })
  // async createEndpoint(
  //   @Param('organizationId') organizationId: string,
  //   @Body() createWebhookEndpointDto: CreateWebhookEndpointDto,
  // ): Promise<WebhookEndpointDto> {
  //   // Check if user has access to this organization
  //   const organization = await this.organizationService.findOne(organizationId)
  //   if (!organization) {
  //     throw new NotFoundException('Organization not found')
  //   }

  //   // TODO: Add proper authorization check here
  //   // For now, we'll assume the user has access if they can see the organization

  //   const endpoint = await this.webhookService.createEndpoint(
  //     organizationId,
  //     createWebhookEndpointDto.url,
  //     createWebhookEndpointDto.description,
  //   )

  //   return endpoint
  // }

  @Get('organizations/:organizationId/endpoints')
  @ApiOperation({ summary: 'List all webhook endpoints for an organization' })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'List of webhook endpoints',
    type: [WebhookEndpointDto],
  })
  @ApiResponse({
    status: HttpStatus.FORBIDDEN,
    description: 'User does not have access to this organization',
  })
  @ApiResponse({
    status: HttpStatus.NOT_FOUND,
    description: 'Organization not found',
  })
  async listEndpoints(@Param('organizationId') organizationId: string): Promise<WebhookEndpointDto[]> {
    // Check if user has access to this organization
    const organization = await this.organizationService.findOne(organizationId)
    if (!organization) {
      throw new NotFoundException('Organization not found')
    }

    // TODO: Add proper authorization check here
    // For now, we'll assume the user has access if they can see the organization

    return this.webhookService.listEndpoints(organizationId)
  }

  // @Delete('organizations/:organizationId/endpoints/:endpointId')
  // @HttpCode(HttpStatus.NO_CONTENT)
  // @ApiOperation({ summary: 'Delete a webhook endpoint' })
  // @ApiResponse({
  //   status: HttpStatus.NO_CONTENT,
  //   description: 'Webhook endpoint deleted successfully',
  // })
  // @ApiResponse({
  //   status: HttpStatus.FORBIDDEN,
  //   description: 'User does not have access to this organization',
  // })
  // @ApiResponse({
  //   status: HttpStatus.NOT_FOUND,
  //   description: 'Organization or endpoint not found',
  // })
  // async deleteEndpoint(
  //   @Param('organizationId') organizationId: string,
  //   @Param('endpointId') endpointId: string,
  // ): Promise<void> {
  //   // Check if user has access to this organization
  //   const organization = await this.organizationService.findOne(organizationId)
  //   if (!organization) {
  //     throw new NotFoundException('Organization not found')
  //   }

  //   // TODO: Add proper authorization check here
  //   // For now, we'll assume the user has access if they can see the organization

  //   await this.webhookService.deleteEndpoint(organizationId, endpointId)
  // }

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

  @Post('organizations/:organizationId/preconfigured-endpoints')
  @ApiOperation({ summary: 'Create preconfigured webhook endpoints for an organization' })
  @ApiResponse({
    status: HttpStatus.CREATED,
    description: 'Preconfigured webhook endpoints created successfully',
  })
  @ApiResponse({
    status: HttpStatus.FORBIDDEN,
    description: 'User does not have access to this organization',
  })
  @ApiResponse({
    status: HttpStatus.NOT_FOUND,
    description: 'Organization not found',
  })
  async createPreconfiguredEndpoints(@Param('organizationId') organizationId: string): Promise<void> {
    // Check if user has access to this organization
    const organization = await this.organizationService.findOne(organizationId)
    if (!organization) {
      throw new NotFoundException('Organization not found')
    }

    // TODO: Add proper authorization check here
    // For now, we'll assume the user has access if they can see the organization

    await this.webhookEndpointInitializerService.createPreconfiguredEndpointsForOrganization(organizationId)
  }

  @Get('preconfigured-endpoints')
  @ApiOperation({ summary: 'Get list of preconfigured webhook endpoint configurations' })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'List of preconfigured endpoint configurations',
    schema: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          url: { type: 'string' },
          description: { type: 'string' },
          eventTypes: {
            type: 'array',
            items: { type: 'string' },
          },
        },
      },
    },
  })
  async getPreconfiguredEndpointConfigs(): Promise<
    Array<{
      url: string
      description: string
      eventTypes: string[]
    }>
  > {
    return this.webhookEndpointInitializerService.getPreconfiguredEndpointConfigs()
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

    return this.webhookInitializationManagerService.getInitializationStatus(organizationId)
  }

  @Post('organizations/:organizationId/initialize')
  @ApiOperation({ summary: 'Initialize webhooks for an organization' })
  @ApiResponse({
    status: HttpStatus.CREATED,
    description: 'Webhooks initialized successfully',
  })
  @ApiResponse({
    status: HttpStatus.FORBIDDEN,
    description: 'User does not have access to this organization',
  })
  @ApiResponse({
    status: HttpStatus.NOT_FOUND,
    description: 'Organization not found',
  })
  async initializeWebhooks(@Param('organizationId') organizationId: string): Promise<void> {
    // Check if user has access to this organization
    const organization = await this.organizationService.findOne(organizationId)
    if (!organization) {
      throw new NotFoundException('Organization not found')
    }

    // TODO: Add proper authorization check here
    // For now, we'll assume the user has access if they can see the organization

    await this.webhookInitializationManagerService.initializeWebhooks(organizationId)
  }

  @Get('initialization-stats')
  @ApiOperation({ summary: 'Get webhook initialization statistics' })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'Webhook initialization statistics',
    schema: {
      type: 'object',
      properties: {
        totalOrganizations: { type: 'number' },
        initializedOrganizations: { type: 'number' },
        failedInitializations: { type: 'number' },
        pendingInitializations: { type: 'number' },
      },
    },
  })
  async getInitializationStats(): Promise<any> {
    return this.webhookInitializationManagerService.getInitializationStats()
  }

  @Post('retry-failed-initializations')
  @ApiOperation({ summary: 'Retry failed webhook initializations' })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'Retry process completed',
  })
  async retryFailedInitializations(): Promise<void> {
    await this.webhookInitializationManagerService.retryFailedInitializations()
  }

  @Post('update-all-endpoints')
  @ApiOperation({ summary: 'Update webhook endpoints for all organizations (for future updates)' })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'Update process completed',
  })
  async updateAllEndpoints(): Promise<void> {
    await this.webhookInitializationManagerService.updateWebhookEndpointsForAllOrganizations()
  }
}
