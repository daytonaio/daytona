/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Controller,
  Get,
  Post,
  Body,
  Patch,
  Param,
  Delete,
  UseGuards,
  HttpCode,
  Query,
  BadRequestException,
  NotFoundException,
} from '@nestjs/common'
import { ApiTags, ApiOperation, ApiResponse, ApiOAuth2, ApiParam, ApiBearerAuth, ApiQuery } from '@nestjs/swagger'
import { AdminCreateDockerRegistryDto } from '../dto/create-docker-registry.dto'
import { AdminUpdateDockerRegistryDto } from '../dto/update-docker-registry.dto'
import { Audit, MASKED_AUDIT_VALUE, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { SystemActionGuard } from '../../auth/system-action.guard'
import { RequiredSystemRole } from '../../common/decorators/required-role.decorator'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { DockerRegistryDto } from '../../docker-registry/dto/docker-registry.dto'
import { SystemRole } from '../../user/enums/system-role.enum'

@ApiTags('admin/registries')
@Controller('admin/registries')
@UseGuards(CombinedAuthGuard, SystemActionGuard)
@RequiredSystemRole([SystemRole.ADMIN])
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class AdminDockerRegistryController {
  constructor(private readonly dockerRegistryService: DockerRegistryService) {}

  @Post()
  @HttpCode(201)
  @ApiOperation({
    summary: 'Create registry',
    operationId: 'adminCreateRegistry',
  })
  @ApiResponse({
    status: 201,
    type: DockerRegistryDto,
  })
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.DOCKER_REGISTRY,
    targetIdFromResult: (result: DockerRegistryDto) => result?.id,
    requestMetadata: {
      body: (req: TypedRequest<AdminCreateDockerRegistryDto>) => ({
        name: req.body?.name,
        username: req.body?.username,
        password: req.body?.password ? MASKED_AUDIT_VALUE : undefined,
        url: req.body?.url,
        project: req.body?.project,
        registryType: req.body?.registryType,
        isActive: req.body?.isActive,
        isFallback: req.body?.isFallback,
      }),
    },
  })
  async create(@Body() createDto: AdminCreateDockerRegistryDto): Promise<DockerRegistryDto> {
    const dockerRegistry = await this.dockerRegistryService.create(createDto)
    return DockerRegistryDto.fromDockerRegistry(dockerRegistry)
  }

  @Get()
  @HttpCode(200)
  @ApiOperation({
    summary: 'List registries',
    operationId: 'adminListRegistries',
  })
  @ApiResponse({
    status: 200,
    type: [DockerRegistryDto],
  })
  @ApiQuery({
    name: 'organizationId',
    description: 'Filter registries by organization ID',
    type: String,
    required: false,
  })
  @ApiQuery({
    name: 'region',
    description: 'Filter registries by region name (organization ID is required)',
    type: String,
    required: false,
  })
  async findAll(
    @Query('organizationId') organizationId?: string,
    @Query('region') region?: string,
  ): Promise<DockerRegistryDto[]> {
    if (!organizationId && region) {
      throw new BadRequestException('Must provide organization ID when filtering by region name')
    }
    const dockerRegistries = await this.dockerRegistryService.findAll(organizationId, region)
    return dockerRegistries.map(DockerRegistryDto.fromDockerRegistry)
  }

  @Get(':id')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Get registry',
    operationId: 'adminGetRegistry',
  })
  @ApiParam({
    name: 'id',
    description: 'Registry ID',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    type: DockerRegistryDto,
  })
  async findOne(@Param('id') registryId: string): Promise<DockerRegistryDto> {
    const registry = await this.dockerRegistryService.findOne(registryId)
    if (!registry) {
      throw new NotFoundException('Registry not found')
    }
    return DockerRegistryDto.fromDockerRegistry(registry)
  }

  @Patch(':id')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Update registry',
    operationId: 'adminUpdateRegistry',
  })
  @ApiParam({
    name: 'id',
    description: 'Registry ID',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    type: DockerRegistryDto,
  })
  @Audit({
    action: AuditAction.UPDATE,
    targetType: AuditTarget.DOCKER_REGISTRY,
    targetIdFromRequest: (req) => req.params.id,
    requestMetadata: {
      body: (req: TypedRequest<AdminUpdateDockerRegistryDto>) => ({
        name: req.body?.name,
        url: req.body?.url,
        username: req.body?.username,
        password: req.body?.password ? MASKED_AUDIT_VALUE : undefined,
        project: req.body?.project,
        isActive: req.body?.isActive,
        isFallback: req.body?.isFallback,
      }),
    },
  })
  async update(
    @Param('id') registryId: string,
    @Body() updateDto: AdminUpdateDockerRegistryDto,
  ): Promise<DockerRegistryDto> {
    const dockerRegistry = await this.dockerRegistryService.update(registryId, updateDto)
    return DockerRegistryDto.fromDockerRegistry(dockerRegistry)
  }

  @Delete(':id')
  @HttpCode(204)
  @ApiOperation({
    summary: 'Delete registry',
    operationId: 'adminDeleteRegistry',
  })
  @ApiParam({
    name: 'id',
    description: 'Registry ID',
    type: 'string',
  })
  @ApiResponse({
    status: 204,
  })
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.DOCKER_REGISTRY,
    targetIdFromRequest: (req) => req.params.id,
  })
  async remove(@Param('id') registryId: string): Promise<void> {
    return this.dockerRegistryService.remove(registryId)
  }
}
