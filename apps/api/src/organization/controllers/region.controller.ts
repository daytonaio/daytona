/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Get, Logger, UseGuards, HttpCode, Query } from '@nestjs/common'
import { ApiOAuth2, ApiResponse, ApiOperation, ApiTags, ApiBearerAuth, ApiHeader, ApiQuery } from '@nestjs/swagger'
import { OrganizationResourceActionGuard } from '../guards/organization-resource-action.guard'
import { OrganizationService } from '../services/organization.service'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { CustomHeaders } from '../../common/constants/header.constants'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { RegionDto } from '../../region/dto/region.dto'
import { RegionService } from '../../region/services/region.service'
import { Region } from '../../region/entities/region.entity'

@ApiTags('regions')
@Controller('regions')
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@UseGuards(CombinedAuthGuard, OrganizationResourceActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class RegionController {
  private readonly logger = new Logger(RegionController.name)

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
  @ApiQuery({
    name: 'includeShared',
    required: false,
    type: Boolean,
    description: 'Include shared regions',
  })
  async listRegions(
    @AuthContext() authContext: OrganizationAuthContext,
    @Query('includeShared') includeShared?: boolean,
  ): Promise<RegionDto[]> {
    const regions: Region[] = []

    if (includeShared) {
      const sharedRegions = await this.regionService.findAll(null)
      const regionQuotas = await this.organizationService.getRegionQuotas(authContext.organizationId)
      const availableSharedRegions = sharedRegions.filter(
        (region) => !region.hidden || regionQuotas.some((quota) => quota.regionId === region.id),
      )
      regions.push(...availableSharedRegions)
    }

    const organizationRegions = await this.regionService.findAll(authContext.organizationId)
    regions.push(...organizationRegions)

    return regions.map(RegionDto.fromRegion)
  }
}
