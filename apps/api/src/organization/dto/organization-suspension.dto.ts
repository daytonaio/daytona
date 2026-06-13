/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { Type } from 'class-transformer'
import { IsDate, IsNumber, IsOptional, IsString, Min } from 'class-validator'

@ApiSchema({ name: 'OrganizationSuspension' })
export class OrganizationSuspensionDto {
  @ApiProperty({
    description: 'Suspension reason',
  })
  @IsString()
  reason: string

  @ApiProperty({
    description: 'Suspension until',
  })
  @IsOptional()
  @Type(() => Date)
  @IsDate()
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
