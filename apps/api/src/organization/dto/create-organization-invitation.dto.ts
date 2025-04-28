/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { Type } from 'class-transformer'
import { IsArray, IsDate, IsEmail, IsEnum, IsOptional, IsString } from 'class-validator'
import { GlobalOrganizationRolesIds } from '../constants/global-organization-roles.constant'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'

@ApiSchema({ name: 'CreateOrganizationInvitation' })
export class CreateOrganizationInvitationDto {
  @ApiProperty({
    description: 'Email address of the invitee',
    example: 'mail@example.com',
    required: true,
  })
  @IsString()
  @IsEmail()
  email: string

  @ApiProperty({
    description: 'Organization member role for the invitee',
    enum: OrganizationMemberRole,
    default: OrganizationMemberRole.MEMBER,
  })
  @IsEnum(OrganizationMemberRole)
  role: OrganizationMemberRole

  @ApiProperty({
    description: 'Array of assigned role IDs for the invitee',
    type: [String],
    default: [GlobalOrganizationRolesIds.DEVELOPER],
  })
  @IsArray()
  @IsString({ each: true })
  assignedRoleIds: string[]

  @ApiPropertyOptional({
    description: 'Expiration date of the invitation',
    example: '2021-12-31T23:59:59Z',
  })
  @IsOptional()
  @Type(() => Date)
  @IsDate()
  expiresAt?: Date
}
