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

  constructor() {
    super()
  }

  static fromSandbox(sandbox: Sandbox): WorkspaceDto {
    // Send empty string for toolboxProxyUrl as it is not needed in deprecated DTO
    const dto = super.fromSandbox(sandbox, '')
    return this.fromSandboxDto(dto)
  }

  static fromSandboxDto(sandboxDto: SandboxDto): WorkspaceDto {
    return {
      ...sandboxDto,
      image: sandboxDto.snapshot,
      snapshotState: sandboxDto.backupState,
      snapshotCreatedAt: sandboxDto.backupCreatedAt,
      info: {
        name: sandboxDto.name,
        created: sandboxDto.createdAt,
        providerMetadata: JSON.stringify({
          state: sandboxDto.state,
          region: sandboxDto.target,
          class: sandboxDto.class,
          updatedAt: sandboxDto.updatedAt,
          lastSnapshot: sandboxDto.backupCreatedAt,
          cpu: sandboxDto.cpu,
          gpu: sandboxDto.gpu,
          memory: sandboxDto.memory,
          disk: sandboxDto.disk,
          autoStopInterval: sandboxDto.autoStopInterval,
          autoArchiveInterval: sandboxDto.autoArchiveInterval,
          daemonVersion: sandboxDto.daemonVersion,
        }),
      },
    }
  }
}
