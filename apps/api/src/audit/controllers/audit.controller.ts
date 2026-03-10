/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Get, Param, Query, UseGuards } from '@nestjs/common'
import { ApiBearerAuth, ApiOAuth2, ApiOperation, ApiParam, ApiResponse, ApiTags } from '@nestjs/swagger'
import { AuditLogDto } from '../dto/audit-log.dto'
import { PaginatedAuditLogsDto } from '../dto/paginated-audit-logs.dto'
import { AuditService } from '../services/audit.service'
import { OrganizationResourceActionGuard } from '../../organization/guards/organization-resource-action.guard'
import { RequiredOrganizationResourcePermissions } from '../../organization/decorators/required-organization-resource-permissions.decorator'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { ListAuditLogsQueryDto } from '../dto/list-audit-logs-query.dto'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'

@ApiTags('audit')
@Controller('audit')
@AuthStrategy([AuthStrategyType.API_KEY, AuthStrategyType.JWT])
@UseGuards(OrganizationResourceActionGuard, AuthenticatedRateLimitGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class AuditController {
  constructor(private readonly auditService: AuditService) {}

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
  @RequiredOrganizationResourcePermissions([OrganizationResourcePermission.READ_AUDIT_LOGS])
  async getOrganizationLogs(
    @Param('organizationId') organizationId: string,
    @Query() query: ListAuditLogsQueryDto,
  ): Promise<PaginatedAuditLogsDto> {
    const result = await this.auditService.getOrganizationLogs(
      organizationId,
      query.page,
      query.limit,
      {
        from: query.from,
        to: query.to,
      },
      query.nextToken,
    )
    return {
      items: result.items.map(AuditLogDto.fromAuditLog),
      total: result.total,
      page: result.page,
      totalPages: result.totalPages,
      nextToken: result.nextToken,
    }
  }
}
