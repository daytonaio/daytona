/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { OrganizationRole } from '../entities/organization-role.entity'
import { OrganizationResourcePermission } from '../enums/organization-resource-permission.enum'

@ApiSchema({ name: 'OrganizationRole' })
export class OrganizationRoleDto {
  @ApiProperty({
    description: 'Role ID',
  })
  id: string

  @ApiProperty({
    description: 'Role name',
  })
  name: string

  @ApiProperty({
    description: 'Role description',
  })
  description: string

  @ApiProperty({
    description: 'Roles assigned to the user',
    enum: OrganizationResourcePermission,
    isArray: true,
  })
  permissions: OrganizationResourcePermission[]

  @ApiProperty({
    description: 'Global role flag',
  })
  isGlobal: boolean

  @ApiProperty({
    description: 'Creation timestamp',
  })
  createdAt: Date

  @ApiProperty({
    description: 'Last update timestamp',
  })
  updatedAt: Date

  static fromOrganizationRole(role: OrganizationRole): OrganizationRoleDto {
    const dto: OrganizationRoleDto = {
      id: role.id,
      name: role.name,
      description: role.description,
      permissions: role.permissions,
      isGlobal: role.isGlobal,
      createdAt: role.createdAt,
      updatedAt: role.updatedAt,
    }

    return dto
  }
}
