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
  HttpCode,
  NotFoundException,
  Param,
  Patch,
  Post,
  Put,
  UseGuards,
} from '@nestjs/common'
import { AuthGuard } from '@nestjs/passport'
import { ApiOAuth2, ApiTags, ApiOperation, ApiResponse, ApiParam, ApiBody, ApiBearerAuth } from '@nestjs/swagger'
import { RequiredOrganizationMemberRole } from '../decorators/required-organization-member-role.decorator'
import { CreateOrganizationDto } from '../dto/create-organization.dto'
import { OrganizationDto } from '../dto/organization.dto'
import { OrganizationInvitationDto } from '../dto/organization-invitation.dto'
import { OrganizationUsageOverviewDto } from '../dto/organization-usage-overview.dto'
import { UpdateOrganizationQuotaDto } from '../dto/update-organization-quota.dto'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { OrganizationActionGuard } from '../guards/organization-action.guard'
import { OrganizationService } from '../services/organization.service'
import { OrganizationUserService } from '../services/organization-user.service'
import { OrganizationInvitationService } from '../services/organization-invitation.service'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { AuthContext as IAuthContext } from '../../common/interfaces/auth-context.interface'
import { SystemActionGuard } from '../../auth/system-action.guard'
import { RequiredApiRole, RequiredSystemRole } from '../../common/decorators/required-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'
import { OrganizationSuspensionDto } from '../dto/organization-suspension.dto'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { UserService } from '../../user/user.service'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { EmailUtils } from '../../common/utils/email.util'
import { OrganizationUsageService } from '../services/organization-usage.service'
import { OrganizationSandboxDefaultLimitedNetworkEgressDto } from '../dto/organization-sandbox-default-limited-network-egress.dto'
import { TypedConfigService } from '../../config/typed-config.service'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'
import { UpdateOrganizationRegionQuotaDto } from '../dto/update-organization-region-quota.dto'
import { UpdateOrganizationDefaultRegionDto } from '../dto/update-organization-default-region.dto'
import { RegionQuotaDto } from '../dto/region-quota.dto'
import { RequireFlagsEnabled } from '@openfeature/nestjs-sdk'
import { OrGuard } from '../../auth/or.guard'
import { OtelProxyGuard } from '../../auth/otel-proxy.guard'
import { OtelConfigDto } from '../dto/otel-config.dto'

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
    private readonly organizationUsageService: OrganizationUsageService,
    private readonly userService: UserService,
    private readonly configService: TypedConfigService,
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
    type: OrganizationInvitationDto,
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
  ): Promise<OrganizationInvitationDto> {
    try {
      const invitation = await this.organizationInvitationService.findOneOrFail(invitationId)
      if (!EmailUtils.areEqual(invitation.email, authContext.email)) {
        throw new ForbiddenException('User email does not match invitation email')
      }
    } catch (error) {
      throw new NotFoundException(`Organization invitation with ID ${invitationId} not found`)
    }

    const acceptedInvitation = await this.organizationInvitationService.accept(invitationId, authContext.userId)
    return OrganizationInvitationDto.fromOrganizationInvitation(acceptedInvitation)
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
        defaultRegionId: req.body?.defaultRegionId,
      }),
    },
  })
  async create(
    @AuthContext() authContext: IAuthContext,
    @Body() createOrganizationDto: CreateOrganizationDto,
  ): Promise<OrganizationDto> {
    const user = await this.userService.findOne(authContext.userId)
    if (!user.emailVerified && !this.configService.get('skipUserEmailVerification')) {
      throw new ForbiddenException('Please verify your email address')
    }

    const organization = await this.organizationService.create(createOrganizationDto, authContext.userId, false, true)
    return OrganizationDto.fromOrganization(organization)
  }

  @Patch('/:organizationId/default-region')
  @HttpCode(204)
  @ApiOperation({
    summary: 'Set default region for organization',
    operationId: 'setOrganizationDefaultRegion',
  })
  @ApiResponse({
    status: 204,
    description: 'Default region set successfully',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiBody({
    type: UpdateOrganizationDefaultRegionDto,
    required: true,
  })
  @UseGuards(AuthGuard('jwt'), AuthenticatedRateLimitGuard, OrganizationActionGuard)
  @RequiredOrganizationMemberRole(OrganizationMemberRole.OWNER)
  @Audit({
    action: AuditAction.UPDATE,
    targetType: AuditTarget.ORGANIZATION,
    targetIdFromRequest: (req) => req.params.organizationId,
    requestMetadata: {
      body: (req: TypedRequest<UpdateOrganizationDefaultRegionDto>) => ({
        defaultRegionId: req.body?.defaultRegionId,
      }),
    },
  })
  async setDefaultRegion(
    @Param('organizationId') organizationId: string,
    @Body() updateDto: UpdateOrganizationDefaultRegionDto,
  ): Promise<void> {
    await this.organizationService.setDefaultRegion(organizationId, updateDto.defaultRegionId)
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
    type: OrganizationUsageOverviewDto,
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @UseGuards(AuthGuard('jwt'), OrganizationActionGuard)
  async getUsageOverview(@Param('organizationId') organizationId: string): Promise<OrganizationUsageOverviewDto> {
    return this.organizationUsageService.getUsageOverview(organizationId)
  }

  @Patch('/:organizationId/quota')
  @HttpCode(204)
  @ApiOperation({
    summary: 'Update organization quota',
    operationId: 'updateOrganizationQuota',
  })
  @ApiResponse({
    status: 204,
    description: 'Organization quota updated successfully',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @RequiredSystemRole(SystemRole.ADMIN)
  @UseGuards(CombinedAuthGuard, AuthenticatedRateLimitGuard, SystemActionGuard)
  @Audit({
    action: AuditAction.UPDATE_QUOTA,
    targetType: AuditTarget.ORGANIZATION,
    targetIdFromRequest: (req) => req.params.organizationId,
    requestMetadata: {
      body: (req: TypedRequest<UpdateOrganizationQuotaDto>) => ({
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
    @Body() updateDto: UpdateOrganizationQuotaDto,
  ): Promise<void> {
    await this.organizationService.updateQuota(organizationId, updateDto)
  }

  @Patch('/:organizationId/quota/:regionId')
  @HttpCode(204)
  @ApiOperation({
    summary: 'Update organization region quota',
    operationId: 'updateOrganizationRegionQuota',
  })
  @ApiResponse({
    status: 204,
    description: 'Region quota updated successfully',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiParam({
    name: 'regionId',
    description: 'ID of the region where the updated quota will be applied',
    type: 'string',
  })
  @RequiredSystemRole(SystemRole.ADMIN)
  @UseGuards(CombinedAuthGuard, AuthenticatedRateLimitGuard, SystemActionGuard)
  @Audit({
    action: AuditAction.UPDATE_REGION_QUOTA,
    targetType: AuditTarget.ORGANIZATION,
    targetIdFromRequest: (req) => req.params.organizationId,
    requestMetadata: {
      params: (req) => ({
        regionId: req.params.regionId,
      }),
      body: (req: TypedRequest<UpdateOrganizationRegionQuotaDto>) => ({
        totalCpuQuota: req.body?.totalCpuQuota,
        totalMemoryQuota: req.body?.totalMemoryQuota,
        totalDiskQuota: req.body?.totalDiskQuota,
      }),
    },
  })
  async updateOrganizationRegionQuota(
    @Param('organizationId') organizationId: string,
    @Param('regionId') regionId: string,
    @Body() updateDto: UpdateOrganizationRegionQuotaDto,
  ): Promise<void> {
    await this.organizationService.updateRegionQuota(organizationId, regionId, updateDto)
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
  @UseGuards(CombinedAuthGuard, AuthenticatedRateLimitGuard, SystemActionGuard)
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
  @UseGuards(CombinedAuthGuard, AuthenticatedRateLimitGuard, SystemActionGuard)
  @Audit({
    action: AuditAction.UNSUSPEND,
    targetType: AuditTarget.ORGANIZATION,
    targetIdFromRequest: (req) => req.params.organizationId,
  })
  async unsuspend(@Param('organizationId') organizationId: string): Promise<void> {
    return this.organizationService.unsuspend(organizationId)
  }

  @Get('/by-sandbox-id/:sandboxId')
  @ApiOperation({
    summary: 'Get organization by sandbox ID',
    operationId: 'getOrganizationBySandboxId',
  })
  @ApiResponse({
    status: 200,
    description: 'Organization',
    type: OrganizationDto,
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'Sandbox ID',
    type: 'string',
  })
  @RequiredApiRole([SystemRole.ADMIN, 'proxy'])
  @UseGuards(CombinedAuthGuard, AuthenticatedRateLimitGuard, SystemActionGuard)
  async getBySandboxId(@Param('sandboxId') sandboxId: string): Promise<OrganizationDto> {
    const organization = await this.organizationService.findBySandboxId(sandboxId)
    if (!organization) {
      throw new NotFoundException(`Organization with sandbox ID ${sandboxId} not found`)
    }

    return OrganizationDto.fromOrganization(organization)
  }

  @Get('/region-quota/by-sandbox-id/:sandboxId')
  @ApiOperation({
    summary: 'Get region quota by sandbox ID',
    operationId: 'getRegionQuotaBySandboxId',
  })
  @ApiResponse({
    status: 200,
    description: 'Region quota',
    type: RegionQuotaDto,
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'Sandbox ID',
    type: 'string',
  })
  @RequiredApiRole([SystemRole.ADMIN, 'proxy'])
  @UseGuards(CombinedAuthGuard, AuthenticatedRateLimitGuard, SystemActionGuard)
  async getRegionQuotaBySandboxId(@Param('sandboxId') sandboxId: string): Promise<RegionQuotaDto> {
    const regionQuota = await this.organizationService.getRegionQuotaBySandboxId(sandboxId)
    if (!regionQuota) {
      throw new NotFoundException(`Region quota for sandbox with ID ${sandboxId} not found`)
    }

    return regionQuota
  }

  @Get('/otel-config/by-sandbox-auth-token/:authToken')
  @ApiOperation({
    summary: 'Get organization OTEL config by sandbox auth token',
    operationId: 'getOrganizationOtelConfigBySandboxAuthToken',
  })
  @ApiResponse({
    status: 200,
    description: 'OTEL Config',
    type: OtelConfigDto,
  })
  @ApiParam({
    name: 'authToken',
    description: 'Sandbox Auth Token',
    type: 'string',
  })
  @RequiredApiRole([SystemRole.ADMIN, 'otel-proxy'])
  @UseGuards(CombinedAuthGuard, OrGuard([SystemActionGuard, OtelProxyGuard]))
  async getOtelConfigBySandboxAuthToken(@Param('authToken') authToken: string): Promise<OtelConfigDto> {
    const otelConfigDto = await this.organizationService.getOtelConfigBySandboxAuthToken(authToken)
    if (!otelConfigDto) {
      throw new NotFoundException(`Organization OTEL config with sandbox auth token ${authToken} not found`)
    }

    return otelConfigDto
  }

  @Post('/:organizationId/sandbox-default-limited-network-egress')
  @ApiOperation({
    summary: 'Update sandbox default limited network egress',
    operationId: 'updateSandboxDefaultLimitedNetworkEgress',
  })
  @ApiResponse({
    status: 204,
    description: 'Sandbox default limited network egress updated successfully',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @RequiredSystemRole(SystemRole.ADMIN)
  @UseGuards(CombinedAuthGuard, SystemActionGuard)
  @Audit({
    action: AuditAction.UPDATE_SANDBOX_DEFAULT_LIMITED_NETWORK_EGRESS,
    targetType: AuditTarget.ORGANIZATION,
    targetIdFromRequest: (req) => req.params.organizationId,
    requestMetadata: {
      body: (req: TypedRequest<OrganizationSandboxDefaultLimitedNetworkEgressDto>) => ({
        sandboxDefaultLimitedNetworkEgress: req.body?.sandboxDefaultLimitedNetworkEgress,
      }),
    },
  })
  async updateSandboxDefaultLimitedNetworkEgress(
    @Param('organizationId') organizationId: string,
    @Body() body: OrganizationSandboxDefaultLimitedNetworkEgressDto,
  ): Promise<void> {
    return this.organizationService.updateSandboxDefaultLimitedNetworkEgress(
      organizationId,
      body.sandboxDefaultLimitedNetworkEgress,
    )
  }

  @Put('/:organizationId/experimental-config')
  @ApiOperation({
    summary: 'Update experimental configuration',
    operationId: 'updateExperimentalConfig',
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiBody({
    description: 'Experimental configuration as a JSON object. Set to null to clear the configuration.',
    required: false,
    schema: {
      additionalProperties: true,
      example: {
        otel: {
          endpoint: 'http://otel-collector:4317',
          headers: {
            'api-key': 'XXX',
          },
        },
      },
    },
  })
  @RequiredOrganizationMemberRole(OrganizationMemberRole.OWNER)
  @UseGuards(AuthGuard('jwt'), OrganizationActionGuard)
  @RequireFlagsEnabled({ flags: [{ flagKey: 'organization_experiments', defaultValue: false }] })
  async updateExperimentalConfig(
    @Param('organizationId') organizationId: string,
    @Body() experimentalConfig: Record<string, any> | null,
  ): Promise<void> {
    await this.organizationService.updateExperimentalConfig(organizationId, experimentalConfig)
  }
}
