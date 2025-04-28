/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, Get, NotFoundException, Param, Post, UseGuards } from '@nestjs/common'
import { User } from './user.entity'
import { UserService } from './user.service'
import { CreateUserDto } from './dto/create-user.dto'
import { ApiOAuth2, ApiTags, ApiOperation, ApiResponse } from '@nestjs/swagger'
import { CombinedAuthGuard } from '../auth/combined-auth.guard'
import { AuthContext } from '../common/decorators/auth-context.decorator'
import { AuthContext as IAuthContext } from '../common/interfaces/auth-context.interface'
import { UserDto } from './dto/user.dto'
import { SystemActionGuard } from '../auth/system-action.guard'
import { RequiredSystemRole } from '../common/decorators/required-system-role.decorator'
import { SystemRole } from './enums/system-role.enum'

@ApiTags('users')
@Controller('users')
@UseGuards(CombinedAuthGuard, SystemActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
export class UserController {
  constructor(private readonly userService: UserService) {}

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
}
