/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger'
import { IsDateString, IsOptional, IsArray, IsString, IsNumber, Min } from 'class-validator'
import { Type, Transform } from 'class-transformer'

export class TelemetryQueryParamsDto {
  @ApiProperty({ type: String, format: 'date-time', description: 'Start of time range (ISO 8601)' })
  @IsDateString()
  from: string

  @ApiProperty({ type: String, format: 'date-time', description: 'End of time range (ISO 8601)' })
  @IsDateString()
  to: string

  @ApiPropertyOptional({ type: Number, default: 1, description: 'Page number (1-indexed)' })
  @IsOptional()
  @Type(() => Number)
  @IsNumber()
  @Min(1)
  page?: number = 1

  @ApiPropertyOptional({ type: Number, default: 100, description: 'Number of items per page' })
  @IsOptional()
  @Type(() => Number)
  @IsNumber()
  @Min(1)
  limit?: number = 100
}

export class LogsQueryParamsDto extends TelemetryQueryParamsDto {
  @ApiPropertyOptional({
    type: [String],
    description: 'Filter by severity levels (DEBUG, INFO, WARN, ERROR)',
  })
  @IsOptional()
  @IsArray()
  @IsString({ each: true })
  @Transform(({ value }) => (Array.isArray(value) ? value : [value]))
  severities?: string[]

  @ApiPropertyOptional({ type: String, description: 'Search in log body' })
  @IsOptional()
  @IsString()
  search?: string
}

export class MetricsQueryParamsDto {
  @ApiProperty({ type: String, format: 'date-time', description: 'Start of time range (ISO 8601)' })
  @IsDateString()
  from: string

  @ApiProperty({ type: String, format: 'date-time', description: 'End of time range (ISO 8601)' })
  @IsDateString()
  to: string

  @ApiPropertyOptional({
    type: [String],
    description: 'Filter by metric names',
  })
  @IsOptional()
  @IsArray()
  @IsString({ each: true })
  @Transform(({ value }) => (Array.isArray(value) ? value : [value]))
  metricNames?: string[]
}
