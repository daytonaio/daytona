/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, Get, Post, Param, UseGuards, Query, HttpCode, BadRequestException } from '@nestjs/common'
import { ApiOAuth2, ApiTags, ApiOperation, ApiBearerAuth, ApiResponse, ApiQuery } from '@nestjs/swagger'
import { AdminCreateRunnerDto } from '../dto/create-runner.dto'
import { Audit, MASKED_AUDIT_VALUE, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { ProxyGuard } from '../../auth/proxy.guard'
import { SystemActionGuard } from '../../auth/system-action.guard'
import { RequiredApiRole } from '../../common/decorators/required-role.decorator'
import { RunnerDto } from '../../sandbox/dto/runner.dto'
import { RunnerSnapshotDto } from '../../sandbox/dto/runner-snapshot.dto'
import { RunnerService } from '../../sandbox/services/runner.service'
import { SystemRole } from '../../user/enums/system-role.enum'
import { OrGuard } from '../../auth/or.guard'
import { SshGatewayGuard } from '../../auth/ssh-gateway.guard'

@ApiTags('admin/runners')
@Controller('admin/runners')
@UseGuards(CombinedAuthGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class AdminRunnerController {
  constructor(private readonly runnerService: RunnerService) {}

  @Post()
  @HttpCode(201)
  @ApiOperation({
    summary: 'Create runner',
    operationId: 'adminCreateRunner',
  })
  @ApiResponse({
    status: 201,
    type: RunnerDto,
  })
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.RUNNER,
    targetIdFromResult: (result: RunnerDto) => result?.id,
    requestMetadata: {
      body: (req: TypedRequest<AdminCreateRunnerDto>) => ({
        domain: req.body?.domain,
        apiUrl: req.body?.apiUrl,
        apiKey: MASKED_AUDIT_VALUE,
        cpu: req.body?.cpu,
        memoryGiB: req.body?.memoryGiB,
        diskGiB: req.body?.diskGiB,
        gpu: req.body?.gpu,
        gpuType: req.body?.gpuType,
        class: req.body?.class,
        regionId: req.body?.regionId,
      }),
    },
  })
  @UseGuards(SystemActionGuard)
  @RequiredApiRole([SystemRole.ADMIN])
  async create(@Body() createRunnerDto: AdminCreateRunnerDto): Promise<RunnerDto> {
    const runner = await this.runnerService.create(createRunnerDto)
    return RunnerDto.fromRunner(runner)
  }

  @Get()
  @HttpCode(200)
  @ApiOperation({
    summary: 'List all runners',
    operationId: 'adminListRunners',
  })
  @ApiResponse({
    status: 200,
    type: [RunnerDto],
  })
  @ApiQuery({
    name: 'organizationId',
    description: 'Filter runners by organization ID',
    type: String,
    required: false,
  })
  @ApiQuery({
    name: 'region',
    description: 'Filter runners by region name (organization ID is required)',
    type: String,
    required: false,
  })
  @UseGuards(SystemActionGuard)
  @RequiredApiRole([SystemRole.ADMIN])
  async findAll(
    @Query('organizationId') organizationId?: string,
    @Query('region') region?: string,
  ): Promise<RunnerDto[]> {
    if (!organizationId && region) {
      throw new BadRequestException('Must provide organization ID when filtering by region name')
    }
    const runners = await this.runnerService.findAll(organizationId, region)
    return runners.map(RunnerDto.fromRunner)
  }

  @Get('/by-sandbox/:sandboxId')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Get runner by sandbox ID',
    operationId: 'getRunnerBySandboxId',
  })
  @ApiResponse({
    status: 200,
    type: RunnerDto,
  })
  @UseGuards(OrGuard([SystemActionGuard, ProxyGuard, SshGatewayGuard]))
  @RequiredApiRole([SystemRole.ADMIN, 'proxy', 'ssh-gateway'])
  async getRunnerBySandboxId(@Param('sandboxId') sandboxId: string): Promise<RunnerDto> {
    const runner = await this.runnerService.findBySandboxId(sandboxId)
    return RunnerDto.fromRunner(runner)
  }

  @Get('/by-snapshot-ref')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Get runners by snapshot ref',
    operationId: 'getRunnersBySnapshotRef',
  })
  @ApiResponse({
    status: 200,
    type: [RunnerSnapshotDto],
  })
  @ApiQuery({
    name: 'ref',
    description: 'Snapshot ref',
    type: String,
    required: true,
  })
  @UseGuards(SystemActionGuard)
  @RequiredApiRole([SystemRole.ADMIN])
  async getRunnersBySnapshotRef(@Query('ref') ref: string): Promise<RunnerSnapshotDto[]> {
    return this.runnerService.getRunnersBySnapshotRef(ref)
  }
}
