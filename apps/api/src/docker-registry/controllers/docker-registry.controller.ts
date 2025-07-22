/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Get, Post, Body, Patch, Param, Delete, UseGuards, HttpCode } from '@nestjs/common'
import { ApiTags, ApiOperation, ApiResponse, ApiOAuth2, ApiHeader, ApiParam, ApiBearerAuth } from '@nestjs/swagger'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { DockerRegistryService } from '../services/docker-registry.service'
import { CreateDockerRegistryDto } from '../dto/create-docker-registry.dto'
import { UpdateDockerRegistryDto } from '../dto/update-docker-registry.dto'
import { DockerRegistryDto } from '../dto/docker-registry.dto'
import { RegistryPushAccessDto } from '../../sandbox/dto/registry-push-access-dto'
import { DockerRegistryAccessGuard } from '../guards/docker-registry-access.guard'
import { DockerRegistry } from '../decorators/docker-registry.decorator'
import { DockerRegistry as DockerRegistryEntity } from '../entities/docker-registry.entity'
import { CustomHeaders } from '../../common/constants/header.constants'
import { AuthContext } from '../../common/decorators/auth-context.decorator'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { OrganizationResourceActionGuard } from '../../organization/guards/organization-resource-action.guard'
import { SystemActionGuard } from '../../auth/system-action.guard'
import { RequiredSystemRole } from '../../common/decorators/required-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'
import { Audit, MASKED_AUDIT_VALUE, TypedRequest } from '../../audit/decorators/audit.decorator'
import { AuditAction } from '../../audit/enums/audit-action.enum'
import { AuditTarget } from '../../audit/enums/audit-target.enum'

@ApiTags('docker-registry')
@Controller('docker-registry')
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@UseGuards(CombinedAuthGuard, SystemActionGuard, OrganizationResourceActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
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
        registryType: req.body?.registryType,
        isDefault: req.body?.isDefault,
      }),
    },
  })
  create(
    @AuthContext() authContext: OrganizationAuthContext,
    @Body() createDockerRegistryDto: CreateDockerRegistryDto,
  ): Promise<DockerRegistryDto> {
    return this.dockerRegistryService.create(createDockerRegistryDto, authContext.organizationId)
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
  findAll(@AuthContext() authContext: OrganizationAuthContext): Promise<DockerRegistryDto[]> {
    return this.dockerRegistryService.findAll(authContext.organizationId)
  }

  @Get('registry-push-access')
  @HttpCode(200)
  @ApiOperation({
    summary: 'Get temporary registry access for pushing snapshots',
    operationId: 'getTransientPushAccess',
  })
  @ApiResponse({
    status: 200,
    description: 'Temporary registry access has been generated',
    type: RegistryPushAccessDto,
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_REGISTRIES])
  async getTransientPushAccess(@AuthContext() authContext: OrganizationAuthContext): Promise<RegistryPushAccessDto> {
    return this.dockerRegistryService.getRegistryPushAccess(authContext.organizationId, authContext.userId)
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
    return registry
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
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_REGISTRIES])
  @UseGuards(DockerRegistryAccessGuard)
  @Audit({
    action: AuditAction.UPDATE,
    targetType: AuditTarget.DOCKER_REGISTRY,
    targetIdFromRequest: (req) => req.params.id,
    requestMetadata: {
      body: (req: TypedRequest<UpdateDockerRegistryDto>) => ({
        name: req.body?.name,
        username: req.body?.username,
        password: req.body?.password ? MASKED_AUDIT_VALUE : undefined,
      }),
    },
  })
  async update(
    @Param('id') registryId: string,
    @Body() updateDockerRegistryDto: UpdateDockerRegistryDto,
  ): Promise<DockerRegistryDto> {
    return this.dockerRegistryService.update(registryId, updateDockerRegistryDto)
  }

  @Delete(':id')
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
  @HttpCode(204)
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.DELETE_REGISTRIES])
  @UseGuards(DockerRegistryAccessGuard)
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.DOCKER_REGISTRY,
    targetIdFromRequest: (req) => req.params.id,
  })
  async remove(@Param('id') registryId: string): Promise<void> {
    return this.dockerRegistryService.remove(registryId)
  }

  @Post(':id/set-default')
  @ApiOperation({
    summary: 'Set default registry',
    operationId: 'setDefaultRegistry',
  })
  @ApiParam({
    name: 'id',
    description: 'ID of the docker registry',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'The docker registry has been set as default.',
    type: DockerRegistryDto,
  })
  @RequiredSystemRole(SystemRole.ADMIN)
  @UseGuards(DockerRegistryAccessGuard)
  @Audit({
    action: AuditAction.SET_DEFAULT,
    targetType: AuditTarget.DOCKER_REGISTRY,
    targetIdFromRequest: (req) => req.params.id,
  })
  async setDefault(@Param('id') registryId: string): Promise<DockerRegistryDto> {
    return this.dockerRegistryService.setDefault(registryId)
  }
}
