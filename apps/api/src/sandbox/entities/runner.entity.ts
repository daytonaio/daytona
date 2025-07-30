/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, PrimaryGeneratedColumn, UpdateDateColumn } from 'typeorm'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { RunnerState } from '../enums/runner-state.enum'

@Entity()
export class Runner {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({ unique: true })
  domain: string

  @Column()
  apiUrl: string

  @Column()
  proxyUrl: string

  @Column()
  apiKey: string

  @Column()
  cpu: number

  @Column()
  memoryGiB: number

  @Column()
  diskGiB: number

  @Column()
  gpu: number

  @Column()
  gpuType: string

  @Column({
    type: 'enum',
    enum: SandboxClass,
    default: SandboxClass.SMALL,
  })
  class: SandboxClass

  @Column({
    default: 0,
  })
  used: number

  @Column()
  capacity: number

  @Column({
    type: 'float',
    default: 0,
  })
  currentCpuUsagePercentage: number

  @Column({
    type: 'float',
    default: 0,
  })
  currentMemoryUsagePercentage: number

  @Column({
    type: 'float',
    default: 0,
  })
  currentDiskUsagePercentage: number

  @Column({
    default: 0,
  })
  currentAllocatedCpu: number

  @Column({
    default: 0,
  })
  currentAllocatedMemoryGiB: number

  @Column({
    default: 0,
  })
  currentAllocatedDiskGiB: number

  @Column({
    default: 0,
  })
  currentSnapshotCount: number

  @Column({
    default: 0,
  })
  availabilityScore: number

  @Column()
  region: string

  @Column({
    type: 'enum',
    enum: RunnerState,
    default: RunnerState.INITIALIZING,
  })
  state: RunnerState

  @Column({
    default: '0',
  })
  version: string

  @Column({
    nullable: true,
    type: 'timestamp with time zone',
  })
  lastChecked: Date

  @Column({
    default: false,
  })
  unschedulable: boolean

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date
}
