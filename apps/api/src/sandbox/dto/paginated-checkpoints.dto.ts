/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { CheckpointDto } from './checkpoint.dto'

@ApiSchema({ name: 'PaginatedCheckpoints' })
export class PaginatedCheckpointsDto {
  @ApiProperty({ type: [CheckpointDto] })
  items: CheckpointDto[]

  @ApiProperty()
  total: number

  @ApiProperty()
  page: number

  @ApiProperty()
  totalPages: number
}
