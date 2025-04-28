/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsEnum, IsNumber, IsString } from 'class-validator'
import { WorkspaceClass } from '../enums/workspace-class.enum'
import { NodeRegion } from '../enums/node-region.enum'
import { ApiProperty, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'CreateNode' })
export class CreateNodeDto {
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

  @IsEnum(NodeRegion)
  @ApiProperty({
    enum: NodeRegion,
    example: Object.values(NodeRegion)[0],
  })
  region: NodeRegion
}
