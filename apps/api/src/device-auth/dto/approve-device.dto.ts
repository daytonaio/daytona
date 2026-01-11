/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsString } from 'class-validator'
import { ApiProperty } from '@nestjs/swagger'

export class ApproveDeviceDto {
  @ApiProperty({
    description: 'User code to approve',
    example: 'WDJB-MJHT',
  })
  @IsString()
  user_code: string

  @ApiProperty({
    description: 'Organization ID to associate with the authorization',
  })
  @IsString()
  organization_id: string
}
