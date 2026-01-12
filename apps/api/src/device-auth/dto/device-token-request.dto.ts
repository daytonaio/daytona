/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'
import { IsString } from 'class-validator'

export class DeviceTokenRequestDto {
  @ApiProperty({ example: 'urn:ietf:params:oauth:grant-type:device_code' })
  @IsString()
  grant_type: string

  @ApiProperty({ example: 'GmRhmhcxhwAzkoEqiMEg_DnyEysNkuNhszIySk9eS' })
  @IsString()
  device_code: string

  @ApiProperty({ example: 'daytona-cli' })
  @IsString()
  client_id: string
}
