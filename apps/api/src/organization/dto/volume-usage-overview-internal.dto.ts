/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export type VolumeUsageOverviewInternalDto = {
  currentVolumeUsage: number
}

export type PendingVolumeUsageOverviewInternalDto = {
  pendingVolumeUsage: number | null
}

export type VolumeUsageOverviewWithPendingInternalDto = VolumeUsageOverviewInternalDto &
  PendingVolumeUsageOverviewInternalDto
