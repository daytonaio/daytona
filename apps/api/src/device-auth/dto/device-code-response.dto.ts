/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'

export class DeviceCodeResponseDto {
  @ApiProperty({ example: 'GmRhmhcxhwAzkoEqiMEg_DnyEysNkuNhszIySk9eS' })
  device_code: string

  @ApiProperty({ example: 'WDJB-MJHT' })
  user_code: string

  @ApiProperty({ example: 'https://app.daytona.io/device' })
  verification_uri: string

  @ApiProperty({ example: 'https://app.daytona.io/device?user_code=WDJB-MJHT' })
  verification_uri_complete: string

  @ApiProperty({ example: 900 })
  expires_in: number

  @ApiProperty({ example: 5 })
  interval: number
}
