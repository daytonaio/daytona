/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsEnum, IsObject, IsOptional, IsString, IsNumber, IsBoolean } from 'class-validator'
import { ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { SandboxVolume } from './sandbox.dto'
import { CreateBuildInfoDto } from './create-build-info.dto'

enum RunnerRegion {
  EU = 'eu',
  US = 'us',
  ASIA = 'asia',
}
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
    enum: SandboxClass,
    example: Object.values(SandboxClass)[0],
  })
  @IsOptional()
  @IsEnum(SandboxClass)
  class?: SandboxClass

  @ApiPropertyOptional({
    description: 'The target (region) where the workspace will be created',
    enum: RunnerRegion,
    example: Object.values(RunnerRegion)[0],
  })
  @IsOptional()
  @IsEnum(RunnerRegion)
  target?: RunnerRegion

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
    description: 'Memory allocated to the workspace in GB',
    example: 1,
    type: 'integer',
  })
  @IsOptional()
  @IsNumber()
  memory?: number

  @ApiPropertyOptional({
    description: 'Disk space allocated to the workspace in GB',
    example: 3,
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
    description: 'Auto-archive interval in minutes (0 means the maximum interval will be used)',
    example: 7 * 24 * 60,
    type: 'integer',
  })
  @IsOptional()
  @IsNumber()
  autoArchiveInterval?: number

  @ApiPropertyOptional({
    description: 'Array of volumes to attach to the workspace',
    type: [SandboxVolume],
    required: false,
  })
  @IsOptional()
  volumes?: SandboxVolume[]

  @ApiPropertyOptional({
    description: 'Build information for the workspace',
    type: CreateBuildInfoDto,
  })
  @IsOptional()
  @IsObject()
  buildInfo?: CreateBuildInfoDto
}
