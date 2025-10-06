/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { AuditLogDto } from './audit-log.dto'
import { type PaginatedList } from '../../common/interfaces/paginated-list.interface'

@ApiSchema({ name: 'PaginatedAuditLogs' })
export class PaginatedAuditLogsDto {
  @ApiProperty({ type: [AuditLogDto] })
  items: AuditLogDto[]

  @ApiProperty()
  total: number

  @ApiProperty()
  page: number

  @ApiProperty()
  totalPages: number

  @ApiProperty({ required: false, description: 'Token for next page in cursor-based pagination' })
  nextToken?: string

  constructor(auditLogs: PaginatedList<AuditLogDto>) {
    this.items = auditLogs.items
    this.total = auditLogs.total
    this.page = auditLogs.page
    this.totalPages = auditLogs.totalPages
    this.nextToken = auditLogs.nextToken
  }
}
