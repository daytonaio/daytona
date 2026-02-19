/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsNotEmpty, IsString, MaxLength } from 'class-validator'

@ApiSchema({ name: 'CreateSnapshotFromCheckpoint' })
export class CreateSnapshotFromCheckpointDto {
  @ApiProperty({
    description: 'The name for the new snapshot',
    example: 'my-snapshot',
  })
  @IsString()
  @IsNotEmpty()
  @MaxLength(255)
  name: string
}
