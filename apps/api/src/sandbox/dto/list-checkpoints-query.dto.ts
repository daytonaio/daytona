/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsOptional, IsString, IsEnum } from 'class-validator'
import { PageNumber } from '../../common/decorators/page-number.decorator'
import { PageLimit } from '../../common/decorators/page-limit.decorator'

export enum CheckpointSortField {
  NAME = 'name',
  STATE = 'state',
  LAST_USED_AT = 'lastUsedAt',
  CREATED_AT = 'createdAt',
}

export enum CheckpointSortDirection {
  ASC = 'asc',
  DESC = 'desc',
}

@ApiSchema({ name: 'ListCheckpointsQuery' })
export class ListCheckpointsQueryDto {
  @PageNumber(1)
  page = 1

  @PageLimit(100)
  limit = 100

  @ApiProperty({
    name: 'sandboxId',
    description: 'Filter by sandbox ID',
    required: false,
    type: String,
  })
  @IsOptional()
  @IsString()
  sandboxId?: string

  @ApiProperty({
    name: 'sort',
    description: 'Field to sort by',
    required: false,
    enum: CheckpointSortField,
    default: CheckpointSortField.CREATED_AT,
  })
  @IsOptional()
  @IsEnum(CheckpointSortField)
  sort = CheckpointSortField.CREATED_AT

  @ApiProperty({
    name: 'order',
    description: 'Direction to sort by',
    required: false,
    enum: CheckpointSortDirection,
    default: CheckpointSortDirection.DESC,
  })
  @IsOptional()
  @IsEnum(CheckpointSortDirection)
  order = CheckpointSortDirection.DESC
}
