/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { WorkspaceState } from '../enums/workspace-state.enum'
import { IsEnum, IsOptional } from 'class-validator'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { Workspace } from '../entities/workspace.entity'
import { WorkspaceDesiredState } from '../enums/workspace-desired-state.enum'

@ApiSchema({ name: 'WorkspaceInfo' })
export class WorkspaceInfoDto {
  @ApiProperty({
    description: 'The creation timestamp of the project',
    example: '2023-10-01T12:00:00Z',
  })
  created: string

  @ApiProperty({
    description: 'Deprecated: The name of the workspace',
    example: 'MyWorkspace',
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

@ApiSchema({ name: 'WorkspaceVolume' })
export class WorkspaceVolume {
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

@ApiSchema({ name: 'Workspace' })
export class WorkspaceDto {
  @ApiProperty({
    description: 'The ID of the workspace',
    example: 'workspace123',
  })
  id: string

  @ApiProperty({
    description: 'The name of the workspace',
    example: 'MyWorkspace',
    deprecated: true,
    default: '',
  })
  name: string

  @ApiProperty({
    description: 'The organization ID of the workspace',
    example: 'organization123',
  })
  organizationId: string

  @ApiPropertyOptional({
    description: 'The image used for the workspace',
    example: 'daytonaio/workspace:latest',
  })
  image: string

  @ApiProperty({
    description: 'The user associated with the project',
    example: 'daytona',
  })
  user: string

  @ApiProperty({
    description: 'Environment variables for the workspace',
    type: 'object',
    additionalProperties: { type: 'string' },
    example: { NODE_ENV: 'production' },
  })
  env: Record<string, string>

  @ApiProperty({
    description: 'Labels for the workspace',
    type: 'object',
    additionalProperties: { type: 'string' },
    example: { 'daytona.io/public': 'true' },
  })
  labels: { [key: string]: string }

  @ApiProperty({
    description: 'Whether the workspace http preview is public',
    example: false,
  })
  public: boolean

  @ApiProperty({
    description: 'The target environment for the workspace',
    example: 'local',
  })
  target: string

  @ApiPropertyOptional({
    description: 'Additional information about the workspace',
    type: WorkspaceInfoDto,
    required: false,
  })
  @IsOptional()
  info?: WorkspaceInfoDto

  @ApiPropertyOptional({
    description: 'The CPU quota for the workspace',
    example: 2,
    required: false,
  })
  @IsOptional()
  cpu?: number

  @ApiPropertyOptional({
    description: 'The GPU quota for the workspace',
    example: 0,
    required: false,
  })
  @IsOptional()
  gpu?: number

  @ApiPropertyOptional({
    description: 'The memory quota for the workspace',
    example: 4,
    required: false,
  })
  @IsOptional()
  memory?: number

  @ApiPropertyOptional({
    description: 'The disk quota for the workspace',
    example: 10,
    required: false,
  })
  @IsOptional()
  disk?: number

  @ApiPropertyOptional({
    description: 'The state of the workspace',
    enum: WorkspaceState,
    enumName: 'WorkspaceState',
    example: Object.values(WorkspaceState)[0],
    required: false,
  })
  @IsEnum(WorkspaceState)
  @IsOptional()
  state?: WorkspaceState

  @ApiPropertyOptional({
    description: 'The error reason of the workspace',
    example: 'The workspace is not running',
    required: false,
  })
  @IsOptional()
  errorReason?: string

  @ApiPropertyOptional({
    description: 'The state of the snapshot',
    enum: SnapshotState,
    example: Object.values(SnapshotState)[0],
    required: false,
  })
  @IsEnum(SnapshotState)
  @IsOptional()
  snapshotState?: SnapshotState

  @ApiPropertyOptional({
    description: 'The creation timestamp of the last snapshot',
    example: '2024-10-01T12:00:00Z',
    required: false,
  })
  @IsOptional()
  snapshotCreatedAt?: string

  @ApiPropertyOptional({
    description: 'Auto-stop interval in minutes (0 means disabled)',
    example: 30,
    required: false,
  })
  @IsOptional()
  autoStopInterval?: number

  @ApiPropertyOptional({
    description: 'Array of volumes attached to the workspace',
    type: [WorkspaceVolume],
    required: false,
  })
  @IsOptional()
  volumes?: WorkspaceVolume[]

  constructor() {
    if (this.name === '') {
      this.name = this.id
    }
  }

  static fromWorkspace(workspace: Workspace, nodeDomain: string): WorkspaceDto {
    return {
      id: workspace.id,
      name: workspace.id,
      organizationId: workspace.organizationId,
      target: workspace.region,
      image: workspace.image,
      user: workspace.osUser,
      env: workspace.env,
      cpu: workspace.cpu,
      gpu: workspace.gpu,
      memory: workspace.mem,
      disk: workspace.disk,
      public: workspace.public,
      labels: workspace.labels,
      volumes: workspace.volumes,
      state: this.getWorkspaceState(workspace),
      errorReason: workspace.errorReason,
      snapshotState: workspace.snapshotState,
      snapshotCreatedAt: workspace.lastSnapshotAt?.toISOString(),
      autoStopInterval: workspace.autoStopInterval,
      info: {
        name: workspace.id,
        created: workspace.createdAt?.toISOString(),
        providerMetadata: JSON.stringify({
          state: this.getWorkspaceState(workspace),
          nodeDomain,
          region: workspace.region,
          class: workspace.class,
          updatedAt: workspace.updatedAt?.toISOString(),
          lastSnapshot: workspace.lastSnapshotAt,
          cpu: workspace.cpu,
          gpu: workspace.gpu,
          memory: workspace.mem,
          disk: workspace.disk,
          autoStopInterval: workspace.autoStopInterval,
        }),
      },
    }
  }

  private static getWorkspaceState(workspace: Workspace): WorkspaceState {
    switch (workspace.state) {
      case WorkspaceState.STARTED:
        if (workspace.desiredState === WorkspaceDesiredState.STOPPED) {
          return WorkspaceState.STOPPING
        }
        if (workspace.desiredState === WorkspaceDesiredState.DESTROYED) {
          return WorkspaceState.DESTROYING
        }
        break
      case WorkspaceState.STOPPED:
        if (workspace.desiredState === WorkspaceDesiredState.STARTED) {
          return WorkspaceState.STARTING
        }
        if (workspace.desiredState === WorkspaceDesiredState.DESTROYED) {
          return WorkspaceState.DESTROYING
        }
        break
      case WorkspaceState.UNKNOWN:
        if (workspace.desiredState === WorkspaceDesiredState.STARTED) {
          return WorkspaceState.CREATING
        }
        break
    }
    return workspace.state
  }
}

@ApiSchema({ name: 'WorkspaceLabels' })
export class WorkspaceLabelsDto {
  @ApiProperty({
    description: 'Key-value pairs of labels',
    example: { environment: 'dev', team: 'backend' },
    type: 'object',
    additionalProperties: { type: 'string' },
  })
  labels: { [key: string]: string }
}
