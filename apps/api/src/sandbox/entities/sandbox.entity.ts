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
  PrimaryColumn,
  OneToOne,
  Unique,
  UpdateDateColumn,
} from 'typeorm'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { v4 as uuidv4 } from 'uuid'
import { SandboxVolume } from '../dto/sandbox.dto'
import { BuildInfo } from './build-info.entity'
import { nanoid } from 'nanoid'
import { SandboxLastActivity } from './sandbox-last-activity.entity'
import { SandboxStateEntity } from './sandbox-state.entity'
import { SandboxBackupEntity } from './sandbox-backup.entity'

@Entity()
@Unique(['organizationId', 'name'])
@Index('sandbox_snapshot_idx', ['snapshot'])
@Index('sandbox_organizationid_idx', ['organizationId'])
@Index('sandbox_region_idx', ['region'])
@Index('sandbox_resources_idx', ['cpu', 'mem', 'disk', 'gpu'])
@Index('idx_sandbox_authtoken', ['authToken'])
@Index('sandbox_labels_gin_full_idx', { synchronize: false })
@Index('idx_sandbox_volumes_gin', { synchronize: false })
export class Sandbox {
  @PrimaryColumn({ default: () => 'uuid_generate_v4()' })
  id: string

  @Column({
    type: 'uuid',
  })
  organizationId: string

  @Column()
  name: string

  @Column()
  region: string

  @Column({
    type: 'enum',
    enum: SandboxClass,
    default: SandboxClass.SMALL,
  })
  class = SandboxClass.SMALL

  @Column({ nullable: true })
  snapshot?: string

  @Column()
  osUser: string

  @Column({
    type: 'jsonb',
    default: {},
  })
  env: { [key: string]: string } = {}

  @Column({ default: false, type: 'boolean' })
  public = false

  @Column({ default: false, type: 'boolean' })
  networkBlockAll = false

  @Column({ nullable: true })
  networkAllowList?: string

  @Column('jsonb', { nullable: true })
  labels: { [key: string]: string }

  @Column({ type: 'int', default: 2 })
  cpu = 2

  @Column({ type: 'int', default: 0 })
  gpu = 0

  @Column({ type: 'int', default: 4 })
  mem = 4

  @Column({ type: 'int', default: 10 })
  disk = 10

  @Column({
    type: 'jsonb',
    default: [],
  })
  volumes: SandboxVolume[] = []

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date

  @OneToOne(() => SandboxLastActivity, (lastActivity) => lastActivity.sandbox)
  lastActivityAt?: SandboxLastActivity

  @Column({ default: 15, type: 'int' })
  autoStopInterval: number | undefined = 15

  @Column({ default: 7 * 24 * 60, type: 'int' })
  autoArchiveInterval: number | undefined = 7 * 24 * 60

  @Column({ default: -1, type: 'int' })
  autoDeleteInterval: number | undefined = -1

  @Column({ type: 'character varying' })
  authToken = nanoid(32).toLowerCase()

  @ManyToOne(() => BuildInfo, (buildInfo) => buildInfo.sandboxes, {
    nullable: true,
  })
  @JoinColumn()
  buildInfo?: BuildInfo

  @OneToOne(() => SandboxStateEntity, (s) => s.sandbox)
  sandboxState!: SandboxStateEntity

  @OneToOne(() => SandboxBackupEntity, (b) => b.sandbox)
  sandboxBackup!: SandboxBackupEntity

  constructor(region: string, name?: string) {
    this.id = uuidv4()
    this.name = name || this.id
    this.region = region
  }
}
