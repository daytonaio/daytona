/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, JoinColumn, ManyToMany, ManyToOne, PrimaryGeneratedColumn } from 'typeorm'
import { Organization } from './organization.entity'
import { OrganizationUser } from './organization-user.entity'
import { OrganizationResourcePermission } from '../enums/organization-resource-permission.enum'
import { OrganizationInvitation } from './organization-invitation.entity'
import { v4 } from 'uuid'

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

  constructor(createParams: {
    organization: Organization
    name: string
    description: string
    permissions: OrganizationResourcePermission[]
    isGlobal?: boolean
    users?: OrganizationUser[]
    invitations?: OrganizationInvitation[]
  }) {
    this.id = v4()
    this.name = createParams.name
    this.description = createParams.description
    this.permissions = createParams.permissions
    this.isGlobal = createParams.isGlobal ?? false
    this.organization = createParams.organization
    this.organizationId = createParams.organization.id
    this.users = createParams.users || []
    this.invitations = createParams.invitations || []
    this.createdAt = new Date()
    this.updatedAt = new Date()
  }
}
