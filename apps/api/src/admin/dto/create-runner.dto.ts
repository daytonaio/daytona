/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsEnum, IsNumber, IsString } from 'class-validator'
import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { CreateRunnerDto } from '../../sandbox/dto/create-runner.dto'
import { SandboxClass } from '../../sandbox/enums/sandbox-class.enum'

@ApiSchema({ name: 'AdminCreateRunner' })
export class AdminCreateRunnerDto extends CreateRunnerDto {
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

  @IsString()
  @ApiProperty()
  version: string
}
