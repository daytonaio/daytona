/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Controller,
  Get,
  Post,
  Delete,
  Body,
  Param,
  Logger,
  UseGuards,
  HttpCode,
  UseInterceptors,
  NotFoundException,
} from '@nestjs/common'
import { ApiOAuth2, ApiResponse, ApiOperation, ApiParam, ApiTags, ApiHeader, ApiBearerAuth } from '@nestjs/swagger'
import { CreateRegionDto } from '../dto/create-region.dto'
import { RegionDto } from '../dto/region.dto'
import { RegionAccessGuard } from '../guards/region-access.guard'
import { RegionService } from '../services/region.service'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { CustomHeaders } from '../../common/constants/header.constants'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { ContentTypeInterceptor } from '../../common/interceptors/content-type.interceptors'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { OrganizationResourceActionGuard } from '../../organization/guards/organization-resource-action.guard'

@ApiTags('regions')
@Controller('regions')
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@UseGuards(CombinedAuthGuard, OrganizationResourceActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class RegionController {
  private readonly logger = new Logger(RegionController.name)

  constructor(private readonly regionService: RegionService) {}

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
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.READ_REGIONS])
  async listRegions(@AuthContext() authContext: OrganizationAuthContext): Promise<RegionDto[]> {
    const regions = await this.regionService.findAll(authContext.organizationId)
    return regions.map(RegionDto.fromRegion)
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
    type: RegionDto,
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
  ): Promise<RegionDto> {
    const region = await this.regionService.create(createRegionDto, authContext.organization)
    return RegionDto.fromRegion(region)
  }

  @Get('/by-name/:name')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Get region by name',
    operationId: 'getRegionByName',
  })
  @ApiResponse({
    status: 200,
    type: RegionDto,
  })
  @ApiParam({
    name: 'name',
    description: 'Region name',
    type: String,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.READ_REGIONS])
  async getRegionByName(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('name') name: string,
  ): Promise<RegionDto> {
    const region = await this.regionService.findOneByName(name, authContext.organizationId)
    if (!region) {
      throw new NotFoundException('Region not found')
    }
    return RegionDto.fromRegion(region)
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
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.READ_REGIONS])
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
}
