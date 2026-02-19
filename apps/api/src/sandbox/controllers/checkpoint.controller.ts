/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Delete, Get, Post, Body, Param, Query, UseGuards, HttpCode, Logger } from '@nestjs/common'
import {
  ApiOAuth2,
  ApiTags,
  ApiOperation,
  ApiResponse,
  ApiParam,
  ApiHeader,
  ApiBearerAuth,
} from '@nestjs/swagger'
import { CustomHeaders } from '../../common/constants/header.constants'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { OrganizationResourceActionGuard } from '../../organization/guards/organization-resource-action.guard'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { SystemActionGuard } from '../../auth/system-action.guard'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { CheckpointAccessGuard } from '../guards/checkpoint-access.guard'
import { CheckpointService } from '../services/checkpoint.service'
import { SnapshotService } from '../services/snapshot.service'
import { CheckpointDto } from '../dto/checkpoint.dto'
import { SnapshotDto } from '../dto/snapshot.dto'
import { CreateSnapshotFromCheckpointDto } from '../dto/create-snapshot-from-checkpoint.dto'
import { ListCheckpointsQueryDto } from '../dto/list-checkpoints-query.dto'
import { PaginatedCheckpointsDto } from '../dto/paginated-checkpoints.dto'

@ApiTags('checkpoints')
@Controller('checkpoints')
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@UseGuards(CombinedAuthGuard, SystemActionGuard, OrganizationResourceActionGuard, AuthenticatedRateLimitGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class CheckpointController {
  private readonly logger = new Logger(CheckpointController.name)

  constructor(
    private readonly checkpointService: CheckpointService,
    private readonly snapshotService: SnapshotService,
  ) {}

  @Get()
  @ApiOperation({
    summary: 'List checkpoints',
    operationId: 'listCheckpoints',
  })
  @ApiResponse({
    status: 200,
    description: 'Paginated list of checkpoints',
    type: PaginatedCheckpointsDto,
  })
  async listCheckpoints(
    @AuthContext() authContext: OrganizationAuthContext,
    @Query() queryParams: ListCheckpointsQueryDto,
  ): Promise<PaginatedCheckpointsDto> {
    const { page, limit, sandboxId, sort, order } = queryParams

    const result = await this.checkpointService.list(
      authContext.organizationId,
      page,
      limit,
      sandboxId,
      { field: sort, direction: order },
    )

    return {
      items: result.items.map(CheckpointDto.fromCheckpoint),
      total: result.total,
      page: result.page,
      totalPages: result.totalPages,
    }
  }

  @Get(':checkpointId')
  @ApiOperation({
    summary: 'Get a checkpoint',
    operationId: 'getCheckpoint',
  })
  @ApiParam({
    name: 'checkpointId',
    description: 'Checkpoint ID',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'The checkpoint',
    type: CheckpointDto,
  })
  @ApiResponse({
    status: 404,
    description: 'Checkpoint not found',
  })
  @UseGuards(CheckpointAccessGuard)
  async getCheckpoint(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('checkpointId') checkpointId: string,
  ): Promise<CheckpointDto> {
    const checkpoint = await this.checkpointService.getCheckpoint(checkpointId, authContext.organizationId)
    return CheckpointDto.fromCheckpoint(checkpoint)
  }

  @Delete(':checkpointId')
  @ApiOperation({
    summary: 'Delete a checkpoint',
    operationId: 'deleteCheckpoint',
  })
  @ApiParam({
    name: 'checkpointId',
    description: 'Checkpoint ID',
    type: 'string',
  })
  @ApiResponse({
    status: 204,
    description: 'Checkpoint has been deleted',
  })
  @HttpCode(204)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.DELETE_CHECKPOINTS])
  @UseGuards(CheckpointAccessGuard)
  async deleteCheckpoint(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('checkpointId') checkpointId: string,
  ): Promise<void> {
    await this.checkpointService.deleteCheckpoint(checkpointId, authContext.organizationId)
  }

  @Post(':checkpointId/promote-to-snapshot')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Promote a checkpoint to a snapshot',
    operationId: 'promoteCheckpointToSnapshot',
  })
  @ApiParam({
    name: 'checkpointId',
    description: 'Checkpoint ID',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Snapshot created from checkpoint',
    type: SnapshotDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SNAPSHOTS])
  @UseGuards(CheckpointAccessGuard)
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.SNAPSHOT,
    targetIdFromResult: (result: SnapshotDto) => result?.id,
    requestMetadata: {
      body: (req: TypedRequest<CreateSnapshotFromCheckpointDto>) => ({
        name: req.body?.name,
      }),
    },
  })
  async promoteCheckpointToSnapshot(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('checkpointId') checkpointId: string,
    @Body() dto: CreateSnapshotFromCheckpointDto,
  ): Promise<SnapshotDto> {
    const checkpoint = await this.checkpointService.getCheckpoint(checkpointId, authContext.organizationId)
    const snapshot = await this.snapshotService.createFromCheckpoint(checkpoint, dto.name)
    return SnapshotDto.fromSnapshot(snapshot)
  }
}
