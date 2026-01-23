/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsBoolean, IsOptional } from 'class-validator'
import { Transform } from 'class-transformer'

@ApiSchema({ name: 'GetForkChildrenQuery' })
export class GetForkChildrenQueryDto {
  @ApiPropertyOptional({
    description: 'Include destroyed sandboxes in the result',
    default: false,
  })
  @IsOptional()
  @IsBoolean()
  @Transform(({ value }) => value === 'true' || value === true)
  includeDestroyed?: boolean
}

@ApiSchema({ name: 'GetForkParentQuery' })
export class GetForkParentQueryDto {
  @ApiPropertyOptional({
    description: 'If true, returns the full ancestor chain up to the root. If false, returns only the direct parent.',
    default: false,
  })
  @IsOptional()
  @IsBoolean()
  @Transform(({ value }) => value === 'true' || value === true)
  ancestors?: boolean
}
