/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'
import { ApiErrorCode } from './api-error-code.enum'

/**
 * Approach A — uniform error envelope for all API endpoints.
 * source+code identify the error; the full ApiErrorCode enum lists all possible codes.
 */
export class ApiErrorResponseDto {
  @ApiProperty({ example: 404 })
  statusCode: number

  @ApiProperty({ example: 'DAYTONA_API' })
  source: string

  @ApiProperty({
    enum: ApiErrorCode,
    enumName: 'ApiErrorCode',
    example: ApiErrorCode.NOT_FOUND,
  })
  code: ApiErrorCode

  @ApiProperty({ example: 'Snapshot not found' })
  message: string

  @ApiProperty({ example: '2026-01-01T00:00:00.000Z' })
  timestamp: string

  @ApiProperty({ example: '/api/snapshots/abc-123' })
  path: string

  @ApiProperty({ example: 'GET' })
  method: string
}
