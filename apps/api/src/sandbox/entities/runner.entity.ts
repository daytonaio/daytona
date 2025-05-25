/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, PrimaryGeneratedColumn, UpdateDateColumn } from 'typeorm'
import { RunnerRegion } from '../enums/runner-region.enum'
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
  apiKey: string

  @Column()
  cpu: number

  @Column()
  memory: number

  @Column()
  disk: number

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
    type: 'enum',
    enum: RunnerRegion,
  })
  region: RunnerRegion

  @Column({
    type: 'enum',
    enum: RunnerState,
    default: RunnerState.INITIALIZING,
  })
  state: RunnerState

  @Column({
    nullable: true,
  })
  lastChecked: Date

  @Column({
    default: false,
  })
  unschedulable: boolean

  @CreateDateColumn()
  createdAt: Date

  @UpdateDateColumn()
  updatedAt: Date
}
