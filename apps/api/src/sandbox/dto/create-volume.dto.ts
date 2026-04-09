/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsOptional, IsString } from 'class-validator'
import { IsSafeDisplayString } from '../../common/validators'

@ApiSchema({ name: 'CreateVolume' })
export class CreateVolumeDto {
  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  @IsSafeDisplayString()
  name?: string
}
