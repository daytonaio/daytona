/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsString, IsUrl, IsEnum, IsOptional, IsBoolean } from 'class-validator'
import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { RegistryType } from './../../docker-registry/enums/registry-type.enum'

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
    enum: RegistryType,
    default: RegistryType.ORGANIZATION,
  })
  @IsEnum(RegistryType)
  registryType: RegistryType

  @ApiPropertyOptional({ description: 'Set as default registry' })
  @IsBoolean()
  @IsOptional()
  isDefault?: boolean
}
