/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsBoolean, IsOptional, IsString } from 'class-validator'
import { IsSafeDisplayString } from '../../common/validators'

@ApiSchema({ name: 'CreateSandboxSnapshot' })
export class CreateSandboxSnapshotDto {
  @ApiProperty({
    description: 'Name for the new snapshot',
    example: 'my-dev-env-v1',
  })
  @IsString()
  @IsSafeDisplayString()
  name: string

  @ApiPropertyOptional({
    description:
      "Include the VM's memory in the snapshot. VM sandboxes only. When true the sandbox must be STARTED; when false (default) VM sandboxes must be STOPPED. Container sandboxes do not support memory snapshots.",
    default: false,
  })
  @IsBoolean()
  @IsOptional()
  includeMemory?: boolean
}
