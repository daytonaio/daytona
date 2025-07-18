/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, JoinColumn, JoinTable, ManyToMany, ManyToOne, PrimaryColumn } from 'typeorm'
import { Organization } from './organization.entity'
import { OrganizationRole } from './organization-role.entity'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'

@Entity()
export class OrganizationUser {
  @PrimaryColumn()
  organizationId: string

  @PrimaryColumn()
  userId: string

  @Column({
    type: 'enum',
    enum: OrganizationMemberRole,
    default: OrganizationMemberRole.MEMBER,
  })
  role: OrganizationMemberRole

  @ManyToOne(() => Organization, (organization) => organization.users, {
    onDelete: 'CASCADE',
  })
  @JoinColumn({ name: 'organizationId' })
  organization: Organization

  @ManyToMany(() => OrganizationRole, (role) => role.users, {
    cascade: true,
    onDelete: 'CASCADE',
  })
  @JoinTable({
    name: 'organization_role_assignment',
    joinColumns: [
      { name: 'organizationId', referencedColumnName: 'organizationId' },
      { name: 'userId', referencedColumnName: 'userId' },
    ],
    inverseJoinColumns: [{ name: 'roleId', referencedColumnName: 'id' }],
  })
  assignedRoles: OrganizationRole[]

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date
}
