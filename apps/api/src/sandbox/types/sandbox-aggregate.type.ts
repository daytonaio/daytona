/* Copyright 2025 Daytona Platforms Inc. SPDX-License-Identifier: AGPL-3.0 */

import { Sandbox } from '../entities/sandbox.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { BackupState } from '../enums/backup-state.enum'
import { SandboxStateEntity } from '../entities/sandbox-state.entity'
import { SandboxBackupEntity } from '../entities/sandbox-backup.entity'

export interface SandboxStateFields {
  state: SandboxState
  desiredState: SandboxDesiredState
  pending: boolean
  errorReason?: string
  recoverable: boolean
  runnerId?: string
  prevRunnerId?: string
  daemonVersion?: string
}

export interface SandboxBackupFields {
  backupState: BackupState
  backupSnapshot?: string | null
  backupRegistryId?: string | null
  lastBackupAt?: Date | null
  backupErrorReason?: string | null
  existingBackupSnapshots: Array<{ snapshotName: string; createdAt: Date }>
}

/**
 * Assembled domain aggregate combining data from all 3 tables.
 * Preserves the same shape as the old Sandbox entity for API compatibility.
 * NOT a TypeORM entity — assembled by the repository facade.
 */
export type SandboxAggregate = Sandbox & SandboxStateFields & SandboxBackupFields
