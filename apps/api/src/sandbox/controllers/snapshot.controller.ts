/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Body,
  Controller,
  Delete,
  Get,
  Param,
  Post,
  Query,
  UseGuards,
  HttpCode,
  Logger,
  NotFoundException,
  Res,
  Request,
  RawBodyRequest,
  Next,
  ParseBoolPipe,
  ParseUUIDPipe,
} from '@nestjs/common'
import { IncomingMessage, ServerResponse } from 'http'
import { NextFunction } from 'express'
import { SnapshotService } from '../services/snapshot.service'
import { RunnerService } from '../services/runner.service'
import {
  ApiOAuth2,
  ApiTags,
  ApiOperation,
  ApiResponse,
  ApiParam,
  ApiQuery,
  ApiHeader,
  ApiBearerAuth,
} from '@nestjs/swagger'
import { CreateSnapshotDto } from '../dto/create-snapshot.dto'
import { SnapshotDto } from '../dto/snapshot.dto'
import { PaginatedSnapshotsDto } from '../dto/paginated-snapshots.dto'
import { SnapshotAccessGuard } from '../guards/snapshot-access.guard'
import { SnapshotReadAccessGuard } from '../guards/snapshot-read-access.guard'
import { CustomHeaders } from '../../common/constants/header.constants'
import { IsOrganizationAuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { OrganizationAuthContextGuard } from '../../organization/guards/organization-auth-context.guard'
import { LogProxy } from '../proxy/log-proxy'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { ListSnapshotsQueryDto } from '../dto/list-snapshots-query.dto'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'
import { UrlDto } from '../../common/dto/url.dto'
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'

@Controller('snapshots')
@ApiTags('snapshots')
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@AuthStrategy([AuthStrategyType.API_KEY, AuthStrategyType.JWT])
@UseGuards(AuthenticatedRateLimitGuard)
@UseGuards(OrganizationAuthContextGuard)
export class SnapshotController {
  private readonly logger = new Logger(SnapshotController.name)

  constructor(
    private readonly snapshotService: SnapshotService,
    private readonly runnerService: RunnerService,
  ) {}

  @Post()
  @HttpCode(200)
  @ApiOperation({
    summary: 'Create a new snapshot',
    operationId: 'createSnapshot',
  })
  @ApiResponse({
    status: 200,
    description: 'The snapshot has been successfully created.',
    type: SnapshotDto,
  })
  @ApiResponse({
    status: 400,
    description: 'Bad request - Snapshots with tag ":latest" are not allowed',
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SNAPSHOTS])
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.SNAPSHOT,
    targetIdFromResult: (result: SnapshotDto) => result?.id,
    requestMetadata: {
      body: (req: TypedRequest<CreateSnapshotDto>) => ({
        name: req.body?.name,
        imageName: req.body?.imageName,
        entrypoint: req.body?.entrypoint,
        cpu: req.body?.cpu,
        memory: req.body?.memory,
        disk: req.body?.disk,
        gpu: req.body?.gpu,
        buildInfo: req.body?.buildInfo,
      }),
    },
  })
  async createSnapshot(
    @IsOrganizationAuthContext() authContext: OrganizationAuthContext,
    @Body() createSnapshotDto: CreateSnapshotDto,
  ): Promise<SnapshotDto> {
    if (createSnapshotDto.buildInfo) {
      if (createSnapshotDto.imageName) {
        throw new BadRequestError('Cannot specify an image name when using a build info entry')
      }
      if (createSnapshotDto.entrypoint) {
        throw new BadRequestError('Cannot specify an entrypoint when using a build info entry')
      }
    } else {
      if (!createSnapshotDto.imageName) {
        throw new BadRequestError('Must specify an image name when not using a build info entry')
      }
    }

    // TODO: consider - if using transient registry, prepend the snapshot name with the username
    const snapshot = createSnapshotDto.buildInfo
      ? await this.snapshotService.createFromBuildInfo(authContext.organization, createSnapshotDto)
      : await this.snapshotService.createFromPull(authContext.organization, createSnapshotDto)
    return SnapshotDto.fromSnapshot(snapshot)
  }

  @Get(':id')
  @ApiOperation({
    summary: 'Get snapshot by ID or name',
    operationId: 'getSnapshot',
  })
  @ApiParam({
    name: 'id',
    description: 'Snapshot ID or name',
  })
  @ApiResponse({
    status: 200,
    description: 'The snapshot',
    type: SnapshotDto,
  })
  @ApiResponse({
    status: 404,
    description: 'Snapshot not found',
  })
  @UseGuards(SnapshotReadAccessGuard)
  async getSnapshot(
    @Param('id') snapshotIdOrName: string,
    @IsOrganizationAuthContext() authContext: OrganizationAuthContext,
  ): Promise<SnapshotDto> {
    const snapshot = await this.snapshotService.getSnapshotWithRegions(snapshotIdOrName, authContext.organizationId)
    return SnapshotDto.fromSnapshot(snapshot)
  }

  @Delete(':id')
  @ApiOperation({
    summary: 'Delete snapshot',
    operationId: 'removeSnapshot',
  })
  @ApiParam({
    name: 'id',
    description: 'Snapshot ID',
  })
  @ApiResponse({
    status: 200,
    description: 'Snapshot has been deleted',
  })
  @UseGuards(SnapshotAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.DELETE_SNAPSHOTS])
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.SNAPSHOT,
    targetIdFromRequest: (req) => req.params.id,
  })
  async removeSnapshot(@Param('id', ParseUUIDPipe) snapshotId: string): Promise<void> {
    await this.snapshotService.removeSnapshot(snapshotId)
  }

  @Get()
  @ApiOperation({
    summary: 'List all snapshots',
    operationId: 'getAllSnapshots',
  })
  @ApiResponse({
    status: 200,
    description: 'Paginated list of all snapshots',
    type: PaginatedSnapshotsDto,
  })
  async getAllSnapshots(
    @IsOrganizationAuthContext() authContext: OrganizationAuthContext,
    @Query() queryParams: ListSnapshotsQueryDto,
  ): Promise<PaginatedSnapshotsDto> {
    const { page, limit, name, sort, order } = queryParams

    const result = await this.snapshotService.getAllSnapshots(
      authContext.organizationId,
      page,
      limit,
      { name },
      { field: sort, direction: order },
    )

    return {
      items: result.items.map(SnapshotDto.fromSnapshot),
      total: result.total,
      page: result.page,
      totalPages: result.totalPages,
    }
  }

  @Get(':id/build-logs')
  @ApiOperation({
    summary: 'Get snapshot build logs',
    operationId: 'getSnapshotBuildLogs',
    deprecated: true,
    description: 'This endpoint is deprecated. Use `getSnapshotBuildLogsUrl` instead.',
  })
  @ApiParam({
    name: 'id',
    description: 'Snapshot ID',
  })
  @ApiQuery({
    name: 'follow',
    required: false,
    type: Boolean,
    description: 'Whether to follow the logs stream',
  })
  @UseGuards(SnapshotAccessGuard)
  async getSnapshotBuildLogs(
    @Request() req: RawBodyRequest<IncomingMessage>,
    @Res() res: ServerResponse<IncomingMessage>,
    @Next() next: NextFunction,
    @Param('id') snapshotId: string,
    @Query('follow', new ParseBoolPipe({ optional: true })) follow?: boolean,
  ): Promise<void> {
    let snapshot = await this.snapshotService.getSnapshot(snapshotId)

    // Check if the snapshot has build info
    if (!snapshot.buildInfo) {
      throw new NotFoundException(`Snapshot ${snapshotId} has no build info`)
    }

    if (snapshot.state == SnapshotState.ACTIVE) {
      // Close the connection
      res.end()
      return
    }

    // Retry until a runner is assigned or timeout after 30 seconds
    const startTime = Date.now()
    const timeoutMs = 30 * 1000

    while (!snapshot.initialRunnerId) {
      if (Date.now() - startTime > timeoutMs) {
        throw new NotFoundException(`Timeout waiting for build runner assignment for snapshot ${snapshotId}`)
      }
      await new Promise((resolve) => setTimeout(resolve, 1000))
      snapshot = await this.snapshotService.getSnapshot(snapshotId)
    }

    const runner = await this.runnerService.findOneOrFail(snapshot.initialRunnerId)

    if (!runner.apiUrl) {
      throw new NotFoundException(`Build runner for snapshot ${snapshotId} has no API URL`)
    }

    const logProxy = new LogProxy(
      runner.apiUrl,
      snapshot.buildInfo.snapshotRef,
      runner.apiKey,
      follow === true,
      req,
      res,
      next,
    )
    return logProxy.create()
  }

  @Get(':id/build-logs-url')
  @ApiOperation({
    summary: 'Get snapshot build logs URL',
    operationId: 'getSnapshotBuildLogsUrl',
  })
  @ApiParam({
    name: 'id',
    description: 'Snapshot ID',
  })
  @ApiResponse({
    status: 200,
    description: 'The snapshot build logs URL',
    type: UrlDto,
  })
  @UseGuards(SnapshotAccessGuard)
  async getSnapshotBuildLogsUrl(@Param('id') snapshotId: string): Promise<UrlDto> {
    let snapshot = await this.snapshotService.getSnapshot(snapshotId)

    // Check if the snapshot has build info
    if (!snapshot.buildInfo) {
      throw new NotFoundException(`Snapshot ${snapshotId} has no build info`)
    }

    // Retry until a runner is assigned or timeout after 30 seconds
    const startTime = Date.now()
    const timeoutMs = 30 * 1000

    while (!snapshot.initialRunnerId) {
      if (Date.now() - startTime > timeoutMs) {
        throw new NotFoundException(`Timeout waiting for build runner assignment for snapshot ${snapshotId}`)
      }
      await new Promise((resolve) => setTimeout(resolve, 1000))
      snapshot = await this.snapshotService.getSnapshot(snapshotId)
    }

    const url = await this.snapshotService.getBuildLogsUrl(snapshot)
    return new UrlDto(url)
  }

  @Post(':id/activate')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Activate a snapshot',
    operationId: 'activateSnapshot',
  })
  @ApiParam({
    name: 'id',
    description: 'Snapshot ID',
  })
  @ApiResponse({
    status: 200,
    description: 'The snapshot has been successfully activated.',
    type: SnapshotDto,
  })
  @ApiResponse({
    status: 400,
    description: 'Bad request - Snapshot is already active, not in inactive state, or has associated snapshot runners',
  })
  @ApiResponse({
    status: 404,
    description: 'Snapshot not found',
  })
  @UseGuards(SnapshotAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SNAPSHOTS])
  @Audit({
    action: AuditAction.ACTIVATE,
    targetType: AuditTarget.SNAPSHOT,
    targetIdFromRequest: (req) => req.params.id,
  })
  async activateSnapshot(
    @Param('id', ParseUUIDPipe) snapshotId: string,
    @IsOrganizationAuthContext() authContext: OrganizationAuthContext,
  ): Promise<SnapshotDto> {
    const snapshot = await this.snapshotService.activateSnapshot(snapshotId, authContext.organization)
    return SnapshotDto.fromSnapshot(snapshot)
  }

  @Post(':id/deactivate')
  @HttpCode(204)
  @ApiOperation({
    summary: 'Deactivate a snapshot',
    operationId: 'deactivateSnapshot',
  })
  @ApiParam({
    name: 'id',
    description: 'Snapshot ID',
  })
  @ApiResponse({
    status: 204,
    description: 'The snapshot has been successfully deactivated.',
  })
  @UseGuards(SnapshotAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SNAPSHOTS])
  @Audit({
    action: AuditAction.DEACTIVATE,
    targetType: AuditTarget.SNAPSHOT,
    targetIdFromRequest: (req) => req.params.id,
  })
  async deactivateSnapshot(@Param('id', ParseUUIDPipe) snapshotId: string) {
    await this.snapshotService.deactivateSnapshot(snapshotId)
  }
}
