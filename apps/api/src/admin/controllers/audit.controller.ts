/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Get, Query, UseGuards } from '@nestjs/common'
import { ApiBearerAuth, ApiOAuth2, ApiOperation, ApiResponse, ApiTags } from '@nestjs/swagger'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { SystemActionGuard } from '../../user/guards/system-action.guard'
import { RequiredSystemRole } from '../../user/decorators/required-system-role.decorator'
import { SystemRole } from '../../user/enums/system-role.enum'
import { AuditService } from '../../audit/services/audit.service'
import { AuditLogDto } from '../../audit/dto/audit-log.dto'
import { PaginatedAuditLogsDto } from '../../audit/dto/paginated-audit-logs.dto'
import { ListAuditLogsQueryDto } from '../../audit/dto/list-audit-logs-query.dto'

@ApiTags('admin')
@Controller('admin/audit')
@UseGuards(CombinedAuthGuard, SystemActionGuard)
@RequiredSystemRole(SystemRole.ADMIN)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
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
