/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Body,
  Controller,
  Delete,
  Get,
  HttpCode,
  NotFoundException,
  Param,
  ParseUUIDPipe,
  Patch,
  Post,
  Query,
  UseGuards,
} from '@nestjs/common'
import { ApiBearerAuth, ApiOAuth2, ApiOperation, ApiParam, ApiQuery, ApiResponse, ApiTags } from '@nestjs/swagger'
import { AdminCreateRunnerDto } from '../dto/create-runner.dto'
import { Audit, MASKED_AUDIT_VALUE, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { SystemActionGuard } from '../../auth/system-action.guard'
import { RequiredApiRole } from '../../common/decorators/required-role.decorator'
import { RegionService } from '../../region/services/region.service'
import { CreateRunnerResponseDto } from '../../sandbox/dto/create-runner-response.dto'
import { RunnerFullDto } from '../../sandbox/dto/runner-full.dto'
import { RunnerDto } from '../../sandbox/dto/runner.dto'
import { RunnerService } from '../../sandbox/services/runner.service'
import { SystemRole } from '../../user/enums/system-role.enum'

@ApiTags('admin')
@Controller('admin/runners')
@UseGuards(CombinedAuthGuard, SystemActionGuard)
@RequiredApiRole([SystemRole.ADMIN])
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class AdminRunnerController {
  constructor(
    private readonly runnerService: RunnerService,
    private readonly regionService: RegionService,
  ) {}

  @Post()
  @HttpCode(201)
  @ApiOperation({
    summary: 'Create runner',
    operationId: 'adminCreateRunner',
  })
  @ApiResponse({
    status: 201,
    type: CreateRunnerResponseDto,
  })
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.RUNNER,
    targetIdFromResult: (result: RunnerDto) => result?.id,
    requestMetadata: {
      body: (req: TypedRequest<AdminCreateRunnerDto>) => ({
        domain: req.body?.domain,
        apiUrl: req.body?.apiUrl,
        proxyUrl: req.body?.proxyUrl,
        regionId: req.body?.regionId,
        name: req.body?.name,
        apiKey: MASKED_AUDIT_VALUE,
        apiVersion: req.body?.apiVersion,
      }),
    },
  })
  async create(@Body() createRunnerDto: AdminCreateRunnerDto): Promise<CreateRunnerResponseDto> {
    const region = await this.regionService.findOne(createRunnerDto.regionId)

    if (!region) {
      throw new NotFoundException('Region not found')
    }

    const { runner, apiKey } = await this.runnerService.create({
      domain: createRunnerDto.domain,
      apiUrl: createRunnerDto.apiUrl,
      proxyUrl: createRunnerDto.proxyUrl,
      regionId: createRunnerDto.regionId,
      name: createRunnerDto.name,
      apiKey: createRunnerDto.apiKey,
      apiVersion: createRunnerDto.apiVersion,
      cpu: createRunnerDto.cpu,
      memoryGiB: createRunnerDto.memoryGiB,
      diskGiB: createRunnerDto.diskGiB,
    })

    return CreateRunnerResponseDto.fromRunner(runner, apiKey)
  }

  @Get(':id')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Get runner by ID',
    operationId: 'adminGetRunnerById',
  })
  @ApiResponse({
    status: 200,
    type: RunnerFullDto,
  })
  @ApiParam({
    name: 'id',
    description: 'Runner ID',
    type: String,
  })
  async getRunnerById(@Param('id', ParseUUIDPipe) id: string): Promise<RunnerFullDto> {
    return this.runnerService.findOneFullOrFail(id)
  }

  @Get()
  @HttpCode(200)
  @ApiOperation({
    summary: 'List all runners',
    operationId: 'adminListRunners',
  })
  @ApiResponse({
    status: 200,
    type: [RunnerFullDto],
  })
  @ApiQuery({
    name: 'regionId',
    description: 'Filter runners by region ID',
    type: String,
    required: false,
  })
  async findAll(@Query('regionId') regionId?: string): Promise<RunnerFullDto[]> {
    if (regionId) {
      return this.runnerService.findAllByRegionFull(regionId)
    }
    return this.runnerService.findAllFull()
  }

  @Patch(':id/scheduling')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Update runner scheduling status',
    operationId: 'adminUpdateRunnerScheduling',
  })
  @ApiResponse({
    status: 204,
  })
  @Audit({
    action: AuditAction.UPDATE_SCHEDULING,
    targetType: AuditTarget.RUNNER,
    targetIdFromRequest: (req) => req.params.id,
    requestMetadata: {
      body: (req: TypedRequest<{ unschedulable: boolean }>) => ({
        unschedulable: req.body?.unschedulable,
      }),
    },
  })
  async updateSchedulingStatus(
    @Param('id', ParseUUIDPipe) id: string,
    @Body('unschedulable') unschedulable: boolean,
  ): Promise<void> {
    await this.runnerService.updateSchedulingStatus(id, unschedulable)
  }

  @Delete(':id')
  @HttpCode(204)
  @ApiOperation({
    summary: 'Delete runner',
    operationId: 'adminDeleteRunner',
  })
  @ApiResponse({
    status: 204,
  })
  @ApiParam({
    name: 'id',
    description: 'Runner ID',
    type: String,
  })
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.RUNNER,
    targetIdFromRequest: (req) => req.params.id,
  })
  async delete(@Param('id', ParseUUIDPipe) id: string): Promise<void> {
    return this.runnerService.remove(id)
  }
}
