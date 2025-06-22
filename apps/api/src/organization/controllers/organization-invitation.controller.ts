/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, Get, Param, Post, Put, UseGuards } from '@nestjs/common'
import { AuthGuard } from '@nestjs/passport'
import { ApiOAuth2, ApiTags, ApiOperation, ApiResponse, ApiParam, ApiBearerAuth } from '@nestjs/swagger'
import { RequiredOrganizationMemberRole } from '../decorators/required-organization-member-role.decorator'
import { CreateOrganizationInvitationDto } from '../dto/create-organization-invitation.dto'
import { UpdateOrganizationInvitationDto } from '../dto/update-organization-invitation.dto'
import { OrganizationInvitationDto } from '../dto/organization-invitation.dto'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { OrganizationActionGuard } from '../guards/organization-action.guard'
import { OrganizationInvitationService } from '../services/organization-invitation.service'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { AuthContext as IAuthContext } from '../../common/interfaces/auth-context.interface'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'

@ApiTags('organizations')
@Controller('organizations/:organizationId/invitations')
@UseGuards(AuthGuard('jwt'), OrganizationActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class OrganizationInvitationController {
  constructor(private readonly organizationInvitationService: OrganizationInvitationService) {}

  @Post()
  @ApiOperation({
    summary: 'Create organization invitation',
    operationId: 'createOrganizationInvitation',
  })
  @ApiResponse({
    status: 201,
    description: 'Organization invitation created successfully',
    type: OrganizationInvitationDto,
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @RequiredOrganizationMemberRole(OrganizationMemberRole.OWNER)
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.ORGANIZATION_INVITATION,
    targetIdFromResult: (result: OrganizationInvitationDto) => result?.id,
    requestMetadata: {
      body: (req: TypedRequest<CreateOrganizationInvitationDto>) => ({
        email: req.body?.email,
        role: req.body?.role,
        assignedRoleIds: req.body?.assignedRoleIds,
        expiresAt: req.body?.expiresAt,
      }),
    },
  })
  async create(
    @AuthContext() authContext: IAuthContext,
    @Param('organizationId') organizationId: string,
    @Body() createOrganizationInvitationDto: CreateOrganizationInvitationDto,
  ): Promise<OrganizationInvitationDto> {
    const invitation = await this.organizationInvitationService.create(
      organizationId,
      createOrganizationInvitationDto,
      authContext.email,
    )
    return OrganizationInvitationDto.fromOrganizationInvitation(invitation)
  }

  @Put('/:invitationId')
  @ApiOperation({
    summary: 'Update organization invitation',
    operationId: 'updateOrganizationInvitation',
  })
  @ApiResponse({
    status: 200,
    description: 'Organization invitation updated successfully',
    type: OrganizationInvitationDto,
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiParam({
    name: 'invitationId',
    description: 'Invitation ID',
    type: 'string',
  })
  @RequiredOrganizationMemberRole(OrganizationMemberRole.OWNER)
  @Audit({
    action: AuditAction.UPDATE,
    targetType: AuditTarget.ORGANIZATION_INVITATION,
    targetIdFromRequest: (req) => req.params.invitationId,
    requestMetadata: {
      body: (req: TypedRequest<UpdateOrganizationInvitationDto>) => ({
        role: req.body?.role,
        assignedRoleIds: req.body?.assignedRoleIds,
        expiresAt: req.body?.expiresAt,
      }),
    },
  })
  async update(
    @Param('organizationId') organizationId: string,
    @Param('invitationId') invitationId: string,
    @Body() updateOrganizationInvitationDto: UpdateOrganizationInvitationDto,
  ): Promise<OrganizationInvitationDto> {
    const invitation = await this.organizationInvitationService.update(invitationId, updateOrganizationInvitationDto)
    return OrganizationInvitationDto.fromOrganizationInvitation(invitation)
  }

  @Get()
  @ApiOperation({
    summary: 'List pending organization invitations',
    operationId: 'listOrganizationInvitations',
  })
  @ApiResponse({
    status: 200,
    description: 'List of pending organization invitations',
    type: [OrganizationInvitationDto],
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  async findPending(@Param('organizationId') organizationId: string): Promise<OrganizationInvitationDto[]> {
    const invitations = await this.organizationInvitationService.findPending(organizationId)
    return invitations.map(OrganizationInvitationDto.fromOrganizationInvitation)
  }

  @Post('/:invitationId/cancel')
  @ApiOperation({
    summary: 'Cancel organization invitation',
    operationId: 'cancelOrganizationInvitation',
  })
  @ApiResponse({
    status: 204,
    description: 'Organization invitation cancelled successfully',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiParam({
    name: 'invitationId',
    description: 'Invitation ID',
    type: 'string',
  })
  @RequiredOrganizationMemberRole(OrganizationMemberRole.OWNER)
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.ORGANIZATION_INVITATION,
    targetIdFromRequest: (req) => req.params.invitationId,
  })
  async cancel(
    @Param('organizationId') organizationId: string,
    @Param('invitationId') invitationId: string,
  ): Promise<void> {
    return this.organizationInvitationService.cancel(invitationId)
  }
}
