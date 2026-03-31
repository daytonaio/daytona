/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, Get, HttpCode, HttpStatus, NotFoundException, Param, Post, UseGuards } from '@nestjs/common'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'
import { ApiBearerAuth, ApiOAuth2, ApiOperation, ApiResponse, ApiTags } from '@nestjs/swagger'
import { RequiredSystemRole } from '../../user/decorators/required-system-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'
import { WebhookService } from '../../webhook/services/webhook.service'
import { OrganizationService } from '../../organization/services/organization.service'
import { SendWebhookDto } from '../../webhook/dto/send-webhook.dto'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'

@Controller('admin/webhooks')
@ApiTags('admin')
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
@AuthStrategy([AuthStrategyType.API_KEY, AuthStrategyType.JWT])
@RequiredSystemRole(SystemRole.ADMIN)
@UseGuards(AuthenticatedRateLimitGuard)
export class AdminWebhookController {
  constructor(
    private readonly webhookService: WebhookService,
    private readonly organizationService: OrganizationService,
  ) {}

  @Post('organizations/:organizationId/send')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Send a webhook message to an organization',
    operationId: 'adminSendWebhook',
  })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'Webhook message sent successfully',
  })
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
  @ApiOperation({
    summary: 'Get delivery attempts for a webhook message',
    operationId: 'adminGetMessageAttempts',
  })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'List of delivery attempts',
    type: [Object],
  })
  async getMessageAttempts(
    @Param('organizationId') organizationId: string,
    @Param('messageId') messageId: string,
  ): Promise<any[]> {
    return this.webhookService.getMessageAttempts(organizationId, messageId)
  }

  @Get('status')
  @ApiOperation({
    summary: 'Get webhook service status',
    operationId: 'adminGetWebhookStatus',
  })
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

  @Post('organizations/:organizationId/initialize')
  @ApiOperation({
    summary: 'Initialize webhooks for an organization',
    operationId: 'adminInitializeWebhooks',
  })
  @ApiResponse({
    status: HttpStatus.CREATED,
    description: 'Webhooks initialized successfully',
  })
  @ApiResponse({
    status: HttpStatus.NOT_FOUND,
    description: 'Organization not found',
  })
  @Audit({
    action: AuditAction.INITIALIZE_WEBHOOKS,
    targetType: AuditTarget.ORGANIZATION,
    targetIdFromRequest: (req) => req.params.organizationId,
  })
  async initializeWebhooks(@Param('organizationId') organizationId: string): Promise<void> {
    const organization = await this.organizationService.findOne(organizationId)
    if (!organization) {
      throw new NotFoundException('Organization not found')
    }

    await this.webhookService.createSvixApplication(organization)
  }
}
