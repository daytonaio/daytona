/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Body,
  Controller,
  Delete,
  Get,
  Logger,
  NotFoundException,
  Param,
  Post,
  UnauthorizedException,
  UseGuards,
} from '@nestjs/common'
import { User } from './user.entity'
import { UserService } from './user.service'
import { CreateUserDto } from './dto/create-user.dto'
import { ApiOAuth2, ApiTags, ApiOperation, ApiResponse, ApiBearerAuth } from '@nestjs/swagger'
import { CombinedAuthGuard } from '../auth/combined-auth.guard'
import { AuthContext } from '../common/decorators/auth-context.decorator'
import { AuthContext as IAuthContext } from '../common/interfaces/auth-context.interface'
import { UserDto } from './dto/user.dto'
import { SystemActionGuard } from '../auth/system-action.guard'
import { RequiredSystemRole } from '../common/decorators/required-system-role.decorator'
import { SystemRole } from './enums/system-role.enum'
import { TypedConfigService } from '../config/typed-config.service'
import axios from 'axios'
import { AccountProviderDto } from './dto/account-provider.dto'
import { ACCOUNT_PROVIDER_DISPLAY_NAME } from './constants/acount-provider-display-name.constant'
import { AccountProvider } from './enums/account-provider.enum'

@ApiTags('users')
@Controller('users')
@UseGuards(CombinedAuthGuard, SystemActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
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
  async getAuthenticatedUser(@AuthContext() authContext: IAuthContext): Promise<UserDto> {
    const user = await this.userService.findOne(authContext.userId)
    if (!user) {
      throw new NotFoundException(`User with ID ${authContext.userId} not found`)
    }

    return UserDto.fromUser(user)
  }

  @Post()
  @ApiOperation({
    summary: 'Create user',
    operationId: 'createUser',
  })
  @RequiredSystemRole(SystemRole.ADMIN)
  async create(@Body() createUserDto: CreateUserDto): Promise<User> {
    return this.userService.create(createUserDto)
  }

  @Get()
  @ApiOperation({
    summary: 'List all users',
    operationId: 'listUsers',
  })
  @RequiredSystemRole(SystemRole.ADMIN)
  async findAll(): Promise<User[]> {
    return this.userService.findAll()
  }

  @Post('/:id/regenerate-key-pair')
  @ApiOperation({
    summary: 'Regenerate user key pair',
    operationId: 'regenerateKeyPair',
  })
  @RequiredSystemRole(SystemRole.ADMIN)
  async regenerateKeyPair(@Param('id') id: string): Promise<User> {
    return this.userService.regenerateKeyPair(id)
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

    const token = await this.getManagementApiToken('read:connections')

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
      this.logger.error('Failed to get available account providers', error)
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
  async unlinkAccount(
    @AuthContext() authContext: IAuthContext,
    @Param('provider') provider: string,
    @Param('providerUserId') providerUserId: string,
  ): Promise<void> {
    if (!this.configService.get('oidc.managementApi.enabled')) {
      this.logger.warn('OIDC Management API is not enabled')
      throw new NotFoundException()
    }

    const token = await this.getManagementApiToken('update:users')

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
      this.logger.error('Failed to unlink account', error)
      throw new UnauthorizedException()
    }
  }

  private async getManagementApiToken(scope: string): Promise<string> {
    try {
      const tokenResponse = await axios.post(`${this.configService.getOrThrow('oidc.issuer')}/oauth/token`, {
        grant_type: 'client_credentials',
        client_id: this.configService.getOrThrow('oidc.managementApi.clientId'),
        client_secret: this.configService.getOrThrow('oidc.managementApi.clientSecret'),
        audience: this.configService.getOrThrow('oidc.managementApi.audience'),
        scope,
      })
      return tokenResponse.data.access_token
    } catch (error) {
      this.logger.error('Failed to get OIDC Management API token', error)
      throw new UnauthorizedException()
    }
  }
}
