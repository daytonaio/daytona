/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsArray, IsEnum, IsString } from 'class-validator'
import { GlobalOrganizationRolesIds } from '../constants/global-organization-roles.constant'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'

@ApiSchema({ name: 'UpdateOrganizationMemberAccess' })
export class UpdateOrganizationMemberAccessDto {
  @ApiProperty({
    description: 'Organization member role',
    enum: OrganizationMemberRole,
    default: OrganizationMemberRole.MEMBER,
  })
  @IsEnum(OrganizationMemberRole)
  role: OrganizationMemberRole

  @ApiProperty({
    description: 'Array of assigned role IDs',
    type: [String],
    default: [GlobalOrganizationRolesIds.DEVELOPER],
  })
  @IsArray()
  @IsString({ each: true })
  assignedRoleIds: string[]
}
