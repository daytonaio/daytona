/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger'
import { CheckpointState } from '../enums/checkpoint-state.enum'
import { Checkpoint } from '../entities/checkpoint.entity'

export class CheckpointDto {
  @ApiProperty({
    description: 'The ID of the checkpoint',
    example: 'checkpoint-uuid',
  })
  id: string

  @ApiProperty({
    description: 'The sandbox ID this checkpoint belongs to',
    example: 'sandbox-uuid',
  })
  sandboxId: string

  @ApiProperty({
    description: 'The organization ID',
    example: 'org-uuid',
  })
  organizationId: string

  @ApiProperty({
    description: 'The name of the checkpoint',
    example: 'my-checkpoint',
  })
  name: string

  @ApiPropertyOptional({
    description: 'The image reference in the registry',
    example: 'registry:5000/daytona/org-123/my-checkpoint:latest',
  })
  ref?: string

  @ApiProperty({
    description: 'The state of the checkpoint',
    enum: CheckpointState,
    enumName: 'CheckpointState',
  })
  state: CheckpointState

  @ApiPropertyOptional({
    description: 'Error reason if the checkpoint is in error state',
  })
  errorReason?: string

  @ApiPropertyOptional({
    description: 'Size of the checkpoint image in bytes',
  })
  sizeBytes?: number

  @ApiProperty({
    description: 'When the checkpoint was created',
  })
  createdAt: Date

  @ApiProperty({
    description: 'When the checkpoint was last updated',
  })
  updatedAt: Date

  static fromCheckpoint(checkpoint: Checkpoint): CheckpointDto {
    return {
      id: checkpoint.id,
      sandboxId: checkpoint.sandboxId,
      organizationId: checkpoint.organizationId,
      name: checkpoint.name,
      ref: checkpoint.ref,
      state: checkpoint.state,
      errorReason: checkpoint.errorReason,
      sizeBytes: checkpoint.sizeBytes ? Number(checkpoint.sizeBytes) : undefined,
      createdAt: checkpoint.createdAt,
      updatedAt: checkpoint.updatedAt,
    }
  }
}
