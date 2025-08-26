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
import { Audit, MASKED_AUDIT_VALUE, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { OrganizationResourceActionGuard } from '../../organization/guards/organization-resource-action.guard'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'

@ApiTags('runners')
@Controller('runners')
@UseGuards(CombinedAuthGuard, SystemActionGuard, ProxyGuard, OrganizationResourceActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class RunnerController {
  constructor(private readonly runnerService: RunnerService) {}

  @Post()
  @ApiOperation({
    summary: 'Create runner',
    operationId: 'createRunner',
  })
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.RUNNER,
    targetIdFromResult: (result: Runner) => result?.id,
    requestMetadata: {
      body: (req: TypedRequest<CreateRunnerDto>) => ({
        domain: req.body?.domain,
        apiUrl: req.body?.apiUrl,
        apiKey: MASKED_AUDIT_VALUE,
        cpu: req.body?.cpu,
        memoryGiB: req.body?.memoryGiB,
        diskGiB: req.body?.diskGiB,
        gpu: req.body?.gpu,
        gpuType: req.body?.gpuType,
        class: req.body?.class,
        capacity: req.body?.capacity,
        region: req.body?.region,
      }),
    },
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_RUNNERS])
  async create(
    @Body() createRunnerDto: CreateRunnerDto,
    @AuthContext() authContext: OrganizationAuthContext,
  ): Promise<Runner> {
    return this.runnerService.create(createRunnerDto, authContext.organization)
  }

  @Get()
  @ApiOperation({
    summary: 'List all runners',
    operationId: 'listRunners',
  })
  @ApiQuery({
    name: 'region',
    description: 'Filter runners by region code',
    type: String,
    required: false,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.READ_RUNNERS])
  @RequiredApiRole([SystemRole.ADMIN, 'proxy'])
  async findAll(
    @Query('region') region: string,
    @AuthContext() authContext: OrganizationAuthContext,
  ): Promise<Runner[]> {
    if (authContext.role === 'proxy') {
      return this.runnerService.findAll(null, region)
    }

    return this.runnerService.findAll(authContext.organization, region)
  }

  @Patch(':id/scheduling')
  @ApiOperation({
    summary: 'Update runner scheduling status',
    operationId: 'updateRunnerScheduling',
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
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_RUNNERS])
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
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.READ_RUNNERS])
  @RequiredApiRole([SystemRole.ADMIN, 'proxy'])
  async getRunnerBySandboxId(@Param('sandboxId') sandboxId: string): Promise<RunnerDto> {
    const runner = await this.runnerService.findBySandboxId(sandboxId)
    return RunnerDto.fromRunner(runner)
  }

  @Get('/by-snapshot-ref')
  @ApiOperation({
    summary: 'Get runners by snapshot ref',
    operationId: 'getRunnersBySnapshotRef',
  })
  @ApiQuery({
    name: 'ref',
    description: 'Snapshot ref',
    type: String,
    required: true,
  })
  @ApiResponse({
    status: 200,
    description: 'Runners found for the snapshot',
    type: [RunnerSnapshotDto],
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.READ_RUNNERS])
  @RequiredApiRole([SystemRole.ADMIN, 'proxy'])
  async getRunnersBySnapshotRef(@Query('ref') ref: string): Promise<RunnerSnapshotDto[]> {
    return this.runnerService.getRunnersBySnapshotRef(ref)
  }
}
