/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  BeforeUpdate,
  Column,
  CreateDateColumn,
  Entity,
  JoinColumn,
  ManyToOne,
  PrimaryGeneratedColumn,
  Unique,
  UpdateDateColumn,
} from 'typeorm'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { BackupState } from '../enums/backup-state.enum'
import { nanoid } from 'nanoid'
import { v4 as uuidv4 } from 'uuid'
import { SandboxVolume } from '../dto/sandbox.dto'
import { BuildInfo } from './build-info.entity'

@Entity()
@Unique(['organizationId', 'name'])
export class Sandbox {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({
    type: 'uuid',
  })
  organizationId: string

  @Column()
  name: string

  @Column({
    default: 'us',
  })
  region: string

  @Column({
    type: 'uuid',
    nullable: true,
  })
  runnerId: string | null

  //  this is the runnerId of the runner that was previously assigned to the sandbox
  //  if something goes wrong with new runner assignment, we can revert to the previous runner
  @Column({
    type: 'uuid',
    nullable: true,
  })
  prevRunnerId: string | null

  @Column({
    type: 'enum',
    enum: SandboxClass,
    default: SandboxClass.SMALL,
  })
  class: SandboxClass

  @Column({
    type: 'enum',
    enum: SandboxState,
    default: SandboxState.UNKNOWN,
  })
  state: SandboxState

  @Column({
    type: 'enum',
    enum: SandboxDesiredState,
    default: SandboxDesiredState.STARTED,
  })
  desiredState: SandboxDesiredState

  @Column({ nullable: true, type: String })
  snapshot: string | null

  @Column()
  osUser: string

  @Column({ nullable: true, type: String })
  errorReason: string | null

  @Column({
    type: 'jsonb',
    default: {},
  })
  env: { [key: string]: string }

  @Column({ default: false })
  public: boolean

  @Column({ default: false })
  networkBlockAll: boolean

  @Column({ nullable: true, type: String })
  networkAllowList: string | null

  @Column('jsonb', { nullable: true })
  labels: { [key: string]: string } | null

  @Column({ nullable: true, type: String })
  backupRegistryId: string | null

  @Column({ nullable: true, type: String })
  backupSnapshot: string | null

  @Column({ nullable: true, type: 'timestamp with time zone' })
  lastBackupAt: Date | null

  @Column({
    type: 'enum',
    enum: BackupState,
    default: BackupState.NONE,
  })
  backupState: BackupState

  @Column({
    type: 'text',
    nullable: true,
  })
  backupErrorReason: string | null

  @Column({
    type: 'jsonb',
    default: [],
  })
  existingBackupSnapshots: Array<{
    snapshotName: string
    createdAt: Date
  }>

  @Column({ type: 'int', default: 2 })
  cpu: number

  @Column({ type: 'int', default: 0 })
  gpu: number

  @Column({ type: 'int', default: 4 })
  mem: number

  @Column({ type: 'int', default: 10 })
  disk: number

  @Column({
    type: 'jsonb',
    default: [],
  })
  volumes: SandboxVolume[]

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date

  @Column({ nullable: true, type: 'timestamp with time zone' })
  lastActivityAt: Date | null

  //  this is the interval in minutes after which the sandbox will be stopped if lastActivityAt is not updated
  //  if set to 0, auto stop will be disabled
  @Column({ default: 15 })
  autoStopInterval: number

  //  this is the interval in minutes after which a continuously stopped workspace will be automatically archived
  @Column({ default: 7 * 24 * 60 })
  autoArchiveInterval: number

  //  this is the interval in minutes after which a continuously stopped workspace will be automatically deleted
  //  if set to negative value, auto delete will be disabled
  //  if set to 0, sandbox will be immediately deleted upon stopping
  @Column({ default: -1 })
  autoDeleteInterval: number

  @Column({ default: false })
  pending: boolean

  @Column({ default: () => 'MD5(random()::text)' })
  authToken: string

  @ManyToOne(() => BuildInfo, (buildInfo) => buildInfo.sandboxes, {
    nullable: true,
    eager: true,
  })
  @JoinColumn()
  buildInfo: BuildInfo | null

  @Column({ nullable: true, type: String })
  daemonVersion: string | null

