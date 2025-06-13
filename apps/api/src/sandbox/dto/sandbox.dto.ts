/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { SandboxState } from '../enums/sandbox-state.enum'
import { IsEnum, IsOptional } from 'class-validator'
import { BackupState } from '../enums/backup-state.enum'
import { Sandbox } from '../entities/sandbox.entity'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { BuildInfoDto } from './build-info.dto'

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

@ApiSchema({ name: 'SandboxVolume' })
export class SandboxVolume {
  @ApiProperty({
    description: 'The ID of the volume',
    example: 'volume123',
  })
  volumeId: string

  @ApiProperty({
    description: 'The mount path for the volume',
    example: '/data',
  })
  mountPath: string
}

@ApiSchema({ name: 'Sandbox' })
export class SandboxDto {
  @ApiProperty({
    description: 'The ID of the sandbox',
    example: 'sandbox123',
  })
  id: string

  @ApiProperty({
    description: 'The name of the sandbox',
    example: 'MySandbox',
    deprecated: true,
    default: '',
  })
  name: string

  @ApiProperty({
    description: 'The organization ID of the sandbox',
    example: 'organization123',
  })
  organizationId: string

  @ApiPropertyOptional({
    description: 'The snapshot used for the sandbox',
    example: 'daytonaio/sandbox:latest',
  })
  snapshot: string

  @ApiProperty({
    description: 'The user associated with the project',
    example: 'daytona',
  })
  user: string

  @ApiProperty({
    description: 'Environment variables for the sandbox',
    type: 'object',
    additionalProperties: { type: 'string' },
    example: { NODE_ENV: 'production' },
  })
  env: Record<string, string>

  @ApiProperty({
    description: 'Labels for the sandbox',
    type: 'object',
    additionalProperties: { type: 'string' },
    example: { 'daytona.io/public': 'true' },
  })
  labels: { [key: string]: string }

  @ApiProperty({
    description: 'Whether the sandbox http preview is public',
    example: false,
  })
  public: boolean

  @ApiProperty({
    description: 'The target environment for the sandbox',
    example: 'local',
  })
  target: string

  @ApiPropertyOptional({
    description: 'Additional information about the sandbox',
    type: SandboxInfoDto,
    required: false,
  })
  @IsOptional()
  info?: SandboxInfoDto

  @ApiPropertyOptional({
    description: 'The CPU quota for the sandbox',
    example: 2,
    required: false,
  })
  @IsOptional()
  cpu?: number

  @ApiPropertyOptional({
    description: 'The GPU quota for the sandbox',
    example: 0,
    required: false,
  })
  @IsOptional()
  gpu?: number

  @ApiPropertyOptional({
    description: 'The memory quota for the sandbox',
    example: 4,
    required: false,
  })
  @IsOptional()
  memory?: number

  @ApiPropertyOptional({
    description: 'The disk quota for the sandbox',
    example: 10,
    required: false,
  })
  @IsOptional()
  disk?: number

  @ApiPropertyOptional({
    description: 'The state of the sandbox',
    enum: SandboxState,
    enumName: 'SandboxState',
    example: Object.values(SandboxState)[0],
    required: false,
  })
  @IsEnum(SandboxState)
  @IsOptional()
  state?: SandboxState

  @ApiPropertyOptional({
    description: 'The error reason of the sandbox',
    example: 'The sandbox is not running',
    required: false,
  })
  @IsOptional()
  errorReason?: string

  @ApiPropertyOptional({
    description: 'The state of the backup',
    enum: BackupState,
    example: Object.values(BackupState)[0],
    required: false,
  })
  @IsEnum(BackupState)
  @IsOptional()
  backupState?: BackupState

  @ApiPropertyOptional({
    description: 'The creation timestamp of the last backup',
    example: '2024-10-01T12:00:00Z',
    required: false,
  })
  @IsOptional()
  backupCreatedAt?: string

  @ApiPropertyOptional({
    description: 'Auto-stop interval in minutes (0 means disabled)',
    example: 30,
    required: false,
  })
  @IsOptional()
  autoStopInterval?: number

