/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'
import { IsUUID } from 'class-validator'

export class AttachDiskDto {
  @ApiProperty({
    description: 'Sandbox ID to attach the disk to',
    example: '123e4567-e89b-12d3-a456-426614174000',
  })
  @IsUUID()
  sandboxId: string
}