  constructor(createSandboxParams: {
    organizationId: string
    osUser: string
    // Optional params (all have defaults or are nullable)
    name?: string
    region?: string
    runnerId?: string | null
    class?: SandboxClass
    snapshot?: string | null
    env?: { [key: string]: string }
    public?: boolean
    networkBlockAll?: boolean
    networkAllowList?: string | null
    labels?: { [key: string]: string } | null
    cpu?: number
    gpu?: number
    mem?: number
    disk?: number
    volumes?: SandboxVolume[]
    autoStopInterval?: number
    autoArchiveInterval?: number
    autoDeleteInterval?: number
    buildInfo?: BuildInfo | null
    daemonVersion?: string | null
  }) {
    this.id = uuidv4()
    this.name = createSandboxParams.name || this.id
    this.organizationId = createSandboxParams.organizationId
    this.osUser = createSandboxParams.osUser

    this.region = createSandboxParams.region ?? 'us'
    this.runnerId = createSandboxParams.runnerId ?? null
    this.prevRunnerId = null
    this.class = createSandboxParams.class ?? SandboxClass.SMALL
    this.state = SandboxState.UNKNOWN
    this.desiredState = SandboxDesiredState.STARTED
    this.errorReason = null
    this.snapshot = createSandboxParams.snapshot ?? null
    this.env = createSandboxParams.env ?? {}
    this.public = createSandboxParams.public ?? false
    this.networkBlockAll = createSandboxParams.networkBlockAll ?? false
    this.networkAllowList = createSandboxParams.networkAllowList ?? null
    this.labels = createSandboxParams.labels ?? null
    this.backupRegistryId = null
    this.backupSnapshot = null
    this.lastBackupAt = null
    this.backupState = BackupState.NONE
    this.backupErrorReason = null
    this.existingBackupSnapshots = []
    this.cpu = createSandboxParams.cpu ?? 2
    this.gpu = createSandboxParams.gpu ?? 0
    this.mem = createSandboxParams.mem ?? 4
    this.disk = createSandboxParams.disk ?? 10
    this.volumes = createSandboxParams.volumes ?? []
    this.createdAt = new Date()
    this.updatedAt = new Date()
    this.lastActivityAt = null
    this.autoStopInterval = createSandboxParams.autoStopInterval ?? 15
    this.autoArchiveInterval = createSandboxParams.autoArchiveInterval ?? 7 * 24 * 60
    this.autoDeleteInterval = createSandboxParams.autoDeleteInterval ?? -1
    this.pending = false
    this.authToken = ''
    this.buildInfo = createSandboxParams.buildInfo ?? null
    this.daemonVersion = createSandboxParams.daemonVersion ?? null
  }

  public setBackupState(
    state: BackupState,
    backupSnapshot?: string | null,
    backupRegistryId?: string | null,
    backupErrorReason?: string | null,
  ) {
    this.backupState = state
    switch (state) {
      case BackupState.NONE:
        this.backupSnapshot = null
        break
      case BackupState.COMPLETED: {
        const now = new Date()
        this.lastBackupAt = now
        if (this.backupSnapshot) {
          this.existingBackupSnapshots = [
            ...(this.existingBackupSnapshots ?? []),
            {
              snapshotName: this.backupSnapshot,
              createdAt: now,
            },
          ]
        }
        this.backupErrorReason = null
        if (this.desiredState === SandboxDesiredState.ARCHIVED) {
          if (this.state === SandboxState.ARCHIVING || this.state === SandboxState.STOPPED) {
            this.state = SandboxState.ARCHIVED
            this.runnerId = null
          }
        }
        break
      }
    }
    if (backupSnapshot !== undefined) {
      this.backupSnapshot = backupSnapshot
    }
    if (backupRegistryId !== undefined) {
      this.backupRegistryId = backupRegistryId
    }
    if (backupErrorReason !== undefined) {
      this.backupErrorReason = backupErrorReason
    }
  }

  @BeforeUpdate()
  updateAccessToken() {
    if (this.state === SandboxState.STARTED) {
      this.authToken = nanoid(32).toLocaleLowerCase()
    }
  }

  @BeforeUpdate()
  updateLastActivityAt() {
    this.lastActivityAt = new Date()
  }

  @BeforeUpdate()
  validateDesiredState() {
    switch (this.desiredState) {
      case SandboxDesiredState.STARTED:
        if (
          [
            SandboxState.STARTED,
            SandboxState.STOPPED,
            SandboxState.STARTING,
            SandboxState.ARCHIVED,
            SandboxState.CREATING,
            SandboxState.UNKNOWN,
            SandboxState.RESTORING,
            SandboxState.PENDING_BUILD,
            SandboxState.BUILDING_SNAPSHOT,
            SandboxState.PULLING_SNAPSHOT,
            SandboxState.ARCHIVING,
            SandboxState.ERROR,
            SandboxState.BUILD_FAILED,
          ].includes(this.state)
        ) {
          break
        }
        throw new Error(`Sandbox ${this.id} is not in a valid state to be started. State: ${this.state}`)
      case SandboxDesiredState.STOPPED:
        if (
          [
            SandboxState.STARTED,
            SandboxState.STOPPING,
            SandboxState.STOPPED,
            SandboxState.ERROR,
            SandboxState.BUILD_FAILED,
          ].includes(this.state)
        ) {
          break
        }
        throw new Error(`Sandbox ${this.id} is not in a valid state to be stopped. State: ${this.state}`)
      case SandboxDesiredState.ARCHIVED:
        if (
          [
            SandboxState.ARCHIVED,
            SandboxState.ARCHIVING,
            SandboxState.STOPPED,
            SandboxState.ERROR,
            SandboxState.BUILD_FAILED,
          ].includes(this.state)
        ) {
          break
        }
        throw new Error(`Sandbox ${this.id} is not in a valid state to be archived. State: ${this.state}`)
      case SandboxDesiredState.DESTROYED:
        if (
          [
            SandboxState.DESTROYED,
            SandboxState.DESTROYING,
            SandboxState.STOPPED,
            SandboxState.STARTED,
            SandboxState.ARCHIVED,
            SandboxState.ERROR,
            SandboxState.BUILD_FAILED,
            SandboxState.ARCHIVING,
          ].includes(this.state)
        ) {
          break
        }
        throw new Error(`Sandbox ${this.id} is not in a valid state to be destroyed. State: ${this.state}`)
    }
  }

  @BeforeUpdate()
  updatePendingFlag() {
    if (String(this.state) === String(this.desiredState)) {
      this.pending = false
    }
    if (this.state === SandboxState.ERROR || this.state === SandboxState.BUILD_FAILED) {
      this.pending = false
    }
  }
}
