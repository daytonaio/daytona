/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'
import { IsBoolean, IsInt, IsOptional, IsString, IsArray, IsEnum, IsDate, Min } from 'class-validator'
import { Type } from 'class-transformer'
import { SandboxState } from '../enums/sandbox-state.enum'
import { ToArray } from '../../common/decorators/to-array.decorator'
import { PageNumber } from '../../common/decorators/page-number.decorator'
import { PageLimit } from '../../common/decorators/page-limit.decorator'

export enum SandboxSortField {
  ID = 'id',
  NAME = 'name',
  STATE = 'state',
  SNAPSHOT = 'snapshot',
  REGION = 'region',
  UPDATED_AT = 'updatedAt',
  CREATED_AT = 'createdAt',
}

export enum SandboxSortDirection {
  ASC = 'asc',
  DESC = 'desc',
}

export const DEFAULT_SANDBOX_SORT_FIELD = SandboxSortField.CREATED_AT
export const DEFAULT_SANDBOX_SORT_DIRECTION = SandboxSortDirection.DESC

const VALID_QUERY_STATES = Object.values(SandboxState).filter((state) => state !== SandboxState.DESTROYED)

export class ListSandboxesQueryDeprecatedDto {
  @PageNumber(1)
  page = 1

  @PageLimit(100)
  limit = 100

  @ApiProperty({
    name: 'id',
    description: 'Filter by partial ID match',
    required: false,
    type: String,
    example: 'abc123',
  })
  @IsOptional()
  @IsString()
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
    name: 'minCpu',
    description: 'Minimum CPU',
    required: false,
    type: Number,
    minimum: 1,
  })
  @IsOptional()
  @Type(() => Number)
  @IsInt()
  @Min(1)
  minCpu?: number

  @ApiProperty({
    name: 'maxCpu',
    description: 'Maximum CPU',
    required: false,
    type: Number,
    minimum: 1,
  })
  @IsOptional()
  @Type(() => Number)
  @IsInt()
  @Min(1)
  maxCpu?: number

  @ApiProperty({
    name: 'minMemoryGiB',
    description: 'Minimum memory in GiB',
    required: false,
    type: Number,
    minimum: 1,
  })
  @IsOptional()
  @Type(() => Number)
  @IsInt()
  @Min(1)
  minMemoryGiB?: number

  @ApiProperty({
    name: 'maxMemoryGiB',
    description: 'Maximum memory in GiB',
    required: false,
    type: Number,
    minimum: 1,
  })
  @IsOptional()
  @Type(() => Number)
  @IsInt()
  @Min(1)
  maxMemoryGiB?: number

  @ApiProperty({
    name: 'minDiskGiB',
    description: 'Minimum disk space in GiB',
    required: false,
    type: Number,
    minimum: 1,
  })
  @IsOptional()
  @Type(() => Number)
  @IsInt()
  @Min(1)
  minDiskGiB?: number

  @ApiProperty({
    name: 'maxDiskGiB',
    description: 'Maximum disk space in GiB',
    required: false,
    type: Number,
    minimum: 1,
  })
  @IsOptional()
  @Type(() => Number)
  @IsInt()
  @Min(1)
  maxDiskGiB?: number

  @ApiProperty({
    name: 'lastEventAfter',
    description: 'Include items with last event after this timestamp',
    required: false,
    type: String,
    format: 'date-time',
    example: '2024-01-01T00:00:00Z',
  })
  @IsOptional()
  @Type(() => Date)
  @IsDate()
  lastEventAfter?: Date

  @ApiProperty({
    name: 'lastEventBefore',
    description: 'Include items with last event before this timestamp',
    required: false,
    type: String,
    format: 'date-time',
    example: '2024-12-31T23:59:59Z',
  })
  @IsOptional()
  @Type(() => Date)
  @IsDate()
  lastEventBefore?: Date

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
