/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { SandboxDto } from './sandbox.dto'
import { IsEnum, IsOptional } from 'class-validator'
import { BackupState as SnapshotState } from '../enums/backup-state.enum'
import { Sandbox } from '../entities/sandbox.entity'

@ApiSchema({ name: 'SandboxInfo' })
export class SandboxInfoDto {
  @ApiProperty({
    description: 'The creation timestamp of the project',
    example: '2023-10-01T12:00:00Z',
  })
  created: string

  @ApiProperty({
    description: 'Deprecated: The name of the sandbox',
    example: 'MySandbox',
    deprecated: true,
    default: '',
  })
  name: string

  @ApiPropertyOptional({
    description: 'Additional metadata provided by the provider',
    example: '{"key": "value"}',
    required: false,
  })
  @IsOptional()
  providerMetadata?: string

  constructor(created: string, name: string, providerMetadata?: string) {
    this.created = created
    this.name = name
    this.providerMetadata = providerMetadata
  }
}

@ApiSchema({ name: 'Workspace' })
export class WorkspaceDto extends SandboxDto {
  @ApiPropertyOptional({
    description: 'The image used for the workspace',
    example: 'daytonaio/workspace:latest',
  })
  image: string

  @ApiPropertyOptional({
    description: 'The state of the snapshot',
    enum: SnapshotState,
    example: Object.values(SnapshotState)[0],
    required: false,
  })
  @IsEnum(SnapshotState)
  snapshotState?: SnapshotState

  @ApiPropertyOptional({
    description: 'The creation timestamp of the last snapshot',
    example: '2024-10-01T12:00:00Z',
    required: false,
  })
  snapshotCreatedAt?: string

  @ApiPropertyOptional({
    description: 'Additional information about the sandbox',
    type: SandboxInfoDto,
    required: false,
  })
  @IsOptional()
  info?: SandboxInfoDto

  constructor(sandbox: Sandbox) {
    super(sandbox)
    this.image = sandbox.snapshot || ''
  }
}
