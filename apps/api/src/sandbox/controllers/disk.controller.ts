/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Controller,
  Get,
  Post,
  Delete,
  Body,
  Param,
  Logger,
  UseGuards,
  HttpCode,
  UseInterceptors,
} from '@nestjs/common'
import { ApiOAuth2, ApiResponse, ApiOperation, ApiParam, ApiTags, ApiHeader, ApiBearerAuth } from '@nestjs/swagger'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { DiskService } from '../services/disk.service'
import { CreateDiskDto } from '../dto/create-disk.dto'
import { AttachDiskDto } from '../dto/attach-disk.dto'
import { DetachDiskDto } from '../dto/detach-disk.dto'
import { ContentTypeInterceptor } from '../../common/interceptors/content-type.interceptors'
import { CustomHeaders } from '../../common/constants/header.constants'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { OrganizationResourceActionGuard } from '../../organization/guards/organization-resource-action.guard'
import { DiskDto } from '../dto/disk.dto'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'

@ApiTags('disks')
@Controller('disks')
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@UseGuards(CombinedAuthGuard, OrganizationResourceActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class DiskController {
  private readonly logger = new Logger(DiskController.name)

  constructor(private readonly diskService: DiskService) {}

  @Get()
  @ApiOperation({
    summary: 'List all disks',
    operationId: 'listDisks',
  })
  @ApiResponse({
    status: 200,
    description: 'List of all disks',
    type: [DiskDto],
  })
  async listDisks(@AuthContext() authContext: OrganizationAuthContext): Promise<DiskDto[]> {
    const disks = await this.diskService.findAll(authContext.organizationId)
    return disks.map(DiskDto.fromDisk)
  }

  @Post()
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Create a new disk',
    operationId: 'createDisk',
  })
  @ApiResponse({
    status: 200,
    description: 'The disk has been successfully created.',
    type: DiskDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.DISK,
    targetIdFromResult: (result: DiskDto) => result?.id,
    requestMetadata: {
      body: (req: TypedRequest<CreateDiskDto>) => ({
        name: req.body?.name,
        size: req.body?.size,
      }),
    },
  })
  async createDisk(
    @AuthContext() authContext: OrganizationAuthContext,
    @Body() createDiskDto: CreateDiskDto,
  ): Promise<DiskDto> {
    const disk = await this.diskService.create(authContext.organization, createDiskDto)
    return DiskDto.fromDisk(disk)
  }

  @Get(':diskId')
  @ApiOperation({
    summary: 'Get disk details',
    operationId: 'getDisk',
  })
  @ApiParam({
    name: 'diskId',
    description: 'ID of the disk',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Disk details',
    type: DiskDto,
  })
  async getDisk(@Param('diskId') diskId: string): Promise<DiskDto> {
    const disk = await this.diskService.findOne(diskId)
    return DiskDto.fromDisk(disk)
  }

  @Delete(':diskId')
  @ApiOperation({
    summary: 'Delete disk',
    operationId: 'deleteDisk',
  })
  @ApiParam({
    name: 'diskId',
    description: 'ID of the disk',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Disk has been marked for deletion',
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.DELETE_SANDBOXES])
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.DISK,
    targetIdFromRequest: (req) => req.params.diskId,
  })
  async deleteDisk(@Param('diskId') diskId: string): Promise<void> {
    return this.diskService.delete(diskId)
  }

  @Post(':diskId/attach')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Attach disk to sandbox',
    operationId: 'attachDisk',
  })
  @ApiParam({
    name: 'diskId',
    description: 'ID of the disk',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Disk has been successfully attached to sandbox',
    type: DiskDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @Audit({
    action: AuditAction.UPDATE,
    targetType: AuditTarget.DISK,
    targetIdFromRequest: (req) => req.params.diskId,
    requestMetadata: {
      body: (req: TypedRequest<AttachDiskDto>) => ({
        sandboxId: req.body?.sandboxId,
      }),
    },
  })
  async attachDisk(@Param('diskId') diskId: string, @Body() attachDiskDto: AttachDiskDto): Promise<DiskDto> {
    const disk = await this.diskService.attachToSandbox(diskId, attachDiskDto.sandboxId)
    return DiskDto.fromDisk(disk)
  }

  @Post(':diskId/detach')
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
  @ApiOperation({
    summary: 'Detach disk from sandbox',
    operationId: 'detachDisk',
  })
  @ApiParam({
    name: 'diskId',
    description: 'ID of the disk',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Disk has been successfully detached from sandbox',
    type: DiskDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
  @Audit({
    action: AuditAction.UPDATE,
    targetType: AuditTarget.DISK,
    targetIdFromRequest: (req) => req.params.diskId,
  })
  async detachDisk(@Param('diskId') diskId: string): Promise<DiskDto> {
    const disk = await this.diskService.detachFromSandbox(diskId)
    return DiskDto.fromDisk(disk)
  }
}
