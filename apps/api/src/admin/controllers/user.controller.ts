/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, Get, Param, Post, UseGuards } from '@nestjs/common'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'
import { ApiBearerAuth, ApiOAuth2, ApiOperation, ApiResponse, ApiTags } from '@nestjs/swagger'
import { RequiredSystemRole } from '../../user/decorators/required-system-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'
import { UserService } from '../../user/user.service'
import { User } from '../../user/user.entity'
import { UserDto } from '../../user/dto/user.dto'
import { CreateUserDto } from '../../user/dto/create-user.dto'
import { NotFoundException } from '@nestjs/common'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'

@Controller('admin/users')
@ApiTags('admin')
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
@AuthStrategy([AuthStrategyType.API_KEY, AuthStrategyType.JWT])
@RequiredSystemRole(SystemRole.ADMIN)
@UseGuards(AuthenticatedRateLimitGuard)
export class AdminUserController {
  constructor(private readonly userService: UserService) {}

  @Post()
  @ApiOperation({
    summary: 'Create user',
    operationId: 'adminCreateUser',
  })
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.USER,
    targetIdFromResult: (result: User) => result?.id,
    requestMetadata: {
      body: (req: TypedRequest<CreateUserDto>) => ({
        id: req.body?.id,
        name: req.body?.name,
        email: req.body?.email,
        personalOrganizationQuota: req.body?.personalOrganizationQuota,
        role: req.body?.role,
        emailVerified: req.body?.emailVerified,
      }),
    },
  })
  async create(@Body() createUserDto: CreateUserDto): Promise<User> {
    return this.userService.create(createUserDto)
  }

  @Get()
  @ApiOperation({
    summary: 'List all users',
    operationId: 'adminListUsers',
  })
  async findAll(): Promise<UserDto[]> {
    const users = await this.userService.findAll()
    return users.map(UserDto.fromUser)
  }

  @Post('/:id/regenerate-key-pair')
  @ApiOperation({
    summary: 'Regenerate user key pair',
    operationId: 'adminRegenerateKeyPair',
  })
  @Audit({
    action: AuditAction.REGENERATE_KEY_PAIR,
    targetType: AuditTarget.USER,
    targetIdFromRequest: (req) => req.params.id,
  })
  async regenerateKeyPair(@Param('id') id: string): Promise<User> {
    return this.userService.regenerateKeyPair(id)
  }

  @Get('/:id')
  @ApiOperation({
    summary: 'Get user by ID',
    operationId: 'adminGetUser',
  })
  @ApiResponse({
    status: 200,
    description: 'User details',
    type: UserDto,
  })
  async getUserById(@Param('id') id: string): Promise<UserDto> {
    const user = await this.userService.findOne(id)
    if (!user) {
      throw new NotFoundException(`User with ID ${id} not found`)
    }

    return UserDto.fromUser(user)
  }
}
