/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Body,
  Controller,
  Delete,
  ForbiddenException,
  Get,
  NotFoundException,
  Param,
  Patch,
  Post,
  UseGuards,
} from '@nestjs/common'
import { AuthGuard } from '@nestjs/passport'
import { ApiOAuth2, ApiTags, ApiOperation, ApiResponse, ApiParam, ApiBody, ApiBearerAuth } from '@nestjs/swagger'
import { RequiredOrganizationMemberRole } from '../decorators/required-organization-member-role.decorator'
import { CreateOrganizationDto } from '../dto/create-organization.dto'
import { OrganizationDto } from '../dto/organization.dto'
import { OrganizationInvitationDto } from '../dto/organization-invitation.dto'
import { OverviewDto } from '../dto/overview.dto'
import { UpdateOrganizationQuotaDto } from '../dto/update-organization-quota.dto'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { OrganizationActionGuard } from '../guards/organization-action.guard'
import { OrganizationService } from '../services/organization.service'
import { OrganizationUserService } from '../services/organization-user.service'
import { OrganizationInvitationService } from '../services/organization-invitation.service'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { AuthContext as IAuthContext } from '../../common/interfaces/auth-context.interface'
import { SystemActionGuard } from '../../auth/system-action.guard'
import { RequiredSystemRole } from '../../common/decorators/required-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'
import { OrganizationSuspensionDto } from '../dto/organization-suspension.dto'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { UserService } from '../../user/user.service'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { EmailUtils } from '../../common/utils/email.util'

