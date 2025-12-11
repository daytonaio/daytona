/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsEnum, IsOptional } from 'class-validator'
import { Runner } from '../entities/runner.entity'
import { RunnerDto } from './runner.dto'
import { RegionType } from '../../region/enums/region-type.enum'

@ApiSchema({ name: 'RunnerFull' })
export class RunnerFullDto extends RunnerDto {
  @ApiProperty({
    description: 'The API key for the runner',
    example: 'dtn_1234567890',
  })
  apiKey: string

  @ApiPropertyOptional({
    description: 'The region type of the runner',
    enum: RegionType,
    enumName: 'RegionType',
    example: Object.values(RegionType)[0],
  })
  @IsOptional()
  @IsEnum(RegionType)
  regionType?: RegionType

  static fromRunner(runner: Runner, regionType?: RegionType): RunnerFullDto {
    return {
      ...RunnerDto.fromRunner(runner),
      apiKey: runner.apiKey,
      regionType,
    }
  }
}
