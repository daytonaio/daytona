/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Column,
  CreateDateColumn,
  Entity,
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

@Entity()
@Unique(['organizationId', 'name'])
export class Snapshot {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({
    nullable: true,
    type: 'uuid',
  })
  organizationId?: string

  //  general snapshot is available to all organizations
  @Column({ default: false })
  general: boolean

  @Column()
  name: string

  @Column()
  imageName: string

  @Column({ nullable: true })
  internalName?: string

  @Column({ default: true })
  enabled: boolean

  @Column({
    type: 'enum',
    enum: SnapshotState,
    default: SnapshotState.PENDING,
  })
  state: SnapshotState

  @Column({ nullable: true })
  errorReason?: string

  @Column({ type: 'float', nullable: true })
  size?: number

  @Column({ type: 'int', default: 1 })
  cpu: number

  @Column({ type: 'int', default: 0 })
  gpu: number

  @Column({ type: 'int', default: 1 })
  mem: number

  @Column({ type: 'int', default: 3 })
  disk: number

  @Column({ default: false })
  hideFromUsers: boolean

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
  buildRunnerId?: string
}
