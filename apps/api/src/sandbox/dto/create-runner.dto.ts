/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsArray, IsOptional, IsString } from 'class-validator'
import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsSafeDisplayString } from '../../common/validators'

@ApiSchema({ name: 'CreateRunner' })
export class CreateRunnerDto {
  @IsString()
  @ApiProperty()
  regionId: string

  @IsString()
  @IsSafeDisplayString()
  @ApiProperty()
  name: string

  @ApiPropertyOptional({
    description: 'Tags to associate with the runner',
    example: ['gpu', 'us-east'],
    type: [String],
  })
  @IsOptional()
  @IsArray()
  @IsString({ each: true })
  tags?: string[]
}
