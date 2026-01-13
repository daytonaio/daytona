/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Body,
  Controller,
  Get,
  Post,
  Param,
  Patch,
  UseGuards,
  Query,
  Delete,
  HttpCode,
  NotFoundException,
  ForbiddenException,
  ParseUUIDPipe,
} from '@nestjs/common'
import { CreateRunnerDto } from '../dto/create-runner.dto'
import { RunnerService } from '../services/runner.service'
import {
  ApiOAuth2,
  ApiTags,
  ApiOperation,
  ApiBearerAuth,
  ApiResponse,
  ApiQuery,
  ApiParam,
  ApiHeader,
} from '@nestjs/swagger'
import { SystemActionGuard } from '../../auth/system-action.guard'
import { RequiredApiRole } from '../../common/decorators/required-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'
import { ProxyGuard } from '../../auth/proxy.guard'
import { RunnerDto } from '../dto/runner.dto'
import { RunnerSnapshotDto } from '../dto/runner-snapshot.dto'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { SshGatewayGuard } from '../../auth/ssh-gateway.guard'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { OrGuard } from '../../auth/or.guard'
import { RunnerAuthGuard } from '../../auth/runner-auth.guard'
import { RunnerContextDecorator } from '../../common/decorators/runner-context.decorator'
import { RunnerContext } from '../../common/interfaces/runner-context.interface'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'
import { RunnerAccessGuard } from '../guards/runner-access.guard'
import { CustomHeaders } from '../../common/constants/header.constants'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { OrganizationResourceActionGuard } from '../../organization/guards/organization-resource-action.guard'
import { CreateRunnerResponseDto } from '../dto/create-runner-response.dto'
import { SandboxAccessGuard } from '../guards/sandbox-access.guard'
import { RunnerFullDto } from '../dto/runner-full.dto'
import { RegionType } from '../../region/enums/region-type.enum'
import { RegionService } from '../../region/services/region.service'
import { RequireFlagsEnabled } from '@openfeature/nestjs-sdk'
import { FeatureFlags } from '../../common/constants/feature-flags'
import { RunnerHealthcheckDto } from '../dto/runner-health.dto'

