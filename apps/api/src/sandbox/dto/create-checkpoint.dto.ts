/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiPropertyOptional } from '@nestjs/swagger'
import { IsOptional, IsString, MaxLength } from 'class-validator'

export class CreateCheckpointDto {
  @ApiPropertyOptional({
    description: 'Optional name for the checkpoint. If not provided, a name will be auto-generated.',
    example: 'my-checkpoint',
  })
  @IsOptional()
  @IsString()
  @MaxLength(255)
  name?: string
}
