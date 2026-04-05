/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export type SnapshotUsageOverviewInternalDto = {
  currentSnapshotUsage: number
}

export type PendingSnapshotUsageOverviewInternalDto = {
  pendingSnapshotUsage: number | null
}

export type SnapshotUsageOverviewWithPendingInternalDto = SnapshotUsageOverviewInternalDto &
  PendingSnapshotUsageOverviewInternalDto

export type TotalSnapshotUsageOverviewInternalDto = {
  currentTotalSnapshotUsage: number
}

export type PendingTotalSnapshotUsageOverviewInternalDto = {
  pendingTotalSnapshotUsage: number | null
}

export type TotalSnapshotUsageOverviewWithPendingInternalDto = TotalSnapshotUsageOverviewInternalDto &
  PendingTotalSnapshotUsageOverviewInternalDto
