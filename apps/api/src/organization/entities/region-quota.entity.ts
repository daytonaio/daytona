/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, JoinColumn, ManyToOne, PrimaryColumn, UpdateDateColumn } from 'typeorm'
import { Organization } from './organization.entity'
import { SandboxClass } from '../../sandbox/enums/sandbox-class.enum'
import { GpuType } from '../../sandbox/enums/gpu-type.enum'

const DEFAULT_MAX_CPU_PER_GPU_SANDBOX = 16
const DEFAULT_MAX_MEMORY_PER_GPU_SANDBOX = 192
const DEFAULT_MAX_DISK_PER_GPU_SANDBOX = 512

@Entity()
export class RegionQuota {
  @PrimaryColumn()
  organizationId: string

  @PrimaryColumn()
  regionId: string

  @PrimaryColumn({ type: 'character varying', default: SandboxClass.CONTAINER })
  sandboxClass: SandboxClass

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
    default: 0,
    name: 'total_gpu_quota',
  })
  totalGpuQuota: number

  /**
   * List of GPU types permitted in this region.
   * `null` = no restriction.
   */
  @Column({
    type: 'text',
    array: true,
    nullable: true,
    name: 'allowed_gpu_types',
  })
  allowedGpuTypes: GpuType[] | null

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

  /**
   * If `null`, fallback to `maxCpuPerSandbox`.
   * Typically larger than `maxCpuPerSandbox` since GPU sandboxes own the whole runner exclusively.
   */
  @Column({
    type: 'int',
    nullable: true,
    default: DEFAULT_MAX_CPU_PER_GPU_SANDBOX,
    name: 'max_cpu_per_gpu_sandbox',
  })
  maxCpuPerGpuSandbox: number | null

  /**
   * If `null`, fallback to `maxMemoryPerSandbox`.
   * Typically larger than `maxMemoryPerSandbox` since GPU sandboxes own the whole runner exclusively.
   */
  @Column({
    type: 'int',
    nullable: true,
    default: DEFAULT_MAX_MEMORY_PER_GPU_SANDBOX,
    name: 'max_memory_per_gpu_sandbox',
  })
  maxMemoryPerGpuSandbox: number | null

  /**
   * If `null`, fallback to `maxDiskPerSandbox`.
   * Typically larger than `maxDiskPerSandbox` since GPU sandboxes own the whole runner exclusively.
   */
  @Column({
    type: 'int',
    nullable: true,
    default: DEFAULT_MAX_DISK_PER_GPU_SANDBOX,
    name: 'max_disk_per_gpu_sandbox',
  })
  maxDiskPerGpuSandbox: number | null

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date

  constructor(params?: {
    organizationId: string
    regionId: string
    sandboxClass: SandboxClass
    totalCpuQuota: number
    totalMemoryQuota: number
    totalDiskQuota: number
    totalGpuQuota?: number
    allowedGpuTypes?: GpuType[] | null
    maxCpuPerSandbox?: number | null
    maxMemoryPerSandbox?: number | null
    maxDiskPerSandbox?: number | null
    maxDiskPerNonEphemeralSandbox?: number | null
    maxCpuPerGpuSandbox?: number | null
    maxMemoryPerGpuSandbox?: number | null
    maxDiskPerGpuSandbox?: number | null
  }) {
    if (!params) return
    this.organizationId = params.organizationId
    this.regionId = params.regionId
    this.sandboxClass = params.sandboxClass
    this.totalCpuQuota = params.totalCpuQuota
    this.totalMemoryQuota = params.totalMemoryQuota
    this.totalDiskQuota = params.totalDiskQuota
    this.totalGpuQuota = params.totalGpuQuota ?? 0
    this.allowedGpuTypes = params.allowedGpuTypes ?? null
    this.maxCpuPerSandbox = params.maxCpuPerSandbox ?? null
    this.maxMemoryPerSandbox = params.maxMemoryPerSandbox ?? null
    this.maxDiskPerSandbox = params.maxDiskPerSandbox ?? null
    this.maxDiskPerNonEphemeralSandbox = params.maxDiskPerNonEphemeralSandbox ?? null
    this.maxCpuPerGpuSandbox = params.maxCpuPerGpuSandbox ?? DEFAULT_MAX_CPU_PER_GPU_SANDBOX
    this.maxMemoryPerGpuSandbox = params.maxMemoryPerGpuSandbox ?? DEFAULT_MAX_MEMORY_PER_GPU_SANDBOX
    this.maxDiskPerGpuSandbox = params.maxDiskPerGpuSandbox ?? DEFAULT_MAX_DISK_PER_GPU_SANDBOX
  }
}
