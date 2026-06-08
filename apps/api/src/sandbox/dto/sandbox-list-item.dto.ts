/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsEnum, IsOptional } from 'class-validator'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { BackupState } from '../enums/backup-state.enum'
import { GpuType } from '../enums/gpu-type.enum'

interface SandboxListItemDtoFields {
  id: string
  organizationId: string
  name: string
  target: string
  runnerId?: string
  sandboxClass?: SandboxClass
  state: SandboxState
  desiredState?: SandboxDesiredState
  snapshot?: string
  user: string
  errorReason?: string
  recoverable?: boolean
  public: boolean
  cpu: number
  gpu: number
  gpuType?: GpuType
  memory: number
  disk: number
  labels: { [key: string]: string }
  backupState?: BackupState
  autoStopInterval?: number
  autoArchiveInterval?: number
  autoDeleteInterval?: number
  createdAt?: string
  updatedAt?: string
  lastActivityAt?: string
  daemonVersion?: string
}

@ApiSchema({ name: 'SandboxListItem' })
export class SandboxListItemDto {
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

  @ApiProperty({
    description: 'The target environment for the sandbox',
    example: 'local',
  })
  target: string

  @ApiPropertyOptional({
    description: 'The runner ID of the sandbox',
    example: 'runner123',
    required: false,
  })
  @IsOptional()
  runnerId?: string

  @ApiPropertyOptional({
    description: 'The class of the sandbox',
    enum: SandboxClass,
    enumName: 'SandboxClass',
    example: Object.values(SandboxClass)[0],
    required: false,
  })
  @IsEnum(SandboxClass)
  @IsOptional()
  sandboxClass?: SandboxClass

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
    description: 'The snapshot used for the sandbox',
    example: 'daytonaio/sandbox:latest',
  })
  snapshot?: string

  @ApiProperty({
    description: 'The user associated with the project',
    example: 'daytona',
  })
  user: string

  @ApiPropertyOptional({
    description: 'The error reason of the sandbox',
    example: 'The sandbox is not running',
    required: false,
  })
  @IsOptional()
  errorReason?: string

  @ApiPropertyOptional({
    description: 'Whether the sandbox error is recoverable.',
    example: true,
    required: false,
  })
  @IsOptional()
  recoverable?: boolean

  @ApiProperty({
    description: 'Whether the sandbox http preview is public',
    example: false,
  })
  public: boolean

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

  @ApiPropertyOptional({
    description: 'The GPU type assigned to the sandbox',
    enum: GpuType,
    enumName: 'GpuType',
    example: GpuType.H100,
  })
  @IsEnum(GpuType)
  @IsOptional()
  gpuType?: GpuType

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

  @ApiProperty({
    description: 'Labels for the sandbox',
    type: 'object',
    additionalProperties: { type: 'string' },
    example: { 'daytona.io/public': 'true' },
  })
  labels: { [key: string]: string }

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
    description: 'The last activity timestamp of the sandbox',
    example: '2024-10-01T12:00:00Z',
    required: false,
  })
  @IsOptional()
  lastActivityAt?: string

  @ApiPropertyOptional({
    description: 'The version of the daemon running in the sandbox',
    example: '1.0.0',
    required: false,
  })
  @IsOptional()
  daemonVersion?: string

  @ApiProperty({
    description: 'The toolbox proxy URL for the sandbox',
    example: 'https://proxy.app.daytona.io/toolbox',
  })
  toolboxProxyUrl: string

  constructor({
    id,
    organizationId,
    name,
    target,
    runnerId,
    sandboxClass,
    state,
    desiredState,
    snapshot,
    user,
    errorReason,
    recoverable,
    public: isPublic,
    cpu,
    gpu,
    gpuType,
    memory,
    disk,
    labels,
    backupState,
    autoStopInterval,
    autoArchiveInterval,
    autoDeleteInterval,
    createdAt,
    updatedAt,
    lastActivityAt,
    daemonVersion,
  }: SandboxListItemDtoFields) {
    this.id = id
    this.organizationId = organizationId
    this.name = name
    this.target = target
    this.runnerId = runnerId
    this.sandboxClass = sandboxClass
    this.state = SandboxListItemDto.deriveState(state, desiredState)
    this.desiredState = desiredState
    this.snapshot = snapshot
    this.user = user
    this.errorReason = errorReason
    this.recoverable = recoverable
    this.public = isPublic
    this.cpu = cpu
    this.gpu = gpu
    this.gpuType = gpuType
    this.memory = memory
    this.disk = disk
    this.labels = labels
    this.backupState = backupState
    this.autoStopInterval = autoStopInterval
    this.autoArchiveInterval = autoArchiveInterval
    this.autoDeleteInterval = autoDeleteInterval
    this.createdAt = createdAt
    this.updatedAt = updatedAt
    this.lastActivityAt = lastActivityAt
    this.daemonVersion = daemonVersion
    this.toolboxProxyUrl = ''
  }

  private static deriveState(state: SandboxState, desiredState?: SandboxDesiredState): SandboxState {
    switch (state) {
      case SandboxState.STARTED:
        if (desiredState === SandboxDesiredState.STOPPED) {
          return SandboxState.STOPPING
        }
        if (desiredState === SandboxDesiredState.DESTROYED) {
          return SandboxState.DESTROYING
        }
        break
      case SandboxState.STOPPED:
        if (desiredState === SandboxDesiredState.STARTED) {
          return SandboxState.STARTING
        }
        if (desiredState === SandboxDesiredState.DESTROYED) {
          return SandboxState.DESTROYING
        }
        if (desiredState === SandboxDesiredState.ARCHIVED) {
          return SandboxState.ARCHIVING
        }
        break
      case SandboxState.ARCHIVED:
        if (desiredState === SandboxDesiredState.STARTED) {
          return SandboxState.RESTORING
        }
        break
      case SandboxState.UNKNOWN:
        if (desiredState === SandboxDesiredState.STARTED) {
          return SandboxState.CREATING
        }
        break
    }
    return state
  }
}