  @ApiPropertyOptional({
    description: 'Auto-archive interval in minutes',
    example: 7 * 24 * 60,
    required: false,
  })
  @IsOptional()
  autoArchiveInterval?: number

  @ApiPropertyOptional({
    description: 'Array of volumes attached to the sandbox',
    type: [SandboxVolume],
    required: false,
  })
  @IsOptional()
  volumes?: SandboxVolume[]

  @ApiPropertyOptional({
    description: 'Build information for the sandbox',
    type: BuildInfoDto,
    required: false,
  })
  @IsOptional()
  buildInfo?: BuildInfoDto

  constructor() {
    if (this.name === '') {
      this.name = this.id
    }
  }

  static fromSandbox(sandbox: Sandbox, runnerDomain: string): SandboxDto {
    return {
      id: sandbox.id,
      name: sandbox.id,
      organizationId: sandbox.organizationId,
      target: sandbox.region,
      snapshot: sandbox.snapshot,
      user: sandbox.osUser,
      env: sandbox.env,
      cpu: sandbox.cpu,
      gpu: sandbox.gpu,
      memory: sandbox.mem,
      disk: sandbox.disk,
      public: sandbox.public,
      labels: sandbox.labels,
      volumes: sandbox.volumes,
      state: this.getSandboxState(sandbox),
      errorReason: sandbox.errorReason,
      backupState: sandbox.backupState,
      backupCreatedAt: sandbox.lastBackupAt?.toISOString(),
      autoStopInterval: sandbox.autoStopInterval,
      autoArchiveInterval: sandbox.autoArchiveInterval,
      buildInfo: sandbox.buildInfo
        ? {
            dockerfileContent: sandbox.buildInfo.dockerfileContent,
            contextHashes: sandbox.buildInfo.contextHashes,
            createdAt: sandbox.buildInfo.createdAt,
            updatedAt: sandbox.buildInfo.updatedAt,
          }
        : undefined,
      info: {
        name: sandbox.id,
        created: sandbox.createdAt?.toISOString(),
        providerMetadata: JSON.stringify({
          state: this.getSandboxState(sandbox),
          runnerDomain: runnerDomain,
          region: sandbox.region,
          class: sandbox.class,
          updatedAt: sandbox.updatedAt?.toISOString(),
          lastBackup: sandbox.lastBackupAt,
          cpu: sandbox.cpu,
          gpu: sandbox.gpu,
          memory: sandbox.mem,
          disk: sandbox.disk,
          autoStopInterval: sandbox.autoStopInterval,
          autoArchiveInterval: sandbox.autoArchiveInterval,
        }),
      },
    }
  }

  private static getSandboxState(sandbox: Sandbox): SandboxState {
    switch (sandbox.state) {
      case SandboxState.STARTED:
        if (sandbox.desiredState === SandboxDesiredState.STOPPED) {
          return SandboxState.STOPPING
        }
        if (sandbox.desiredState === SandboxDesiredState.DESTROYED) {
          return SandboxState.DESTROYING
        }
        break
      case SandboxState.STOPPED:
        if (sandbox.desiredState === SandboxDesiredState.STARTED) {
          return SandboxState.STARTING
        }
        if (sandbox.desiredState === SandboxDesiredState.DESTROYED) {
          return SandboxState.DESTROYING
        }
        if (sandbox.desiredState === SandboxDesiredState.ARCHIVED) {
          return SandboxState.ARCHIVING
        }
        break
      case SandboxState.UNKNOWN:
        if (sandbox.desiredState === SandboxDesiredState.STARTED) {
          return SandboxState.CREATING
        }
        break
    }
    return sandbox.state
  }
}

@ApiSchema({ name: 'SandboxLabels' })
export class SandboxLabelsDto {
  @ApiProperty({
    description: 'Key-value pairs of labels',
    example: { environment: 'dev', team: 'backend' },
    type: 'object',
    additionalProperties: { type: 'string' },
  })
  labels: { [key: string]: string }
}
