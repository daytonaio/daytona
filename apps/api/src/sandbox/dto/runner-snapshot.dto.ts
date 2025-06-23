/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'

export class RunnerSnapshotDto {
  @ApiProperty({
    description: 'Runner snapshot ID',
    example: '123e4567-e89b-12d3-a456-426614174000',
  })
  runnerSnapshotId: string

  @ApiProperty({
    description: 'Runner ID',
    example: '123e4567-e89b-12d3-a456-426614174000',
  })
  runnerId: string

  @ApiProperty({
    description: 'Runner domain',
    example: 'runner.example.com',
  })
  runnerDomain: string

  constructor(runnerSnapshotId: string, runnerId: string, runnerDomain: string) {
    this.runnerSnapshotId = runnerSnapshotId
    this.runnerId = runnerId
    this.runnerDomain = runnerDomain
  }
}
