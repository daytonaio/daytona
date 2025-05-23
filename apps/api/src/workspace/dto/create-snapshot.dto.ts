/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsArray, IsBoolean, IsOptional, IsString } from 'class-validator'

@ApiSchema({ name: 'CreateSnapshot' })
export class CreateSnapshotDto {
  @ApiProperty({
    description: 'The name of the snapshot',
    example: 'my-docker-snapshot:8.0.1-alpha',
  })
  @IsString()
  name: string

  @ApiPropertyOptional({
    description: 'The entrypoint command for the snapshot',
    example: 'sleep infinity',
  })
  @IsString({
    each: true,
  })
  @IsArray()
  @IsOptional()
  entrypoint?: string[]

  @ApiPropertyOptional({
    description: 'Whether the snapshot is general',
  })
  @IsBoolean()
  @IsOptional()
  general?: boolean
}
