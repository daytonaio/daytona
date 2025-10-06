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
import { SandboxClass } from '../enums/sandbox-class.enum'

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

  private constructor() {
    this.volumeId = ''
    this.mountPath = ''
  }
}

@ApiSchema({ name: 'Sandbox' })
export class SandboxDto {
  @ApiProperty({
    description: 'The ID of the sandbox',
    example: 'sandbox123',
  })
  id: string

  @ApiProperty({
    description: 'The organization ID of the sandbox',
    example: 'organization123',
  })
  organizationId: string

  @ApiProperty({
    description: 'The name of the sandbox',
    example: 'MySandbox',
  })
  name: string

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
    description: 'Whether to block all network access for the sandbox',
    example: false,
  })
  networkBlockAll: boolean

  @ApiPropertyOptional({
    description: 'Comma-separated list of allowed CIDR network addresses for the sandbox',
    example: '192.168.1.0/16,10.0.0.0/24',
  })
  networkAllowList?: string

  @ApiProperty({
    description: 'The target environment for the sandbox',
    example: 'local',
  })
  target: string

  @ApiProperty({
    description: 'The CPU quota for the sandbox',
    example: 2,
  })
  cpu: number

  @ApiProperty({
    description: 'The GPU quota for the sandbox',
    example: 0,
  })
  gpu: number

  @ApiProperty({
    description: 'The memory quota for the sandbox',
    example: 4,
  })
  memory: number

  @ApiProperty({
    description: 'The disk quota for the sandbox',
    example: 10,
  })
  disk: number

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
    description: 'The desired state of the sandbox',
    enum: SandboxDesiredState,
    enumName: 'SandboxDesiredState',
    example: Object.values(SandboxDesiredState)[0],
    required: false,
  })
  @IsEnum(SandboxDesiredState)
  @IsOptional()
  desiredState?: SandboxDesiredState

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
    description:
      'Auto-delete interval in minutes (negative value means disabled, 0 means delete immediately upon stopping)',
    example: 30,
    required: false,
  })
  @IsOptional()
  autoDeleteInterval?: number

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

  @ApiPropertyOptional({
    description: 'The creation timestamp of the sandbox',
    example: '2024-10-01T12:00:00Z',
    required: false,
  })
  @IsOptional()
  createdAt?: string

  @ApiPropertyOptional({
    description: 'The last update timestamp of the sandbox',
    example: '2024-10-01T12:00:00Z',
    required: false,
  })
  @IsOptional()
  updatedAt?: string

  @ApiPropertyOptional({
    description: 'The class of the sandbox',
    enum: SandboxClass,
    example: Object.values(SandboxClass)[0],
    required: false,
    deprecated: true,
  })
  @IsEnum(SandboxClass)
  @IsOptional()
  class?: SandboxClass

  @ApiPropertyOptional({
    description: 'The version of the daemon running in the sandbox',
    example: '1.0.0',
    required: false,
  })
  @IsOptional()
  daemonVersion?: string

  constructor(sandbox: Sandbox) {
    this.id = sandbox.id
    this.organizationId = sandbox.organizationId
    this.name = sandbox.name
    this.target = sandbox.region
    this.snapshot = sandbox.snapshot ?? ''
    this.user = sandbox.osUser
    this.env = sandbox.env
    this.cpu = sandbox.cpu
    this.gpu = sandbox.gpu
    this.memory = sandbox.mem
    this.disk = sandbox.disk
    this.public = sandbox.public
    this.networkBlockAll = sandbox.networkBlockAll
    this.networkAllowList = sandbox.networkAllowList ?? undefined
    this.labels = sandbox.labels ?? {}
    this.volumes = sandbox.volumes
    this.state = SandboxDto.getSandboxState(sandbox)
    this.desiredState = sandbox.desiredState
    this.errorReason = sandbox.errorReason ?? undefined
    this.backupState = sandbox.backupState
    this.backupCreatedAt = sandbox.lastBackupAt?.toISOString()
    this.autoStopInterval = sandbox.autoStopInterval
    this.autoArchiveInterval = sandbox.autoArchiveInterval
    this.autoDeleteInterval = sandbox.autoDeleteInterval
    this.class = sandbox.class
    this.createdAt = sandbox.createdAt?.toISOString()
    this.updatedAt = sandbox.updatedAt?.toISOString()
    this.buildInfo = sandbox.buildInfo ? new BuildInfoDto(sandbox.buildInfo) : undefined
    this.daemonVersion = sandbox.daemonVersion ?? undefined
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

  constructor(labels?: { [key: string]: string }) {
    this.labels = labels ?? {}
  }
}