@ApiTags('organizations')
@Controller('organizations')
// TODO: Rethink this. Can we allow access to these methods with API keys as well?
// @UseGuards(AuthGuard('jwt'))
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class OrganizationController {
  constructor(
    private readonly organizationService: OrganizationService,
    private readonly organizationUserService: OrganizationUserService,
    private readonly organizationInvitationService: OrganizationInvitationService,
    private readonly userService: UserService,
  ) {}

  @Get('/invitations')
  @ApiOperation({
    summary: 'List organization invitations for authenticated user',
    operationId: 'listOrganizationInvitationsForAuthenticatedUser',
  })
  @ApiResponse({
    status: 200,
    description: 'List of organization invitations',
    type: [OrganizationInvitationDto],
  })
  @UseGuards(AuthGuard('jwt'))
  async findInvitationsByUser(@AuthContext() authContext: IAuthContext): Promise<OrganizationInvitationDto[]> {
    const invitations = await this.organizationInvitationService.findByUser(authContext.userId)
    return invitations.map(OrganizationInvitationDto.fromOrganizationInvitation)
  }

  @Get('/invitations/count')
  @ApiOperation({
    summary: 'Get count of organization invitations for authenticated user',
    operationId: 'getOrganizationInvitationsCountForAuthenticatedUser',
  })
  @ApiResponse({
    status: 200,
    description: 'Count of organization invitations',
    type: Number,
  })
  @UseGuards(AuthGuard('jwt'))
  async getInvitationsCountByUser(@AuthContext() authContext: IAuthContext): Promise<number> {
    return this.organizationInvitationService.getCountByUser(authContext.userId)
  }

  @Post('/invitations/:invitationId/accept')
  @ApiOperation({
    summary: 'Accept organization invitation',
    operationId: 'acceptOrganizationInvitation',
  })
  @ApiResponse({
    status: 200,
    description: 'Organization invitation accepted successfully',
  })
  @ApiParam({
    name: 'invitationId',
    description: 'Invitation ID',
    type: 'string',
  })
  @UseGuards(AuthGuard('jwt'))
  @Audit({
    action: AuditAction.ACCEPT,
    targetType: AuditTarget.ORGANIZATION_INVITATION,
    targetIdFromRequest: (req) => req.params.invitationId,
  })
  async acceptInvitation(
    @AuthContext() authContext: IAuthContext,
    @Param('invitationId') invitationId: string,
  ): Promise<void> {
    try {
      const invitation = await this.organizationInvitationService.findOneOrFail(invitationId)
      if (!EmailUtils.areEqual(invitation.email, authContext.email)) {
        throw new ForbiddenException('User email does not match invitation email')
      }
    } catch (error) {
      throw new NotFoundException(`Organization invitation with ID ${invitationId} not found`)
    }

    return this.organizationInvitationService.accept(invitationId, authContext.userId)
  }

  @Post('/invitations/:invitationId/decline')
  @ApiOperation({
    summary: 'Decline organization invitation',
    operationId: 'declineOrganizationInvitation',
  })
  @ApiResponse({
    status: 200,
    description: 'Organization invitation declined successfully',
  })
  @ApiParam({
    name: 'invitationId',
    description: 'Invitation ID',
    type: 'string',
  })
  @UseGuards(AuthGuard('jwt'))
  @Audit({
    action: AuditAction.DECLINE,
    targetType: AuditTarget.ORGANIZATION_INVITATION,
    targetIdFromRequest: (req) => req.params.invitationId,
  })
  async declineInvitation(
    @AuthContext() authContext: IAuthContext,
    @Param('invitationId') invitationId: string,
  ): Promise<void> {
    try {
      const invitation = await this.organizationInvitationService.findOneOrFail(invitationId)
      if (!EmailUtils.areEqual(invitation.email, authContext.email)) {
        throw new ForbiddenException('User email does not match invitation email')
      }
    } catch (error) {
      throw new NotFoundException(`Organization invitation with ID ${invitationId} not found`)
    }

    return this.organizationInvitationService.decline(invitationId)
  }

  @Post()
  @ApiOperation({
    summary: 'Create organization',
    operationId: 'createOrganization',
  })
  @ApiResponse({
    status: 201,
    description: 'Organization created successfully',
    type: OrganizationDto,
  })
  @UseGuards(AuthGuard('jwt'))
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.ORGANIZATION,
    targetIdFromResult: (result: OrganizationDto) => result?.id,
    requestMetadata: {
      body: (req: TypedRequest<CreateOrganizationDto>) => ({
        name: req.body?.name,
      }),
    },
  })
  async create(
    @AuthContext() authContext: IAuthContext,
    @Body() createOrganizationDto: CreateOrganizationDto,
  ): Promise<OrganizationDto> {
    const user = await this.userService.findOne(authContext.userId)
    if (!user.emailVerified) {
      throw new ForbiddenException('Please verify your email address')
    }

    const organization = await this.organizationService.create(createOrganizationDto, authContext.userId, false, true)
    return OrganizationDto.fromOrganization(organization)
  }

  @Get()
  @ApiOperation({
    summary: 'List organizations',
    operationId: 'listOrganizations',
  })
  @ApiResponse({
    status: 200,
    description: 'List of organizations',
    type: [OrganizationDto],
  })
  @UseGuards(AuthGuard('jwt'))
  async findAll(@AuthContext() authContext: IAuthContext): Promise<OrganizationDto[]> {
    const organizations = await this.organizationService.findByUser(authContext.userId)
    return organizations.map(OrganizationDto.fromOrganization)
  }

  @Get('/:organizationId')
  @ApiOperation({
    summary: 'Get organization by ID',
    operationId: 'getOrganization',
  })
  @ApiResponse({
    status: 200,
    description: 'Organization details',
    type: OrganizationDto,
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @UseGuards(AuthGuard('jwt'), OrganizationActionGuard)
  async findOne(@Param('organizationId') organizationId: string): Promise<OrganizationDto> {
    const organization = await this.organizationService.findOne(organizationId)
    if (!organization) {
      throw new NotFoundException(`Organization with ID ${organizationId} not found`)
    }

    return OrganizationDto.fromOrganization(organization)
  }

  @Delete('/:organizationId')
  @ApiOperation({
    summary: 'Delete organization',
    operationId: 'deleteOrganization',
  })
  @ApiResponse({
    status: 204,
    description: 'Organization deleted successfully',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @UseGuards(AuthGuard('jwt'), OrganizationActionGuard)
  @RequiredOrganizationMemberRole(OrganizationMemberRole.OWNER)
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.ORGANIZATION,
    targetIdFromRequest: (req) => req.params.organizationId,
  })
  async delete(@Param('organizationId') organizationId: string): Promise<void> {
    return this.organizationService.delete(organizationId)
  }

  @Get('/:organizationId/usage')
  @ApiOperation({
    summary: 'Get organization current usage overview',
    operationId: 'getOrganizationUsageOverview',
  })
  @ApiResponse({
    status: 200,
    description: 'Current usage overview',
    type: OverviewDto,
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @UseGuards(AuthGuard('jwt'), OrganizationActionGuard)
  async getUsageOverview(@Param('organizationId') organizationId: string): Promise<OverviewDto> {
    return this.organizationService.getUsageOverview(organizationId)
  }

  @Patch('/:organizationId/quota')
  @ApiOperation({
    summary: 'Update organization quota',
    operationId: 'updateOrganizationQuota',
  })
  @ApiResponse({
    status: 200,
    description: 'Organization details',
    type: OrganizationDto,
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @RequiredSystemRole(SystemRole.ADMIN)
  @UseGuards(CombinedAuthGuard, SystemActionGuard)
  @Audit({
    action: AuditAction.UPDATE_QUOTA,
    targetType: AuditTarget.ORGANIZATION,
    targetIdFromRequest: (req) => req.params.organizationId,
    requestMetadata: {
      body: (req: TypedRequest<UpdateOrganizationQuotaDto>) => ({
        totalCpuQuota: req.body?.totalCpuQuota,
        totalMemoryQuota: req.body?.totalMemoryQuota,
        totalDiskQuota: req.body?.totalDiskQuota,
        maxCpuPerSandbox: req.body?.maxCpuPerSandbox,
        maxMemoryPerSandbox: req.body?.maxMemoryPerSandbox,
        maxDiskPerSandbox: req.body?.maxDiskPerSandbox,
        snapshotQuota: req.body?.snapshotQuota,
        maxSnapshotSize: req.body?.maxSnapshotSize,
        volumeQuota: req.body?.volumeQuota,
      }),
    },
  })
  async updateOrganizationQuota(
    @Param('organizationId') organizationId: string,
    @Body() updateOrganizationQuotaDto: UpdateOrganizationQuotaDto,
  ): Promise<OrganizationDto> {
    const organization = await this.organizationService.updateQuota(organizationId, updateOrganizationQuotaDto)
    return OrganizationDto.fromOrganization(organization)
  }

  @Post('/:organizationId/leave')
  @ApiOperation({
    summary: 'Leave organization',
    operationId: 'leaveOrganization',
  })
  @ApiResponse({
    status: 204,
    description: 'Organization left successfully',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @UseGuards(AuthGuard('jwt'), OrganizationActionGuard)
  @Audit({
    action: AuditAction.LEAVE_ORGANIZATION,
  })
  async leave(
    @AuthContext() authContext: IAuthContext,
    @Param('organizationId') organizationId: string,
  ): Promise<void> {
    return this.organizationUserService.delete(organizationId, authContext.userId)
  }

  @Post('/:organizationId/suspend')
  @ApiOperation({
    summary: 'Suspend organization',
    operationId: 'suspendOrganization',
  })
  @ApiResponse({
    status: 204,
    description: 'Organization suspended successfully',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiBody({
    type: OrganizationSuspensionDto,
    required: false,
  })
  @RequiredSystemRole(SystemRole.ADMIN)
  @UseGuards(CombinedAuthGuard, SystemActionGuard)
  @Audit({
    action: AuditAction.SUSPEND,
    targetType: AuditTarget.ORGANIZATION,
    targetIdFromRequest: (req) => req.params.organizationId,
    requestMetadata: {
      body: (req: TypedRequest<OrganizationSuspensionDto>) => ({
        reason: req.body?.reason,
        until: req.body?.until,
      }),
    },
  })
  async suspend(
    @Param('organizationId') organizationId: string,
    @Body() organizationSuspensionDto?: OrganizationSuspensionDto,
  ): Promise<void> {
    return this.organizationService.suspend(
      organizationId,
      organizationSuspensionDto?.reason,
      organizationSuspensionDto?.until,
    )
  }

  @Post('/:organizationId/unsuspend')
  @ApiOperation({
    summary: 'Unsuspend organization',
    operationId: 'unsuspendOrganization',
  })
  @ApiResponse({
    status: 204,
    description: 'Organization unsuspended successfully',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @RequiredSystemRole(SystemRole.ADMIN)
  @UseGuards(CombinedAuthGuard, SystemActionGuard)
  @Audit({
    action: AuditAction.UNSUSPEND,
    targetType: AuditTarget.ORGANIZATION,
    targetIdFromRequest: (req) => req.params.organizationId,
  })
  async unsuspend(@Param('organizationId') organizationId: string): Promise<void> {
    return this.organizationService.unsuspend(organizationId)
  }
}
