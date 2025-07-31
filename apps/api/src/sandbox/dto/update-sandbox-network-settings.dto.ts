/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsOptional, IsString, IsBoolean } from 'class-validator'
import { ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsValidNetworkAllowList } from '../decorators/is-valid-network-allow-list.decorator'

@ApiSchema({ name: 'UpdateSandboxNetworkSettings' })
export class UpdateSandboxNetworkSettingsDto {
  @ApiPropertyOptional({
    description: 'Whether to block all network access for the sandbox',
    example: false,
  })
  @IsOptional()
  @IsBoolean()
  networkBlockAll?: boolean

  @ApiPropertyOptional({
    description: 'Comma-separated list of allowed CIDR network addresses for the sandbox',
    example: '192.168.1.0/16,10.0.0.0/24',
  })
  @IsOptional()
  @IsString()
  @IsValidNetworkAllowList()
  networkAllowList?: string
}
