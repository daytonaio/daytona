/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { applyDecorators } from '@nestjs/common'
import { ApiProperty } from '@nestjs/swagger'
import { IsOptional, IsInt, Min, Max } from 'class-validator'
import { Type } from 'class-transformer'

export function PageLimit(defaultValue = 100) {
  return applyDecorators(
    ApiProperty({
      name: 'limit',
      description: 'Number of results per page',
      required: false,
      type: Number,
      minimum: 1,
      maximum: 200,
      default: defaultValue,
    }),
    IsOptional(),
    Type(() => Number),
    IsInt(),
    Min(1),
    Max(200),
  )
}
