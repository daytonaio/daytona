/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsBoolean, IsArray, ValidateNested, IsString, IsNumber, Min } from 'class-validator'
import { Type } from 'class-transformer'

@ApiSchema({ name: 'SetSnapshotGeneralStatusDto' })
export class SetSnapshotGeneralStatusDto {
  @ApiProperty({
    description: 'Whether the snapshot is general',
    example: true,
  })
  @IsBoolean()
  general: boolean
}

@ApiSchema({ name: 'SetSnapshotTargetPropagationDto' })
export class SetSnapshotTargetPropagationDto {
  @ApiProperty({
    description: 'The target environment for the snapshot',
    example: 'local',
  })
  @IsString()
  target: string

  @ApiProperty({
    description: 'User minimum value for the target',
    example: 0,
  })
  @IsNumber()
  @Min(0, { message: 'User minimum cannot be negative' })
  userOverride: number
}

@ApiSchema({ name: 'SetSnapshotTargetPropagationsDto' })
export class SetSnapshotTargetPropagationsDto {
  @ApiProperty({
    description: 'Target propagation settings',
    type: [SetSnapshotTargetPropagationDto],
  })
  @IsArray()
  @ValidateNested({ each: true })
  @Type(() => SetSnapshotTargetPropagationDto)
  targetPropagations: SetSnapshotTargetPropagationDto[]
}
