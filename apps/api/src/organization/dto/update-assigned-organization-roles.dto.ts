/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsArray, IsString } from 'class-validator'

@ApiSchema({ name: 'UpdateAssignedOrganizationRoles' })
export class UpdateAssignedOrganizationRolesDto {
  @ApiProperty({
    description: 'Array of role IDs',
    type: [String],
  })
  @IsArray()
  @IsString({ each: true })
  roleIds: string[]
}
