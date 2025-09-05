/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, Delete, ForbiddenException, Get, Param, Post, UseGuards } from '@nestjs/common'
import { AuthGuard } from '@nestjs/passport'
import { ApiOAuth2, ApiTags, ApiOperation, ApiResponse, ApiParam, ApiBearerAuth } from '@nestjs/swagger'
import { RequiredOrganizationMemberRole } from '../decorators/required-organization-member-role.decorator'
import { UpdateOrganizationMemberAccessDto } from '../dto/update-organization-member-access.dto'
import { OrganizationUserDto } from '../dto/organization-user.dto'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { OrganizationActionGuard } from '../guards/organization-action.guard'
import { OrganizationUserService } from '../services/organization-user.service'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { AuthContext as IAuthContext } from '../../common/interfaces/auth-context.interface'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'

@ApiTags('organizations')
@Controller('organizations/:organizationId/users')
@UseGuards(AuthGuard('jwt'), OrganizationActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class OrganizationUserController {
  constructor(private readonly organizationUserService: OrganizationUserService) {}

  @Get()
  @ApiOperation({
    summary: 'List organization members',
    operationId: 'listOrganizationMembers',
  })
  @ApiResponse({
    status: 200,
    description: 'List of organization members',
    type: [OrganizationUserDto],
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  async findAll(@Param('organizationId') organizationId: string): Promise<OrganizationUserDto[]> {
    return this.organizationUserService.findAll(organizationId)
  }

  @Post('/:userId/access')
  @ApiOperation({
    summary: 'Update access for organization member',
    operationId: 'updateAccessForOrganizationMember',
  })
  @ApiResponse({
    status: 200,
    description: 'Access updated successfully',
    type: OrganizationUserDto,
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiParam({
    name: 'userId',
    description: 'User ID',
    type: 'string',
  })
  @RequiredOrganizationMemberRole(OrganizationMemberRole.OWNER)
  @Audit({
    action: AuditAction.UPDATE_ACCESS,
    targetType: AuditTarget.ORGANIZATION_USER,
    targetIdFromRequest: (req) => req.params.userId,
    requestMetadata: {
      body: (req: TypedRequest<UpdateOrganizationMemberAccessDto>) => ({
        role: req.body?.role,
        assignedRoleIds: req.body?.assignedRoleIds,
      }),
    },
  })
  async updateAccess(
    @AuthContext() authContext: IAuthContext,
    @Param('organizationId') organizationId: string,
    @Param('userId') userId: string,
    @Body() dto: UpdateOrganizationMemberAccessDto,
  ): Promise<OrganizationUserDto> {
    if (authContext.userId === userId) {
      throw new ForbiddenException('You cannot update your own access')
    }

    return this.organizationUserService.updateAccess(organizationId, userId, dto.role, dto.assignedRoleIds)
  }

  @Delete('/:userId')
  @ApiOperation({
    summary: 'Delete organization member',
    operationId: 'deleteOrganizationMember',
  })
  @ApiResponse({
    status: 204,
    description: 'User removed from organization successfully',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiParam({
    name: 'userId',
    description: 'User ID',
    type: 'string',
  })
  @RequiredOrganizationMemberRole(OrganizationMemberRole.OWNER)
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.ORGANIZATION_USER,
    targetIdFromRequest: (req) => req.params.userId,
  })
  async delete(@Param('organizationId') organizationId: string, @Param('userId') userId: string): Promise<void> {
    return this.organizationUserService.delete(organizationId, userId)
  }
}
