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
import { WorkspaceState } from '../enums/workspace-state.enum'
import { WorkspaceDesiredState } from '../enums/workspace-desired-state.enum'
import { WorkspaceClass } from '../enums/workspace-class.enum'
import { BackupState } from '../enums/backup-state.enum'
import { RunnerRegion } from '../enums/runner-region.enum'
import { nanoid } from 'nanoid'
import { WorkspaceVolume } from '../dto/workspace.dto'
import { BuildInfo } from './build-info.entity'

@Entity()
export class Workspace {
  @PrimaryColumn()
  @Generated('uuid')
  id: string

  @Column({
    type: 'uuid',
  })
  organizationId: string

  @Column({
    type: 'enum',
    enum: RunnerRegion,
    default: RunnerRegion.EU,
  })
  region: RunnerRegion

  @Column({
    type: 'uuid',
    nullable: true,
  })
  runnerId?: string

  //  this is the runnerId of the runner that was previously assigned to the workspace
  //  if something goes wrong with new runner assignment, we can revert to the previous runner
  @Column({
    type: 'uuid',
    nullable: true,
  })
  prevRunnerId?: string

  @Column({
    type: 'enum',
    enum: WorkspaceClass,
    default: WorkspaceClass.SMALL,
  })
  class: WorkspaceClass

  @Column({
    type: 'enum',
    enum: WorkspaceState,
    default: WorkspaceState.UNKNOWN,
  })
  state: WorkspaceState

  @Column({
    type: 'enum',
    enum: WorkspaceDesiredState,
    default: WorkspaceDesiredState.STARTED,
  })
  desiredState: WorkspaceDesiredState

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

  @Column({ nullable: true })
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
  volumes: WorkspaceVolume[]

  @CreateDateColumn()
  createdAt: Date

  @UpdateDateColumn()
  updatedAt: Date

  @Column({ nullable: true, type: 'timestamp' })
  lastActivityAt?: Date

  //  this is the interval in minutes after which the workspace will be stopped if lastActivityAt is not updated
  //  if set to 0, auto stop will be disabled
  @Column({ default: 15 })
  autoStopInterval?: number

  //  this is the interval in minutes after which a continuously stopped workspace will be automatically archived
  @Column({ default: 7 * 24 * 60 })
  autoArchiveInterval?: number

  @Column({ default: false })
  pending?: boolean

  @Column({ default: () => 'MD5(random()::text)' })
  authToken: string

  @ManyToOne(() => BuildInfo, (buildInfo) => buildInfo.workspaces, {
    nullable: true,
    eager: true,
  })
  @JoinColumn()
  buildInfo?: BuildInfo

  @BeforeUpdate()
  updateAccessToken() {
    if (this.state === WorkspaceState.STARTED) {
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
      case WorkspaceDesiredState.STARTED:
        if (
          [
            WorkspaceState.STARTED,
            WorkspaceState.STOPPED,
            WorkspaceState.STARTING,
            WorkspaceState.ARCHIVED,
            WorkspaceState.CREATING,
            WorkspaceState.UNKNOWN,
            WorkspaceState.RESTORING,
            WorkspaceState.PENDING_BUILD,
            WorkspaceState.BUILDING_SNAPSHOT,
            WorkspaceState.PULLING_SNAPSHOT,
            WorkspaceState.ERROR,
          ].includes(this.state)
        ) {
          break
        }
        throw new Error(`Workspace ${this.id} is not in a valid state to be started. State: ${this.state}`)
      case WorkspaceDesiredState.STOPPED:
        if (
          [WorkspaceState.STARTED, WorkspaceState.STOPPING, WorkspaceState.STOPPED, WorkspaceState.ERROR].includes(
            this.state,
          )
        ) {
          break
        }
        throw new Error(`Workspace ${this.id} is not in a valid state to be stopped. State: ${this.state}`)
      case WorkspaceDesiredState.ARCHIVED:
        if (
          [WorkspaceState.ARCHIVED, WorkspaceState.ARCHIVING, WorkspaceState.STOPPED, WorkspaceState.ERROR].includes(
            this.state,
          )
        ) {
          break
        }
        throw new Error(`Workspace ${this.id} is not in a valid state to be archived. State: ${this.state}`)
      case WorkspaceDesiredState.DESTROYED:
        if (
          [
            WorkspaceState.DESTROYED,
            WorkspaceState.DESTROYING,
            WorkspaceState.STOPPED,
            WorkspaceState.STARTED,
            WorkspaceState.ARCHIVED,
            WorkspaceState.ERROR,
          ].includes(this.state)
        ) {
          break
        }
        throw new Error(`Workspace ${this.id} is not in a valid state to be destroyed. State: ${this.state}`)
    }
  }

  @BeforeUpdate()
  updatePendingFlag() {
    if (String(this.state) === String(this.desiredState)) {
      this.pending = false
    }
    if (this.state === WorkspaceState.ERROR) {
      this.pending = false
    }
  }
}
