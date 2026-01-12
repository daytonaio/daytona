/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Controller,
  Post,
  Get,
  Body,
  Query,
  UseGuards,
  BadRequestException,
  HttpCode,
  HttpStatus,
} from '@nestjs/common'
import { ApiTags, ApiOperation, ApiResponse, ApiOAuth2, ApiBearerAuth, ApiHeader, ApiQuery } from '@nestjs/swagger'
import { DeviceAuthService } from './device-auth.service'
import { DeviceCodeRequestDto } from './dto/device-code-request.dto'
import { DeviceCodeResponseDto } from './dto/device-code-response.dto'
import { DeviceTokenRequestDto } from './dto/device-token-request.dto'
import { DeviceTokenResponseDto, DeviceTokenErrorDto } from './dto/device-token-response.dto'
import { DeviceApproveRequestDto } from './dto/device-approve-request.dto'
import { DeviceStatusResponseDto } from './dto/device-status-response.dto'
import { CombinedAuthGuard } from '../auth/combined-auth.guard'
import { AuthContext } from '../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../common/interfaces/auth-context.interface'
import { CustomHeaders } from '../common/constants/header.constants'
import { TypedConfigService } from '../config/typed-config.service'

@ApiTags('device-auth')
@Controller('auth/device')
export class DeviceAuthController {
  constructor(
    private readonly deviceAuthService: DeviceAuthService,
    private readonly configService: TypedConfigService,
  ) {}

  @Post('code')
  @ApiOperation({
    summary: 'Request device authorization code',
    description: 'Initiates a device authorization flow by generating a device code and user code',
    operationId: 'requestDeviceCode',
  })
  @ApiResponse({
    status: 200,
    description: 'Device authorization codes generated successfully',
    type: DeviceCodeResponseDto,
  })
  @HttpCode(HttpStatus.OK)
  async requestDeviceCode(@Body() request: DeviceCodeRequestDto): Promise<DeviceCodeResponseDto> {
    const result = await this.deviceAuthService.createDeviceAuthorizationRequest(request.client_id, request.scope)

    const dashboardUrl = this.configService.get('dashboardUrl') || 'https://app.daytona.io'
    // Strip /dashboard suffix if present since /device is a top-level route
    const baseUrl = dashboardUrl.replace(/\/dashboard\/?$/, '')

    return {
      device_code: result.deviceCode,
      user_code: result.userCode,
      verification_uri: `${baseUrl}/device`,
      verification_uri_complete: `${baseUrl}/device?user_code=${result.userCode}`,
      expires_in: result.expiresIn,
      interval: result.interval,
    }
  }

  @Post('token')
  @ApiOperation({
    summary: 'Poll for device token',
    description: 'Polls for the token after user has authenticated via browser',
    operationId: 'pollDeviceToken',
  })
  @ApiResponse({
    status: 200,
    description: 'Token response (success or pending/error status)',
    type: DeviceTokenResponseDto,
  })
  @ApiResponse({
    status: 400,
    description: 'Error response',
    type: DeviceTokenErrorDto,
  })
  @HttpCode(HttpStatus.OK)
  async pollDeviceToken(@Body() request: DeviceTokenRequestDto): Promise<DeviceTokenResponseDto | DeviceTokenErrorDto> {
    if (request.grant_type !== 'urn:ietf:params:oauth:grant-type:device_code') {
      throw new BadRequestException('Invalid grant type')
    }

    const result = await this.deviceAuthService.pollForToken(request.device_code, request.client_id)

    switch (result.status) {
      case 'approved':
        return {
          access_token: result.accessToken!,
          token_type: 'Bearer',
          expires_in: 31536000, // 1 year
          scope: result.scope || '',
          organization_id: result.organizationId!,
          organization_name: result.organizationName!,
        }
      case 'pending':
        return { error: 'authorization_pending' }
      case 'denied':
        return { error: 'access_denied', error_description: 'The user denied the authorization request' }
      case 'expired':
        return { error: 'expired_token', error_description: 'The device code has expired' }
      case 'slow_down':
        return { error: 'slow_down', error_description: 'Please slow down the polling rate' }
      default:
        return { error: 'authorization_pending' }
    }
  }

  @Get('status')
  @ApiOperation({
    summary: 'Get device authorization status',
    description: 'Gets the current status of a device authorization request by user code',
    operationId: 'getDeviceStatus',
  })
  @ApiQuery({ name: 'user_code', required: true, description: 'The user code to check status for' })
  @ApiResponse({
    status: 200,
    description: 'Device authorization status',
    type: DeviceStatusResponseDto,
  })
  async getDeviceStatus(@Query('user_code') userCode: string): Promise<DeviceStatusResponseDto> {
    const request = await this.deviceAuthService.getDeviceAuthorizationByUserCode(userCode)

    if (!request) {
      throw new BadRequestException('Invalid or expired user code')
    }

    const now = new Date()
    const expiresIn = Math.max(0, Math.floor((request.expiresAt.getTime() - now.getTime()) / 1000))

    return {
      user_code: request.userCode,
      client_id: request.clientId,
      scope: request.scope || '',
      status: request.status,
      expires_in: expiresIn,
    }
  }

  @Post('approve')
  @UseGuards(CombinedAuthGuard)
  @ApiHeader(CustomHeaders.ORGANIZATION_ID)
  @ApiOAuth2(['openid', 'profile', 'email'])
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Approve or deny device authorization',
    description: 'Approves or denies a device authorization request. Requires authentication.',
    operationId: 'approveDeviceAuthorization',
  })
  @ApiResponse({
    status: 200,
    description: 'Device authorization processed successfully',
  })
  @HttpCode(HttpStatus.OK)
  async approveDeviceAuthorization(
    @AuthContext() authContext: OrganizationAuthContext,
    @Body() request: DeviceApproveRequestDto,
  ): Promise<{ success: boolean; message: string }> {
    const organizationId = request.organization_id || authContext.organizationId

    if (request.action === 'approve') {
      return this.deviceAuthService.approveDeviceAuthorization(request.user_code, authContext.userId, organizationId)
    } else {
      return this.deviceAuthService.denyDeviceAuthorization(request.user_code)
    }
  }
}
