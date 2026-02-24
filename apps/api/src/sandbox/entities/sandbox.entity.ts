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
  PrimaryGeneratedColumn,
  Unique,
  UpdateDateColumn,
} from 'typeorm'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { BackupState } from '../enums/backup-state.enum'
import { v4 as uuidv4 } from 'uuid'
import { SandboxVolume } from '../dto/sandbox.dto'
import { BuildInfo } from './build-info.entity'
import { nanoid } from 'nanoid'

@Entity()
@Unique(['organizationId', 'name'])
@Index('sandbox_state_idx', ['state'])
@Index('sandbox_desiredstate_idx', ['desiredState'])
@Index('sandbox_snapshot_idx', ['snapshot'])
@Index('sandbox_runnerid_idx', ['runnerId'])
@Index('sandbox_runner_state_idx', ['runnerId', 'state'])
@Index('sandbox_organizationid_idx', ['organizationId'])
@Index('sandbox_region_idx', ['region'])
@Index('sandbox_resources_idx', ['cpu', 'mem', 'disk', 'gpu'])
@Index('sandbox_backupstate_idx', ['backupState'])
@Index('sandbox_runner_state_desired_idx', ['runnerId', 'state', 'desiredState'], {
  where: '"pending" = false',
})
@Index('sandbox_active_only_idx', ['id'], {
  where: `"state" <> ALL (ARRAY['destroyed'::sandbox_state_enum, 'archived'::sandbox_state_enum])`,
})
@Index('sandbox_pending_idx', ['id'], {
  where: `"pending" = true`,
})
@Index('idx_sandbox_authtoken', ['authToken'])
@Index('sandbox_labels_gin_full_idx', { synchronize: false })
export class Sandbox {
  @PrimaryGeneratedColumn('uuid')
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
  class = SandboxClass.SMALL

  @Column({
    type: 'enum',
    enum: SandboxState,
    default: SandboxState.UNKNOWN,
  })
  state = SandboxState.UNKNOWN

  @Column({
    type: 'enum',
    enum: SandboxDesiredState,
    default: SandboxDesiredState.STARTED,
  })
  desiredState = SandboxDesiredState.STARTED

  @Column({ nullable: true })
  snapshot?: string

  @Column()
  osUser: string

  @Column({ nullable: true })
  errorReason?: string

  @Column({ default: false, type: 'boolean' })
  recoverable = false

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

  @Column({ nullable: true })
  backupRegistryId: string | null

  @Column({ nullable: true })
  backupSnapshot: string | null

  @Column({ nullable: true, type: 'timestamp with time zone' })
  lastBackupAt: Date | null

  @Column({
    type: 'enum',
    enum: BackupState,
    default: BackupState.NONE,
  })
  backupState = BackupState.NONE

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
  }> = []

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

  @Column({ nullable: true, type: 'timestamp with time zone' })
  lastActivityAt?: Date

  //  this is the interval in minutes after which the sandbox will be stopped if lastActivityAt is not updated
  //  if set to 0, auto stop will be disabled
  @Column({ default: 15, type: 'int' })
  autoStopInterval: number | undefined = 15

  //  this is the interval in minutes after which a continuously stopped workspace will be automatically archived
  @Column({ default: 7 * 24 * 60, type: 'int' })
  autoArchiveInterval: number | undefined = 7 * 24 * 60

  //  this is the interval in minutes after which a continuously stopped workspace will be automatically deleted
  //  if set to negative value, auto delete will be disabled
  //  if set to 0, sandbox will be immediately deleted upon stopping
  @Column({ default: -1, type: 'int' })
  autoDeleteInterval: number | undefined = -1

  @Column({ default: false, type: 'boolean' })
  pending: boolean | undefined = false

  @Column({ default: () => 'MD5(random()::text)', type: 'text' })
  authToken = nanoid(32).toLocaleLowerCase()

  @ManyToOne(() => BuildInfo, (buildInfo) => buildInfo.sandboxes, {
    nullable: true,
  })
  @JoinColumn()
  buildInfo?: BuildInfo

  @Column({ nullable: true })
  daemonVersion?: string

  constructor(region: string, name?: string) {
    this.id = uuidv4()
    // Set name - use provided name or fallback to ID
    this.name = name || this.id
    this.region = region
  }

  /**
   * Helper method that returns the update data needed for a backup state update.
   */
  static getBackupStateUpdate(
    sandbox: Sandbox,
    backupState: BackupState,
    backupSnapshot?: string | null,
    backupRegistryId?: string | null,
    backupErrorReason?: string | null,
  ): Partial<Sandbox> {
    const update: Partial<Sandbox> = {
      backupState,
    }
    switch (backupState) {
      case BackupState.NONE:
        update.backupSnapshot = null
        break
      case BackupState.COMPLETED: {
        const now = new Date()
        update.lastBackupAt = now
        update.existingBackupSnapshots = [
          ...sandbox.existingBackupSnapshots,
          {
            snapshotName: sandbox.backupSnapshot,
            createdAt: now,
          },
        ]
        update.backupErrorReason = null
        break
      }
    }
    if (backupSnapshot !== undefined) {
      update.backupSnapshot = backupSnapshot
    }
    if (backupRegistryId !== undefined) {
      update.backupRegistryId = backupRegistryId
    }
    if (backupErrorReason !== undefined) {
      update.backupErrorReason = backupErrorReason
    }
    return update
  }

  /**
   * Helper method that returns the update data needed for a soft delete operation.
   */
  static getSoftDeleteUpdate(sandbox: Sandbox): Partial<Sandbox> {
    return {
      pending: true,
      desiredState: SandboxDesiredState.DESTROYED,
      backupState: BackupState.NONE,
      name: 'DESTROYED_' + sandbox.name + '_' + Date.now(),
    }
  }

  /**
   * Asserts that the current entity state is valid.
   */
  assertValid(): void {
    this.validateDesiredStateTransition()
  }

  private validateDesiredStateTransition(): void {
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
            SandboxState.RESIZING,
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
            SandboxState.RESIZING,
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
            SandboxState.PENDING_BUILD,
          ].includes(this.state)
        ) {
          break
        }
        throw new Error(`Sandbox ${this.id} is not in a valid state to be destroyed. State: ${this.state}`)
    }
  }

  /**
   * Enforces domain invariants on the current entity state.
   *
   * @returns Additional field changes that invariant enforcement produced.
   */
  enforceInvariants(): Partial<Sandbox> {
    const changes = this.getInvariantChanges()
    Object.assign(this, changes)
    return changes
  }

  private getInvariantChanges(): Partial<Sandbox> {
    const changes: Partial<Sandbox> = {}

    if (!this.pending && String(this.state) !== String(this.desiredState)) {
      changes.pending = true
    }
    if (this.pending && String(this.state) === String(this.desiredState)) {
      changes.pending = false
    }
    if (
      this.state === SandboxState.ERROR ||
      this.state === SandboxState.BUILD_FAILED ||
      this.desiredState === SandboxDesiredState.ARCHIVED
    ) {
      changes.pending = false
    }

    if (this.state === SandboxState.DESTROYED) {
      changes.runnerId = null
      changes.backupState = BackupState.NONE
    }

    return changes
  }
}
