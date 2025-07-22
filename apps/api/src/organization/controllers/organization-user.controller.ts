/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, Delete, ForbiddenException, Get, Param, Post, UseGuards } from '@nestjs/common'
import { AuthGuard } from '@nestjs/passport'
import { ApiOAuth2, ApiTags, ApiOperation, ApiResponse, ApiParam, ApiBearerAuth } from '@nestjs/swagger'
import { RequiredOrganizationMemberRole } from '../decorators/required-organization-member-role.decorator'
import { UpdateAssignedOrganizationRolesDto } from '../dto/update-assigned-organization-roles.dto'
import { UpdateOrganizationMemberRoleDto } from '../dto/update-organization-member-role.dto'
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

  @Post('/:userId/role')
  @ApiOperation({
    summary: 'Update role for organization member',
    operationId: 'updateRoleForOrganizationMember',
  })
  @ApiResponse({
    status: 200,
    description: 'Role updated successfully',
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
    action: AuditAction.UPDATE_ROLE,
    targetType: AuditTarget.ORGANIZATION_USER,
    targetIdFromRequest: (req) => req.params.userId,
    requestMetadata: {
      body: (req: TypedRequest<UpdateOrganizationMemberRoleDto>) => ({
        role: req.body?.role,
      }),
    },
  })
  async updateRole(
    @AuthContext() authContext: IAuthContext,
    @Param('organizationId') organizationId: string,
    @Param('userId') userId: string,
    @Body() dto: UpdateOrganizationMemberRoleDto,
  ): Promise<OrganizationUserDto> {
    if (authContext.userId === userId) {
      throw new ForbiddenException('You cannot update your own role')
    }

    return this.organizationUserService.updateRole(organizationId, userId, dto.role)
  }

  @Post('/:userId/assigned-roles')
  @ApiOperation({
    summary: 'Update assigned roles to organization member',
    operationId: 'updateAssignedOrganizationRoles',
  })
  @ApiResponse({
    status: 200,
    description: 'Assigned roles updated successfully',
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
    action: AuditAction.UPDATE_ASSIGNED_ROLES,
    targetType: AuditTarget.ORGANIZATION_USER,
    targetIdFromRequest: (req) => req.params.userId,
    requestMetadata: {
      body: (req: TypedRequest<UpdateAssignedOrganizationRolesDto>) => ({
        roleIds: req.body?.roleIds,
      }),
    },
  })
  async updateAssignedRoles(
    @Param('organizationId') organizationId: string,
    @Param('userId') userId: string,
    @Body() dto: UpdateAssignedOrganizationRolesDto,
  ): Promise<OrganizationUserDto> {
    return this.organizationUserService.updateAssignedRoles(organizationId, userId, dto.roleIds)
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
