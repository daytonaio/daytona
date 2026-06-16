/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, Entity, JoinColumn, OneToOne, PrimaryColumn, UpdateDateColumn } from 'typeorm'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { BackupState } from '../enums/backup-state.enum'
import { SandboxLifecyclePhase } from '../enums/sandbox-lifecycle-phase.enum'
import { Sandbox } from './sandbox.entity'

/**
 * Hot, write-heavy state of a sandbox.
 *
 * The underlying table is LIST-partitioned by {@link SandboxLifecyclePhase}.
 *
 * Composite PK `(sandboxId, lifecyclePhase)` is a Postgres requirement for partitioned tables (partition key must be part of the PK).
 * Logical uniqueness is on `sandboxId` alone — the application maintains the invariant that a sandbox has exactly one lifecycle row at a time.
 *
 * `organizationId` is denormalized from {@link Sandbox} so the optimistic-concurrency predicate stays single-table.
 */
@Entity('sandbox_lifecycle')
export class SandboxLifecycle {
  @PrimaryColumn()
  sandboxId: string

  @PrimaryColumn({
    type: 'text',
    enum: SandboxLifecyclePhase,
  })
  lifecyclePhase: SandboxLifecyclePhase

  @Column({ type: 'uuid' })
  organizationId: string

  @Column({
    type: 'enum',
    enum: SandboxState,
    default: SandboxState.UNKNOWN,
  })
  state: SandboxState = SandboxState.UNKNOWN

  @Column({
    type: 'enum',
    enum: SandboxDesiredState,
    default: SandboxDesiredState.STARTED,
  })
  desiredState: SandboxDesiredState = SandboxDesiredState.STARTED

  @Column({ default: false, type: 'boolean' })
  pending: boolean | undefined = false

  @Column({ nullable: true })
  errorReason?: string

  @Column({ default: false, type: 'boolean' })
  recoverable = false

  @Column({ nullable: true })
  daemonVersion?: string

  @Column({ type: 'uuid', nullable: true })
  runnerId?: string

  @Column({ type: 'uuid', nullable: true })
  prevRunnerId?: string

  @Column({
    type: 'enum',
    enum: BackupState,
    default: BackupState.NONE,
  })
  backupState: BackupState = BackupState.NONE

  @Column({ nullable: true, type: 'timestamp with time zone' })
  lastBackupAt: Date | null

  @Column({ nullable: true })
  backupSnapshot: string | null

  @Column({ nullable: true })
  backupRegistryId: string | null

  @Column({ type: 'text', nullable: true })
  backupErrorReason: string | null

  @Column({
    type: 'jsonb',
    default: [],
  })
  existingBackupSnapshots: Array<{
    snapshotName: string
    createdAt: Date
  }> = []

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date

  @OneToOne(() => Sandbox, { onDelete: 'CASCADE' })
  @JoinColumn({ name: 'sandboxId' })
  sandbox?: Sandbox

  /**
   * Derive the {@link SandboxLifecyclePhase} for a given {@link SandboxState}.
   * Terminal states (DESTROYED, ARCHIVED) map to TERMINAL; everything else maps to ACTIVE.
   * Used by `Sandbox.enforceInvariants()` to keep the partition key in sync with the state machine.
   */
  static phaseFor(state: SandboxState): SandboxLifecyclePhase {
    return state === SandboxState.DESTROYED || state === SandboxState.ARCHIVED
      ? SandboxLifecyclePhase.TERMINAL
      : SandboxLifecyclePhase.ACTIVE
  }

  /**
   * Constructs the lifecycle row that the AFTER-INSERT trigger would have written
   * when active. Mirrors every state-machine column from the sandbox entity verbatim.
   */
  static fromSandbox(sandbox: Sandbox): Partial<SandboxLifecycle> {
    return {
      sandboxId: sandbox.id,
      lifecyclePhase: SandboxLifecycle.phaseFor(sandbox.state),
      organizationId: sandbox.organizationId,
      state: sandbox.state,
      desiredState: sandbox.desiredState,
      pending: sandbox.pending,
      errorReason: sandbox.errorReason,
      recoverable: sandbox.recoverable,
      daemonVersion: sandbox.daemonVersion,
      runnerId: sandbox.runnerId,
      prevRunnerId: sandbox.prevRunnerId,
      backupState: sandbox.backupState,
      lastBackupAt: sandbox.lastBackupAt,
      backupSnapshot: sandbox.backupSnapshot,
      backupRegistryId: sandbox.backupRegistryId,
      backupErrorReason: sandbox.backupErrorReason,
      existingBackupSnapshots: sandbox.existingBackupSnapshots,
      updatedAt: sandbox.updatedAt,
    }
  }
}
