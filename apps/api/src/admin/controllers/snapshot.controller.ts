/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, Get, Param, Patch, Query } from '@nestjs/common'
import { ApiBearerAuth, ApiOAuth2, ApiOperation, ApiParam, ApiQuery, ApiResponse, ApiTags } from '@nestjs/swagger'
import { RequiredSystemRole } from '../../user/decorators/required-system-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'
import { SnapshotService } from '../../sandbox/services/snapshot.service'
import { SnapshotDto } from '../../sandbox/dto/snapshot.dto'
import { SetSnapshotGeneralStatusDto } from '../../sandbox/dto/update-snapshot.dto'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'

@ApiTags('admin')
@Controller('admin/snapshots')
@AuthStrategy([AuthStrategyType.API_KEY, AuthStrategyType.JWT])
@RequiredSystemRole(SystemRole.ADMIN)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class AdminSnapshotController {
  constructor(private readonly snapshotService: SnapshotService) {}

  @Get('can-cleanup-image')
  @ApiOperation({
    summary: 'Check if an image can be cleaned up',
    operationId: 'adminCanCleanupImage',
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
  async canCleanupImage(@Query('imageName') imageName: string): Promise<boolean> {
    return this.snapshotService.canCleanupImage(imageName)
  }

  @Patch(':id/general')
  @ApiOperation({
    summary: 'Set snapshot general status',
    operationId: 'adminSetSnapshotGeneralStatus',
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
}
