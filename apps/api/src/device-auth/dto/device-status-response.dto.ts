/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'

export class DeviceStatusResponseDto {
  @ApiProperty({ example: 'WDJB-MJHT' })
  user_code: string

  @ApiProperty({ example: 'daytona-cli' })
  client_id: string

  @ApiProperty({ example: 'workspaces:read workspaces:write' })
  scope: string

  @ApiProperty({ example: 'pending' })
  status: string

  @ApiProperty({ example: 300 })
  expires_in: number
}
