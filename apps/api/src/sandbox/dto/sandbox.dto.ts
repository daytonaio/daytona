/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { SandboxState } from '../enums/sandbox-state.enum'
import { IsEnum, IsNotEmpty, IsOptional, IsString } from 'class-validator'
import { BackupState } from '../enums/backup-state.enum'
import { Sandbox } from '../entities/sandbox.entity'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { BuildInfoDto } from './build-info.dto'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { GpuType } from '../enums/gpu-type.enum'

@ApiSchema({ name: 'SandboxVolume' })
export class SandboxVolume {
  @ApiProperty({
    description: 'The ID or name of the volume. Resolved to the volume ID on sandbox create.',
    example: '3fa85f64-5717-4562-b3fc-2c963f66afa6',
  })
  // Kept as volumeId (not volumeIdOrName): the schema is shared with responses/storage
  // where it is always the UUID. SandboxService.resolveVolumes swaps names for UUIDs
  // before storage/forwarding; the runner rejects non-UUIDs.
  @IsString()
  @IsNotEmpty()
  volumeId: string

  @ApiProperty({
    description: 'The mount path for the volume',
    example: '/data',
  })
  @IsString()
  mountPath: string

  @ApiPropertyOptional({
    description:
      'Optional subpath within the volume to mount. When specified, only this S3 prefix will be accessible. When omitted, the entire volume is mounted.',
    example: 'users/alice',
  })
  @IsOptional()
  @IsString()
  subpath?: string
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

  @ApiPropertyOptional({
    description: 'Comma-separated list of allowed domains for the sandbox',
    example: 'example.com,*.daytona.io',
  })
  domainAllowList?: string

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
    description: 'Whether the sandbox error is recoverable.',
    example: true,
    required: false,
  })
  @IsOptional()
  recoverable?: boolean

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
    description: 'The last activity timestamp of the sandbox',
    example: '2024-10-01T12:00:00Z',
    required: false,
  })
  @IsOptional()
  lastActivityAt?: string

  @ApiPropertyOptional({
    description: 'The class of the sandbox',
    enum: SandboxClass,
    example: SandboxClass.LINUX_VM,
    required: false,
  })
  @IsEnum(SandboxClass)
  @IsOptional()
  sandboxClass?: SandboxClass

  @ApiPropertyOptional({
    description: 'The version of the daemon running in the sandbox',
    example: '1.0.0',
    required: false,
  })
  @IsOptional()
  daemonVersion?: string

  @ApiPropertyOptional({
    description: 'The runner ID of the sandbox',
    example: 'runner123',
    required: false,
  })
  @IsOptional()
  runnerId?: string

  @ApiPropertyOptional({
    description:
      'ID of the sandbox this sandbox is linked to. When set, the sandbox is co-located on the same runner as the linked sandbox.',
    example: 'sandbox123',
    required: false,
  })
  @IsOptional()
  linkedSandboxId?: string

  @ApiProperty({
    description: 'The toolbox proxy URL for the sandbox',
    example: 'https://proxy.app.daytona.io/toolbox',
  })
  toolboxProxyUrl: string

  static fromSandbox(sandbox: Sandbox, toolboxProxyUrl: string): SandboxDto {
    return {
      id: sandbox.id,
      organizationId: sandbox.organizationId,
      name: sandbox.name,
      target: sandbox.region,
      snapshot: sandbox.snapshot,
      user: sandbox.osUser,
      env: sandbox.env,
      cpu: sandbox.cpu,
      gpu: sandbox.gpu,
      gpuType: sandbox.gpuType ?? undefined,
      memory: sandbox.mem,
      disk: sandbox.disk,
      public: sandbox.public,
      networkBlockAll: sandbox.networkBlockAll,
      networkAllowList: sandbox.networkAllowList,
      domainAllowList: sandbox.domainAllowList,
      labels: sandbox.labels,
      volumes: sandbox.volumes,
      state: this.getSandboxState(sandbox),
      desiredState: sandbox.desiredState,
      errorReason: sandbox.errorReason,
      recoverable: sandbox.recoverable,
      backupState: sandbox.backupState,
      backupCreatedAt: sandbox.lastBackupAt ? new Date(sandbox.lastBackupAt).toISOString() : undefined,
      autoStopInterval: sandbox.autoStopInterval,
      autoArchiveInterval: sandbox.autoArchiveInterval,
      autoDeleteInterval: sandbox.autoDeleteInterval,
      sandboxClass: sandbox.sandboxClass,
      createdAt: sandbox.createdAt ? new Date(sandbox.createdAt).toISOString() : undefined,
      updatedAt: sandbox.updatedAt ? new Date(sandbox.updatedAt).toISOString() : undefined,
      lastActivityAt: sandbox.lastActivityAt?.lastActivityAt
        ? new Date(sandbox.lastActivityAt.lastActivityAt).toISOString()
        : undefined,
      buildInfo: sandbox.buildInfo
        ? {
            dockerfileContent: sandbox.buildInfo.dockerfileContent,
            contextHashes: sandbox.buildInfo.contextHashes,
            createdAt: sandbox.buildInfo.createdAt,
            updatedAt: sandbox.buildInfo.updatedAt,
            snapshotRef: sandbox.buildInfo.snapshotRef,
          }
        : undefined,
      daemonVersion: sandbox.daemonVersion,
      runnerId: sandbox.runnerId,
      linkedSandboxId: sandbox.linkedSandboxId ?? undefined,
      toolboxProxyUrl,
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
