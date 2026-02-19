/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Column,
  CreateDateColumn,
  Entity,
  Index,
  OneToMany,
  PrimaryGeneratedColumn,
  Unique,
  UpdateDateColumn,
} from 'typeorm'
import { CheckpointState } from '../enums/checkpoint-state.enum'
import { SandboxVolume } from '../dto/sandbox.dto'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { SnapshotRunner } from './snapshot-runner.entity'
import { Sandbox } from './sandbox.entity'
import { Snapshot } from './snapshot.entity'

@Entity()
@Unique(['organizationId', 'name'])
@Index('checkpoint_origin_sandboxid_idx', ['originSandboxId']) // TODO: DZ - check indices
@Index('checkpoint_organizationid_idx', ['organizationId'])
@Index('checkpoint_state_idx', ['state'])
export class Checkpoint {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({ type: 'uuid' })
  originSandboxId: string

  @Column({ type: 'uuid' })
  organizationId: string

  @Column()
  name: string

  @OneToMany(() => Sandbox, (sandbox) => sandbox.checkpoint)
  sandboxes: Sandbox[]

  @OneToMany(() => Snapshot, (snapshot) => snapshot.checkpoint)
  snapshots: Snapshot[]

  @OneToMany(() => SnapshotRunner, (runner) => runner.snapshotRef)
  runners: SnapshotRunner[]

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

  @Column({ type: 'float', nullable: true })
  size?: number

  @Column({ type: 'int', default: 1 })
  cpu = 1

  @Column({ type: 'int', default: 0 })
  gpu = 0

  @Column({ type: 'int', default: 1 })
  mem = 1

  @Column({ type: 'int', default: 3 })
  disk = 3

  @Column({ array: true, type: 'text', nullable: true })
  entrypoint?: string[]

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date

  @Column({ nullable: true })
  lastUsedAt?: Date

  @Column({ nullable: true })
  initialRunnerId?: string

  // --- Sandbox creation specific params captured at checkpoint time ---

  @Column()
  region: string

  @Column({
    type: 'enum',
    enum: SandboxClass,
    default: SandboxClass.SMALL,
  })
  class: SandboxClass

  @Column()
  osUser: string

  @Column({
    type: 'jsonb',
    default: {},
  })
  env: { [key: string]: string }

  @Column({ default: false })
  public: boolean

  @Column({ default: false })
  networkBlockAll: boolean

  @Column({ nullable: true })
  networkAllowList?: string

  @Column('jsonb', { nullable: true })
  labels: { [key: string]: string }

  @Column({
    type: 'jsonb',
    default: [],
  })
  volumes: SandboxVolume[]

  @Column({ default: 15 })
  autoStopInterval?: number

  @Column({ default: 7 * 24 * 60 })
  autoArchiveInterval?: number

  @Column({ default: -1 })
  autoDeleteInterval?: number

  @Column({ nullable: true })
  buildInfoSnapshotRef?: string
}
