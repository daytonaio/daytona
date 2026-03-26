/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { TraceSummaryDto } from './trace-summary.dto'

@ApiSchema({ name: 'PaginatedTraces' })
export class PaginatedTracesDto {
  @ApiProperty({ type: [TraceSummaryDto], description: 'List of trace summaries' })
  items: TraceSummaryDto[]

  @ApiProperty({ description: 'Total number of traces matching the query' })
  total: number

  @ApiProperty({ description: 'Current page number' })
  page: number

  @ApiProperty({ description: 'Total number of pages' })
  totalPages: number
}
