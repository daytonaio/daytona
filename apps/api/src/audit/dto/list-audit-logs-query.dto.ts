/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { PageNumber } from '../../common/decorators/page-number.decorator'
import { PageLimit } from '../../common/decorators/page-limit.decorator'
import { IsDate, IsOptional, IsString } from 'class-validator'
import { Type } from 'class-transformer'

@ApiSchema({ name: 'ListAuditLogsQuery' })
export class ListAuditLogsQueryDto {
  @PageNumber(1)
  page = 1

  @PageLimit(100)
  limit = 100

  @ApiPropertyOptional({ type: String, format: 'date-time', description: 'From date (ISO 8601 format)' })
  @IsOptional()
  @Type(() => Date)
  @IsDate()
  from?: Date

  @ApiPropertyOptional({ type: String, format: 'date-time', description: 'To date (ISO 8601 format)' })
  @IsOptional()
  @Type(() => Date)
  @IsDate()
  to?: Date

  @ApiPropertyOptional({
    type: String,
    description: 'Token for cursor-based pagination. When provided, takes precedence over page parameter.',
  })
  @IsOptional()
  @IsString()
  nextToken?: string
}
