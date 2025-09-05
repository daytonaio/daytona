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
} from '@nestjs/common'
import { ApiOAuth2, ApiResponse, ApiOperation, ApiParam, ApiTags, ApiHeader, ApiBearerAuth } from '@nestjs/swagger'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { RegionService } from '../services/region.service'
import { CreateRegionDto } from '../dto/create-region.dto'
import { ContentTypeInterceptor } from '../../common/interceptors/content-type.interceptors'
import { CustomHeaders } from '../../common/constants/header.constants'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { OrganizationResourceActionGuard } from '../../organization/guards/organization-resource-action.guard'
import { RegionDto } from '../dto/region.dto'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'

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
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Create a new region',
    operationId: 'createRegion',
  })
  @ApiResponse({
    status: 200,
    description: 'The region has been successfully created.',
    type: RegionDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_REGIONS])
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.REGION,
    targetIdFromResult: (result: RegionDto) => result?.code,
    requestMetadata: {
      body: (req: TypedRequest<CreateRegionDto>) => ({
        name: req.body?.name,
      }),
    },
  })
  async createRegion(
    @AuthContext() authContext: OrganizationAuthContext,
    @Body() createRegionDto: CreateRegionDto,
  ): Promise<RegionDto> {
    const region = await this.regionService.create(authContext.organization, createRegionDto)
    return RegionDto.fromRegion(region)
  }

  @Delete(':code')
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
    name: 'code',
    description: 'Region code',
    example: 'abc12345',
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_REGIONS])
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.REGION,
    targetIdFromRequest: (req) => req.params.code,
  })
  async deleteRegion(@Param('code') code: string): Promise<void> {
    await this.regionService.delete(code)
  }
}
