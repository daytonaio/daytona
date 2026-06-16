/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, Entity, PrimaryGeneratedColumn } from 'typeorm'
import { SandboxUsagePeriod } from './sandbox-usage-period.entity'
import { SandboxClass } from '../../sandbox/enums/sandbox-class.enum'
import { GpuType } from '../../sandbox/enums/gpu-type.enum'
import { RegionType } from '../../region/enums/region-type.enum'

// Duplicate of SandboxUsagePeriod
// Used to archive usage periods and keep the original table lightweight
// Will only contain closed usage periods
@Entity('sandbox_usage_periods_archive')
export class SandboxUsagePeriodArchive {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column()
  sandboxId: string

  @Column()
  // Redundant property to optimize billing queries
  organizationId: string

  @Column({ type: 'timestamp with time zone' })
  startAt: Date

  @Column({ type: 'timestamp with time zone' })
  endAt: Date

  @Column({ type: 'float' })
  cpu: number

  @Column({ type: 'float' })
  gpu: number

  @Column({
    type: 'character varying',
    nullable: true,
    name: 'gpu_type',
  })
  gpuType: GpuType | null

  @Column({ type: 'float' })
  mem: number

  @Column({ type: 'float' })
  disk: number

  @Column()
  region: string

  @Column({
    type: 'character varying',
    default: SandboxClass.CONTAINER,
  })
  sandboxClass: SandboxClass = SandboxClass.CONTAINER

  @Column({ type: 'character varying', default: RegionType.SHARED })
  regionType: string

  public static fromUsagePeriod(usagePeriod: SandboxUsagePeriod) {
    const usagePeriodEntity = new SandboxUsagePeriodArchive()
    usagePeriodEntity.sandboxId = usagePeriod.sandboxId
    usagePeriodEntity.organizationId = usagePeriod.organizationId
    usagePeriodEntity.startAt = usagePeriod.startAt
    usagePeriodEntity.endAt = usagePeriod.endAt
    usagePeriodEntity.cpu = usagePeriod.cpu
    usagePeriodEntity.gpu = usagePeriod.gpu
    usagePeriodEntity.gpuType = usagePeriod.gpuType
    usagePeriodEntity.mem = usagePeriod.mem
    usagePeriodEntity.disk = usagePeriod.disk
    usagePeriodEntity.region = usagePeriod.region
    usagePeriodEntity.sandboxClass = usagePeriod.sandboxClass
    usagePeriodEntity.regionType = usagePeriod.regionType
    return usagePeriodEntity
  }
}
