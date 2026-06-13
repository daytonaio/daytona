/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsObject, IsOptional, IsString } from 'class-validator'

@ApiSchema({ name: 'OtelConfig' })
export class OtelConfigDto {
  @ApiProperty({
    description: 'Endpoint',
  })
  @IsString()
  endpoint: string

  @ApiProperty({
    description: 'Headers',
    example: {
      'x-api-key': 'my-api-key',
    },
    nullable: true,
    required: false,
    additionalProperties: { type: 'string' },
  })
  @IsObject()
  @IsOptional()
  headers?: Record<string, string>
}
