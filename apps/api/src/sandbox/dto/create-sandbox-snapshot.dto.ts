/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsBoolean, IsOptional, IsString } from 'class-validator'

@ApiSchema({ name: 'CreateSandboxSnapshot' })
export class CreateSandboxSnapshotDto {
  @ApiProperty({
    description: 'Name for the new snapshot',
    example: 'my-dev-env-v1',
  })
  @IsString()
  name: string

  @ApiPropertyOptional({
    description: 'Use live mode (optimistic snapshot without pausing the VM). Default is false (safe mode with pause).',
    example: false,
  })
  @IsOptional()
  @IsBoolean()
  live?: boolean
}
