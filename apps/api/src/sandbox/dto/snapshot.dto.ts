/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { Snapshot } from '../entities/snapshot.entity'
import { BuildInfoDto } from './build-info.dto'

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

  constructor(snapshot: Snapshot) {
    this.id = snapshot.id
    this.organizationId = snapshot.organizationId ?? undefined
    this.general = snapshot.general
    this.name = snapshot.name
    this.imageName = snapshot.imageName
    this.state = snapshot.state
    this.size = snapshot.size ?? undefined
    this.entrypoint = snapshot.entrypoint ?? undefined
    this.cpu = snapshot.cpu
    this.gpu = snapshot.gpu
    this.mem = snapshot.mem
    this.disk = snapshot.disk
    this.errorReason = snapshot.errorReason ?? undefined
    this.createdAt = snapshot.createdAt
    this.updatedAt = snapshot.updatedAt
    this.lastUsedAt = snapshot.lastUsedAt ?? undefined
    if (snapshot.buildInfo) {
      this.buildInfo = new BuildInfoDto(snapshot.buildInfo)
    }
  }
}
