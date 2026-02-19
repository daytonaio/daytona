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
import { SnapshotRunner } from './snapshot-runner.entity'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { BuildInfo } from './build-info.entity'
import { SnapshotRegion } from './snapshot-region.entity'
import { Checkpoint } from './checkpoint.entity'

@Entity()
@Unique(['organizationId', 'name'])
@Index('snapshot_name_idx', ['name'])
@Index('snapshot_state_idx', ['state'])
export class Snapshot {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({
    nullable: true,
    type: 'uuid',
  })
  organizationId?: string

  //  general snapshot is available to all organizations
  @Column({
    type: 'boolean',
    default: false,
  })
  general = false

  @Column()
  name: string

  @Column()
  imageName: string

  @Column({ nullable: true })
  ref?: string

  @Column({
    type: 'enum',
    enum: SnapshotState,
    default: SnapshotState.PENDING,
  })
  state = SnapshotState.PENDING

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

  @Column({ type: 'boolean', default: false })
  hideFromUsers = false

  @OneToMany(() => SnapshotRunner, (runner) => runner.snapshotRef)
  runners: SnapshotRunner[]

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

  @ManyToOne(() => BuildInfo, (buildInfo) => buildInfo.snapshots, {
    nullable: true,
    eager: true,
  })
  @JoinColumn()
  buildInfo?: BuildInfo

  @Column({ nullable: true })
  initialRunnerId?: string

  @OneToMany(() => SnapshotRegion, (snapshotRegion) => snapshotRegion.snapshot, {
    cascade: true,
    onDelete: 'CASCADE',
  })
  snapshotRegions: SnapshotRegion[]

  @ManyToOne(() => Checkpoint, (checkpoint) => checkpoint.snapshots, {
    nullable: true,
  })
  @JoinColumn()
  checkpoint?: Checkpoint
}
