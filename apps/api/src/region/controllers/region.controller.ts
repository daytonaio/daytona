/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Get, Logger, UseGuards, HttpCode } from '@nestjs/common'
import { ApiOAuth2, ApiResponse, ApiOperation, ApiTags, ApiBearerAuth, ApiHeader } from '@nestjs/swagger'
import { RegionDto } from '../dto/region.dto'
import { RegionService } from '../services/region.service'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { CustomHeaders } from '../../common/constants/header.constants'
import { OrganizationResourceActionGuard } from '../../organization/guards/organization-resource-action.guard'
import { OrganizationService } from '../../organization/services/organization.service'

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
  async listRegions(@AuthContext() authContext: OrganizationAuthContext): Promise<RegionDto[]> {
    const [sharedRegions, organizationRegions, regionQuotas] = await Promise.all([
      this.regionService.findAll(null),
      this.regionService.findAll(authContext.organizationId),
      this.organizationService.getRegionQuotas(authContext.organizationId),
    ])

    const availableSharedRegions = sharedRegions.filter(
      (region) => !region.hidden || regionQuotas.some((quota) => quota.regionId === region.id),
    )

    const regions = [...availableSharedRegions, ...organizationRegions]
    return regions.map(RegionDto.fromRegion)
  }
}
