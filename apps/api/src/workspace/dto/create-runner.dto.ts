/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsEnum, IsNumber, IsString } from 'class-validator'
import { WorkspaceClass } from '../enums/workspace-class.enum'
import { RunnerRegion } from '../enums/runner-region.enum'
import { ApiProperty, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'CreateRunner' })
export class CreateRunnerDto {
  @ApiProperty()
  @IsString()
  domain: string

  @IsString()
  @ApiProperty()
  apiUrl: string

  @IsString()
  @ApiProperty()
  apiKey: string

  @IsNumber()
  @ApiProperty()
  cpu: number

  @IsNumber()
  @ApiProperty()
  memory: number

  @IsNumber()
  @ApiProperty()
  disk: number

  @IsNumber()
  @ApiProperty()
  gpu: number

  @IsString()
  @ApiProperty()
  gpuType: string

  @IsEnum(WorkspaceClass)
  @ApiProperty({
    enum: WorkspaceClass,
    example: Object.values(WorkspaceClass)[0],
  })
  class: WorkspaceClass

  @IsNumber()
  @ApiProperty()
  capacity: number

  @IsEnum(RunnerRegion)
  @ApiProperty({
    enum: RunnerRegion,
    example: Object.values(RunnerRegion)[0],
  })
  region: RunnerRegion
}
