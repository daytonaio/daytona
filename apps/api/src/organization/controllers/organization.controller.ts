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
import { OrganizationUsageOverviewDto } from '../dto/organization-usage-overview.dto'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { OrganizationActionGuard } from '../guards/organization-action.guard'
import { OrganizationService } from '../services/organization.service'
import { OrganizationUserService } from '../services/organization-user.service'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { AuthContext as IAuthContext } from '../../common/interfaces/auth-context.interface'
import { SystemActionGuard } from '../../auth/system-action.guard'
import { RequiredApiRole, RequiredSystemRole } from '../../common/decorators/required-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { UserService } from '../../user/user.service'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { OrganizationUsageService } from '../services/organization-usage.service'
import { OrganizationSandboxDefaultLimitedNetworkEgressDto } from '../dto/organization-sandbox-default-limited-network-egress.dto'
import { TypedConfigService } from '../../config/typed-config.service'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'
import { UpdateOrganizationDefaultRegionDto } from '../dto/update-organization-default-region.dto'
import { RegionQuotaDto } from '../dto/region-quota.dto'
import { RequireFlagsEnabled } from '@openfeature/nestjs-sdk'

@ApiTags('organizations')
@Controller('organizations')
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class OrganizationController {
  constructor(
    private readonly organizationService: OrganizationService,
    private readonly organizationUserService: OrganizationUserService,
    private readonly organizationUsageService: OrganizationUsageService,
    private readonly userService: UserService,
    private readonly configService: TypedConfigService,
  ) {}

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
    },
  })
  @RequiredOrganizationMemberRole(OrganizationMemberRole.OWNER)
  @UseGuards(AuthGuard('jwt'), OrganizationActionGuard)
  @RequireFlagsEnabled({ flags: [{ flagKey: 'organization_experiments', defaultValue: true }] })
  async updateExperimentalConfig(
    @Param('organizationId') organizationId: string,
    @Body() experimentalConfig: Record<string, any> | null,
  ): Promise<void> {
    await this.organizationService.updateExperimentalConfig(organizationId, experimentalConfig)
  }
}
