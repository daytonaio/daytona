/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger'
import { IsBoolean, IsEnum, IsOptional, IsString } from 'class-validator'
import { SandboxState } from '../enums/sandbox-state.enum'

export class UpdateSandboxStateDto {
  @IsEnum(SandboxState)
  @ApiProperty({
    description: 'The new state for the sandbox',
    enum: SandboxState,
    example: SandboxState.STARTED,
  })
  state: SandboxState

  @IsOptional()
  @IsString()
  @ApiPropertyOptional({
    description: 'Optional error message when reporting an error state',
    example: 'Failed to pull snapshot image',
  })
  errorReason?: string

  @IsOptional()
  @IsBoolean()
  @ApiPropertyOptional({
    description: 'Whether the sandbox is recoverable',
    example: true,
  })
  recoverable?: boolean
}
