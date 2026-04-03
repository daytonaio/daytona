/* Copyright 2025 Daytona Platforms Inc. SPDX-License-Identifier: AGPL-3.0 */

import { Entity, PrimaryColumn, Column, UpdateDateColumn, OneToOne, JoinColumn, Index } from 'typeorm'
import { Sandbox } from './sandbox.entity'
import { BackupState } from '../enums/backup-state.enum'

@Entity('sandbox_backup')
@Index('sb_backupstate_idx', ['backupState'])
export class SandboxBackupEntity {
  @PrimaryColumn()
  sandboxId: string

  @Column({ type: 'enum', enum: BackupState, default: BackupState.NONE })
  backupState: BackupState

  @Column({ nullable: true })
  backupSnapshot: string | null

  @Column({ nullable: true })
  backupRegistryId: string | null

  @Column({ nullable: true, type: 'timestamp with time zone' })
  lastBackupAt: Date | null

  @Column({ type: 'text', nullable: true })
  backupErrorReason: string | null

  @Column({ type: 'jsonb', default: [] })
  existingBackupSnapshots: Array<{ snapshotName: string; createdAt: Date }>

  @UpdateDateColumn({ type: 'timestamp with time zone' })
  updatedAt: Date

  @OneToOne(() => Sandbox, { onDelete: 'CASCADE' })
  @JoinColumn({ name: 'sandboxId' })
  sandbox?: Sandbox

  enforceInvariants(): Partial<SandboxBackupEntity> {
    return {}
  }

  static getBackupStateUpdate(
    current: Pick<SandboxBackupEntity, 'backupSnapshot' | 'existingBackupSnapshots'>,
    backupState: BackupState,
    backupSnapshot?: string | null,
    backupRegistryId?: string | null,
    backupErrorReason?: string | null,
  ): Partial<SandboxBackupEntity> {
    const update: Partial<SandboxBackupEntity> = { backupState }

    switch (backupState) {
      case BackupState.NONE:
        update.backupSnapshot = null
        break
      case BackupState.COMPLETED: {
        const now = new Date()
        update.lastBackupAt = now
        if (current.backupSnapshot) {
          update.existingBackupSnapshots = [
            ...current.existingBackupSnapshots,
            { snapshotName: current.backupSnapshot, createdAt: now },
          ]
        }
        update.backupErrorReason = null
        break
      }
    }

    if (backupSnapshot !== undefined) update.backupSnapshot = backupSnapshot
    if (backupRegistryId !== undefined) update.backupRegistryId = backupRegistryId
    if (backupErrorReason !== undefined) update.backupErrorReason = backupErrorReason

    return update
  }
}
