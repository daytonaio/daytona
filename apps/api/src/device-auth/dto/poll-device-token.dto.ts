/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsString } from 'class-validator'
import { ApiProperty } from '@nestjs/swagger'

export class PollDeviceTokenDto {
  @ApiProperty({
    description: 'Grant type',
    example: 'urn:ietf:params:oauth:grant-type:device_code',
  })
  @IsString()
  grant_type: string

  @ApiProperty({
    description: 'Device code from initial request',
  })
  @IsString()
  device_code: string

  @ApiProperty({
    description: 'Client ID',
    example: 'daytona-cli',
  })
  @IsString()
  client_id: string
}
