/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsEnum, IsObject, IsOptional, IsString, IsNumber, IsBoolean } from 'class-validator'
import { ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { WorkspaceClass } from '../enums/workspace-class.enum'
import { NodeRegion } from '../enums/node-region.enum'
import { WorkspaceVolume } from './workspace.dto'
import { CreateBuildInfoDto } from './create-build-info.dto'

@ApiSchema({ name: 'CreateWorkspace' })
export class CreateWorkspaceDto {
  @ApiPropertyOptional({
    description: 'The image used for the workspace',
    example: 'daytonaio/workspace:latest',
  })
  @IsOptional()
  @IsString()
  image?: string

  @ApiPropertyOptional({
    description: 'The user associated with the project',
    example: 'daytona',
  })
  @IsOptional()
  @IsString()
  user?: string

  @ApiPropertyOptional({
    description: 'Environment variables for the workspace',
    type: 'object',
    additionalProperties: { type: 'string' },
    example: { NODE_ENV: 'production' },
  })
  @IsOptional()
  @IsObject()
  env?: { [key: string]: string }

  @ApiPropertyOptional({
    description: 'Labels for the workspace',
    type: 'object',
    additionalProperties: { type: 'string' },
    example: { 'daytona.io/public': 'true' },
  })
  @IsOptional()
  @IsObject()
  labels?: { [key: string]: string }

  @ApiPropertyOptional({
    description: 'Whether the workspace http preview is publicly accessible',
    example: false,
  })
  @IsOptional()
  @IsBoolean()
  public?: boolean

  @ApiPropertyOptional({
    description: 'The workspace class type',
    enum: WorkspaceClass,
    example: Object.values(WorkspaceClass)[0],
  })
  @IsOptional()
  @IsEnum(WorkspaceClass)
  class?: WorkspaceClass

  @ApiPropertyOptional({
    description: 'The target (region) where the workspace will be created',
    enum: NodeRegion,
    example: Object.values(NodeRegion)[0],
  })
  @IsOptional()
  @IsEnum(NodeRegion)
  target?: NodeRegion

  @ApiPropertyOptional({
    description: 'CPU cores allocated to the workspace',
    example: 2,
    type: 'integer',
  })
  @IsOptional()
  @IsNumber()
  cpu?: number

  @ApiPropertyOptional({
    description: 'GPU units allocated to the workspace',
    example: 1,
    type: 'integer',
  })
  @IsOptional()
  @IsNumber()
  gpu?: number

  @ApiPropertyOptional({
    description: 'Memory allocated to the workspace in MB',
    example: 4096,
    type: 'integer',
  })
  @IsOptional()
  @IsNumber()
  memory?: number

  @ApiPropertyOptional({
    description: 'Disk space allocated to the workspace in GB',
    example: 20,
    type: 'integer',
  })
  @IsOptional()
  @IsNumber()
  disk?: number

  @ApiPropertyOptional({
    description: 'Auto-stop interval in minutes (0 means disabled)',
    example: 30,
    type: 'integer',
  })
  @IsOptional()
  @IsNumber()
  autoStopInterval?: number

  @ApiPropertyOptional({
    description: 'Array of volumes to attach to the workspace',
    type: [WorkspaceVolume],
    required: false,
  })
  @IsOptional()
  volumes?: WorkspaceVolume[]

  @ApiPropertyOptional({
    description: 'Build information for the workspace',
    type: CreateBuildInfoDto,
  })
  @IsOptional()
  @IsObject()
  buildInfo?: CreateBuildInfoDto
}