@ApiTags('runners')
@Controller('runners')
@UseGuards(CombinedAuthGuard, AuthenticatedRateLimitGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class RunnerController {
  constructor(
    private readonly runnerService: RunnerService,
    private readonly regionService: RegionService,
  ) {}

  @Post()
  @HttpCode(201)
  @ApiOperation({
    summary: 'Create runner',
    operationId: 'createRunner',
  })
  @ApiResponse({
    status: 201,
    type: CreateRunnerResponseDto,
  })
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.RUNNER,
    targetIdFromResult: (result: CreateRunnerResponseDto) => result?.id,
    requestMetadata: {
      body: (req: TypedRequest<CreateRunnerDto>) => ({
        regionId: req.body?.regionId,
        name: req.body?.name,
      }),
    },
  })
  @ApiHeader(CustomHeaders.ORGANIZATION_ID)
  @UseGuards(OrganizationResourceActionGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_RUNNERS])
  @RequireFlagsEnabled({ flags: [{ flagKey: FeatureFlags.ORGANIZATION_INFRASTRUCTURE, defaultValue: false }] })
  async create(
    @Body() createRunnerDto: CreateRunnerDto,
    @AuthContext() authContext: OrganizationAuthContext,
  ): Promise<CreateRunnerResponseDto> {
    // validate that the runner region is a custom region owned by the organization
    const region = await this.regionService.findOne(createRunnerDto.regionId)

    if (!region || region.organizationId !== authContext.organizationId) {
      throw new NotFoundException('Region not found')
    }

    if (region.regionType !== RegionType.CUSTOM) {
      throw new ForbiddenException('Runner can only be created in a custom region')
    }

    // create the runner
    const { runner, apiKey } = await this.runnerService.create({
      regionId: createRunnerDto.regionId,
      name: createRunnerDto.name,
      apiVersion: '2',
    })

    return CreateRunnerResponseDto.fromRunner(runner, apiKey)
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
    type: RunnerFullDto,
  })
  async getInfoForAuthenticatedRunner(@RunnerContextDecorator() runnerContext: RunnerContext): Promise<RunnerFullDto> {
    return this.runnerService.findOneFullOrFail(runnerContext.runnerId)
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
  @ApiHeader(CustomHeaders.ORGANIZATION_ID)
  @UseGuards(OrganizationResourceActionGuard, RunnerAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.READ_RUNNERS])
  async getRunnerById(@Param('id', ParseUUIDPipe) id: string): Promise<RunnerDto> {
    const runner = await this.runnerService.findOne(id)

    if (!runner) {
      throw new NotFoundException('Runner not found')
    }

    return RunnerDto.fromRunner(runner)
  }

  @Get(':id/full')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Get runner by ID',
    operationId: 'getRunnerFullById',
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
  @UseGuards(OrGuard([SystemActionGuard, ProxyGuard, SshGatewayGuard, RunnerAccessGuard]))
  @RequiredApiRole([SystemRole.ADMIN, 'proxy', 'ssh-gateway', 'region-proxy', 'region-ssh-gateway'])
  async getRunnerByIdFull(@Param('id', ParseUUIDPipe) id: string): Promise<RunnerFullDto> {
    const runner = await this.runnerService.findOne(id)

    if (!runner) {
      throw new NotFoundException('Runner not found')
    }

    return RunnerFullDto.fromRunner(runner)
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
  @ApiHeader(CustomHeaders.ORGANIZATION_ID)
  @UseGuards(OrganizationResourceActionGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.READ_RUNNERS])
  @RequireFlagsEnabled({ flags: [{ flagKey: FeatureFlags.ORGANIZATION_INFRASTRUCTURE, defaultValue: false }] })
  async findAll(@AuthContext() authContext: OrganizationAuthContext): Promise<RunnerDto[]> {
    return this.runnerService.findAllByOrganization(authContext.organizationId, RegionType.CUSTOM)
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
  @ApiHeader(CustomHeaders.ORGANIZATION_ID)
  @UseGuards(OrganizationResourceActionGuard, RunnerAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_RUNNERS])
  @RequireFlagsEnabled({ flags: [{ flagKey: FeatureFlags.ORGANIZATION_INFRASTRUCTURE, defaultValue: false }] })
  async updateSchedulingStatus(
    @Param('id', ParseUUIDPipe) id: string,
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
  @ApiHeader(CustomHeaders.ORGANIZATION_ID)
  @UseGuards(OrganizationResourceActionGuard, RunnerAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.DELETE_RUNNERS])
  @RequireFlagsEnabled({ flags: [{ flagKey: FeatureFlags.ORGANIZATION_INFRASTRUCTURE, defaultValue: false }] })
  async delete(@Param('id', ParseUUIDPipe) id: string): Promise<void> {
    return this.runnerService.remove(id)
  }

  @Get('/by-sandbox/:sandboxId')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Get runner by sandbox ID',
    operationId: 'getRunnerBySandboxId',
  })
  @ApiResponse({
    status: 200,
    type: RunnerFullDto,
  })
  @UseGuards(OrGuard([SystemActionGuard, ProxyGuard, SshGatewayGuard, SandboxAccessGuard]))
  @RequiredApiRole([SystemRole.ADMIN, 'proxy', 'ssh-gateway', 'region-proxy', 'region-ssh-gateway'])
  async getRunnerBySandboxId(@Param('sandboxId', ParseUUIDPipe) sandboxId: string): Promise<RunnerFullDto> {
    const runner = await this.runnerService.findBySandboxId(sandboxId)

    if (!runner) {
      throw new NotFoundException('Runner not found')
    }

    return RunnerFullDto.fromRunner(runner)
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
  @UseGuards(OrGuard([SystemActionGuard, ProxyGuard, SshGatewayGuard]))
  @RequiredApiRole([SystemRole.ADMIN, 'proxy', 'ssh-gateway'])
  async getRunnersBySnapshotRef(@Query('ref') ref: string): Promise<RunnerSnapshotDto[]> {
    return this.runnerService.getRunnersBySnapshotRef(ref)
  }

  @Get('/inital-by-snapshot-id/:snapshotId')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Get initial runner by snapshot ID',
    operationId: 'getInitialRunnerBySnapshotId',
  })
  @ApiResponse({
    status: 200,
    type: RunnerFullDto,
  })
  @ApiParam({
    name: 'snapshotId',
    description: 'Snapshot ID',
    type: String,
    required: true,
  })
  @UseGuards(OrGuard([SystemActionGuard, ProxyGuard, SshGatewayGuard]))
  @RequiredApiRole([SystemRole.ADMIN, 'proxy', 'ssh-gateway'])
  async getInitialRunnerBySnapshotId(@Param('snapshotId') snapshotId: string): Promise<RunnerFullDto> {
    const runner = await this.runnerService.getInitialRunnerBySnapshotId(snapshotId)
    return RunnerFullDto.fromRunner(runner)
  }

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
      healthcheck.apiUrl,
      healthcheck.proxyUrl,
      healthcheck.metrics,
      healthcheck.appVersion,
    )
  }
}
