/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'LogEntry' })
export class LogEntryDto {
  @ApiProperty({ description: 'Timestamp of the log entry' })
  timestamp: string

  @ApiProperty({ description: 'Log message body' })
  body: string

  @ApiProperty({ description: 'Severity level text (e.g., INFO, WARN, ERROR)' })
  severityText: string

  @ApiPropertyOptional({ description: 'Severity level number' })
  severityNumber?: number

  @ApiProperty({ description: 'Service name that generated the log' })
  serviceName: string

  @ApiProperty({
    type: 'object',
    description: 'Resource attributes from OTEL',
    additionalProperties: { type: 'string' },
  })
  resourceAttributes: Record<string, string>

  @ApiProperty({ type: 'object', description: 'Log-specific attributes', additionalProperties: { type: 'string' } })
  logAttributes: Record<string, string>

  @ApiPropertyOptional({ description: 'Associated trace ID if available' })
  traceId?: string

  @ApiPropertyOptional({ description: 'Associated span ID if available' })
  spanId?: string
}
