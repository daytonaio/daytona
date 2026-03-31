/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Get, Post, Body, Patch, Param, Delete, UseGuards, HttpCode, Query } from '@nestjs/common'
import {
  ApiTags,
  ApiOperation,
  ApiResponse,
  ApiOAuth2,
  ApiHeader,
  ApiParam,
  ApiBearerAuth,
  ApiQuery,
} from '@nestjs/swagger'
import { DockerRegistryService } from '../services/docker-registry.service'
import { CreateDockerRegistryDto } from '../dto/create-docker-registry.dto'
import { UpdateDockerRegistryDto } from '../dto/update-docker-registry.dto'
import { DockerRegistryDto } from '../dto/docker-registry.dto'
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import { RegistryPushAccessDto } from '../../sandbox/dto/registry-push-access-dto'
import { DockerRegistryAccessGuard } from '../guards/docker-registry-access.guard'
import { DockerRegistry } from '../decorators/docker-registry.decorator'
import { DockerRegistry as DockerRegistryEntity } from '../entities/docker-registry.entity'
import { CustomHeaders } from '../../common/constants/header.constants'
import { IsOrganizationAuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { OrganizationAuthContextGuard } from '../../organization/guards/organization-auth-context.guard'
import { Audit, MASKED_AUDIT_VALUE, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'
import { RegistryType } from '../enums/registry-type.enum'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'

@Controller('docker-registry')
@ApiTags('docker-registry')
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@AuthStrategy([AuthStrategyType.API_KEY, AuthStrategyType.JWT])
@UseGuards(AuthenticatedRateLimitGuard)
@UseGuards(OrganizationAuthContextGuard)
export class DockerRegistryController {
  constructor(private readonly dockerRegistryService: DockerRegistryService) {}

  @Post()
  @ApiOperation({
    summary: 'Create registry',
    operationId: 'createRegistry',
  })
  @ApiResponse({
    status: 201,
    description: 'The docker registry has been successfully created.',
    type: DockerRegistryDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_REGISTRIES])
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.DOCKER_REGISTRY,
    targetIdFromResult: (result: DockerRegistryDto) => result?.id,
    requestMetadata: {
      body: (req: TypedRequest<CreateDockerRegistryDto>) => ({
        name: req.body?.name,
        username: req.body?.username,
        password: req.body?.password ? MASKED_AUDIT_VALUE : undefined,
        url: req.body?.url,
        project: req.body?.project,
      }),
    },
  })
  async create(
    @IsOrganizationAuthContext() authContext: OrganizationAuthContext,
    @Body() createDockerRegistryDto: CreateDockerRegistryDto,
  ): Promise<DockerRegistryDto> {
    const dockerRegistry = await this.dockerRegistryService.create(
      {
        ...createDockerRegistryDto,
        registryType: RegistryType.ORGANIZATION,
      },
      authContext.organizationId,
    )
    return DockerRegistryDto.fromDockerRegistry(dockerRegistry)
  }

  @Get()
  @ApiOperation({
    summary: 'List registries',
    operationId: 'listRegistries',
  })
  @ApiResponse({
    status: 200,
    description: 'List of all docker registries',
    type: [DockerRegistryDto],
  })
  async findAll(@IsOrganizationAuthContext() authContext: OrganizationAuthContext): Promise<DockerRegistryDto[]> {
    const dockerRegistries = await this.dockerRegistryService.findAll(
      authContext.organizationId,
      // only include registries manually created by the organization
      RegistryType.ORGANIZATION,
    )
    return dockerRegistries.map(DockerRegistryDto.fromDockerRegistry)
  }

  @Get('registry-push-access')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Get temporary registry access for pushing snapshots',
    operationId: 'getTransientPushAccess',
  })
  @ApiQuery({
    name: 'regionId',
    required: false,
    description: 'ID of the region where the snapshot will be available (defaults to organization default region)',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Temporary registry access has been generated',
    type: RegistryPushAccessDto,
  })
  async getTransientPushAccess(
    @IsOrganizationAuthContext() authContext: OrganizationAuthContext,
    @Query('regionId') regionId?: string,
  ): Promise<RegistryPushAccessDto> {
    return this.dockerRegistryService.getRegistryPushAccess(authContext.organizationId, authContext.userId, regionId)
  }

  @Get(':id')
  @ApiOperation({
    summary: 'Get registry',
    operationId: 'getRegistry',
  })
  @ApiParam({
    name: 'id',
    description: 'ID of the docker registry',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'The docker registry',
    type: DockerRegistryDto,
  })
  @UseGuards(DockerRegistryAccessGuard)
  async findOne(@DockerRegistry() registry: DockerRegistryEntity): Promise<DockerRegistryDto> {
    return DockerRegistryDto.fromDockerRegistry(registry)
  }

  @Patch(':id')
  @ApiOperation({
    summary: 'Update registry',
    operationId: 'updateRegistry',
  })
  @ApiParam({
    name: 'id',
    description: 'ID of the docker registry',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'The docker registry has been successfully updated.',
    type: DockerRegistryDto,
  })
  @UseGuards(DockerRegistryAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_REGISTRIES])
  @Audit({
    action: AuditAction.UPDATE,
    targetType: AuditTarget.DOCKER_REGISTRY,
    targetIdFromRequest: (req) => req.params.id,
    requestMetadata: {
      body: (req: TypedRequest<UpdateDockerRegistryDto>) => ({
        name: req.body?.name,
        url: req.body?.url,
        username: req.body?.username,
        password: req.body?.password ? MASKED_AUDIT_VALUE : undefined,
        project: req.body?.project,
      }),
    },
  })
  async update(
    @Param('id') registryId: string,
    @Body() updateDockerRegistryDto: UpdateDockerRegistryDto,
  ): Promise<DockerRegistryDto> {
    const dockerRegistry = await this.dockerRegistryService.update(registryId, updateDockerRegistryDto)
    return DockerRegistryDto.fromDockerRegistry(dockerRegistry)
  }

  @Delete(':id')
  @HttpCode(204)
  @ApiOperation({
    summary: 'Delete registry',
    operationId: 'deleteRegistry',
  })
  @ApiParam({
    name: 'id',
    description: 'ID of the docker registry',
    type: 'string',
  })
  @ApiResponse({
    status: 204,
    description: 'The docker registry has been successfully deleted.',
  })
  @UseGuards(DockerRegistryAccessGuard)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.DELETE_REGISTRIES])
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.DOCKER_REGISTRY,
    targetIdFromRequest: (req) => req.params.id,
  })
  async remove(@Param('id') registryId: string): Promise<void> {
    return this.dockerRegistryService.remove(registryId)
  }
}
