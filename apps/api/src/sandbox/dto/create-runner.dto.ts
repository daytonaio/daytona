/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsEnum, IsNumber, IsString } from 'class-validator'
import { SandboxClass } from '../enums/sandbox-class.enum'
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
  proxyUrl: string

  @IsString()
  @ApiProperty()
  apiKey: string

  @IsNumber()
  @ApiProperty()
  cpu: number

  @IsNumber()
  @ApiProperty()
  memoryGiB: number

  @IsNumber()
  @ApiProperty()
  diskGiB: number

  @IsNumber()
  @ApiProperty()
  gpu: number

  @IsString()
  @ApiProperty()
  gpuType: string

  @IsEnum(SandboxClass)
  @ApiProperty({
    enum: SandboxClass,
    example: Object.values(SandboxClass)[0],
  })
  class: SandboxClass

  @IsNumber()
  @ApiProperty()
  capacity: number

  @IsString()
  @ApiProperty({
    example: 'us',
  })
  region: string

  @IsString()
  @ApiProperty()
  version: string
}
