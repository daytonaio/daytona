/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsIn, IsOptional, IsString } from 'class-validator'
import { IsSafeDisplayString } from '../../common/validators'

@ApiSchema({ name: 'CreateVolume' })
export class CreateVolumeDto {
  @ApiProperty()
  @IsString()
  @IsSafeDisplayString()
  name: string

  @ApiPropertyOptional({
    description:
      "Storage backend for this volume. 's3fuse' (default) mounts a dedicated S3 bucket on the runner host. 'layered' mounts inside the sandbox via the layered control plane and requires the volume_backend_picker feature flag. When omitted, the organization's default backend is used.",
    enum: ['s3fuse', 'layered'],
  })
  @IsOptional()
  @IsString()
  @IsIn(['s3fuse', 'layered'])
  backend?: string

  @ApiPropertyOptional({
    description:
      "Daytona Region ID the volume should live in. Only honored for the layered backend; rejected for s3fuse. When omitted, defaults to the organization's default region.",
  })
  @IsOptional()
  @IsString()
  regionId?: string
}
