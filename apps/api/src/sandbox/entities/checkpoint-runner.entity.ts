/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Column,
  CreateDateColumn,
  Entity,
  Index,
  JoinColumn,
  ManyToOne,
  PrimaryGeneratedColumn,
  UpdateDateColumn,
} from 'typeorm'
import { CheckpointRunnerState } from '../enums/checkpoint-runner-state.enum'
import { Checkpoint } from './checkpoint.entity'

@Entity()
@Index('checkpoint_runner_checkpointid_idx', ['checkpointId'])
@Index('checkpoint_runner_runnerid_idx', ['runnerId'])
@Index('checkpoint_runner_state_idx', ['state'])
export class CheckpointRunner {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({ type: 'uuid' })
  checkpointId: string

  @Column()
  runnerId: string

  @Column({
    type: 'enum',
    enum: CheckpointRunnerState,
    default: CheckpointRunnerState.PULLING,
  })
  state: CheckpointRunnerState

  @Column({ nullable: true })
  errorReason?: string

  @ManyToOne(() => Checkpoint, (c) => c.runners, { onDelete: 'CASCADE' })
  @JoinColumn({ name: 'checkpointId' })
  checkpoint: Checkpoint

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date
}
