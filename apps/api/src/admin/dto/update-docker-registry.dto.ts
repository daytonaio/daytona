/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsOptional, IsBoolean } from 'class-validator'
import { ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { UpdateDockerRegistryDto } from '../../docker-registry/dto/update-docker-registry.dto'

@ApiSchema({ name: 'AdminUpdateDockerRegistry' })
export class AdminUpdateDockerRegistryDto extends UpdateDockerRegistryDto {
  @ApiPropertyOptional({ description: 'Whether the registry can be used as a fallback registry' })
  @IsBoolean()
  @IsOptional()
  isFallback?: boolean
}
