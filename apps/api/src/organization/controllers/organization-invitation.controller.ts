/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, Get, Param, Post, Put, UseGuards } from '@nestjs/common'
import { ApiOAuth2, ApiTags, ApiOperation, ApiResponse, ApiParam, ApiBearerAuth } from '@nestjs/swagger'
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import { RequiredOrganizationMemberRole } from '../decorators/required-organization-member-role.decorator'
import { CreateOrganizationInvitationDto } from '../dto/create-organization-invitation.dto'
import { UpdateOrganizationInvitationDto } from '../dto/update-organization-invitation.dto'
import { OrganizationInvitationDto } from '../dto/organization-invitation.dto'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { OrganizationAuthContextGuard } from '../guards/organization-auth-context.guard'
import { OrganizationInvitationService } from '../services/organization-invitation.service'
import { IsOrganizationAuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'

@Controller('organizations/:organizationId/invitations')
@ApiTags('organizations')
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
@AuthStrategy(AuthStrategyType.JWT)
@UseGuards(AuthenticatedRateLimitGuard)
@UseGuards(OrganizationAuthContextGuard)
export class OrganizationInvitationController {
  constructor(private readonly organizationInvitationService: OrganizationInvitationService) {}

  @Post()
  @ApiOperation({
    summary: 'Create organization invitation',
    operationId: 'createOrganizationInvitation',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiResponse({
    status: 201,
    description: 'Organization invitation created successfully',
    type: OrganizationInvitationDto,
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
    @IsOrganizationAuthContext() authContext: OrganizationAuthContext,
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
  @ApiResponse({
    status: 200,
    description: 'Organization invitation updated successfully',
    type: OrganizationInvitationDto,
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
    const invitation = await this.organizationInvitationService.update(
      organizationId,
      invitationId,
      updateOrganizationInvitationDto,
    )
    return OrganizationInvitationDto.fromOrganizationInvitation(invitation)
  }

  @Get()
  @ApiOperation({
    summary: 'List pending organization invitations',
    operationId: 'listOrganizationInvitations',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'List of pending organization invitations',
    type: [OrganizationInvitationDto],
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
  @ApiResponse({
    status: 204,
    description: 'Organization invitation cancelled successfully',
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
    return this.organizationInvitationService.cancel(organizationId, invitationId)
  }
}
