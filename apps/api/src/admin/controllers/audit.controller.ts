/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Get, Query, UseGuards } from '@nestjs/common'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'
import { ApiBearerAuth, ApiOAuth2, ApiOperation, ApiResponse, ApiTags } from '@nestjs/swagger'
import { RequiredSystemRole } from '../../user/decorators/required-system-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'
import { AuditService } from '../../audit/services/audit.service'
import { AuditLogDto } from '../../audit/dto/audit-log.dto'
import { PaginatedAuditLogsDto } from '../../audit/dto/paginated-audit-logs.dto'
import { ListAuditLogsQueryDto } from '../../audit/dto/list-audit-logs-query.dto'
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'

@Controller('admin/audit')
@ApiTags('admin')
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
@AuthStrategy([AuthStrategyType.API_KEY, AuthStrategyType.JWT])
@RequiredSystemRole(SystemRole.ADMIN)
@UseGuards(AuthenticatedRateLimitGuard)
export class AdminAuditController {
  constructor(private readonly auditService: AuditService) {}

  @Get()
  @ApiOperation({
    summary: 'Get all audit logs',
    operationId: 'adminGetAllAuditLogs',
  })
  @ApiResponse({
    status: 200,
    description: 'Paginated list of all audit logs',
    type: PaginatedAuditLogsDto,
  })
  async getAllLogs(@Query() query: ListAuditLogsQueryDto): Promise<PaginatedAuditLogsDto> {
    const result = await this.auditService.getAllLogs(
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
