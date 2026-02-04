/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, Index, OneToMany, PrimaryGeneratedColumn, UpdateDateColumn } from 'typeorm'
import { OrganizationUser } from './organization-user.entity'
import { OrganizationRole } from './organization-role.entity'
import { OrganizationInvitation } from './organization-invitation.entity'
import { RegionQuota } from './region-quota.entity'

@Entity()
@Index('idx_organization_deleted_at', ['deletedAt'], { where: '"deletedAt" IS NOT NULL' })
@Index('idx_organization_created_at_not_deleted', ['createdAt'], { where: '"deletedAt" IS NULL' })
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

  @Column({ nullable: true })
  defaultRegionId?: string

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

  @Column({
    type: 'int',
    nullable: true,
    name: 'authenticated_rate_limit',
  })
  authenticatedRateLimit: number | null

  @Column({
    type: 'int',
    nullable: true,
    name: 'sandbox_create_rate_limit',
  })
  sandboxCreateRateLimit: number | null

  @Column({
    type: 'int',
    nullable: true,
    name: 'sandbox_lifecycle_rate_limit',
  })
  sandboxLifecycleRateLimit: number | null

  @Column({
    type: 'int',
    nullable: true,
    name: 'authenticated_rate_limit_ttl_seconds',
  })
  authenticatedRateLimitTtlSeconds: number | null

  @Column({
    type: 'int',
    nullable: true,
    name: 'sandbox_create_rate_limit_ttl_seconds',
  })
  sandboxCreateRateLimitTtlSeconds: number | null

  @Column({
    type: 'int',
    nullable: true,
    name: 'sandbox_lifecycle_rate_limit_ttl_seconds',
  })
  sandboxLifecycleRateLimitTtlSeconds: number | null

  @OneToMany(() => RegionQuota, (quota) => quota.organization, {
    cascade: true,
    onDelete: 'CASCADE',
  })
  regionQuotas: RegionQuota[]

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

  /**
   * @see {@link isSuspended} to account for temporary suspensions.
   */
  @Column({
    default: false,
  })
  suspended: boolean

  @Column({
    nullable: true,
    type: 'timestamp with time zone',
  })
  suspendedAt?: Date

  @Column({
    nullable: true,
  })
  suspensionReason?: string

  @Column({
    type: 'int',
    default: 24,
  })
  suspensionCleanupGracePeriodHours: number

  @Column({
    nullable: true,
    type: 'timestamp with time zone',
  })
  suspendedUntil?: Date

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

  @Column({
    type: 'jsonb',
    nullable: true,
    name: 'experimentalConfig',
  })
  // configuration for experimental features
  _experimentalConfig: Record<string, any> | null

  get sandboxMetadata(): Record<string, string> {
    return {
      organizationId: this.id,
      organizationName: this.name,
      limitNetworkEgress: String(this.sandboxLimitedNetworkEgress),
    }
  }

  @Column({
    nullable: true,
    type: 'timestamp with time zone',
  })
  deletedAt?: Date

  /**
   * Whether the organization is currently suspended, accounting for temporary suspensions via {@link suspendedUntil}.
   */
  get isSuspended(): boolean {
    // not suspended
    if (!this.suspended) {
      return false
    }

    // permanently suspended
    if (!this.suspendedUntil) {
      return true
    }

    // temporarily suspended, check if suspended until is in the future
    return this.suspendedUntil > new Date()
  }

  constructor(defaultRegionId?: string) {
    if (defaultRegionId) {
      this.defaultRegionId = defaultRegionId
    }
  }
}
