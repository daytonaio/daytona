/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { applyDecorators } from '@nestjs/common'
import { ApiProperty } from '@nestjs/swagger'
import { IsOptional, IsInt, Min } from 'class-validator'
import { Type } from 'class-transformer'

export function PageNumber(defaultValue = 1) {
  return applyDecorators(
    ApiProperty({
      name: 'page',
      description: 'Page number of the results',
      required: false,
      type: Number,
      minimum: 1,
      default: defaultValue,
    }),
    IsOptional(),
    Type(() => Number),
    IsInt(),
    Min(1),
  )
}
