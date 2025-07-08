/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { Snapshot } from '../entities/snapshot.entity'
import { BuildInfoDto } from './build-info.dto'
import { SnapshotTargetPropagationDto } from './snapshot-target-propagation.dto'

export class SnapshotDto {
  @ApiProperty()
  id: string

  @ApiPropertyOptional()
  organizationId?: string

  @ApiProperty()
  general: boolean

  @ApiProperty()
  name: string

  @ApiPropertyOptional()
  imageName?: string

  @ApiProperty()
  enabled: boolean

  @ApiProperty({
    enum: SnapshotState,
    enumName: 'SnapshotState',
  })
  state: SnapshotState

  @ApiProperty({ nullable: true })
  size?: number

  @ApiProperty({ nullable: true })
  entrypoint?: string[]

  @ApiProperty()
  cpu: number

  @ApiProperty()
  gpu: number

  @ApiProperty()
  mem: number

  @ApiProperty()
  disk: number

  @ApiProperty({ nullable: true })
  errorReason?: string

  @ApiProperty()
  createdAt: Date

  @ApiProperty()
  updatedAt: Date

  @ApiProperty({ nullable: true })
  lastUsedAt?: Date

  @ApiPropertyOptional({
    description: 'Build information for the snapshot',
    type: BuildInfoDto,
  })
  buildInfo?: BuildInfoDto

  @ApiProperty({
    description: 'Target propagations for the snapshot',
    type: [SnapshotTargetPropagationDto],
  })
  targetPropagations: SnapshotTargetPropagationDto[]

  static fromSnapshot(snapshot: Snapshot): SnapshotDto {
    return {
      id: snapshot.id,
      organizationId: snapshot.organizationId,
      general: snapshot.general,
      name: snapshot.name,
      imageName: snapshot.imageName,
      enabled: snapshot.enabled,
      state: snapshot.state,
      size: snapshot.size,
      entrypoint: snapshot.entrypoint,
      cpu: snapshot.cpu,
      gpu: snapshot.gpu,
      mem: snapshot.mem,
      disk: snapshot.disk,
      errorReason: snapshot.errorReason,
      createdAt: snapshot.createdAt,
      updatedAt: snapshot.updatedAt,
      lastUsedAt: snapshot.lastUsedAt,
      buildInfo: snapshot.buildInfo
        ? {
            dockerfileContent: snapshot.buildInfo.dockerfileContent,
            contextHashes: snapshot.buildInfo.contextHashes,
            createdAt: snapshot.buildInfo.createdAt,
            updatedAt: snapshot.buildInfo.updatedAt,
          }
        : undefined,
      targetPropagations: snapshot.targetPropagations?.map((propagation) => {
        return {
          id: propagation.id,
          target: propagation.target,
          desiredConcurrentSandboxes: propagation.desiredConcurrentSandboxes,
          userOverride: propagation.userOverride,
          snapshotId: propagation.snapshotId,
          createdAt: propagation.createdAt,
          updatedAt: propagation.updatedAt,
        }
      }),
    }
  }
}
