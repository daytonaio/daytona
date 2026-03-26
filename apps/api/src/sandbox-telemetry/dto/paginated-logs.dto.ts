/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { LogEntryDto } from './log-entry.dto'

@ApiSchema({ name: 'PaginatedLogs' })
export class PaginatedLogsDto {
  @ApiProperty({ type: [LogEntryDto], description: 'List of log entries' })
  items: LogEntryDto[]

  @ApiProperty({ description: 'Total number of log entries matching the query' })
  total: number

  @ApiProperty({ description: 'Current page number' })
  page: number

  @ApiProperty({ description: 'Total number of pages' })
  totalPages: number
}
