/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Post, Get, Delete, Param, Body, UseGuards, ForbiddenException, HttpCode } from '@nestjs/common'
import { ApiKeyService } from './api-key.service'
import { CreateApiKeyDto } from './dto/create-api-key.dto'
import { ApiHeader, ApiOAuth2, ApiOperation, ApiResponse, ApiTags, ApiBearerAuth } from '@nestjs/swagger'
import { ApiKeyResponseDto } from './dto/api-key-response.dto'
import { ApiKeyListDto } from './dto/api-key-list.dto'
import { CombinedAuthGuard } from '../auth/combined-auth.guard'
import { CustomHeaders } from '../common/constants/header.constants'
import { AuthContext } from '../common/decorators/auth-context.decorator'
import { AuthContext as IAuthContext } from '../common/interfaces/auth-context.interface'
import { OrganizationAuthContext } from '../common/interfaces/auth-context.interface'
import { OrganizationMemberRole } from '../organization/enums/organization-member-role.enum'
import { OrganizationResourcePermission } from '../organization/enums/organization-resource-permission.enum'
import { OrganizationResourceActionGuard } from '../organization/guards/organization-resource-action.guard'
import { SystemRole } from '../user/enums/system-role.enum'
import { Audit, TypedRequest } from '../audit/decorators/audit.decorator'
import { AuditAction } from '../audit/enums/audit-action.enum'
import { AuditTarget } from '../audit/enums/audit-target.enum'

@ApiTags('api-keys')
@Controller('api-keys')
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@UseGuards(CombinedAuthGuard, OrganizationResourceActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class ApiKeyController {
  constructor(private readonly apiKeyService: ApiKeyService) {}

  @Post()
  @ApiOperation({
    summary: 'Create API key',
    operationId: 'createApiKey',
  })
  @ApiResponse({
    status: 201,
    description: 'API key created successfully.',
    type: ApiKeyResponseDto,
  })
  @Audit({
    action: AuditAction.CREATE,
    targetType: AuditTarget.API_KEY,
    targetIdFromResult: (result: ApiKeyResponseDto) => result?.name,
    requestMetadata: {
      body: (req: TypedRequest<CreateApiKeyDto>) => ({
        name: req.body?.name,
        permissions: req.body?.permissions,
        expiresAt: req.body?.expiresAt,
      }),
    },
  })
  async createApiKey(
    @AuthContext() authContext: OrganizationAuthContext,
    @Body() createApiKeyDto: CreateApiKeyDto,
  ): Promise<ApiKeyResponseDto> {
    this.validateRequestedApiKeyPermissions(authContext, createApiKeyDto.permissions)

    const { apiKey, value } = await this.apiKeyService.createApiKey(
      authContext.organizationId,
      authContext.userId,
      createApiKeyDto.name,
      createApiKeyDto.permissions,
      createApiKeyDto.expiresAt,
    )

    return ApiKeyResponseDto.fromApiKey(apiKey, value)
  }

  @Get()
  @ApiOperation({
    summary: 'List API keys',
    operationId: 'listApiKeys',
  })
  @ApiResponse({
    status: 200,
    description: 'API keys retrieved successfully.',
    type: [ApiKeyListDto],
  })
  @ApiResponse({ status: 500, description: 'Error fetching API keys.' })
  async getApiKeys(@AuthContext() authContext: OrganizationAuthContext): Promise<ApiKeyListDto[]> {
    const apiKeys = await this.apiKeyService.getApiKeys(authContext.organizationId, authContext.userId)
    return apiKeys.map((apiKey) => ApiKeyListDto.fromApiKey(apiKey))
  }

  @Get('current')
  @ApiOperation({
    summary: "Get current API key's details",
    operationId: 'getCurrentApiKey',
  })
  @ApiResponse({
    status: 200,
    description: 'API key retrieved successfully.',
    type: ApiKeyListDto,
  })
  async getCurrentApiKey(@AuthContext() authContext: IAuthContext): Promise<ApiKeyListDto> {
    if (!authContext.apiKey) {
      throw new ForbiddenException('Authenticate with an API key to use this endpoint')
    }

    return ApiKeyListDto.fromApiKey(authContext.apiKey)
  }

  @Get(':name')
  @ApiOperation({
    summary: 'Get API key',
    operationId: 'getApiKey',
  })
  @ApiResponse({
    status: 200,
    description: 'API key retrieved successfully.',
    type: ApiKeyListDto,
  })
  async getApiKey(
    @AuthContext() authContext: OrganizationAuthContext,
    @Param('name') name: string,
  ): Promise<ApiKeyListDto> {
    const apiKey = await this.apiKeyService.getApiKeyByName(authContext.organizationId, authContext.userId, name)
    return ApiKeyListDto.fromApiKey(apiKey)
  }

  @Delete(':name')
  @ApiOperation({
    summary: 'Delete API key',
    operationId: 'deleteApiKey',
  })
  @ApiResponse({ status: 204, description: 'API key deleted successfully.' })
  @HttpCode(204)
  @Audit({
    action: AuditAction.DELETE,
    targetType: AuditTarget.API_KEY,
    targetIdFromRequest: (req) => req.params.name,
  })
  async deleteApiKey(@AuthContext() authContext: OrganizationAuthContext, @Param('name') name: string) {
    await this.apiKeyService.deleteApiKey(authContext.organizationId, authContext.userId, name)
  }

  private validateRequestedApiKeyPermissions(
    authContext: OrganizationAuthContext,
    requestedPermissions: OrganizationResourcePermission[],
  ): void {
    if (authContext.role === SystemRole.ADMIN) {
      return
    }

    if (!authContext.organizationUser) {
      throw new ForbiddenException(`Insufficient permissions for assigning: ${requestedPermissions.join(', ')}`)
    }

    if (authContext.organizationUser.role === OrganizationMemberRole.OWNER) {
      return
    }

    const organizationUserPermissions = new Set(
      authContext.organizationUser.assignedRoles.flatMap((role) => role.permissions),
    )

    const forbiddenPermissions = requestedPermissions.filter(
      (permission) => !organizationUserPermissions.has(permission),
    )

    if (forbiddenPermissions.length) {
      throw new ForbiddenException(`Insufficient permissions for assigning: ${forbiddenPermissions.join(', ')}`)
    }
  }
}
