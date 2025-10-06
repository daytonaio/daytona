/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, Entity, PrimaryGeneratedColumn } from 'typeorm'
import { v4 } from 'uuid'

@Entity('sandbox_usage_periods')
export class SandboxUsagePeriod {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column()
  sandboxId: string

  @Column()
  // Redundant property to optimize billing queries
  organizationId: string

  @Column({ type: 'timestamp with time zone' })
  startAt: Date

  @Column({ type: 'timestamp with time zone', nullable: true })
  endAt: Date | null

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

  constructor({
    sandboxId,
    organizationId,
    startAt,
    endAt,
    cpu,
    gpu,
    mem,
    disk,
    region,
  }: {
    sandboxId: string
    organizationId: string
    startAt: Date
    endAt: Date | null
    cpu: number
    gpu: number
    mem: number
    disk: number
    region: string
  }) {
    this.id = v4()
    this.sandboxId = sandboxId
    this.organizationId = organizationId
    this.startAt = startAt
    this.endAt = endAt
    this.cpu = cpu
    this.gpu = gpu
    this.mem = mem
    this.disk = disk
    this.region = region
  }
}
