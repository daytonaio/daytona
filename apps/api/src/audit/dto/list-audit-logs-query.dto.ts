/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsInt, IsOptional, Min, Max } from 'class-validator'
import { Type } from 'class-transformer'

@ApiSchema({ name: 'ListAuditLogsQuery' })
export class ListAuditLogsQueryDto {
  @ApiProperty({
    name: 'page',
    description: 'Page number of the results',
    required: false,
    type: Number,
    minimum: 1,
    default: 1,
  })
  @IsOptional()
  @Type(() => Number)
  @IsInt()
  @Min(1)
  page = 1

  @ApiProperty({
    name: 'limit',
    description: 'Number of results per page',
    required: false,
    type: Number,
    minimum: 1,
    maximum: 100,
    default: 10,
  })
  @IsOptional()
  @Type(() => Number)
  @IsInt()
  @Min(1)
  @Max(100)
  limit = 10
}
