/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Column,
  CreateDateColumn,
  Entity,
  JoinColumn,
  JoinTable,
  ManyToMany,
  ManyToOne,
  PrimaryGeneratedColumn,
} from 'typeorm'
import { Organization } from './organization.entity'
import { OrganizationInvitationStatus } from '../enums/organization-invitation-status.enum'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { OrganizationRole } from './organization-role.entity'

@Entity()
export class OrganizationInvitation {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column()
  organizationId: string

  @Column()
  email: string

  @Column({
    default: '',
  })
  invitedBy: string

  @Column({
    type: 'enum',
    enum: OrganizationMemberRole,
    default: OrganizationMemberRole.MEMBER,
  })
  role: OrganizationMemberRole

  @ManyToMany(() => OrganizationRole, (role) => role.invitations, {
    cascade: true,
    onDelete: 'CASCADE',
  })
  @JoinTable({
    name: 'organization_role_assignment_invitation',
    joinColumn: {
      name: 'invitationId',
      referencedColumnName: 'id',
    },
    inverseJoinColumn: {
      name: 'roleId',
      referencedColumnName: 'id',
    },
  })
  assignedRoles: OrganizationRole[]

  @Column({
    type: 'timestamp with time zone',
  })
  expiresAt: Date

  @Column({
    type: 'enum',
    enum: OrganizationInvitationStatus,
    default: OrganizationInvitationStatus.PENDING,
  })
  status: OrganizationInvitationStatus

  @ManyToOne(() => Organization, { onDelete: 'CASCADE' })
  @JoinColumn({ name: 'organizationId' })
  organization: Organization

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date
}
