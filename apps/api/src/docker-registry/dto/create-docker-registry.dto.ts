/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsString, IsUrl, IsEnum } from 'class-validator'
import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { RegistryType } from './../../docker-registry/enums/registry-type.enum'

@ApiSchema({ name: 'CreateDockerRegistry' })
export class CreateDockerRegistryDto {
  @ApiProperty({
    description: 'Registry name',
  })
  @IsString()
  name: string

  @ApiProperty({
    description: 'Registry URL',
  })
  @IsUrl()
  url: string

  @ApiProperty({ description: 'Registry username' })
  @IsString()
  username: string

  @ApiProperty({ description: 'Registry password' })
  @IsString()
  password: string

  @ApiProperty({ description: 'Registry project' })
  @IsString()
  project: string

  @ApiProperty({
    description: 'Registry type',
    enum: RegistryType,
    default: RegistryType.INTERNAL,
  })
  @IsEnum(RegistryType)
  registryType: RegistryType

  @ApiProperty({
    description: 'Set as default registry',
    default: false,
  })
  isDefault?: boolean
}
