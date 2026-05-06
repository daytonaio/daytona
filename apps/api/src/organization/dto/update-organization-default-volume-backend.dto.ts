/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsIn, IsNotEmpty, IsString } from 'class-validator'

@ApiSchema({ name: 'UpdateOrganizationDefaultVolumeBackend' })
export class UpdateOrganizationDefaultVolumeBackendDto {
  @ApiProperty({
    description:
      'The default volume backend for the organization. `s3fuse` mounts on the runner host (existing behavior). `experimental` mounts inside the sandbox via Archil using a per-volume mount token.',
    example: 's3fuse',
    enum: ['s3fuse', 'experimental'],
  })
  @IsString()
  @IsNotEmpty()
  @IsIn(['s3fuse', 'experimental'])
  defaultVolumeBackend: string
}
