/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsString, IsBoolean, ValidateIf } from 'class-validator'
import { ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'UpdateSandboxNetworkSettings' })
export class UpdateSandboxNetworkSettingsDto {
  @ApiPropertyOptional({
    description: 'Whether to block all network access for the sandbox',
    example: false,
  })
  @ValidateIf((_, value) => value !== undefined)
  @IsBoolean()
  networkBlockAll?: boolean

  @ApiPropertyOptional({
    description: 'Comma-separated list of allowed CIDR network addresses for the sandbox',
    example: '192.168.1.0/16,10.0.0.0/24',
  })
  @ValidateIf((_, value) => value !== undefined)
  @IsString()
  networkAllowList?: string
}
