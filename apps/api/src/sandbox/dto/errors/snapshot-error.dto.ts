/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'

/**
 * Approach B — per-endpoint per-status typed error schemas.
 * Each schema has a narrow single-value code enum. The model type itself
 * (SnapshotNotFoundErrorDto vs SnapshotAccessDeniedErrorDto) identifies the error
 * without reading the code field. Generated SDK clients surface this in e.data (Python/Go).
 */

@ApiSchema({ name: 'SnapshotNotFoundError' })
export class SnapshotNotFoundErrorDto {
  @ApiProperty({ example: 404 })
  statusCode: number

  @ApiProperty({ example: 'DAYTONA_API' })
  source: string

  @ApiProperty({ enum: ['SNAPSHOT_NOT_FOUND'], example: 'SNAPSHOT_NOT_FOUND' })
  code: 'SNAPSHOT_NOT_FOUND'

  @ApiProperty({ example: 'Snapshot "abc-123" not found' })
  message: string

  @ApiProperty({ example: '2026-01-01T00:00:00.000Z' })
  timestamp: string

  @ApiProperty({ example: '/api/snapshots/abc-123' })
  path: string

  @ApiProperty({ example: 'GET' })
  method: string
}

@ApiSchema({ name: 'SnapshotAccessDeniedError' })
export class SnapshotAccessDeniedErrorDto {
  @ApiProperty({ example: 403 })
  statusCode: number

  @ApiProperty({ example: 'DAYTONA_API' })
  source: string

  @ApiProperty({ enum: ['SNAPSHOT_ACCESS_DENIED'], example: 'SNAPSHOT_ACCESS_DENIED' })
  code: 'SNAPSHOT_ACCESS_DENIED'

  @ApiProperty({ example: 'You do not have access to this snapshot' })
  message: string

  @ApiProperty({ example: '2026-01-01T00:00:00.000Z' })
  timestamp: string

  @ApiProperty({ example: '/api/snapshots/abc-123' })
  path: string

  @ApiProperty({ example: 'GET' })
  method: string
}
