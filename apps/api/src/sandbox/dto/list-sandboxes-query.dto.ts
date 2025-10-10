/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsBoolean, IsOptional, IsString, IsArray, IsEnum } from 'class-validator'
import { Type } from 'class-transformer'
import { SandboxState } from '../enums/sandbox-state.enum'
import { ToArray } from '../../common/decorators/to-array.decorator'
import { PageNumber } from '../../common/decorators/page-number.decorator'
import { PageLimit } from '../../common/decorators/page-limit.decorator'

export enum SandboxSortField {
  LAST_ACTIVITY_AT = 'lastActivityAt',
  CREATED_AT = 'createdAt',
}

export enum SandboxSortDirection {
  ASC = 'asc',
  DESC = 'desc',
}

export const DEFAULT_SANDBOX_SORT_FIELD = SandboxSortField.CREATED_AT
export const DEFAULT_SANDBOX_SORT_DIRECTION = SandboxSortDirection.DESC

const VALID_QUERY_STATES = Object.values(SandboxState).filter((state) => state !== SandboxState.DESTROYED)

@ApiSchema({ name: 'ListSandboxesQuery' })
export class ListSandboxesQueryDto {
  @PageNumber(1)
  page = 1

  @PageLimit(100)
  limit = 100

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
