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
import { v4 } from 'uuid'

@Entity()
@Unique(['organizationId', 'name'])
export class Snapshot {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({
    nullable: true,
    type: 'uuid',
  })
  organizationId: string | null

  //  general snapshot is available to all organizations
  @Column({ default: false })
  general: boolean

  @Column()
  name: string

  @Column()
  imageName: string

  @Column({ nullable: true, type: String })
  internalName: string | null

  @Column({
    type: 'enum',
    enum: SnapshotState,
    default: SnapshotState.PENDING,
  })
  state: SnapshotState

  @Column({ nullable: true, type: String })
  errorReason: string | null

  @Column({ type: 'float', nullable: true })
  size: number | null

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
  entrypoint: string[] | null

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date

  @Column({ nullable: true, type: 'timestamp with time zone' })
  lastUsedAt: Date | null

  @ManyToOne(() => BuildInfo, (buildInfo) => buildInfo.snapshots, {
    nullable: true,
    eager: true,
  })
  @JoinColumn()
  buildInfo: BuildInfo | null

  @Column({ nullable: true, type: String })
  buildRunnerId: string | null

  constructor(createParams: {
    organizationId: string | null
    name: string
    imageName: string
    cpu?: number
    gpu?: number
    mem?: number
    disk?: number
    entrypoint?: string[] | null
    general?: boolean
    hideFromUsers?: boolean
    buildInfo?: BuildInfo
    buildRunnerId?: string
  }) {
    this.id = v4()
    this.organizationId = createParams.organizationId
    this.name = createParams.name
    this.imageName = createParams.imageName
    this.cpu = createParams.cpu ?? 1
    this.gpu = createParams.gpu ?? 0
    this.mem = createParams.mem ?? 1
    this.disk = createParams.disk ?? 3
    this.entrypoint = createParams.entrypoint ?? null
    this.general = createParams.general ?? false
    this.hideFromUsers = createParams.hideFromUsers ?? false
    this.buildRunnerId = createParams.buildRunnerId ?? null
    this.buildInfo = createParams.buildInfo ?? null
    this.runners = []

    this.createdAt = new Date()
    this.updatedAt = new Date()
    this.state = SnapshotState.PENDING
    this.size = null
    this.errorReason = null
    this.lastUsedAt = null
    this.internalName = null
  }
}
