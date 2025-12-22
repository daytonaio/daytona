/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, Index, PrimaryGeneratedColumn, Unique, UpdateDateColumn } from 'typeorm'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { RunnerState } from '../enums/runner-state.enum'

@Entity()
@Unique(['region', 'name'])
@Index(['state', 'unschedulable', 'region'])
export class Runner {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({
    nullable: true,
    unique: true,
  })
  domain: string | null

  @Column({
    nullable: true,
  })
  apiUrl: string | null

  @Column({
    nullable: true,
  })
  proxyUrl: string | null

  @Column()
  apiKey: string

  @Column({
    type: 'float',
    default: 0,
  })
  cpu: number

  @Column({
    type: 'float',
    default: 0,
  })
  memoryGiB: number

  @Column({
    type: 'float',
    default: 0,
  })
  diskGiB: number

  @Column({
    nullable: true,
  })
  gpu: number | null

  @Column({
    nullable: true,
  })
  gpuType: string | null

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

  @Column()
  name: string

  @Column({
    type: 'enum',
    enum: RunnerState,
    default: RunnerState.INITIALIZING,
  })
  state: RunnerState

  @Column({
    default: 'v0.0.0-dev',
    nullable: true,
  })
  appVersion: string | null

  @Column({
    default: '0',
  })
  apiVersion: string

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

  constructor(params: {
    region: string
    name: string
    apiKey: string
    apiVersion: string
    cpu?: number
    memoryGiB?: number
    diskGiB?: number
    domain?: string | null
    apiUrl?: string
    proxyUrl?: string
    appVersion?: string | null
  }) {
    this.region = params.region
    this.name = params.name
    this.apiKey = params.apiKey
    this.cpu = params.cpu ?? 0
    this.memoryGiB = params.memoryGiB ?? 0
    this.diskGiB = params.diskGiB ?? 0
    this.domain = params.domain ?? null
    this.apiUrl = params.apiUrl
    this.proxyUrl = params.proxyUrl
    this.class = SandboxClass.SMALL
    this.apiVersion = params.apiVersion
    this.appVersion = params.appVersion ?? null
    this.gpu = null
    this.gpuType = null
  }
}
