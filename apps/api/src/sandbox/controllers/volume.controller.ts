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
  Query,
} from '@nestjs/common'
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
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { VolumeService } from '../services/volume.service'
import { CreateVolumeDto } from '../dto/create-volume.dto'
import { ContentTypeInterceptor } from '../../common/interceptors/content-type.interceptors'
import { CustomHeaders } from '../../common/constants/header.constants'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { OrganizationResourceActionGuard } from '../../organization/guards/organization-resource-action.guard'
import { VolumeDto } from '../dto/volume.dto'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { ForbiddenException } from '@nestjs/common'
import { Audit, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'

@ApiTags('volumes')
@Controller('volumes')
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@UseGuards(CombinedAuthGuard, OrganizationResourceActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class VolumeController {
  private readonly logger = new Logger(VolumeController.name)

  constructor(
    @InjectRedis() private readonly redis: Redis,
    private readonly volumeService: VolumeService,
  ) {}

  @Get()
  @ApiOperation({
    summary: 'List all volumes',
    operationId: 'listVolumes',
  })
  @ApiResponse({
    status: 200,
    description: 'List of all volumes',
    type: [VolumeDto],
  })
  @ApiQuery({
    name: 'includeDeleted',
    required: false,
    type: Boolean,
    description: 'Include deleted volumes in the response',
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.READ_VOLUMES])
  async listVolumes(
    @AuthContext() authContext: OrganizationAuthContext,
    @Query('includeDeleted') includeDeleted = false,
  ): Promise<VolumeDto[]> {
    const volumes = await this.volumeService.findAll(authContext.organizationId, includeDeleted)
    return volumes.map(VolumeDto.fromVolume)
  }

  @Post()
  @HttpCode(200)
  @UseInterceptors(ContentTypeInterceptor)
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
    @AuthContext() authContext: OrganizationAuthContext,
    @Body() createVolumeDto: CreateVolumeDto,
  ): Promise<VolumeDto> {
    const organization = authContext.organization

    //  optimistic quota guard
    //  protect against race condition on volume create abuse
    //  not 100% correct when close to quota limit
    const concurrentCreateKey = `volume-concurrent-create-${organization.id}`
    let concurrentCreateCount = parseInt(await this.redis.get(concurrentCreateKey)) || 0
    concurrentCreateCount++
    await this.redis.setex(concurrentCreateKey, 1, concurrentCreateCount)

    const activeVolumeCount = await this.volumeService.countActive(organization.id)

    if (activeVolumeCount + concurrentCreateCount > organization.volumeQuota) {
      throw new ForbiddenException(`Volume quota exceeded. Maximum allowed: ${organization.volumeQuota}`)
    }

    const volume = await this.volumeService.create(organization, createVolumeDto)
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
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.READ_VOLUMES])
  async getVolumeByName(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('name') name: string,
  ): Promise<VolumeDto> {
    const volume = await this.volumeService.findByName(authContext.organizationId, name)
    return VolumeDto.fromVolume(volume)
  }
}
