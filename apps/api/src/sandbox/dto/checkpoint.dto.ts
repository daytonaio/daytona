/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger'
import { CheckpointState } from '../enums/checkpoint-state.enum'
import { Checkpoint } from '../entities/checkpoint.entity'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { SandboxVolume } from './sandbox.dto'

export class CheckpointDto {
  @ApiProperty({ description: 'The ID of the checkpoint' })
  id: string

  @ApiProperty({ description: 'The organization ID' })
  organizationId: string

  @ApiProperty({ description: 'The name of the checkpoint' })
  name: string

  @ApiProperty({ description: 'The ID of the sandbox this checkpoint was created from' })
  originSandboxId: string

  @ApiPropertyOptional({ description: 'The image reference in the registry' })
  ref?: string

  @ApiProperty({ description: 'The state of the checkpoint', enum: CheckpointState, enumName: 'CheckpointState' })
  state: CheckpointState

  @ApiPropertyOptional({ description: 'Error reason if in error state' })
  errorReason?: string

  @ApiPropertyOptional({ description: 'Size of the checkpoint in GB' })
  size?: number

  @ApiPropertyOptional({ description: 'When the checkpoint was last used' })
  lastUsedAt?: Date

  @ApiProperty({ description: 'Region captured at checkpoint time' })
  region: string

  @ApiProperty({ description: 'OS user captured at checkpoint time' })
  osUser: string

  @ApiProperty({ description: 'CPU captured at checkpoint time' })
  cpu: number

  @ApiProperty({ description: 'GPU captured at checkpoint time' })
  gpu: number

  @ApiProperty({ description: 'Memory captured at checkpoint time' })
  mem: number

  @ApiProperty({ description: 'Disk captured at checkpoint time' })
  disk: number

  @ApiProperty({ description: 'Environment variables captured at checkpoint time' })
  env: { [key: string]: string }

  @ApiProperty({ description: 'Public status captured at checkpoint time' })
  public: boolean

  @ApiProperty({ description: 'Network block all captured at checkpoint time' })
  networkBlockAll: boolean

  @ApiPropertyOptional({ description: 'Network allow list captured at checkpoint time' })
  networkAllowList?: string

  @ApiPropertyOptional({ description: 'Labels captured at checkpoint time' })
  labels?: { [key: string]: string }

  @ApiProperty({ description: 'Volumes captured at checkpoint time', type: [SandboxVolume] })
  volumes: SandboxVolume[]

  @ApiProperty({ description: 'Sandbox class captured at checkpoint time', enum: SandboxClass })
  class: SandboxClass

  @ApiPropertyOptional({ description: 'Auto-stop interval in minutes' })
  autoStopInterval?: number

  @ApiPropertyOptional({ description: 'Auto-archive interval in minutes' })
  autoArchiveInterval?: number

  @ApiPropertyOptional({ description: 'Auto-delete interval in minutes' })
  autoDeleteInterval?: number

  @ApiProperty({ description: 'When the checkpoint was created' })
  createdAt: Date

  @ApiProperty({ description: 'When the checkpoint was last updated' })
  updatedAt: Date

  static fromCheckpoint(checkpoint: Checkpoint): CheckpointDto {
    return {
      id: checkpoint.id,
      organizationId: checkpoint.organizationId,
      name: checkpoint.name,
      originSandboxId: checkpoint.originSandboxId,
      ref: checkpoint.ref,
      state: checkpoint.state,
      errorReason: checkpoint.errorReason,
      size: checkpoint.size,
      lastUsedAt: checkpoint.lastUsedAt,
      region: checkpoint.region,
      osUser: checkpoint.osUser,
      cpu: checkpoint.cpu,
      gpu: checkpoint.gpu,
      mem: checkpoint.mem,
      disk: checkpoint.disk,
      env: checkpoint.env,
      public: checkpoint.public,
      networkBlockAll: checkpoint.networkBlockAll,
      networkAllowList: checkpoint.networkAllowList,
      labels: checkpoint.labels,
      volumes: checkpoint.volumes || [],
      class: checkpoint.class,
      autoStopInterval: checkpoint.autoStopInterval,
      autoArchiveInterval: checkpoint.autoArchiveInterval,
      autoDeleteInterval: checkpoint.autoDeleteInterval,
      createdAt: checkpoint.createdAt,
      updatedAt: checkpoint.updatedAt,
    }
  }
}
