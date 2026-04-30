/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger'
import { IsBoolean, IsOptional, IsString, MinLength } from 'class-validator'

export class SetSandboxErrorStateDto {
  @IsString()
  @MinLength(1)
  @ApiProperty({
    description: 'Custom error message stored on the sandbox',
    example: 'Sandbox encountered an unrecoverable runtime error.',
  })
  errorReason: string

  @IsOptional()
  @IsBoolean()
  @ApiPropertyOptional({
    description: 'Whether the sandbox can be recovered from this error',
    example: false,
    default: false,
  })
  recoverable?: boolean
}
