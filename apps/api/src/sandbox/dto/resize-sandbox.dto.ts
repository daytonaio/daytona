/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsOptional, IsNumber, IsBoolean, Min } from 'class-validator'
import { ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'ResizeSandbox' })
export class ResizeSandboxDto {
  @ApiPropertyOptional({
    description: 'CPU cores to allocate to the sandbox (minimum: 1)',
    example: 2,
    type: 'integer',
    minimum: 1,
  })
  @IsOptional()
  @IsNumber()
  @Min(1)
  cpu?: number

  @ApiPropertyOptional({
    description: 'Memory in GB to allocate to the sandbox (minimum: 1)',
    example: 4,
    type: 'integer',
    minimum: 1,
  })
  @IsOptional()
  @IsNumber()
  @Min(1)
  memory?: number

  @ApiPropertyOptional({
    description: 'Disk space in GB to allocate to the sandbox (can only be increased)',
    example: 20,
    type: 'integer',
    minimum: 1,
  })
  @IsOptional()
  @IsNumber()
  @Min(1)
  disk?: number

  @ApiPropertyOptional({
    description: 'GPU units to allocate (not supported - will throw error if provided)',
    example: 0,
    type: 'integer',
  })
  @IsOptional()
  @IsNumber()
  gpu?: number

  @ApiPropertyOptional({
    description:
      'If true, performs hot resize on a running sandbox (only CPU/memory increase allowed). If false, sandbox must be stopped for full resize.',
    example: false,
    default: false,
  })
  @IsOptional()
  @IsBoolean()
  hot?: boolean
}
