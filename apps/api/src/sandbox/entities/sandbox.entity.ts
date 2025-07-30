/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  BeforeUpdate,
  Column,
  CreateDateColumn,
  Entity,
  Generated,
  JoinColumn,
  ManyToOne,
  PrimaryColumn,
  UpdateDateColumn,
} from 'typeorm'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { BackupState } from '../enums/backup-state.enum'
import { nanoid } from 'nanoid'
import { SandboxVolume } from '../dto/sandbox.dto'
import { BuildInfo } from './build-info.entity'

@Entity()
export class Sandbox {
  @PrimaryColumn()
  @Generated('uuid')
  id: string

  @Column({
    type: 'uuid',
  })
  organizationId: string

  @Column({
    default: 'us',
  })
  region: string

  @Column({
    type: 'uuid',
    nullable: true,
  })
  runnerId?: string

  //  this is the runnerId of the runner that was previously assigned to the sandbox
  //  if something goes wrong with new runner assignment, we can revert to the previous runner
  @Column({
    type: 'uuid',
    nullable: true,
  })
  prevRunnerId?: string

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

  @Column({ nullable: true })
  snapshot?: string

  @Column()
  osUser: string

  @Column({ nullable: true })
  errorReason?: string

  @Column({
    type: 'jsonb',
    default: {},
  })
  env: { [key: string]: string }

  @Column({ default: false })
  public: boolean

  @Column('jsonb', { nullable: true })
  labels: { [key: string]: string }

  @Column({ nullable: true })
  backupRegistryId: string

  @Column({ nullable: true })
  backupSnapshot: string

  @Column({ nullable: true, type: 'timestamp with time zone' })
  lastBackupAt: Date

  @Column({
    type: 'enum',
    enum: BackupState,
    default: BackupState.NONE,
  })
  backupState: BackupState

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
  lastActivityAt?: Date

  //  this is the interval in minutes after which the sandbox will be stopped if lastActivityAt is not updated
  //  if set to 0, auto stop will be disabled
  @Column({ default: 15 })
  autoStopInterval?: number

  //  this is the interval in minutes after which a continuously stopped workspace will be automatically archived
  @Column({ default: 7 * 24 * 60 })
  autoArchiveInterval?: number

  //  this is the interval in minutes after which a continuously stopped workspace will be automatically deleted
  //  if set to negative value, auto delete will be disabled
  //  if set to 0, sandbox will be immediately deleted upon stopping
  @Column({ default: -1 })
  autoDeleteInterval?: number

  @Column({ default: false })
  pending?: boolean

  @Column({ default: () => 'MD5(random()::text)' })
  authToken: string

  @ManyToOne(() => BuildInfo, (buildInfo) => buildInfo.sandboxes, {
    nullable: true,
    eager: true,
  })
  @JoinColumn()
  buildInfo?: BuildInfo

  @Column({ nullable: true })
  daemonVersion?: string

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
