/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiPropertyOptional } from '@nestjs/swagger'
import { IsNotEmpty, IsOptional, IsString } from 'class-validator'

export class UpdateSandboxDegradedReasonDto {
  @IsOptional()
  @IsString()
  @IsNotEmpty()
  @ApiPropertyOptional({
    description: 'Degraded reason; omit or send null to clear',
    nullable: true,
    example: 'fd-exhaustion: too many open files',
  })
  degradedReason?: string | null
}
