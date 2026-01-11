/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, Post, HttpCode, HttpStatus, UseGuards, Get, Query, Request } from '@nestjs/common'
import { ApiTags, ApiOperation, ApiResponse } from '@nestjs/swagger'
import { DeviceAuthService } from './device-auth.service'
import { RequestDeviceCodeDto, PollDeviceTokenDto, ApproveDeviceDto } from './dto'
import { CombinedAuthGuard } from '../auth/combined-auth.guard'
import { ApiKeyService } from '../api-key/api-key.service'
import { UserService } from '../user/user.service'

@ApiTags('device-auth')
@Controller('device')
export class DeviceAuthController {
  constructor(
    private readonly deviceAuthService: DeviceAuthService,
    private readonly apiKeyService: ApiKeyService,
    private readonly userService: UserService,
  ) {}

  @Post('code')
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Request device authorization code' })
  @ApiResponse({ status: 200, description: 'Device code generated successfully' })
  async requestDeviceCode(@Body() dto: RequestDeviceCodeDto) {
    return this.deviceAuthService.requestDeviceCode(dto.client_id, dto.scope)
  }

  @Post('token')
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Poll for device authorization token' })
  @ApiResponse({ status: 200, description: 'Token response' })
  async pollDeviceToken(@Body() dto: PollDeviceTokenDto) {
    const result = await this.deviceAuthService.pollForToken(dto.device_code, dto.client_id)

    // If not authorized, return error
    if ('error' in result) {
      return result
    }

    // Generate API key for the user
    const apiKey = await this.apiKeyService.generateApiKey({
      userId: result.userId,
      organizationId: result.organizationId,
      name: 'CLI Device Authorization',
      expiresIn: 365 * 24 * 60 * 60, // 1 year
    })

    const user = await this.userService.findOne(result.userId)

    return {
      access_token: apiKey.key,
      token_type: 'Bearer',
      expires_in: 365 * 24 * 60 * 60, // 1 year in seconds
      scope: result.scope,
      organization_id: result.organizationId,
      user: {
        id: user.id,
        name: user.name,
        email: user.email,
      },
    }
  }

  @Get('info')
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Get device authorization info by user code' })
  @ApiResponse({ status: 200, description: 'Device authorization info' })
  async getDeviceInfo(@Query('user_code') userCode: string) {
    const authorization = await this.deviceAuthService.getByUserCode(userCode)

    if (!authorization) {
      return { error: 'invalid_user_code', message: 'Invalid or expired user code' }
    }

    return {
      client_id: authorization.clientId,
      scope: authorization.scope,
      expires_at: authorization.expiresAt,
    }
  }

  @Post('approve')
  @UseGuards(CombinedAuthGuard)
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Approve device authorization' })
  @ApiResponse({ status: 200, description: 'Authorization approved' })
  async approveDevice(@Request() req, @Body() dto: ApproveDeviceDto) {
    const userId = req.user.id
    const organizationId = dto.organization_id

    await this.deviceAuthService.approve(dto.user_code, userId, organizationId)

    return {
      success: true,
      message: 'Authorization approved successfully',
    }
  }

  @Post('deny')
  @UseGuards(CombinedAuthGuard)
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Deny device authorization' })
  @ApiResponse({ status: 200, description: 'Authorization denied' })
  async denyDevice(@Body() dto: { user_code: string }) {
    await this.deviceAuthService.deny(dto.user_code)

    return {
      success: true,
      message: 'Authorization denied',
    }
  }
}
