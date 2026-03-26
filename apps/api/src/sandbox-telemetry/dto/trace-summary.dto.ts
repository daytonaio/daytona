/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'TraceSummary' })
export class TraceSummaryDto {
  @ApiProperty({ description: 'Unique trace identifier' })
  traceId: string

  @ApiProperty({ description: 'Name of the root span' })
  rootSpanName: string

  @ApiProperty({ description: 'Trace start time' })
  startTime: string

  @ApiProperty({ description: 'Trace end time' })
  endTime: string

  @ApiProperty({ description: 'Total duration in milliseconds' })
  durationMs: number

  @ApiProperty({ description: 'Number of spans in this trace' })
  spanCount: number

  @ApiPropertyOptional({ description: 'Status code of the trace' })
  statusCode?: string
}
