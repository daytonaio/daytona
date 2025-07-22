/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Body, Controller, Get, Param, Post, Query, UseGuards, Request as Req } from '@nestjs/common'
import { ApiTags, ApiOperation, ApiResponse, ApiBearerAuth, ApiOAuth2, ApiParam, ApiQuery } from '@nestjs/swagger'
import { Request } from 'express'
import { AuditLogDto } from '../dto/audit-log.dto'
import { PaginatedAuditLogsDto } from '../dto/paginated-audit-logs.dto'
import { CreateAuditLogDto } from '../dto/create-audit-log.dto'
import { AuditService } from '../services/audit.service'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { SystemActionGuard } from '../../auth/system-action.guard'
import { CustomHeaders } from '../../common/constants/header.constants'
import { RequiredSystemRole } from '../../common/decorators/required-role.decorator'
import { OrganizationResourceActionGuard } from '../../organization/guards/organization-resource-action.guard'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { SystemRole } from '../../user/enums/system-role.enum'

@ApiTags('audit')
@Controller('audit')
@UseGuards(CombinedAuthGuard, SystemActionGuard, OrganizationResourceActionGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class AuditController {
  constructor(private readonly auditService: AuditService) {}

  @Get()
  @ApiOperation({
    summary: 'Get all audit logs',
    operationId: 'getAllAuditLogs',
  })
  @ApiResponse({
    status: 200,
    description: 'Paginated list of all audit logs',
    type: PaginatedAuditLogsDto,
  })
  @ApiQuery({
    name: 'page',
    required: false,
    type: Number,
    description: 'Page number (default: 1)',
  })
  @ApiQuery({
    name: 'limit',
    required: false,
    type: Number,
    description: 'Number of items per page (default: 10)',
  })
  @RequiredSystemRole(SystemRole.ADMIN)
  async getAllLogs(@Query('page') page = 1, @Query('limit') limit = 10): Promise<PaginatedAuditLogsDto> {
    const result = await this.auditService.getLogs(page, limit)
    return {
      items: result.items.map(AuditLogDto.fromAuditLog),
      total: result.total,
      page: result.page,
      totalPages: result.totalPages,
    }
  }

  @Get('/organizations/:organizationId')
  @ApiOperation({
    summary: 'Get audit logs for organization',
    operationId: 'getOrganizationAuditLogs',
  })
  @ApiResponse({
    status: 200,
    description: 'Paginated list of organization audit logs',
    type: PaginatedAuditLogsDto,
  })
  @ApiParam({
    name: 'organizationId',
    description: 'Organization ID',
    type: 'string',
  })
  @ApiQuery({
    name: 'page',
    required: false,
    type: Number,
    description: 'Page number (default: 1)',
  })
  @ApiQuery({
    name: 'limit',
    required: false,
    type: Number,
    description: 'Number of items per page (default: 10)',
  })
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.READ_AUDIT_LOGS])
  async getOrganizationLogs(
    @Param('organizationId') organizationId: string,
    @Query('page') page = 1,
    @Query('limit') limit = 10,
  ): Promise<PaginatedAuditLogsDto> {
    const result = await this.auditService.getLogs(page, limit, organizationId)
    return {
      items: result.items.map(AuditLogDto.fromAuditLog),
      total: result.total,
      page: result.page,
      totalPages: result.totalPages,
    }
  }

  @Post()
  @ApiOperation({
    summary: 'Create audit log entry',
    operationId: 'createAuditLog',
  })
  @ApiResponse({
    status: 201,
    description: 'Audit log entry created successfully',
    type: AuditLogDto,
  })
  @RequiredSystemRole(SystemRole.ADMIN)
  async createLog(@Req() req: Request, @Body() createAuditLogDto: CreateAuditLogDto): Promise<AuditLogDto> {
    const auditLog = await this.auditService.createLog({
      actorId: createAuditLogDto.actorId,
      actorEmail: createAuditLogDto.actorEmail,
      organizationId: createAuditLogDto.organizationId,
      action: createAuditLogDto.action,
      targetType: createAuditLogDto.targetType,
      targetId: createAuditLogDto.targetId,
      ipAddress: req.ip,
      userAgent: req.get('user-agent'),
      source: req.get(CustomHeaders.SOURCE.name),
    })
    return AuditLogDto.fromAuditLog(auditLog)
  }
}
