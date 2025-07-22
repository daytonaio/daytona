/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, PrimaryGeneratedColumn, UpdateDateColumn } from 'typeorm'
import { SnapshotRunnerState } from '../enums/snapshot-runner-state.enum'

@Entity()
export class SnapshotRunner {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({
    type: 'enum',
    enum: SnapshotRunnerState,
    default: SnapshotRunnerState.PULLING_SNAPSHOT,
  })
  state: SnapshotRunnerState

  @Column({ nullable: true })
  errorReason?: string

  @Column({
    //  todo: remove default
    default: '',
  })
  snapshotRef: string

  @Column()
  runnerId: string

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date
}
