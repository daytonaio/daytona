/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsString } from 'class-validator'
import { ApiProperty, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'CreateRunner' })
export class CreateRunnerDto {
  @IsString()
  @ApiProperty()
  regionId: string

  @IsString()
  @ApiProperty()
  name: string
}
