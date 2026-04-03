/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Get, UseGuards } from '@nestjs/common'
import { TypedConfigService } from './typed-config.service'
import { ApiOperation, ApiTags, ApiResponse } from '@nestjs/swagger'
import { ConfigurationDto } from './dto/configuration.dto'
import { Public } from '../auth/decorators/public.decorator'
import { AnonymousRateLimitGuard } from '../common/guards/anonymous-rate-limit.guard'

@ApiTags('config')
@Controller('config')
export class ConfigController {
  constructor(private readonly configService: TypedConfigService) {}

  @Get()
  @Public()
  @UseGuards(AnonymousRateLimitGuard)
  @ApiOperation({ summary: 'Get config' })
  @ApiResponse({
    status: 200,
    description: 'Daytona configuration',
    type: ConfigurationDto,
  })
  getConfig() {
    return new ConfigurationDto(this.configService)
  }
}
