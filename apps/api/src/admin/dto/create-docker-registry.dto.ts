/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsOptional, IsBoolean, IsEnum } from 'class-validator'
import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { CreateDockerRegistryDto } from '../../docker-registry/dto/create-docker-registry.dto'
import { RegistryType } from '../../docker-registry/enums/registry-type.enum'

@ApiSchema({ name: 'AdminCreateDockerRegistry' })
export class AdminCreateDockerRegistryDto extends CreateDockerRegistryDto {
  @ApiProperty({
    description: 'Registry type',
    enum: RegistryType,
    default: RegistryType.SOURCE,
  })
  @IsEnum(RegistryType)
  declare registryType: RegistryType

  @ApiPropertyOptional({ description: 'Whether the registry can be used as a fallback registry' })
  @IsBoolean()
  @IsOptional()
  isFallback?: boolean
}
