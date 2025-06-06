/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiSchema } from '@nestjs/swagger'
import { IsNumber } from 'class-validator'

@ApiSchema({ name: 'ResizeQuota' })
export class ResizeDto {
  @IsNumber()
  cpu: number

  @IsNumber()
  gpu: number

  @IsNumber()
  memory: number
}
