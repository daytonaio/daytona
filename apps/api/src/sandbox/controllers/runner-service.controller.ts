/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Post, Body, UseGuards } from '@nestjs/common'
import { ApiTags, ApiOperation, ApiBearerAuth, ApiResponse } from '@nestjs/swagger'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { RunnerAuthGuard } from '../../auth/runner-auth.guard'
import { RunnerContextDecorator } from '../../common/decorators/runner-context.decorator'
import { RunnerContext } from '../../common/interfaces/runner-context.interface'
import { RequiredApiRole } from '../../common/decorators/required-role.decorator'
import { RunnerService } from '../services/runner.service'
import { RunnerHealthcheckDto } from '../dto/runner-health.dto'

@ApiTags('runner-service')
@Controller('runner-service')
@UseGuards(CombinedAuthGuard, RunnerAuthGuard)
@RequiredApiRole(['runner'])
@ApiBearerAuth()
export class RunnerServiceController {
  constructor(private readonly runnerService: RunnerService) {}

  @Post('healthcheck')
  @ApiOperation({
    summary: 'Runner healthcheck',
    operationId: 'runnerHealthcheck',
    description:
      'Endpoint for version 2 runners to send healthcheck and metrics. Updates lastChecked timestamp and runner metrics.',
  })
  @ApiResponse({
    status: 200,
    description: 'Healthcheck received',
  })
  async runnerHealthcheck(
    @RunnerContextDecorator() runnerContext: RunnerContext,
    @Body() healthcheck: RunnerHealthcheckDto,
  ): Promise<void> {
    await this.runnerService.updateRunnerHealth(
      runnerContext.runnerId,
      healthcheck.domain,
      healthcheck.proxyUrl,
      healthcheck.metrics,
    )
  }
}
