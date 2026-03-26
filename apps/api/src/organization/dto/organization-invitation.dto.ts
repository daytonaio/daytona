/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { OrganizationRoleDto } from './organization-role.dto'
import { OrganizationInvitationStatus } from '../enums/organization-invitation-status.enum'
import { OrganizationInvitation } from '../entities/organization-invitation.entity'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'

@ApiSchema({ name: 'OrganizationInvitation' })
export class OrganizationInvitationDto {
  @ApiProperty({
    description: 'Invitation ID',
  })
  id: string

  @ApiProperty({
    description: 'Email address of the invitee',
  })
  email: string

  @ApiProperty({
    description: 'Email address of the inviter',
  })
  invitedBy: string

  @ApiProperty({
    description: 'Organization ID',
  })
  organizationId: string

  @ApiProperty({
    description: 'Organization name',
  })
  organizationName: string

  @ApiProperty({
    description: 'Expiration date of the invitation',
  })
  expiresAt: Date

  @ApiProperty({
    description: 'Invitation status',
    enum: OrganizationInvitationStatus,
  })
  status: OrganizationInvitationStatus

  @ApiProperty({
    description: 'Member role',
    enum: OrganizationMemberRole,
  })
  role: OrganizationMemberRole

  @ApiProperty({
    description: 'Assigned roles',
    type: [OrganizationRoleDto],
  })
  assignedRoles: OrganizationRoleDto[]

  @ApiProperty({
    description: 'Creation timestamp',
  })
  createdAt: Date

  @ApiProperty({
    description: 'Last update timestamp',
  })
  updatedAt: Date

  static fromOrganizationInvitation(invitation: OrganizationInvitation): OrganizationInvitationDto {
    const dto: OrganizationInvitationDto = {
      id: invitation.id,
      email: invitation.email,
      invitedBy: invitation.invitedBy,
      organizationId: invitation.organizationId,
      organizationName: invitation.organization.name,
      expiresAt: invitation.expiresAt,
      status: invitation.status,
      role: invitation.role,
      assignedRoles: invitation.assignedRoles.map(OrganizationRoleDto.fromOrganizationRole),
      createdAt: invitation.createdAt,
      updatedAt: invitation.updatedAt,
    }

    return dto
  }
}
