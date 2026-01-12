/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'
import { IsString, IsOptional } from 'class-validator'

export class DeviceCodeRequestDto {
  @ApiProperty({ example: 'daytona-cli', description: 'Client identifier' })
  @IsString()
  client_id: string

  @ApiProperty({ example: 'workspaces:read workspaces:write', required: false })
  @IsString()
  @IsOptional()
  scope?: string
}
