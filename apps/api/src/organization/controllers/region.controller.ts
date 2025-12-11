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
  UseInterceptors,
  Body,
  Param,
  NotFoundException,
  Delete,
} from '@nestjs/common'
import { ApiOAuth2, ApiResponse, ApiOperation, ApiTags, ApiBearerAuth, ApiHeader, ApiParam } from '@nestjs/swagger'
import { RequiredOrganizationResourcePermissions } from '../decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../enums/organization-resource-permission.enum'
import { OrganizationResourceActionGuard } from '../guards/organization-resource-action.guard'
import { OrganizationService } from '../services/organization.service'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { CustomHeaders } from '../../common/constants/header.constants'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { ContentTypeInterceptor } from '../../common/interceptors/content-type.interceptors'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { CreateRegionDto, CreateRegionResponseDto } from '../../region/dto/create-region.dto'
import { RegionDto } from '../../region/dto/region.dto'
import { RegionService } from '../../region/services/region.service'
import { RegionAccessGuard } from '../../region/guards/region-access.guard'
import { RegenerateApiKeyResponseDto } from '../../region/dto/regenerate-api-key.dto'
import { RegionType } from '../../region/enums/region-type.enum'

@ApiTags('regions')
@Controller('regions')
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@UseGuards(CombinedAuthGuard, OrganizationResourceActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class OrganizationRegionController {
  private readonly logger = new Logger(OrganizationRegionController.name)

  constructor(
    private readonly regionService: RegionService,
    private readonly organizationService: OrganizationService,
  ) {}

  @Get()
  @HttpCode(200)
  @ApiOperation({
    summary: 'List all regions',
    operationId: 'listRegions',
  })
  @ApiResponse({
    status: 200,
    description: 'List of all regions',
    type: [RegionDto],
  })
  async listRegions(@AuthContext() authContext: OrganizationAuthContext): Promise<RegionDto[]> {
    return this.organizationService.listRegions(authContext.organizationId)
  }

  @Post()
  @HttpCode(201)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Create a new region',
    operationId: 'createRegion',
  })
  @ApiResponse({
    status: 201,
    description: 'The region has been successfully created.',
    type: CreateRegionResponseDto,
  })
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
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_REGIONS])
  async createRegion(
    @AuthContext() authContext: OrganizationAuthContext,
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
  @ApiResponse({
    status: 200,
    type: RegionDto,
  })
  @ApiParam({
    name: 'id',
    description: 'Region ID',
    type: String,
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
  @ApiOperation({
    summary: 'Delete a region',
    operationId: 'deleteRegion',
  })
  @ApiResponse({
    status: 204,
    description: 'The region has been successfully deleted.',
  })
  @ApiParam({
    name: 'id',
    description: 'Region ID',
  })
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.REGION,
    targetIdFromRequest: (req) => req.params.id,
  })
  @UseGuards(RegionAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.DELETE_REGIONS])
  async deleteRegion(@Param('id') id: string): Promise<void> {
    await this.regionService.delete(id)
  }

  @Post(':id/regenerate-proxy-api-key')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Regenerate proxy API key for a region',
    operationId: 'regenerateProxyApiKey',
  })
  @ApiResponse({
    status: 200,
    description: 'The proxy API key has been successfully regenerated.',
    type: RegenerateApiKeyResponseDto,
  })
  @ApiParam({
    name: 'id',
    description: 'Region ID',
    type: String,
  })
  @Audit({
    action: AuditAction.REGENERATE_PROXY_API_KEY,
    targetType: AuditTarget.REGION,
    targetIdFromRequest: (req) => req.params.id,
  })
  @UseGuards(RegionAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_REGIONS])
  async regenerateProxyApiKey(@Param('id') id: string): Promise<RegenerateApiKeyResponseDto> {
    const apiKey = await this.regionService.regenerateProxyApiKey(id)
    return new RegenerateApiKeyResponseDto(apiKey)
  }

  @Post(':id/regenerate-ssh-gateway-api-key')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Regenerate SSH gateway API key for a region',
    operationId: 'regenerateSshGatewayApiKey',
  })
  @ApiResponse({
    status: 200,
    description: 'The SSH gateway API key has been successfully regenerated.',
    type: RegenerateApiKeyResponseDto,
  })
  @ApiParam({
    name: 'id',
    description: 'Region ID',
    type: String,
  })
  @Audit({
    action: AuditAction.REGENERATE_SSH_GATEWAY_API_KEY,
    targetType: AuditTarget.REGION,
    targetIdFromRequest: (req) => req.params.id,
  })
  @UseGuards(RegionAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_REGIONS])
  async regenerateSshGatewayApiKey(@Param('id') id: string): Promise<RegenerateApiKeyResponseDto> {
    const apiKey = await this.regionService.regenerateSshGatewayApiKey(id)
    return new RegenerateApiKeyResponseDto(apiKey)
  }
}
