/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsBoolean, IsOptional, IsString, IsArray, IsEnum, IsUUID } from 'class-validator'
import { Type } from 'class-transformer'
import { SandboxState } from '../enums/sandbox-state.enum'
import { ToArray } from '../../common/decorators/to-array.decorator'
import { PageNumber } from '../../common/decorators/page-number.decorator'
import { PageLimit } from '../../common/decorators/page-limit.decorator'
import {
  DEFAULT_SANDBOX_SORT_DIRECTION,
  DEFAULT_SANDBOX_SORT_FIELD,
  SandboxSortDirection,
  SandboxSortField,
  VALID_QUERY_STATES,
} from './list-sandboxes-query.dto'

@ApiSchema({ name: 'SearchSandboxesQuery' })
export class SearchSandboxesQueryDto {
  @PageNumber(1)
  page = 1

  @PageLimit(100)
  limit = 100

  @ApiProperty({
    name: 'id',
    description: 'Filter by exact ID match',
    required: false,
    type: String,
    example: '00000000-0000-0000-0000-000000000000',
  })
  @IsOptional()
  @IsString()
  @IsUUID()
  id?: string

  @ApiProperty({
    name: 'name',
    description: 'Filter by partial name match',
    required: false,
    type: String,
    example: 'abc123',
  })
  @IsOptional()
  @IsString()
  name?: string

  @ApiProperty({
    name: 'labels',
    description: 'JSON encoded labels to filter by',
    required: false,
    type: String,
    example: '{"label1": "value1", "label2": "value2"}',
    deprecated: true,
  })
  @IsOptional()
  @IsString()
  labels?: string

  @ApiProperty({
    name: 'includeErroredDeleted',
    description: 'Include results with errored state and deleted desired state',
    required: false,
    type: Boolean,
    default: false,
  })
  @IsOptional()
  @Type(() => Boolean)
  @IsBoolean()
  includeErroredDeleted?: boolean

  @ApiProperty({
    name: 'states',
    description: 'List of states to filter by',
    required: false,
    enum: VALID_QUERY_STATES,
    isArray: true,
  })
  @IsOptional()
  @ToArray()
  @IsArray()
  @IsEnum(VALID_QUERY_STATES, {
    each: true,
    message: `each value must be one of the following values: ${VALID_QUERY_STATES.join(', ')}`,
  })
  states?: SandboxState[]

  @ApiProperty({
    name: 'snapshots',
    description: 'List of snapshot names to filter by',
    required: false,
    type: [String],
  })
  @IsOptional()
  @ToArray()
  @IsArray()
  @IsString({ each: true })
  snapshots?: string[]

  @ApiProperty({
    name: 'regions',
    description: 'List of regions to filter by',
    required: false,
    type: [String],
  })
  @IsOptional()
  @ToArray()
  @IsArray()
  @IsString({ each: true })
  regions?: string[]

  @ApiProperty({
    name: 'sort',
    description: 'Field to sort by',
    required: false,
    enum: SandboxSortField,
    default: DEFAULT_SANDBOX_SORT_FIELD,
  })
  @IsOptional()
  @IsEnum(SandboxSortField)
  sort = DEFAULT_SANDBOX_SORT_FIELD

  @ApiProperty({
    name: 'order',
    description: 'Direction to sort by',
    required: false,
    enum: SandboxSortDirection,
    default: DEFAULT_SANDBOX_SORT_DIRECTION,
  })
  @IsOptional()
  @IsEnum(SandboxSortDirection)
  order = DEFAULT_SANDBOX_SORT_DIRECTION
}
