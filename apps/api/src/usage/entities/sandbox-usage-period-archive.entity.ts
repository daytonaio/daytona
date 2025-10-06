/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, Entity, PrimaryGeneratedColumn } from 'typeorm'
import { SandboxUsagePeriod } from './sandbox-usage-period.entity'
import { v4 } from 'uuid'

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

  @Column({ type: 'float' })
  mem: number

  @Column({ type: 'float' })
  disk: number

  @Column()
  region: string

  constructor(usagePeriod: SandboxUsagePeriod) {
    if (!usagePeriod.endAt) {
      throw new Error('Usage period must be closed is required')
    }

    this.id = v4()
    this.sandboxId = usagePeriod.sandboxId
    this.organizationId = usagePeriod.organizationId
    this.startAt = usagePeriod.startAt
    this.endAt = usagePeriod.endAt
    this.cpu = usagePeriod.cpu
    this.gpu = usagePeriod.gpu
    this.mem = usagePeriod.mem
    this.disk = usagePeriod.disk
    this.region = usagePeriod.region
  }
}
