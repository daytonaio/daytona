/* Copyright 2025 Daytona Platforms Inc. SPDX-License-Identifier: AGPL-3.0 */

import { Entity, PrimaryColumn, Column, UpdateDateColumn, OneToOne, JoinColumn, Index } from 'typeorm'
import { Sandbox } from './sandbox.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'

@Entity('sandbox_state')
@Index('ss_state_idx', ['state'])
@Index('ss_desiredstate_idx', ['desiredState'])
@Index('ss_runnerid_idx', ['runnerId'])
@Index('ss_runner_state_idx', ['runnerId', 'state'])
@Index('ss_runner_state_desired_idx', ['runnerId', 'state', 'desiredState'], {
  where: '"pending" = false',
})
@Index('ss_active_only_idx', ['sandboxId'], {
  where: "\"state\" <> ALL (ARRAY['destroyed'::sandbox_state_enum, 'archived'::sandbox_state_enum])",
})
@Index('ss_pending_idx', ['sandboxId'], {
  where: '"pending" = true',
})
export class SandboxStateEntity {
  @PrimaryColumn()
  sandboxId: string

  @Column({ type: 'enum', enum: SandboxState, default: SandboxState.UNKNOWN })
  state: SandboxState

  @Column({ type: 'enum', enum: SandboxDesiredState, default: SandboxDesiredState.STARTED })
  desiredState: SandboxDesiredState

  @Column({ default: false, type: 'boolean' })
  pending: boolean

  @Column({ nullable: true })
  errorReason?: string

  @Column({ default: false, type: 'boolean' })
  recoverable: boolean

  @Column({ type: 'uuid', nullable: true })
  runnerId?: string

  @Column({ type: 'uuid', nullable: true })
  prevRunnerId?: string

  @Column({ nullable: true })
  daemonVersion?: string

  @UpdateDateColumn({ type: 'timestamp with time zone' })
  updatedAt: Date

  @OneToOne(() => Sandbox, { onDelete: 'CASCADE' })
  @JoinColumn({ name: 'sandboxId' })
  sandbox?: Sandbox

  /**
   * Per-table invariant enforcement. Only handles fields within sandbox_state.
   * Cross-table invariants (e.g., DESTROYED → backup) are handled by the repository.
   */
  assertValid(sandboxId: string): void {
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
        throw new Error(`Sandbox ${sandboxId} is not in a valid state to be started. State: ${this.state}`)
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
        throw new Error(`Sandbox ${sandboxId} is not in a valid state to be stopped. State: ${this.state}`)
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
        throw new Error(`Sandbox ${sandboxId} is not in a valid state to be archived. State: ${this.state}`)
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
        throw new Error(`Sandbox ${sandboxId} is not in a valid state to be destroyed. State: ${this.state}`)
    }
  }

  enforceInvariants(): Partial<SandboxStateEntity> {
    const changes: Partial<SandboxStateEntity> = {}

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
    if (this.state === SandboxState.DESTROYED || this.state === SandboxState.ARCHIVED) {
      changes.runnerId = null
    }

    return changes
  }
}
