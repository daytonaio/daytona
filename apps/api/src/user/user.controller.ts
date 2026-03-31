/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  BadRequestException,
  Body,
  Controller,
  Delete,
  ForbiddenException,
  Get,
  Logger,
  NotFoundException,
  Param,
  Post,
  UnauthorizedException,
  UseGuards,
} from '@nestjs/common'
import { UserService } from './user.service'
import { ApiOAuth2, ApiTags, ApiOperation, ApiResponse, ApiBearerAuth } from '@nestjs/swagger'
import { IsUserAuthContext } from '../common/decorators/auth-context.decorator'
import { UserAuthContext } from '../common/interfaces/user-auth-context.interface'
import { UserDto } from './dto/user.dto'
import { TypedConfigService } from '../config/typed-config.service'
import axios from 'axios'
import { AccountProviderDto } from './dto/account-provider.dto'
import { ACCOUNT_PROVIDER_DISPLAY_NAME } from './constants/acount-provider-display-name.constant'
import { AccountProvider } from './enums/account-provider.enum'
import { CreateLinkedAccountDto } from './dto/create-linked-account.dto'
import { Audit, TypedRequest } from '../audit/decorators/audit.decorator'
import { AuditAction } from '../audit/enums/audit-action.enum'
import { AuthenticatedRateLimitGuard } from '../common/guards/authenticated-rate-limit.guard'
import { AuthStrategy } from '../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../auth/enums/auth-strategy-type.enum'

@Controller('users')
@ApiTags('users')
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
@AuthStrategy(AuthStrategyType.JWT)
@UseGuards(AuthenticatedRateLimitGuard)
export class UserController {
  private readonly logger = new Logger(UserController.name)

  constructor(
    private readonly userService: UserService,
    private readonly configService: TypedConfigService,
  ) {}

  @Get('/me')
  @ApiOperation({
    summary: 'Get authenticated user',
    operationId: 'getAuthenticatedUser',
  })
  @ApiResponse({
    status: 200,
    description: 'User details',
    type: UserDto,
  })
  async getAuthenticatedUser(@IsUserAuthContext() authContext: UserAuthContext): Promise<UserDto> {
    const user = await this.userService.findOne(authContext.userId)
    if (!user) {
      throw new NotFoundException(`User with ID ${authContext.userId} not found`)
    }

    return UserDto.fromUser(user)
  }

