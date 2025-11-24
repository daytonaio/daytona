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
  ) {
    this.organizationId = organizationId
    this.regionId = regionId
    this.totalCpuQuota = totalCpuQuota
    this.totalMemoryQuota = totalMemoryQuota
    this.totalDiskQuota = totalDiskQuota
  }
}
