/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Sandbox, SandboxState, SnapshotDto } from '@daytonaio/api-client'
import { Table } from '@tanstack/react-table'

export interface SandboxTableProps {
  data: Sandbox[]
  loadingSandboxes: Record<string, boolean>
  transitioningSandboxes: Record<string, boolean>
  loading: boolean
  snapshots: SnapshotDto[]
  loadingSnapshots: boolean
  handleStart: (id: string) => void
  handleStop: (id: string) => void
  handleDelete: (id: string) => void
  handleBulkDelete: (ids: string[]) => void
  handleArchive: (id: string) => void
  handleVnc: (id: string) => void
  getWebTerminalUrl: (id: string) => Promise<string | null>
  onRowClick?: (sandbox: Sandbox) => void
}

export interface SandboxTableActionsProps {
  sandbox: Sandbox
  writePermitted: boolean
  deletePermitted: boolean
  isLoading: boolean
  onStart: (id: string) => void
  onStop: (id: string) => void
  onDelete: (id: string) => void
  onArchive: (id: string) => void
  onVnc: (id: string) => void
  onOpenWebTerminal: (id: string) => void
}

export interface SandboxTableHeaderProps {
  table: Table<Sandbox>
  labelOptions: FacetedFilterOption[]
  regionOptions: FacetedFilterOption[]
  snapshots: SnapshotDto[]
  loadingSnapshots: boolean
}

export interface FacetedFilterOption {
  label: string
  value: string | SandboxState
  icon?: any
}
