/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  BeforeInsert,
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
import { NodeRegion } from '../enums/node-region.enum'
import { SnapshotState } from '../enums/snapshot-state.enum'
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
    enum: NodeRegion,
    default: NodeRegion.EU,
  })
  region: NodeRegion

  @Column({
    type: 'uuid',
    nullable: true,
  })
  nodeId?: string

  //  this is the nodeId of the node that was previously assigned to the workspace
  //  if something goes wrong with new node assignment, we can revert to the previous node
  @Column({
    type: 'uuid',
    nullable: true,
  })
  prevNodeId?: string

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
  image?: string

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
  snapshotRegistryId: string

  @Column({ nullable: true })
  snapshotImage: string

  @Column({ nullable: true })
  lastSnapshotAt: Date

  @Column({
    type: 'enum',
    enum: SnapshotState,
    default: SnapshotState.NONE,
  })
  snapshotState: SnapshotState

  @Column({
    type: 'jsonb',
    default: [],
  })
  existingSnapshotImages: Array<{
    imageName: string
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
            WorkspaceState.BUILDING_IMAGE,
            WorkspaceState.PULLING_IMAGE,
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
