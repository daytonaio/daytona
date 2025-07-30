/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, JoinColumn, ManyToMany, ManyToOne, PrimaryGeneratedColumn } from 'typeorm'
import { Organization } from './organization.entity'
import { OrganizationUser } from './organization-user.entity'
import { OrganizationResourcePermission } from '../enums/organization-resource-permission.enum'
import { OrganizationInvitation } from './organization-invitation.entity'

@Entity()
export class OrganizationRole {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column()
  name: string

  @Column()
  description: string

  @Column({
    type: 'enum',
    enum: OrganizationResourcePermission,
    array: true,
  })
  permissions: OrganizationResourcePermission[]

  @Column({ default: false })
  isGlobal: boolean

  @Column({
    nullable: true,
  })
  organizationId?: string

  @ManyToOne(() => Organization, { onDelete: 'CASCADE' })
  @JoinColumn({ name: 'organizationId' })
  organization: Organization

  @ManyToMany(() => OrganizationUser, (user) => user.assignedRoles)
  users: OrganizationUser[]

  @ManyToMany(() => OrganizationInvitation, (invitation) => invitation.assignedRoles)
  invitations: OrganizationInvitation[]

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date
}
