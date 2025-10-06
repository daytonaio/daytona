/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, OneToMany, PrimaryGeneratedColumn, UpdateDateColumn } from 'typeorm'
import { OrganizationUser } from './organization-user.entity'
import { OrganizationRole } from './organization-role.entity'
import { OrganizationInvitation } from './organization-invitation.entity'
import { v4 } from 'uuid'

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
    default: 10,
    name: 'total_memory_quota',
  })
  totalMemoryQuota: number

  @Column({
    type: 'int',
    default: 30,
    name: 'total_disk_quota',
  })
  totalDiskQuota: number

  @Column({
    type: 'int',
    default: 4,
    name: 'max_cpu_per_sandbox',
  })
  maxCpuPerSandbox: number

  @Column({
    type: 'int',
    default: 8,
    name: 'max_memory_per_sandbox',
  })
  maxMemoryPerSandbox: number

  @Column({
    type: 'int',
    default: 10,
    name: 'max_disk_per_sandbox',
  })
  maxDiskPerSandbox: number

  @Column({
    type: 'int',
    default: 20,
    name: 'max_snapshot_size',
  })
  maxSnapshotSize: number

  @Column({
    type: 'int',
    default: 100,
    name: 'snapshot_quota',
  })
  snapshotQuota: number

  @Column({
    type: 'int',
    default: 100,
    name: 'volume_quota',
  })
  volumeQuota: number

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
    type: 'timestamp with time zone',
  })
  suspendedAt: Date | null

  @Column({
    nullable: true,
    type: String,
  })
  suspensionReason: string | null

  @Column({
    type: 'int',
    default: 24,
  })
  suspensionCleanupGracePeriodHours: number

  @Column({
    nullable: true,
    type: 'timestamp with time zone',
  })
  suspendedUntil: Date | null

  @Column({
    default: false,
  })
  sandboxLimitedNetworkEgress: boolean

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date

  constructor(createParams: {
    name: string
    createdBy: string
    telemetryEnabled?: boolean
    personal?: boolean
    totalCpuQuota?: number
    totalMemoryQuota?: number
    totalDiskQuota?: number
    maxCpuPerSandbox?: number
    maxMemoryPerSandbox?: number
    maxDiskPerSandbox?: number
    maxSnapshotSize?: number
    snapshotQuota?: number
    volumeQuota?: number
    suspended?: boolean
    suspendedAt?: Date | null
    suspensionReason?: string | null
    suspensionCleanupGracePeriodHours?: number
    suspendedUntil?: Date
    sandboxLimitedNetworkEgress?: boolean
  }) {
    this.id = v4()
    this.name = createParams.name
    this.createdBy = createParams.createdBy
    this.telemetryEnabled = createParams.telemetryEnabled ?? true
    this.personal = createParams.personal ?? false
    this.totalCpuQuota = createParams.totalCpuQuota ?? 10
    this.totalMemoryQuota = createParams.totalMemoryQuota ?? 10
    this.totalDiskQuota = createParams.totalDiskQuota ?? 30
    this.maxCpuPerSandbox = createParams.maxCpuPerSandbox ?? 4
    this.maxMemoryPerSandbox = createParams.maxMemoryPerSandbox ?? 8
    this.maxDiskPerSandbox = createParams.maxDiskPerSandbox ?? 10
    this.maxSnapshotSize = createParams.maxSnapshotSize ?? 20
    this.snapshotQuota = createParams.snapshotQuota ?? 100
    this.volumeQuota = createParams.volumeQuota ?? 100
    this.suspended = createParams.suspended ?? false
    this.suspendedAt = createParams.suspendedAt ?? null
    this.suspensionReason = createParams.suspensionReason ?? null
    this.suspensionCleanupGracePeriodHours = createParams.suspensionCleanupGracePeriodHours ?? 24
    this.suspendedUntil = createParams.suspendedUntil ?? null
    this.sandboxLimitedNetworkEgress = createParams.sandboxLimitedNetworkEgress ?? false
    this.createdAt = new Date()
    this.updatedAt = new Date()
    this.roles = []
    this.users = []
    this.invitations = []
  }
}
