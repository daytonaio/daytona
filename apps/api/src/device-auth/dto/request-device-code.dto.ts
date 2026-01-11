/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsString, IsOptional } from 'class-validator'
import { ApiProperty } from '@nestjs/swagger'

export class RequestDeviceCodeDto {
  @ApiProperty({
    description: 'Client ID requesting authorization',
    example: 'daytona-cli',
  })
  @IsString()
  client_id: string

  @ApiProperty({
    description: 'Requested scope (optional)',
    example: 'workspaces:read workspaces:write',
    required: false,
  })
  @IsOptional()
  @IsString()
  scope?: string
}
