/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsBoolean, IsOptional, IsString, IsArray, IsEnum, IsInt, Min, IsDate } from 'class-validator'
import { Transform, Type } from 'class-transformer'
import { SandboxState } from '../enums/sandbox-state.enum'
import { ToArray } from '../../common/decorators/to-array.decorator'
import { PageLimit } from '../../common/decorators/page-limit.decorator'

export enum SandboxSearchSortField {
  NAME = 'name',
  CPU = 'cpu',
  MEMORY = 'memoryGiB',
  DISK = 'diskGiB',
  LAST_ACTIVITY_AT = 'lastActivityAt',
  CREATED_AT = 'createdAt',
}

export enum SandboxSearchSortDirection {
  ASC = 'asc',
  DESC = 'desc',
}

export const DEFAULT_SANDBOX_SEARCH_SORT_FIELD = SandboxSearchSortField.LAST_ACTIVITY_AT
export const DEFAULT_SANDBOX_SEARCH_SORT_DIRECTION = SandboxSearchSortDirection.DESC

const SEARCH_SANDBOXES_QUERY_VALID_STATES = Object.values(SandboxState).filter(
  (state) => state !== SandboxState.DESTROYED,
)

@ApiSchema({ name: 'SearchSandboxesQuery' })
export class SearchSandboxesQueryDto {
  @ApiProperty({
    name: 'cursor',
    description: 'Pagination cursor from a previous response',
    required: false,
    type: String,
  })
  @IsOptional()
  @IsString()
  cursor?: string

  @PageLimit(100)
  limit = 100

  @ApiProperty({
    name: 'id',
    description: 'Filter by ID prefix (case-insensitive)',
    required: false,
    type: String,
  })
  @IsOptional()
  @IsString()
  id?: string

  @ApiProperty({
    name: 'name',
    description: 'Filter by name prefix (case-insensitive)',
    required: false,
    type: String,
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
  @Transform(({ value }) => value === 'true' || value === true)
  @IsBoolean()
  includeErroredDeleted?: boolean

  @ApiProperty({
    name: 'states',
    description: 'List of states to filter by. Can not be combined with "name"',
    required: false,
    enum: SEARCH_SANDBOXES_QUERY_VALID_STATES,
    isArray: true,
  })
  @IsOptional()
  @ToArray()
  @IsArray()
  @IsEnum(SEARCH_SANDBOXES_QUERY_VALID_STATES, {
    each: true,
    message: `each value must be one of the following values: ${SEARCH_SANDBOXES_QUERY_VALID_STATES.join(', ')}`,
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
    name: 'regionIds',
    description: 'List of regions IDs to filter by',
    required: false,
    type: [String],
  })
  @IsOptional()
  @ToArray()
  @IsArray()
  @IsString({ each: true })
  regionIds?: string[]

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
    name: 'isPublic',
    description: 'Filter by public status',
    required: false,
    type: Boolean,
    default: undefined,
  })
  @IsOptional()
  @Transform(({ value }) => value === 'true' || value === true)
  @IsBoolean()
  isPublic?: boolean

  @ApiProperty({
    name: 'isRecoverable',
    description: 'Filter by recoverable status',
    required: false,
    type: Boolean,
    default: undefined,
  })
  @IsOptional()
  @Transform(({ value }) => value === 'true' || value === true)
  @IsBoolean()
  isRecoverable?: boolean

  @ApiProperty({
    name: 'createdAtAfter',
    description: 'Include items created after this timestamp',
    required: false,
    type: String,
    format: 'date-time',
    example: '2024-01-01T00:00:00Z',
  })
  @IsOptional()
  @Type(() => Date)
  @IsDate()
  createdAtAfter?: Date

  @ApiProperty({
    name: 'createdAtBefore',
    description: 'Include items created before this timestamp',
    required: false,
    type: String,
    format: 'date-time',
    example: '2024-12-31T23:59:59Z',
  })
  @IsOptional()
  @Type(() => Date)
  @IsDate()
  createdAtBefore?: Date

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
    enum: SandboxSearchSortField,
    default: DEFAULT_SANDBOX_SEARCH_SORT_FIELD,
  })
  @IsOptional()
  @IsEnum(SandboxSearchSortField)
  sort = DEFAULT_SANDBOX_SEARCH_SORT_FIELD

  @ApiProperty({
    name: 'order',
    description: 'Direction to sort by',
    required: false,
    enum: SandboxSearchSortDirection,
    default: DEFAULT_SANDBOX_SEARCH_SORT_DIRECTION,
  })
  @IsOptional()
  @IsEnum(SandboxSearchSortDirection)
  order = DEFAULT_SANDBOX_SEARCH_SORT_DIRECTION
}
