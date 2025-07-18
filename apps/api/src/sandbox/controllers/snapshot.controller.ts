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
  Patch,
  Post,
  Query,
  UseGuards,
  HttpCode,
  ForbiddenException,
  Logger,
  NotFoundException,
  Res,
  Request,
  RawBodyRequest,
  Next,
  ParseBoolPipe,
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
import { ToggleStateDto } from '../dto/toggle-state.dto'
import { SnapshotDto } from '../dto/snapshot.dto'
import { PaginatedSnapshotsDto } from '../dto/paginated-snapshots.dto'
import { SnapshotAccessGuard } from '../guards/snapshot-access.guard'
import { CustomHeaders } from '../../common/constants/header.constants'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { OrganizationResourceActionGuard } from '../../organization/guards/organization-resource-action.guard'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { SystemActionGuard } from '../../auth/system-action.guard'
import { RequiredSystemRole } from '../../common/decorators/required-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'
import { SetSnapshotGeneralStatusDto } from '../dto/update-snapshot.dto'
import { LogProxy } from '../proxy/log-proxy'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { Snapshot } from '../entities/snapshot.entity'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'

@ApiTags('snapshots')
@Controller('snapshots')
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@UseGuards(CombinedAuthGuard, SystemActionGuard, OrganizationResourceActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
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
        general: req.body?.general,
        cpu: req.body?.cpu,
        memory: req.body?.memory,
        disk: req.body?.disk,
        gpu: req.body?.gpu,
        buildInfo: req.body?.buildInfo,
      }),
    },
  })
  async createSnapshot(
    @AuthContext() authContext: OrganizationAuthContext,
    @Body() createSnapshotDto: CreateSnapshotDto,
  ): Promise<SnapshotDto> {
    if (createSnapshotDto.general && authContext.role !== SystemRole.ADMIN) {
      throw new ForbiddenException('Insufficient permissions for creating general snapshots')
    }

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
    const snapshot = await this.snapshotService.createSnapshot(authContext.organization, createSnapshotDto)
    return SnapshotDto.fromSnapshot(snapshot)
  }

  @Get('can-cleanup-image')
  @ApiOperation({
    summary: 'Check if an image can be cleaned up',
    operationId: 'canCleanupImage',
  })
  @ApiQuery({
    name: 'imageName',
    required: true,
    type: String,
    description: 'Image name with tag to check',
  })
  @ApiResponse({
    status: 200,
    description: 'Boolean indicating if image can be cleaned up',
    type: Boolean,
  })
  @RequiredSystemRole(SystemRole.ADMIN)
  async canCleanupImage(@Query('imageName') imageName: string): Promise<boolean> {
    return this.snapshotService.canCleanupImage(imageName)
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
  @UseGuards(SnapshotAccessGuard)
  async getSnapshot(
    @Param('id') snapshotIdOrName: string,
    @AuthContext() authContext: OrganizationAuthContext,
  ): Promise<SnapshotDto> {
    let snapshot: Snapshot
    try {
      // Try to get by ID
      snapshot = await this.snapshotService.getSnapshot(snapshotIdOrName)
    } catch (error) {
      // If not found by ID, try by name
      snapshot = await this.snapshotService.getSnapshotByName(snapshotIdOrName, authContext.organizationId)
    }
    return SnapshotDto.fromSnapshot(snapshot)
  }

  @Patch(':id/toggle')
  @ApiOperation({
    summary: 'Toggle snapshot state',
    operationId: 'toggleSnapshotState',
  })
  @ApiParam({
    name: 'id',
    description: 'Snapshot ID',
  })
  @ApiResponse({
    status: 200,
    description: 'Snapshot state has been toggled',
    type: SnapshotDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SNAPSHOTS])
  @UseGuards(SnapshotAccessGuard)
  @Audit({
    action: AuditAction.TOGGLE_STATE,
    targetType: AuditTarget.SNAPSHOT,
    targetIdFromRequest: (req) => req.params.id,
    requestMetadata: {
      body: (req: TypedRequest<ToggleStateDto>) => ({
        enabled: req.body?.enabled,
      }),
    },
  })
  async toggleSnapshotState(@Param('id') snapshotId: string, @Body() toggleDto: ToggleStateDto): Promise<SnapshotDto> {
    const snapshot = await this.snapshotService.toggleSnapshotState(snapshotId, toggleDto.enabled)
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
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.DELETE_SNAPSHOTS])
  @UseGuards(SnapshotAccessGuard)
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.SNAPSHOT,
    targetIdFromRequest: (req) => req.params.id,
  })
  async removeSnapshot(@Param('id') snapshotId: string): Promise<void> {
    await this.snapshotService.removeSnapshot(snapshotId)
  }

  @Get()
  @ApiOperation({
    summary: 'List all snapshots',
    operationId: 'getAllSnapshots',
  })
  @ApiQuery({
    name: 'page',
    required: false,
    type: Number,
    description: 'Page number',
  })
  @ApiQuery({
    name: 'limit',
    required: false,
    type: Number,
    description: 'Number of items per page',
  })
  @ApiResponse({
    status: 200,
    description: 'List of all snapshots with pagination',
    type: PaginatedSnapshotsDto,
  })
  async getAllSnapshots(
    @AuthContext() authContext: OrganizationAuthContext,
    @Query('page') page = 1,
    @Query('limit') limit = 10,
  ): Promise<PaginatedSnapshotsDto> {
    const result = await this.snapshotService.getAllSnapshots(authContext.organizationId, page, limit)
    return {
      items: result.items.map(SnapshotDto.fromSnapshot),
      total: result.total,
      page: result.page,
      totalPages: result.totalPages,
    }
  }

  @Patch(':id/general')
  @ApiOperation({
    summary: 'Set snapshot general status',
    operationId: 'setSnapshotGeneralStatus',
  })
  @ApiParam({
    name: 'id',
    description: 'Snapshot ID',
  })
  @ApiResponse({
    status: 200,
    description: 'Snapshot general status has been set',
    type: SnapshotDto,
  })
  @RequiredSystemRole(SystemRole.ADMIN)
  @Audit({
    action: AuditAction.SET_GENERAL_STATUS,
    targetType: AuditTarget.SNAPSHOT,
    targetIdFromRequest: (req) => req.params.id,
    requestMetadata: {
      body: (req: TypedRequest<SetSnapshotGeneralStatusDto>) => ({
        general: req.body?.general,
      }),
    },
  })
  async setSnapshotGeneralStatus(
    @Param('id') snapshotId: string,
    @Body() dto: SetSnapshotGeneralStatusDto,
  ): Promise<SnapshotDto> {
    const snapshot = await this.snapshotService.setSnapshotGeneralStatus(snapshotId, dto.general)
    return SnapshotDto.fromSnapshot(snapshot)
  }

  @Get(':id/build-logs')
  @ApiOperation({
    summary: 'Get snapshot build logs',
    operationId: 'getSnapshotBuildLogs',
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

    // Retry until a runner is assigned or timeout after 30 seconds
    const startTime = Date.now()
    const timeoutMs = 30 * 1000

    while (!snapshot.buildRunnerId) {
      if (Date.now() - startTime > timeoutMs) {
        throw new NotFoundException(`Timeout waiting for build runner assignment for snapshot ${snapshotId}`)
      }
      await new Promise((resolve) => setTimeout(resolve, 1000))
      snapshot = await this.snapshotService.getSnapshot(snapshotId)
    }

    const runner = await this.runnerService.findOne(snapshot.buildRunnerId)
    if (!runner) {
      throw new NotFoundException(`Build runner for snapshot ${snapshotId} not found`)
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
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SNAPSHOTS])
  @UseGuards(SnapshotAccessGuard)
  @Audit({
    action: AuditAction.ACTIVATE,
    targetType: AuditTarget.SNAPSHOT,
    targetIdFromRequest: (req) => req.params.id,
  })
  async activateSnapshot(@Param('id') snapshotId: string): Promise<SnapshotDto> {
    const snapshot = await this.snapshotService.activateSnapshot(snapshotId)
    return SnapshotDto.fromSnapshot(snapshot)
  }
}
