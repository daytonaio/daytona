/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, Get, Post, Param, Patch, UseGuards, Query } from '@nestjs/common'
import { CreateRunnerDto } from '../dto/create-runner.dto'
import { Runner } from '../entities/runner.entity'
import { RunnerService } from '../services/runner.service'
import { ApiOAuth2, ApiTags, ApiOperation, ApiBearerAuth, ApiResponse, ApiQuery } from '@nestjs/swagger'
import { SystemActionGuard } from '../../auth/system-action.guard'
import { RequiredApiRole } from '../../common/decorators/required-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'
import { ProxyGuard } from '../../auth/proxy.guard'
import { RunnerDto } from '../dto/runner.dto'
import { RunnerSnapshotDto } from '../dto/runner-snapshot.dto'

import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
@ApiTags('runners')
@Controller('runners')
@UseGuards(CombinedAuthGuard, SystemActionGuard, ProxyGuard)
@RequiredApiRole([SystemRole.ADMIN, 'proxy'])
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

  @Get('/by-sandbox/:sandboxId')
  @ApiOperation({
    summary: 'Get runner by sandbox ID',
    operationId: 'getRunnerBySandboxId',
  })
  @ApiResponse({
    status: 200,
    description: 'Runner found',
    type: RunnerDto,
  })
  async getRunnerBySandboxId(@Param('sandboxId') sandboxId: string): Promise<RunnerDto> {
    const runner = await this.runnerService.findBySandboxId(sandboxId)
    return RunnerDto.fromRunner(runner)
  }

  @Get('/by-snapshot')
  @ApiOperation({
    summary: 'Get runners by snapshot internal name',
    operationId: 'getRunnersBySnapshotRef',
  })
  @ApiQuery({
    name: 'ref',
    description: 'Internal name of the snapshot',
    type: String,
    required: true,
  })
  @ApiResponse({
    status: 200,
    description: 'Runners found for the snapshot',
    type: [RunnerSnapshotDto],
  })
  async getRunnersBySnapshotRef(@Query('ref') ref: string): Promise<RunnerSnapshotDto[]> {
    return this.runnerService.getRunnersBySnapshotRef(ref)
  }
}
