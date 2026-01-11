/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'

export class DeviceTokenResponseDto {
  @ApiProperty({ example: 'dtn_xxxxxxxxxxxxxxxxxxxxx' })
  access_token: string

  @ApiProperty({ example: 'Bearer' })
  token_type: string

  @ApiProperty({ example: 31536000 })
  expires_in: number

  @ApiProperty({ example: 'workspaces:read workspaces:write' })
  scope: string

  @ApiProperty({ example: 'org-abc123' })
  organization_id: string

  @ApiProperty({ example: 'my-organization' })
  organization_name: string
}

export class DeviceTokenErrorDto {
  @ApiProperty({ example: 'authorization_pending' })
  error: string

  @ApiProperty({ example: 'The authorization request is still pending', required: false })
  error_description?: string
}
