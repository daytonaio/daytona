/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, JoinColumn, ManyToOne, PrimaryColumn, UpdateDateColumn } from 'typeorm'
import { Organization } from './organization.entity'

@Entity()
export class RegionQuota {
  @PrimaryColumn()
  organizationId: string

  @PrimaryColumn()
  regionId: string

  @ManyToOne(() => Organization, (organization) => organization.regionQuotas, {
    onDelete: 'CASCADE',
  })
  @JoinColumn({ name: 'organizationId' })
  organization: Organization

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
    nullable: true,
    name: 'max_cpu_per_sandbox',
  })
  maxCpuPerSandbox: number | null

  @Column({
    type: 'int',
    nullable: true,
    name: 'max_memory_per_sandbox',
  })
  maxMemoryPerSandbox: number | null

  @Column({
    type: 'int',
    nullable: true,
    name: 'max_disk_per_sandbox',
  })
  maxDiskPerSandbox: number | null

  /**
   * The maximum disk size allowed for non-ephemeral sandboxes.
   * If `null`, fallback to `maxDiskPerSandbox`.
   * If `0`, non-ephemeral sandboxes are not permitted in this region.
   */
  @Column({
    type: 'int',
    nullable: true,
    name: 'max_disk_per_non_ephemeral_sandbox',
  })
  maxDiskPerNonEphemeralSandbox: number | null

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date

  constructor(
    organizationId: string,
    regionId: string,
    totalCpuQuota: number,
    totalMemoryQuota: number,
    totalDiskQuota: number,
    maxCpuPerSandbox: number | null = null,
    maxMemoryPerSandbox: number | null = null,
    maxDiskPerSandbox: number | null = null,
    maxDiskPerNonEphemeralSandbox: number | null = null,
  ) {
    this.organizationId = organizationId
    this.regionId = regionId
    this.totalCpuQuota = totalCpuQuota
    this.totalMemoryQuota = totalMemoryQuota
    this.totalDiskQuota = totalDiskQuota
    this.maxCpuPerSandbox = maxCpuPerSandbox
    this.maxMemoryPerSandbox = maxMemoryPerSandbox
    this.maxDiskPerSandbox = maxDiskPerSandbox
    this.maxDiskPerNonEphemeralSandbox = maxDiskPerNonEphemeralSandbox
  }
}
