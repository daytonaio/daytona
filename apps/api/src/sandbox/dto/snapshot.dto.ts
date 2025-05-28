/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { Snapshot } from '../entities/snapshot.entity'

export class SnapshotDto {
  @ApiProperty()
  id: string

  @ApiPropertyOptional()
  organizationId?: string

  @ApiProperty()
  general: boolean

  @ApiProperty()
  name: string

  @ApiProperty()
  imageName: string

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
  lastUsedAt: Date

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
    }
  }
}
