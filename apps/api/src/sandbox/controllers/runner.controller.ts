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
import { ProxyGuard } from '../guards/proxy.guard'
import { RunnerDto } from '../dto/runner.dto'
import { RunnerSnapshotDto } from '../dto/runner-snapshot.dto'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { SshGatewayGuard } from '../guards/ssh-gateway.guard'
import { OrGuard } from '../../auth/or.guard'
import { RunnerAuthGuard } from '../guards/runner-auth.guard'
import { RunnerAuthContext } from '../../common/interfaces/runner-auth-context.interface'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'
import { RunnerAccessGuard } from '../guards/runner-access.guard'
import { RegionRunnerAccessGuard } from '../guards/region-runner-access.guard'
import { CustomHeaders } from '../../common/constants/header.constants'
import { IsOrganizationAuthContext, IsRunnerAuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { OrganizationResourceActionGuard } from '../../organization/guards/organization-resource-action.guard'
import { CreateRunnerResponseDto } from '../dto/create-runner-response.dto'
import { RegionSandboxAccessGuard } from '../guards/region-sandbox-access.guard'
import { RunnerFullDto } from '../dto/runner-full.dto'
import { RegionType } from '../../region/enums/region-type.enum'
import { RegionService } from '../../region/services/region.service'
import { RequireFlagsEnabled } from '@openfeature/nestjs-sdk'
import { FeatureFlags } from '../../common/constants/feature-flags'
import { RunnerHealthcheckDto } from '../dto/runner-health.dto'
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'

@ApiTags('runners')
@Controller('runners')
@AuthStrategy([AuthStrategyType.API_KEY, AuthStrategyType.JWT])
@UseGuards(AuthenticatedRateLimitGuard)
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
    @IsOrganizationAuthContext() authContext: OrganizationAuthContext,
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
  async getInfoForAuthenticatedRunner(@IsRunnerAuthContext() runnerContext: RunnerAuthContext): Promise<RunnerFullDto> {
    return this.runnerService.findOneFullOrFail(runnerContext.runnerId)
  }

  // TODO: add similar endpoint to admin controller
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
  @UseGuards(OrGuard([ProxyGuard, SshGatewayGuard, RegionSandboxAccessGuard]))
  async getRunnerBySandboxId(@Param('sandboxId') sandboxId: string): Promise<RunnerFullDto> {
    const runner = await this.runnerService.findBySandboxId(sandboxId)

    if (!runner) {
      throw new NotFoundException('Runner not found')
    }

    return RunnerFullDto.fromRunner(runner)
  }

  // TODO: add similar endpoint to admin controller
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
  @UseGuards(OrGuard([ProxyGuard, SshGatewayGuard]))
  async getRunnersBySnapshotRef(@Query('ref') ref: string): Promise<RunnerSnapshotDto[]> {
    return this.runnerService.getRunnersBySnapshotRef(ref)
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
    const runner = await this.runnerService.findOneOrFail(id)
    return RunnerDto.fromRunner(runner)
  }

  // TODO: add similar endpoint to admin controller
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
  @UseGuards(OrGuard([ProxyGuard, SshGatewayGuard, RegionRunnerAccessGuard]))
  async getRunnerByIdFull(@Param('id', ParseUUIDPipe) id: string): Promise<RunnerFullDto> {
    const runner = await this.runnerService.findOneOrFail(id)
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
  async findAll(@IsOrganizationAuthContext() authContext: OrganizationAuthContext): Promise<RunnerDto[]> {
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

  @Patch(':id/draining')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Update runner draining status',
    operationId: 'updateRunnerDraining',
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
    action: AuditAction.UPDATE_DRAINING,
    targetType: AuditTarget.RUNNER,
    targetIdFromRequest: (req) => req.params.id,
    requestMetadata: {
      body: (req: TypedRequest<{ draining: boolean }>) => ({
        draining: req.body?.draining,
      }),
    },
  })
  @ApiHeader(CustomHeaders.ORGANIZATION_ID)
  @UseGuards(OrganizationResourceActionGuard, RunnerAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_RUNNERS])
  @RequireFlagsEnabled({ flags: [{ flagKey: FeatureFlags.ORGANIZATION_INFRASTRUCTURE, defaultValue: false }] })
  async updateDrainingStatus(
    @Param('id', ParseUUIDPipe) id: string,
    @Body('draining') draining: boolean,
  ): Promise<RunnerDto> {
    const updatedRunner = await this.runnerService.updateDrainingStatus(id, draining)
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
    @IsRunnerAuthContext() runnerContext: RunnerAuthContext,
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
