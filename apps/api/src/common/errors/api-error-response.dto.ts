/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'
import { ApiErrorCode } from './api-error-code.enum'

export class ApiErrorResponseDto {
  @ApiProperty({ example: '/api/snapshots/abc-123' })
  declare path: string

  @ApiProperty({ example: '2026-01-01T00:00:00.000Z' })
  declare timestamp: string

  @ApiProperty({ example: 404 })
  declare statusCode: number

  @ApiProperty({ example: 'DAYTONA_API' })
  declare source: string

  @ApiProperty({
    enum: ApiErrorCode,
    enumName: 'ApiErrorCode',
    required: false,
  })
  declare code?: ApiErrorCode

  @ApiProperty({ example: 'Not Found' })
  declare error: string

  @ApiProperty({ example: 'Snapshot not found' })
  declare message: string
}
