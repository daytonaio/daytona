/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'TraceSpan' })
export class TraceSpanDto {
  @ApiProperty({ description: 'Trace identifier' })
  traceId: string

  @ApiProperty({ description: 'Span identifier' })
  spanId: string

  @ApiPropertyOptional({ description: 'Parent span identifier' })
  parentSpanId?: string

  @ApiProperty({ description: 'Span name' })
  spanName: string

  @ApiProperty({ description: 'Span start timestamp' })
  timestamp: string

  @ApiProperty({ description: 'Span duration in nanoseconds' })
  durationNs: number

  @ApiProperty({ type: 'object', description: 'Span attributes', additionalProperties: { type: 'string' } })
  spanAttributes: Record<string, string>

  @ApiPropertyOptional({ description: 'Status code of the span' })
  statusCode?: string

  @ApiPropertyOptional({ description: 'Status message' })
  statusMessage?: string
}
