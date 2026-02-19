/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsNotEmpty, IsOptional, IsString, IsUUID, MaxLength } from 'class-validator'

@ApiSchema({ name: 'CreateSandboxFromCheckpoint' })
export class CreateSandboxFromCheckpointDto {
  @ApiProperty({
    description: 'ID of the checkpoint to create the sandbox from',
    example: '550e8400-e29b-41d4-a716-446655440000',
  })
  @IsNotEmpty()
  @IsUUID()
  checkpointId: string

  @ApiPropertyOptional({
    description: 'Name for the new sandbox. If not provided, a name will be auto-generated.',
    example: 'my-sandbox',
  })
  @IsOptional()
  @IsString()
  @MaxLength(255)
  name?: string
}
