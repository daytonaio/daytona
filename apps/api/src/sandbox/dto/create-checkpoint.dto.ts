/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'
import { IsNotEmpty, IsString, MaxLength } from 'class-validator'

export class CreateCheckpointDto {
  @ApiProperty({
    description: 'Name for the checkpoint',
    example: 'my-checkpoint',
  })
  @IsNotEmpty()
  @IsString()
  @MaxLength(255)
  name: string
}
