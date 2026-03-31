/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Get, Post, Delete, Body, Param, Logger, UseGuards, HttpCode, Query } from '@nestjs/common'
import {
  ApiOAuth2,
  ApiResponse,
  ApiOperation,
  ApiParam,
  ApiTags,
  ApiHeader,
  ApiQuery,
  ApiBearerAuth,
} from '@nestjs/swagger'
import { VolumeService } from '../services/volume.service'
import { CreateVolumeDto } from '../dto/create-volume.dto'
import { CustomHeaders } from '../../common/constants/header.constants'
import { IsOrganizationAuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { OrganizationAuthContextGuard } from '../../organization/guards/organization-auth-context.guard'
import { VolumeDto } from '../dto/volume.dto'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { VolumeAccessGuard } from '../guards/volume-access.guard'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'

@Controller('volumes')
@ApiTags('volumes')
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@AuthStrategy([AuthStrategyType.API_KEY, AuthStrategyType.JWT])
@UseGuards(AuthenticatedRateLimitGuard)
@UseGuards(OrganizationAuthContextGuard)
export class VolumeController {
  private readonly logger = new Logger(VolumeController.name)

  constructor(private readonly volumeService: VolumeService) {}

  @Get()
  @ApiOperation({
    summary: 'List all volumes',
    operationId: 'listVolumes',
  })
  @ApiQuery({
    name: 'includeDeleted',
    required: false,
    type: Boolean,
    description: 'Include deleted volumes in the response',
  })
  @ApiResponse({
    status: 200,
    description: 'List of all volumes',
    type: [VolumeDto],
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.READ_VOLUMES])
  async listVolumes(
    @IsOrganizationAuthContext() authContext: OrganizationAuthContext,
    @Query('includeDeleted') includeDeleted = false,
  ): Promise<VolumeDto[]> {
    const volumes = await this.volumeService.findAll(authContext.organizationId, includeDeleted)
    return volumes.map(VolumeDto.fromVolume)
  }

  @Post()
  @HttpCode(200)
  @ApiOperation({
    summary: 'Create a new volume',
    operationId: 'createVolume',
  })
  @ApiResponse({
    status: 200,
    description: 'The volume has been successfully created.',
    type: VolumeDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_VOLUMES])
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.VOLUME,
    targetIdFromResult: (result: VolumeDto) => result?.id,
    requestMetadata: {
      body: (req: TypedRequest<CreateVolumeDto>) => ({
        name: req.body?.name,
      }),
    },
  })
  async createVolume(
    @IsOrganizationAuthContext() authContext: OrganizationAuthContext,
    @Body() createVolumeDto: CreateVolumeDto,
  ): Promise<VolumeDto> {
    const volume = await this.volumeService.create(authContext.organization, createVolumeDto)
    return VolumeDto.fromVolume(volume)
  }

  @Get(':volumeId')
  @ApiOperation({
    summary: 'Get volume details',
    operationId: 'getVolume',
  })
  @ApiParam({
    name: 'volumeId',
    description: 'ID of the volume',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Volume details',
    type: VolumeDto,
  })
  @UseGuards(VolumeAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.READ_VOLUMES])
  async getVolume(@Param('volumeId') volumeId: string): Promise<VolumeDto> {
    const volume = await this.volumeService.findOne(volumeId)
    return VolumeDto.fromVolume(volume)
  }

  @Delete(':volumeId')
  @ApiOperation({
    summary: 'Delete volume',
    operationId: 'deleteVolume',
  })
  @ApiParam({
    name: 'volumeId',
    description: 'ID of the volume',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Volume has been marked for deletion',
  })
  @ApiResponse({
    status: 409,
    description: 'Volume is in use by one or more sandboxes',
  })
  @UseGuards(VolumeAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.DELETE_VOLUMES])
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.VOLUME,
    targetIdFromRequest: (req) => req.params.volumeId,
  })
  async deleteVolume(@Param('volumeId') volumeId: string): Promise<void> {
    return this.volumeService.delete(volumeId)
  }

  @Get('by-name/:name')
  @ApiOperation({
    summary: 'Get volume details by name',
    operationId: 'getVolumeByName',
  })
  @ApiParam({
    name: 'name',
    description: 'Name of the volume',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Volume details',
    type: VolumeDto,
  })
  @UseGuards(VolumeAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.READ_VOLUMES])
  async getVolumeByName(
    @IsOrganizationAuthContext() authContext: OrganizationAuthContext,
    @Param('name') name: string,
  ): Promise<VolumeDto> {
    const volume = await this.volumeService.findByName(authContext.organizationId, name)
    return VolumeDto.fromVolume(volume)
  }
}
