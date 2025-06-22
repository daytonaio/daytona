/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, Delete, Get, Param, Post, Put, UseGuards } from '@nestjs/common'
import { AuthGuard } from '@nestjs/passport'
import { ApiOAuth2, ApiTags, ApiOperation, ApiResponse, ApiParam, ApiBearerAuth } from '@nestjs/swagger'
import { RequiredOrganizationMemberRole } from '../decorators/required-organization-member-role.decorator'
import { CreateOrganizationRoleDto } from '../dto/create-organization-role.dto'
import { UpdateOrganizationRoleDto } from '../dto/update-organization-role.dto'
import { OrganizationRoleDto } from '../dto/organization-role.dto'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { OrganizationActionGuard } from '../guards/organization-action.guard'
import { OrganizationRoleService } from '../services/organization-role.service'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'

@ApiTags('organizations')
@Controller('organizations/:organizationId/roles')
@UseGuards(AuthGuard('jwt'), OrganizationActionGuard)
@RequiredOrganizationMemberRole(OrganizationMemberRole.OWNER)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class OrganizationRoleController {
  constructor(private readonly organizationRoleService: OrganizationRoleService) {}

  @Post()
  @ApiOperation({
    summary: 'Create organization role',
    operationId: 'createOrganizationRole',
  })
  @ApiResponse({
    status: 201,
    description: 'Organization role created successfully',
    type: OrganizationRoleDto,
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
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
  @ApiResponse({
    status: 200,
    description: 'List of organization roles',
    type: [OrganizationRoleDto],
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
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
  @ApiResponse({
    status: 200,
    description: 'Role updated successfully',
    type: OrganizationRoleDto,
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
  @ApiResponse({
    status: 204,
    description: 'Organization role deleted successfully',
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
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.ORGANIZATION_ROLE,
    targetIdFromRequest: (req) => req.params.roleId,
  })
  async delete(@Param('organizationId') organizationId: string, @Param('roleId') roleId: string): Promise<void> {
    return this.organizationRoleService.delete(roleId)
  }
}
