/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsBoolean, IsInt, IsOptional, IsString } from 'class-validator'
import { Type } from 'class-transformer'
import { Min, Max } from 'class-validator'

@ApiSchema({ name: 'ListSandboxesQuery' })
export class ListSandboxesQueryDto {
  @ApiProperty({
    name: 'page',
    description: 'Page number',
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
    description: 'Number of items per page',
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

  @ApiProperty({
    name: 'labels',
    description: 'JSON encoded labels to filter by',
    required: false,
    type: String,
    example: '{"label1": "value1", "label2": "value2"}',
  })
  @IsOptional()
  @IsString()
  labels?: string

  @ApiProperty({
    name: 'includeErroredDeleted',
    description: 'Include errored sandboxes with deleted desired state',
    required: false,
    type: Boolean,
  })
  @IsOptional()
  @Type(() => Boolean)
  @IsBoolean()
  includeErroredDeleted?: boolean
}
