/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { SandboxDto } from './sandbox.dto'
import { IsEnum } from 'class-validator'
import { BackupState as SnapshotState } from '../enums/backup-state.enum'
import { Sandbox } from '../entities/sandbox.entity'

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

  static fromSandbox(sandbox: Sandbox, runnerDomain: string): WorkspaceDto {
    const dto = super.fromSandbox(sandbox, runnerDomain)
    return this.fromSandboxDto(dto)
  }

  static fromSandboxDto(sandboxDto: SandboxDto): WorkspaceDto {
    const sandboxProviderMetadata = JSON.parse(sandboxDto.info?.providerMetadata ?? '{}')
    return {
      ...sandboxDto,
      image: sandboxDto.snapshot,
      snapshotState: sandboxDto.backupState,
      snapshotCreatedAt: sandboxDto.backupCreatedAt,
      info: {
        ...sandboxDto.info,
        providerMetadata: JSON.stringify({
          ...sandboxProviderMetadata,
          nodeDomain: sandboxProviderMetadata.runnerDomain,
          lastSnapshot: sandboxDto.backupCreatedAt,
        }),
      },
    }
  }
}
