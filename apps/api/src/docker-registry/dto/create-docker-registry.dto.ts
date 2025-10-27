/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsString, IsUrl, IsEnum, IsOptional, IsBoolean } from 'class-validator'
import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { RegistryType } from './../../docker-registry/enums/registry-type.enum'

const VALID_REGISTRY_TYPES = Object.values(RegistryType).filter((type) => type !== RegistryType.TRANSIENT)

@ApiSchema({ name: 'CreateDockerRegistry' })
export class CreateDockerRegistryDto {
  @ApiProperty({ description: 'Registry name' })
  @IsString()
  name: string

  @ApiProperty({ description: 'Registry URL' })
  @IsUrl()
  url: string

  @ApiProperty({ description: 'Registry username' })
  @IsString()
  username: string

  @ApiProperty({ description: 'Registry password' })
  @IsString()
  password: string

  @ApiPropertyOptional({ description: 'Registry project' })
  @IsString()
  @IsOptional()
  project?: string

  @ApiProperty({
    description: 'Registry type',
    enum: VALID_REGISTRY_TYPES,
    default: RegistryType.SOURCE,
  })
  @IsEnum(VALID_REGISTRY_TYPES, {
    message: `value must be one of the following values: ${VALID_REGISTRY_TYPES.join(', ')}`,
  })
  registryType: RegistryType

  @ApiPropertyOptional({ description: 'Whether the registry is active is available for use' })
  @IsBoolean()
  @IsOptional()
  isActive?: boolean
}
