/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsInt, IsOptional, IsString, IsEnum, Min, Max } from 'class-validator'
import { Type } from 'class-transformer'

export enum SnapshotSortField {
  NAME = 'name',
  STATE = 'state',
  LAST_USED_AT = 'lastUsedAt',
  CREATED_AT = 'createdAt',
}

export enum SnapshotSortDirection {
  ASC = 'asc',
  DESC = 'desc',
}

@ApiSchema({ name: 'ListSnapshotsQuery' })
export class ListSnapshotsQueryDto {
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
    name: 'sort',
    description: 'Field to sort by',
    required: false,
    enum: SnapshotSortField,
    default: SnapshotSortField.LAST_USED_AT,
  })
  @IsOptional()
  @IsEnum(SnapshotSortField)
  sort = SnapshotSortField.LAST_USED_AT

  @ApiProperty({
    name: 'order',
    description: 'Direction to sort by',
    required: false,
    enum: SnapshotSortDirection,
    default: SnapshotSortDirection.DESC,
  })
  @IsOptional()
  @IsEnum(SnapshotSortDirection)
  order = SnapshotSortDirection.DESC
}
