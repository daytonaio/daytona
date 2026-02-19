/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsNumber, IsOptional, Min } from 'class-validator'

@ApiSchema({ name: 'OrganizationSuspension' })
export class OrganizationSuspensionDto {
  @ApiProperty({
    description: 'Suspension reason',
  })
  reason: string

  @ApiProperty({
    description: 'Suspension until',
  })
  @IsOptional()
  until?: Date

  @ApiPropertyOptional({
    description: 'Suspension cleanup grace period hours',
    type: 'number',
    minimum: 0,
  })
  @IsOptional()
  @IsNumber()
  @Min(0)
  suspensionCleanupGracePeriodHours?: number
}
