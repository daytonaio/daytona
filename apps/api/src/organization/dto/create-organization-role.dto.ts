/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { ArrayNotEmpty, IsArray, IsEnum, IsString } from 'class-validator'
import { OrganizationResourcePermission } from '../enums/organization-resource-permission.enum'

@ApiSchema({ name: 'CreateOrganizationRole' })
export class CreateOrganizationRoleDto {
  @ApiProperty({
    description: 'The name of the role',
    example: 'Maintainer',
    required: true,
  })
  @IsString()
  name: string

  @ApiProperty({
    description: 'The description of the role',
    example: 'Can manage all resources',
  })
  @IsString()
  description: string

  @ApiProperty({
    description: 'The list of permissions assigned to the role',
    enum: OrganizationResourcePermission,
    isArray: true,
    required: true,
  })
  @IsArray()
  @ArrayNotEmpty()
  @IsEnum(OrganizationResourcePermission, { each: true })
  permissions: OrganizationResourcePermission[]
}
