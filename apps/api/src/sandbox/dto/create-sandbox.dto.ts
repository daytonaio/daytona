/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  ArrayMinSize,
  IsEnum,
  IsObject,
  IsOptional,
  IsString,
  IsNumber,
  IsBoolean,
  IsArray,
  Max,
  Min,
  ValidateNested,
} from 'class-validator'
import { Transform, Type } from 'class-transformer'
import { ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { SandboxVolume } from './sandbox.dto'
import { CreateBuildInfoDto } from './create-build-info.dto'
import { IsSafeDisplayString } from '../../common/validators'
import { GpuType } from '../enums/gpu-type.enum'

@ApiSchema({ name: 'CreateSandbox' })
export class CreateSandboxDto {
  @ApiPropertyOptional({
    description: 'The name of the sandbox. If not provided, the sandbox ID will be used as the name',
    example: 'MySandbox',
  })
  @IsOptional()
  @IsString()
  @IsSafeDisplayString()
  name?: string

  @ApiPropertyOptional({
    description: 'The ID or name of the snapshot used for the sandbox',
    example: 'ubuntu-4vcpu-8ram-100gb',
  })
  @IsOptional()
  @IsString()
  snapshot?: string

  @ApiPropertyOptional({
    description: 'The user associated with the project',
    example: 'daytona',
  })
  @IsOptional()
  @IsString()
  user?: string

  @ApiPropertyOptional({
    description: 'Environment variables for the sandbox',
    type: 'object',
    additionalProperties: { type: 'string' },
    example: { NODE_ENV: 'production' },
  })
  @IsOptional()
  @IsObject()
  env?: { [key: string]: string }

  @ApiPropertyOptional({
    description: 'Labels for the sandbox',
    type: 'object',
    additionalProperties: { type: 'string' },
    example: { 'daytona.io/public': 'true' },
  })
  @IsOptional()
  @IsObject()
  labels?: { [key: string]: string }

  @ApiPropertyOptional({
    description: 'Whether the sandbox http preview is publicly accessible',
    example: false,
  })
  @IsOptional()
  @IsBoolean()
  public?: boolean

  @ApiPropertyOptional({
    description: 'Whether to block all network access for the sandbox',
    example: false,
  })
  @IsOptional()
  @IsBoolean()
  networkBlockAll?: boolean

  @ApiPropertyOptional({
    description: 'Comma-separated list of allowed CIDR network addresses for the sandbox',
    example: '192.168.1.0/16,10.0.0.0/24',
  })
  @IsOptional()
  @IsString()
  networkAllowList?: string

  @ApiPropertyOptional({
    description: 'Comma-separated list of allowed domains for the sandbox',
    example: 'example.com,*.daytona.io',
  })
  @IsOptional()
  @IsString()
  domainAllowList?: string

  @ApiPropertyOptional({
    description: 'The target (region) where the sandbox will be created',
    example: 'us',
  })
  @IsOptional()
  @IsString()
  target?: string

  @ApiPropertyOptional({
    description: 'CPU cores allocated to the sandbox',
    example: 2,
    type: 'integer',
  })
  @IsOptional()
  @IsNumber()
  @Min(0)
  cpu?: number

  @ApiPropertyOptional({
    description: 'GPU units allocated to the sandbox',
    example: 1,
    type: 'integer',
  })
  @IsOptional()
  @IsNumber()
  @Min(0)
  @Max(1)
  gpu?: number

  @ApiPropertyOptional({
    description:
      'Preferred GPU type for the sandbox. Accepts a single value or an ordered preference list — the scheduler tries each in order and pins the sandbox to the first that has capacity.',
    enum: GpuType,
    enumName: 'GpuType',
    isArray: true,
    example: [GpuType.H100],
  })
  @IsOptional()
  @Transform(({ value }) =>
    Array.isArray(value)
      ? value.length === 0
        ? undefined
        : value
      : value === undefined || value === null
        ? value
        : [value],
  )
  @IsArray()
  @ArrayMinSize(1)
  @IsEnum(GpuType, { each: true })
  gpuType?: GpuType[]

  @ApiPropertyOptional({
    description: 'Memory allocated to the sandbox in GB',
    example: 1,
    type: 'integer',
  })
  @IsOptional()
  @IsNumber()
  @Min(0)
  memory?: number

  @ApiPropertyOptional({
    description: 'Disk space allocated to the sandbox in GB',
    example: 3,
    type: 'integer',
  })
  @IsOptional()
  @IsNumber()
  @Min(0)
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
    description:
      'Auto-delete interval in minutes (negative value means disabled, 0 means delete immediately upon stopping)',
    example: 30,
    type: 'integer',
  })
  @IsOptional()
  @IsNumber()
  autoDeleteInterval?: number

  @ApiPropertyOptional({
    description: 'Array of volumes to attach to the sandbox',
    type: [SandboxVolume],
    required: false,
  })
  @IsOptional()
  @IsArray()
  @ValidateNested({ each: true })
  @Type(() => SandboxVolume)
  volumes?: SandboxVolume[]

  @ApiPropertyOptional({
    description: 'Build information for the sandbox',
    type: CreateBuildInfoDto,
  })
  @IsOptional()
  @IsObject()
  buildInfo?: CreateBuildInfoDto

  @ApiPropertyOptional({
    description:
      'ID or name of an existing sandbox to link the new sandbox to. The new sandbox will be scheduled on the same runner as the linked sandbox so a local network can be established between them. Linked sandboxes must be ephemeral (autoDeleteInterval=0) and cannot themselves be linked to another sandbox.',
    example: 'sandbox123',
  })
  @IsOptional()
  @IsString()
  linkedSandbox?: string
}
