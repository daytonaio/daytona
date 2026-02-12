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
  OneToMany,
  PrimaryGeneratedColumn,
  Unique,
  UpdateDateColumn,
} from 'typeorm'
import { CheckpointState } from '../enums/checkpoint-state.enum'
import { CheckpointRunner } from './checkpoint-runner.entity'
import { Sandbox } from './sandbox.entity'

@Entity()
@Unique(['sandboxId', 'name'])
@Index('checkpoint_sandboxid_idx', ['sandboxId'])
@Index('checkpoint_organizationid_idx', ['organizationId'])
@Index('checkpoint_state_idx', ['state'])
export class Checkpoint {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({ type: 'uuid' })
  sandboxId: string

  @Column({ type: 'uuid' })
  organizationId: string

  @Column()
  name: string

  @Column({ nullable: true })
  ref?: string

  @Column({
    type: 'enum',
    enum: CheckpointState,
    default: CheckpointState.CREATING,
  })
  state: CheckpointState = CheckpointState.CREATING

  @Column({ nullable: true })
  errorReason?: string

  @Column({ type: 'bigint', nullable: true })
  sizeBytes?: number

  @Column({ nullable: true })
  hash?: string

  @ManyToOne(() => Sandbox, { onDelete: 'CASCADE' })
  @JoinColumn({ name: 'sandboxId' })
  sandbox: Sandbox

  @OneToMany(() => CheckpointRunner, (cr) => cr.checkpoint, {
    cascade: true,
    onDelete: 'CASCADE',
  })
  runners: CheckpointRunner[]

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date
}
