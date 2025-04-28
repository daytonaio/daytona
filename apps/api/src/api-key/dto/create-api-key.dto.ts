/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ArrayNotEmpty, IsArray, IsEnum, IsNotEmpty, IsString } from 'class-validator'
import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'

@ApiSchema({ name: 'CreateApiKey' })
export class CreateApiKeyDto {
  @ApiProperty({
    description: 'The name of the API key',
    example: 'My API Key',
    required: true,
  })
  @IsNotEmpty()
  @IsString()
  name: string

  @ApiProperty({
    description: 'The list of organization resource permissions assigned to the API key',
    enum: OrganizationResourcePermission,
    isArray: true,
    required: true,
  })
  @IsArray()
  @ArrayNotEmpty()
  @IsEnum(OrganizationResourcePermission, { each: true })
  permissions: OrganizationResourcePermission[]
}
