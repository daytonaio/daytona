/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Get, Logger, HttpCode } from '@nestjs/common'
import { ApiOAuth2, ApiResponse, ApiOperation, ApiTags, ApiBearerAuth } from '@nestjs/swagger'
import { RegionDto } from '../dto/region.dto'
import { RegionType } from '../enums/region-type.enum'
import { RegionService } from '../services/region.service'

@ApiTags('regions')
@Controller('shared-regions')
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class RegionController {
  private readonly logger = new Logger(RegionController.name)

  constructor(private readonly regionService: RegionService) {}

  @Get()
  @HttpCode(200)
  @ApiOperation({
    summary: 'List all shared regions',
    operationId: 'listSharedRegions',
  })
  @ApiResponse({
    status: 200,
    description: 'List of all shared regions',
    type: [RegionDto],
  })
  async listRegions(): Promise<RegionDto[]> {
    return this.regionService.findAllByRegionType(RegionType.SHARED)
  }
}