  @Get('/account-providers')
  @ApiOperation({
    summary: 'Get available account providers',
    operationId: 'getAvailableAccountProviders',
  })
  @ApiResponse({
    status: 200,
    description: 'Available account providers',
    type: [AccountProviderDto],
  })
  async getAvailableAccountProviders(): Promise<AccountProviderDto[]> {
    if (!this.configService.get('oidc.managementApi.enabled')) {
      this.logger.warn('OIDC Management API is not enabled')
      throw new NotFoundException()
    }

    const token = await this.getManagementApiToken()

    try {
      const response = await axios.get<{ name: string }[]>(
        `${this.configService.getOrThrow('oidc.issuer')}/api/v2/connections`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        },
      )

      const supportedProviders = new Set([AccountProvider.GOOGLE, AccountProvider.GITHUB])

      const result: AccountProviderDto[] = response.data
        .filter((connection) => supportedProviders.has(connection.name as AccountProvider))
        .map((connection) => ({
          name: connection.name,
          displayName: ACCOUNT_PROVIDER_DISPLAY_NAME[connection.name as AccountProvider],
        }))

      return result
    } catch (error) {
      this.logger.error('Failed to get available account providers', error?.message || String(error))
      throw new UnauthorizedException()
    }
  }

  @Post('/linked-accounts')
  @ApiOperation({
    summary: 'Link account',
    operationId: 'linkAccount',
  })
  @ApiResponse({
    status: 204,
    description: 'Account linked successfully',
  })
  @Audit({
    action: AuditAction.LINK_ACCOUNT,
    requestMetadata: {
      body: (req: TypedRequest<CreateLinkedAccountDto>) => ({
        provider: req.body?.provider,
        userId: req.body?.userId,
      }),
    },
  })
  async linkAccount(
    @IsUserAuthContext() authContext: UserAuthContext,
    @Body() createLinkedAccountDto: CreateLinkedAccountDto,
  ): Promise<void> {
    if (!this.configService.get('oidc.managementApi.enabled')) {
      this.logger.warn('OIDC Management API is not enabled')
      throw new NotFoundException()
    }

    const authenticatedUser = await this.userService.findOne(authContext.userId)
    if (!authenticatedUser.emailVerified) {
      throw new ForbiddenException('Please verify your email address')
    }

    const userToLinkId = `${createLinkedAccountDto.provider}|${createLinkedAccountDto.userId}`

    // Verify user doesn't already exist in our user table
    const userToLink = await this.userService.findOne(userToLinkId)
    if (userToLink) {
      throw new BadRequestException('This account is already associated with another user')
    }

    const token = await this.getManagementApiToken()

    // Verify account is eligible to be linked (must be reachable via OIDC Management API)
    try {
      await axios.get(
        `${this.configService.getOrThrow('oidc.issuer')}/api/v2/users/${encodeURIComponent(userToLinkId)}`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        },
      )
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.status === 404) {
        throw new BadRequestException('Account not found or already linked to another user')
      }
      throw error
    }

    // Link account
    try {
      await axios.post(
        `${this.configService.getOrThrow('oidc.issuer')}/api/v2/users/${authContext.userId}/identities`,
        {
          provider: createLinkedAccountDto.provider,
          user_id: createLinkedAccountDto.userId,
        },
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        },
      )
    } catch (error) {
      this.logger.error('Failed to link account', error?.message || String(error))
      throw new UnauthorizedException()
    }
  }

  @Delete('/linked-accounts/:provider/:providerUserId')
  @ApiOperation({
    summary: 'Unlink account',
    operationId: 'unlinkAccount',
  })
  @ApiResponse({
    status: 204,
    description: 'Account unlinked successfully',
  })
  @Audit({
    action: AuditAction.UNLINK_ACCOUNT,
    requestMetadata: {
      params: (req) => ({
        provider: req.params.provider,
        providerUserId: req.params.providerUserId,
      }),
    },
  })
  async unlinkAccount(
    @IsUserAuthContext() authContext: UserAuthContext,
    @Param('provider') provider: string,
    @Param('providerUserId') providerUserId: string,
  ): Promise<void> {
    if (!this.configService.get('oidc.managementApi.enabled')) {
      this.logger.warn('OIDC Management API is not enabled')
      throw new NotFoundException()
    }

    const token = await this.getManagementApiToken()

    try {
      await axios.delete(
        `${this.configService.getOrThrow('oidc.issuer')}/api/v2/users/${authContext.userId}/identities/${provider}/${providerUserId}`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        },
      )
    } catch (error) {
      this.logger.error('Failed to unlink account', error?.message || String(error))
      throw new UnauthorizedException()
    }
  }

  @Post('/mfa/sms/enroll')
  @ApiOperation({
    summary: 'Enroll in SMS MFA',
    operationId: 'enrollInSmsMfa',
  })
  @ApiResponse({
    status: 200,
    description: 'SMS MFA enrollment URL',
    type: String,
  })
  async enrollInSmsMfa(@IsUserAuthContext() authContext: UserAuthContext): Promise<string> {
    if (!this.configService.get('oidc.managementApi.enabled')) {
      this.logger.warn('OIDC Management API is not enabled')
      throw new NotFoundException()
    }

    const token = await this.getManagementApiToken()

    try {
      const response = await axios.post(
        `${this.configService.getOrThrow('oidc.issuer')}/api/v2/guardian/enrollments/ticket`,
        {
          user_id: authContext.userId,
        },
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        },
      )

      return response.data.ticket_url
    } catch (error) {
      this.logger.error('Failed to enable SMS MFA', error?.message || String(error))
      throw new UnauthorizedException()
    }
  }

  private async getManagementApiToken(): Promise<string> {
    try {
      const tokenResponse = await axios.post(`${this.configService.getOrThrow('oidc.issuer')}/oauth/token`, {
        grant_type: 'client_credentials',
        client_id: this.configService.getOrThrow('oidc.managementApi.clientId'),
        client_secret: this.configService.getOrThrow('oidc.managementApi.clientSecret'),
        audience: this.configService.getOrThrow('oidc.managementApi.audience'),
      })
      return tokenResponse.data.access_token
    } catch (error) {
      this.logger.error('Failed to get OIDC Management API token', error?.message || String(error))
      throw new UnauthorizedException()
    }
  }
}
