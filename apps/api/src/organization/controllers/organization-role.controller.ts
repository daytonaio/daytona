/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, Delete, Get, Param, Post, Put, UseGuards } from '@nestjs/common'
import { ApiBearerAuth, ApiOAuth2, ApiOperation, ApiParam, ApiResponse, ApiTags } from '@nestjs/swagger'
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import { RequiredOrganizationMemberRole } from '../decorators/required-organization-member-role.decorator'
import { CreateOrganizationRoleDto } from '../dto/create-organization-role.dto'
import { UpdateOrganizationRoleDto } from '../dto/update-organization-role.dto'
import { OrganizationRoleDto } from '../dto/organization-role.dto'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { OrganizationAuthContextGuard } from '../guards/organization-auth-context.guard'
import { OrganizationRoleService } from '../services/organization-role.service'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'

@Controller('organizations/:organizationId/roles')
@ApiTags('organizations')
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
@AuthStrategy(AuthStrategyType.JWT)
@UseGuards(AuthenticatedRateLimitGuard)
@UseGuards(OrganizationAuthContextGuard)
export class OrganizationRoleController {
  constructor(private readonly organizationRoleService: OrganizationRoleService) {}

  @Post()
  @ApiOperation({
    summary: 'Create organization role',
    operationId: 'createOrganizationRole',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiResponse({
    status: 201,
    description: 'Organization role created successfully',
    type: OrganizationRoleDto,
  })
  @RequiredOrganizationMemberRole(OrganizationMemberRole.OWNER)
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.ORGANIZATION_ROLE,
    targetIdFromResult: (result: OrganizationRoleDto) => result?.id,
    requestMetadata: {
      body: (req: TypedRequest<CreateOrganizationRoleDto>) => ({
        name: req.body?.name,
        description: req.body?.description,
        permissions: req.body?.permissions,
      }),
    },
  })
  async create(
    @Param('organizationId') organizationId: string,
    @Body() createOrganizationRoleDto: CreateOrganizationRoleDto,
  ): Promise<OrganizationRoleDto> {
    const role = await this.organizationRoleService.create(organizationId, createOrganizationRoleDto)
    return OrganizationRoleDto.fromOrganizationRole(role)
  }

  @Get()
  @ApiOperation({
    summary: 'List organization roles',
    operationId: 'listOrganizationRoles',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'List of organization roles',
    type: [OrganizationRoleDto],
  })
  async findAll(@Param('organizationId') organizationId: string): Promise<OrganizationRoleDto[]> {
    const roles = await this.organizationRoleService.findAll(organizationId)
    return roles.map(OrganizationRoleDto.fromOrganizationRole)
  }

  @Put('/:roleId')
  @ApiOperation({
    summary: 'Update organization role',
    operationId: 'updateOrganizationRole',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiParam({
    name: 'roleId',
    description: 'Role ID',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Role updated successfully',
    type: OrganizationRoleDto,
  })
  @RequiredOrganizationMemberRole(OrganizationMemberRole.OWNER)
  @Audit({
    action: AuditAction.UPDATE,
    targetType: AuditTarget.ORGANIZATION_ROLE,
    targetIdFromRequest: (req) => req.params.roleId,
    requestMetadata: {
      body: (req: TypedRequest<UpdateOrganizationRoleDto>) => ({
        name: req.body?.name,
        description: req.body?.description,
        permissions: req.body?.permissions,
      }),
    },
  })
  async updateRole(
    @Param('organizationId') organizationId: string,
    @Param('roleId') roleId: string,
    @Body() updateOrganizationRoleDto: UpdateOrganizationRoleDto,
  ): Promise<OrganizationRoleDto> {
    const updatedRole = await this.organizationRoleService.update(roleId, updateOrganizationRoleDto)
    return OrganizationRoleDto.fromOrganizationRole(updatedRole)
  }

  @Delete('/:roleId')
  @ApiOperation({
    summary: 'Delete organization role',
    operationId: 'deleteOrganizationRole',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiParam({
    name: 'roleId',
    description: 'Role ID',
    type: 'string',
  })
  @ApiResponse({
    status: 204,
    description: 'Organization role deleted successfully',
  })
  @RequiredOrganizationMemberRole(OrganizationMemberRole.OWNER)
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.ORGANIZATION_ROLE,
    targetIdFromRequest: (req) => req.params.roleId,
  })
  async delete(@Param('organizationId') organizationId: string, @Param('roleId') roleId: string): Promise<void> {
    return this.organizationRoleService.delete(roleId)
  }
}
