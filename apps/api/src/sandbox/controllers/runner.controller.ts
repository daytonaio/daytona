/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, Get, Post, Param, Patch, UseGuards, Query, Delete, HttpCode } from '@nestjs/common'
import { CreateRunnerDto } from '../dto/create-runner.dto'
import { RunnerService } from '../services/runner.service'
import { ApiOAuth2, ApiTags, ApiOperation, ApiBearerAuth, ApiResponse, ApiQuery, ApiParam } from '@nestjs/swagger'
import { RunnerDto } from '../dto/runner.dto'
import { Audit, MASKED_AUDIT_VALUE, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { OrganizationResourceActionGuard } from '../../organization/guards/organization-resource-action.guard'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { RunnerAccessGuard } from '../guards/runner-access.guard'
import { RunnerAuthGuard } from '../../auth/runner-auth.guard'
import { RunnerContextDecorator } from '../../common/decorators/runner-context.decorator'
import { RunnerContext } from '../../common/interfaces/runner-context.interface'

@ApiTags('runners')
@Controller('runners')
@UseGuards(CombinedAuthGuard, OrganizationResourceActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class RunnerController {
  constructor(private readonly runnerService: RunnerService) {}

  @Post()
  @HttpCode(201)
  @ApiOperation({
    summary: 'Create runner',
    operationId: 'createRunner',
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
      body: (req: TypedRequest<CreateRunnerDto>) => ({
        domain: req.body?.domain,
        apiUrl: req.body?.apiUrl,
        apiKey: MASKED_AUDIT_VALUE,
        cpu: req.body?.cpu,
        memoryGiB: req.body?.memoryGiB,
        diskGiB: req.body?.diskGiB,
        regionId: req.body?.regionId,
      }),
    },
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_RUNNERS])
  async create(
    @Body() createRunnerDto: CreateRunnerDto,
    @AuthContext() authContext: OrganizationAuthContext,
  ): Promise<RunnerDto> {
    const runner = await this.runnerService.create(
      {
        domain: createRunnerDto.domain,
        apiUrl: createRunnerDto.apiUrl,
        proxyUrl: createRunnerDto.proxyUrl,
        apiKey: createRunnerDto.apiKey,
        cpu: createRunnerDto.cpu,
        memoryGiB: createRunnerDto.memoryGiB,
        diskGiB: createRunnerDto.diskGiB,
        gpu: 0,
        gpuType: '',
        class: SandboxClass.SMALL,
        regionId: createRunnerDto.regionId,
        version: '0',
      },
      authContext.organization,
    )
    return RunnerDto.fromRunner(runner)
  }

  @Get('/me')
  @UseGuards(RunnerAuthGuard)
  @ApiOperation({
    summary: 'Get info for authenticated runner',
    operationId: 'getInfoForAuthenticatedRunner',
  })
  @ApiResponse({
    status: 200,
    description: 'Runner info',
    type: RunnerDto,
  })
  async getInfoForAuthenticatedRunner(@RunnerContextDecorator() runnerContext: RunnerContext): Promise<RunnerDto> {
    return RunnerDto.fromRunner(runnerContext.runner)
  }

  @Get()
  @HttpCode(200)
  @ApiOperation({
    summary: 'List all runners',
    operationId: 'listRunners',
  })
  @ApiResponse({
    status: 200,
    type: [RunnerDto],
  })
  @ApiQuery({
    name: 'region',
    description: 'Filter runners by region name',
    type: String,
    required: false,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.READ_RUNNERS])
  async findAll(
    @AuthContext() authContext: OrganizationAuthContext,
    @Query('region') region?: string,
  ): Promise<RunnerDto[]> {
    const runners = await this.runnerService.findAll(authContext.organizationId, region)
    return runners.map(RunnerDto.fromRunner)
  }

  @Get(':id')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Get runner by ID',
    operationId: 'getRunnerById',
  })
  @ApiResponse({
    status: 200,
    type: RunnerDto,
  })
  @ApiParam({
    name: 'id',
    description: 'Runner ID',
    type: String,
  })
  @UseGuards(RunnerAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.READ_RUNNERS])
  async getRunnerById(@Param('id') id: string): Promise<RunnerDto> {
    const runner = await this.runnerService.findOne(id)
    return RunnerDto.fromRunner(runner)
  }

  @Patch(':id/scheduling')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Update runner scheduling status',
    operationId: 'updateRunnerScheduling',
  })
  @ApiResponse({
    status: 200,
    type: RunnerDto,
  })
  @ApiParam({
    name: 'id',
    description: 'Runner ID',
    type: String,
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
  @UseGuards(RunnerAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_RUNNERS])
  async updateSchedulingStatus(
    @Param('id') id: string,
    @Body('unschedulable') unschedulable: boolean,
  ): Promise<RunnerDto> {
    const updatedRunner = await this.runnerService.updateSchedulingStatus(id, unschedulable)
    return RunnerDto.fromRunner(updatedRunner)
  }

  @Delete(':id')
  @HttpCode(204)
  @ApiOperation({
    summary: 'Delete runner',
    operationId: 'deleteRunner',
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
  @UseGuards(RunnerAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.DELETE_RUNNERS])
  async delete(@Param('id') id: string): Promise<void> {
    return this.runnerService.remove(id)
  }
}
