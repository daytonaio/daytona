/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, PrimaryGeneratedColumn, UpdateDateColumn } from 'typeorm'
import { SnapshotRunnerState } from '../enums/snapshot-runner-state.enum'
import { v4 } from 'uuid'

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

  @Column({ nullable: true, type: String })
  errorReason: string | null

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

  constructor(createParams: {
    snapshotRef: string
    runnerId: string
    state?: SnapshotRunnerState
    errorReason?: string
  }) {
    this.id = v4()
    this.snapshotRef = createParams.snapshotRef
    this.runnerId = createParams.runnerId
    this.state = createParams.state ?? SnapshotRunnerState.PULLING_SNAPSHOT
    this.errorReason = createParams.errorReason ?? null
    this.createdAt = new Date()
    this.updatedAt = new Date()
  }
}
