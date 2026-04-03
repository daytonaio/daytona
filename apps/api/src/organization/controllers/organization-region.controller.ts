/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Controller,
  Get,
  Logger,
  UseGuards,
  HttpCode,
  Post,
  Body,
  Param,
  NotFoundException,
  Delete,
  Patch,
} from '@nestjs/common'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'
import { ApiOAuth2, ApiResponse, ApiOperation, ApiTags, ApiBearerAuth, ApiParam, ApiHeader } from '@nestjs/swagger'
import { RequiredOrganizationResourcePermissions } from '../decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../enums/organization-resource-permission.enum'
import { OrganizationAuthContextGuard } from '../guards/organization-auth-context.guard'
import { OrganizationService } from '../services/organization.service'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { CreateRegionDto, CreateRegionResponseDto } from '../../region/dto/create-region.dto'
import { RegionDto } from '../../region/dto/region.dto'
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import { RegionService } from '../../region/services/region.service'
import { RegionAccessGuard } from '../../region/guards/region-access.guard'
import { RegenerateApiKeyResponseDto } from '../../region/dto/regenerate-api-key.dto'
import { RegionType } from '../../region/enums/region-type.enum'
import { RequireFlagsEnabled } from '@openfeature/nestjs-sdk'
import { FeatureFlags } from '../../common/constants/feature-flags'
import { CustomHeaders } from '../../common/constants/header.constants'
import { IsOrganizationAuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { SnapshotManagerCredentialsDto } from '../../region/dto/snapshot-manager-credentials.dto'
import { UpdateRegionDto } from '../../region/dto/update-region.dto'

@Controller('regions')
@ApiTags('organizations')
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@AuthStrategy([AuthStrategyType.API_KEY, AuthStrategyType.JWT])
@UseGuards(AuthenticatedRateLimitGuard)
@UseGuards(OrganizationAuthContextGuard)
export class OrganizationRegionController {
  private readonly logger = new Logger(OrganizationRegionController.name)

  constructor(
    private readonly regionService: RegionService,
    private readonly organizationService: OrganizationService,
  ) {}

  @Get()
  @HttpCode(200)
  @ApiOperation({
    summary: 'List all available regions for the organization',
    operationId: 'listAvailableRegions',
  })
  @ApiResponse({
    status: 200,
    description: 'List of all available regions',
    type: [RegionDto],
  })
  async listAvailableRegions(@IsOrganizationAuthContext() authContext: OrganizationAuthContext): Promise<RegionDto[]> {
    return this.organizationService.listAvailableRegions(authContext.organizationId)
  }

  @Post()
  @HttpCode(201)
  @RequireFlagsEnabled({ flags: [{ flagKey: FeatureFlags.ORGANIZATION_INFRASTRUCTURE, defaultValue: false }] })
  @ApiOperation({
    summary: 'Create a new region',
    operationId: 'createRegion',
  })
  @ApiResponse({
    status: 201,
    description: 'The region has been successfully created.',
    type: CreateRegionResponseDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_REGIONS])
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.REGION,
    targetIdFromResult: (result: RegionDto) => result?.id,
    requestMetadata: {
      body: (req: TypedRequest<CreateRegionDto>) => ({
        name: req.body?.name,
      }),
    },
  })
  async createRegion(
    @IsOrganizationAuthContext() authContext: OrganizationAuthContext,
    @Body() createRegionDto: CreateRegionDto,
  ): Promise<CreateRegionResponseDto> {
    return await this.regionService.create(
      {
        ...createRegionDto,
        enforceQuotas: false,
        regionType: RegionType.CUSTOM,
      },
      authContext.organizationId,
    )
  }

  @Get(':id')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Get region by ID',
    operationId: 'getRegionById',
  })
  @ApiParam({
    name: 'id',
    description: 'Region ID',
    type: String,
  })
  @ApiResponse({
    status: 200,
    type: RegionDto,
  })
  @UseGuards(RegionAccessGuard)
  async getRegionById(@Param('id') id: string): Promise<RegionDto> {
    const region = await this.regionService.findOne(id)
    if (!region) {
      throw new NotFoundException('Region not found')
    }
    return RegionDto.fromRegion(region)
  }

  @Delete(':id')
  @HttpCode(204)
  @RequireFlagsEnabled({ flags: [{ flagKey: FeatureFlags.ORGANIZATION_INFRASTRUCTURE, defaultValue: false }] })
  @ApiOperation({
    summary: 'Delete a region',
    operationId: 'deleteRegion',
  })
  @ApiParam({
    name: 'id',
    description: 'Region ID',
  })
  @ApiResponse({
    status: 204,
    description: 'The region has been successfully deleted.',
  })
  @UseGuards(RegionAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.DELETE_REGIONS])
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.REGION,
    targetIdFromRequest: (req) => req.params.id,
  })
  async deleteRegion(@Param('id') id: string): Promise<void> {
    await this.regionService.delete(id)
  }

  @Post(':id/regenerate-proxy-api-key')
  @HttpCode(200)
  @RequireFlagsEnabled({ flags: [{ flagKey: FeatureFlags.ORGANIZATION_INFRASTRUCTURE, defaultValue: false }] })
  @ApiOperation({
    summary: 'Regenerate proxy API key for a region',
    operationId: 'regenerateProxyApiKey',
  })
  @ApiParam({
    name: 'id',
    description: 'Region ID',
    type: String,
  })
  @ApiResponse({
    status: 200,
    description: 'The proxy API key has been successfully regenerated.',
    type: RegenerateApiKeyResponseDto,
  })
  @UseGuards(RegionAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_REGIONS])
  @Audit({
    action: AuditAction.REGENERATE_PROXY_API_KEY,
    targetType: AuditTarget.REGION,
    targetIdFromRequest: (req) => req.params.id,
  })
  async regenerateProxyApiKey(@Param('id') id: string): Promise<RegenerateApiKeyResponseDto> {
    const apiKey = await this.regionService.regenerateProxyApiKey(id)
    return new RegenerateApiKeyResponseDto(apiKey)
  }

  @Patch(':id')
  @HttpCode(200)
  @RequireFlagsEnabled({ flags: [{ flagKey: FeatureFlags.ORGANIZATION_INFRASTRUCTURE, defaultValue: false }] })
  @ApiOperation({
    summary: 'Update region configuration',
    operationId: 'updateRegion',
  })
  @ApiParam({
    name: 'id',
    description: 'Region ID',
    type: String,
  })
  @UseGuards(RegionAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_REGIONS])
  @Audit({
    action: AuditAction.UPDATE,
    targetType: AuditTarget.REGION,
    targetIdFromRequest: (req) => req.params.id,
    requestMetadata: {
      body: (req: TypedRequest<UpdateRegionDto>) => ({
        ...req.body,
      }),
    },
  })
  async updateRegion(@Param('id') id: string, @Body() updateRegionDto: UpdateRegionDto): Promise<void> {
    return await this.regionService.update(id, updateRegionDto)
  }

  @Post(':id/regenerate-ssh-gateway-api-key')
  @HttpCode(200)
  @RequireFlagsEnabled({ flags: [{ flagKey: FeatureFlags.ORGANIZATION_INFRASTRUCTURE, defaultValue: false }] })
  @ApiOperation({
    summary: 'Regenerate SSH gateway API key for a region',
    operationId: 'regenerateSshGatewayApiKey',
  })
  @ApiParam({
    name: 'id',
    description: 'Region ID',
    type: String,
  })
  @ApiResponse({
    status: 200,
    description: 'The SSH gateway API key has been successfully regenerated.',
    type: RegenerateApiKeyResponseDto,
  })
  @UseGuards(RegionAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_REGIONS])
  @Audit({
    action: AuditAction.REGENERATE_SSH_GATEWAY_API_KEY,
    targetType: AuditTarget.REGION,
    targetIdFromRequest: (req) => req.params.id,
  })
  async regenerateSshGatewayApiKey(@Param('id') id: string): Promise<RegenerateApiKeyResponseDto> {
    const apiKey = await this.regionService.regenerateSshGatewayApiKey(id)
    return new RegenerateApiKeyResponseDto(apiKey)
  }

  @Post(':id/regenerate-snapshot-manager-credentials')
  @HttpCode(200)
  @RequireFlagsEnabled({ flags: [{ flagKey: FeatureFlags.ORGANIZATION_INFRASTRUCTURE, defaultValue: false }] })
  @ApiOperation({
    summary: 'Regenerate snapshot manager credentials for a region',
    operationId: 'regenerateSnapshotManagerCredentials',
  })
  @ApiParam({
    name: 'id',
    description: 'Region ID',
    type: String,
  })
  @ApiResponse({
    status: 200,
    description: 'The snapshot manager credentials have been successfully regenerated.',
    type: SnapshotManagerCredentialsDto,
  })
  @UseGuards(RegionAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_REGIONS])
  @Audit({
    action: AuditAction.REGENERATE_SNAPSHOT_MANAGER_CREDENTIALS,
    targetType: AuditTarget.REGION,
    targetIdFromRequest: (req) => req.params.id,
  })
  async regenerateSnapshotManagerCredentials(@Param('id') id: string): Promise<SnapshotManagerCredentialsDto> {
    return await this.regionService.regenerateSnapshotManagerCredentials(id)
  }
}
