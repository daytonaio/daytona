/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsOptional, IsString, IsEnum } from 'class-validator'
import { PageNumber } from '../../common/decorators/page-number.decorator'
import { PageLimit } from '../../common/decorators/page-limit.decorator'

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
  @PageNumber(1)
  page = 1

  @PageLimit(10)
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
