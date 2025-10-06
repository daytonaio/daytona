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

  constructor(role: OrganizationRole) {
    this.id = role.id
    this.name = role.name
    this.description = role.description
    this.permissions = role.permissions
    this.isGlobal = role.isGlobal
    this.createdAt = role.createdAt
    this.updatedAt = role.updatedAt
  }
}
