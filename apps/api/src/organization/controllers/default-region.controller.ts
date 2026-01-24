/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Get, Logger, UseGuards, HttpCode, NotFoundException } from '@nestjs/common'
import { ApiOAuth2, ApiResponse, ApiOperation, ApiTags, ApiBearerAuth, ApiHeader } from '@nestjs/swagger'
import { OrganizationResourceActionGuard } from '../guards/organization-resource-action.guard'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { CustomHeaders } from '../../common/constants/header.constants'
import { RegionDto } from '../../region/dto/region.dto'
import { RegionService } from '../../region/services/region.service'
import { configuration } from '../../config/configuration'

@ApiTags('regions')
@Controller('region')
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@UseGuards(CombinedAuthGuard, OrganizationResourceActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class DefaultRegionController {
    private readonly logger = new Logger(DefaultRegionController.name)

    constructor(private readonly regionService: RegionService) { }

    @Get()
    @HttpCode(200)
    @ApiOperation({
        summary: 'Get default region',
        operationId: 'getDefaultRegion',
    })
    @ApiResponse({
        status: 200,
        description: 'The default region',
        type: RegionDto,
    })
    async getDefaultRegion(): Promise<RegionDto> {
        const defaultRegionId = configuration.defaultRegion.id || 'us'
        const region = await this.regionService.findOne(defaultRegionId)

        if (!region) {
            throw new NotFoundException(`Default region ${defaultRegionId} not found`)
        }

        return RegionDto.fromRegion(region)
    }
}
