/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsIn, IsNotEmpty, IsString } from 'class-validator'

@ApiSchema({ name: 'ChangeVolumeBackend' })
export class ChangeVolumeBackendDto {
  @ApiProperty({
    description:
      "The backend to switch the volume to. `s3fuse` mounts the volume's bucket on the runner host; `experimental` exposes the same bucket through an Archil disk that is mounted inside the sandbox. The volume's data (its S3 bucket) is preserved across the switch - only the mount strategy changes.",
    example: 'experimental',
    enum: ['s3fuse', 'experimental'],
  })
  @IsString()
  @IsNotEmpty()
  @IsIn(['s3fuse', 'experimental'])
  backend: string
}
