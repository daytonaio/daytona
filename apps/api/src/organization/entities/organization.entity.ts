/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, OneToMany, PrimaryGeneratedColumn } from 'typeorm'
import { OrganizationUser } from './organization-user.entity'
import { OrganizationRole } from './organization-role.entity'
import { OrganizationInvitation } from './organization-invitation.entity'

@Entity()
export class Organization {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column()
  name: string

  @Column()
  createdBy: string

  @Column({
    default: false,
  })
  personal: boolean

  @Column({
    default: true,
  })
  telemetryEnabled: boolean

  @Column({
    type: 'int',
    default: 10,
    name: 'total_cpu_quota',
  })
  totalCpuQuota: number

  @Column({
    type: 'int',
    default: 40,
    name: 'total_memory_quota',
  })
  totalMemoryQuota: number

  @Column({
    type: 'int',
    default: 100,
    name: 'total_disk_quota',
  })
  totalDiskQuota: number

  @Column({
    type: 'int',
    default: 2,
    name: 'max_cpu_per_workspace',
  })
  maxCpuPerWorkspace: number

  @Column({
    type: 'int',
    default: 4,
    name: 'max_memory_per_workspace',
  })
  maxMemoryPerWorkspace: number

  @Column({
    type: 'int',
    default: 10,
    name: 'max_disk_per_workspace',
  })
  maxDiskPerWorkspace: number

  @Column({
    type: 'int',
    default: 10,
    name: 'max_concurrent_workspaces',
  })
  maxConcurrentWorkspaces: number

  @Column({
    type: 'int',
    default: 0,
    name: 'workspace_quota',
  })
  workspaceQuota: number

  @Column({
    type: 'int',
    default: 0,
    name: 'image_quota',
  })
  imageQuota: number

  @Column({
    type: 'int',
    default: 2,
    name: 'max_image_size',
  })
  maxImageSize: number

  @Column({
    type: 'int',
    default: 5,
    name: 'total_image_size',
  })
  totalImageSize: number

  @OneToMany(() => OrganizationRole, (organizationRole) => organizationRole.organization, {
    cascade: true,
    onDelete: 'CASCADE',
  })
  roles: OrganizationRole[]

  @OneToMany(() => OrganizationUser, (user) => user.organization, {
    cascade: true,
    onDelete: 'CASCADE',
  })
  users: OrganizationUser[]

  @OneToMany(() => OrganizationInvitation, (invitation) => invitation.organization, {
    cascade: true,
    onDelete: 'CASCADE',
  })
  invitations: OrganizationInvitation[]

  @Column({
    default: false,
  })
  suspended: boolean

  @Column({
    nullable: true,
    type: 'timestamp',
  })
  suspendedAt?: Date

  @Column({
    nullable: true,
  })
  suspensionReason?: string

  @Column({
    nullable: true,
    type: 'timestamp',
  })
  suspendedUntil?: Date

  @CreateDateColumn()
  createdAt: Date

  @CreateDateColumn()
  updatedAt: Date
}
