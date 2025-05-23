/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, Get, Post, Param, Patch, UseGuards } from '@nestjs/common'
import { CreateRunnerDto } from '../dto/create-runner.dto'
import { Runner } from '../entities/runner.entity'
import { RunnerService } from '../services/runner.service'
import { AuthGuard } from '@nestjs/passport'
import { ApiOAuth2, ApiTags, ApiOperation, ApiBearerAuth } from '@nestjs/swagger'
import { SystemActionGuard } from '../../auth/system-action.guard'
import { RequiredSystemRole } from '../../common/decorators/required-system-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'

@ApiTags('runners')
@Controller('runners')
@UseGuards(AuthGuard('jwt'), SystemActionGuard)
@RequiredSystemRole(SystemRole.ADMIN)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class RunnerController {
  constructor(private readonly runnerService: RunnerService) {}

  @Post()
  @ApiOperation({
    summary: 'Create runner',
    operationId: 'createRunner',
  })
  async create(@Body() createRunnerDto: CreateRunnerDto): Promise<Runner> {
    return this.runnerService.create(createRunnerDto)
  }

  @Get()
  @ApiOperation({
    summary: 'List all runners',
    operationId: 'listRunners',
  })
  async findAll(): Promise<Runner[]> {
    return this.runnerService.findAll()
  }

  @Patch(':id/scheduling')
  @ApiOperation({
    summary: 'Update runner scheduling status',
    operationId: 'updateRunnerScheduling',
  })
  async updateSchedulingStatus(
    @Param('id') id: string,
    @Body('unschedulable') unschedulable: boolean,
  ): Promise<Runner> {
    return this.runnerService.updateSchedulingStatus(id, unschedulable)
  }
}
