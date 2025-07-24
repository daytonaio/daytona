/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'
import { SnapshotTargetPropagation } from '../entities/snapshot-target-propagation.entity'

export class SnapshotTargetPropagationDto {
  @ApiProperty()
  id: string

  @ApiProperty()
  target: string

  @ApiProperty()
  desiredConcurrentSandboxes: number

  @ApiProperty()
  userOverride: number

  @ApiProperty()
  snapshotId: string

  @ApiProperty()
  createdAt: Date

  @ApiProperty()
  updatedAt: Date

  static fromSnapshotTargetPropagation(propagation: SnapshotTargetPropagation): SnapshotTargetPropagationDto {
    return {
      id: propagation.id,
      target: propagation.target,
      desiredConcurrentSandboxes: propagation.desiredConcurrentSandboxes,
      userOverride: propagation.userOverride,
      snapshotId: propagation.snapshotId,
      createdAt: propagation.createdAt,
      updatedAt: propagation.updatedAt,
    }
  }
}
