/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsOptional } from 'class-validator'

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

  @ApiProperty({
    description: 'Suspension cleanup grace period hours',
  })
  @IsOptional()
  suspensionCleanupGracePeriodHours?: number
}
