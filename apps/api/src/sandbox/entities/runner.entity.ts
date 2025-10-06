/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, PrimaryGeneratedColumn, UpdateDateColumn } from 'typeorm'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { RunnerState } from '../enums/runner-state.enum'
import { v4 } from 'uuid'

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
  lastChecked: Date | null

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

  constructor(createParams: {
    domain: string
    apiUrl: string
    proxyUrl: string
    apiKey: string
    cpu: number
    memoryGiB: number
    diskGiB: number
    gpu: number
    gpuType: string
    class: SandboxClass
    region: string
    version?: string
  }) {
    this.id = v4()
    this.domain = createParams.domain
    this.apiUrl = createParams.apiUrl
    this.proxyUrl = createParams.proxyUrl
    this.apiKey = createParams.apiKey
    this.cpu = createParams.cpu
    this.memoryGiB = createParams.memoryGiB
    this.diskGiB = createParams.diskGiB
    this.gpu = createParams.gpu
    this.gpuType = createParams.gpuType
    this.class = createParams.class
    this.region = createParams.region
    this.version = createParams.version ?? '0'
    this.state = RunnerState.INITIALIZING
    this.currentCpuUsagePercentage = 0
    this.currentMemoryUsagePercentage = 0
    this.currentDiskUsagePercentage = 0
    this.currentAllocatedCpu = 0
    this.currentAllocatedMemoryGiB = 0
    this.currentAllocatedDiskGiB = 0
    this.currentSnapshotCount = 0
    this.availabilityScore = 0
    this.unschedulable = false
    this.lastChecked = null
    this.createdAt = new Date()
    this.updatedAt = new Date()
  }
}
