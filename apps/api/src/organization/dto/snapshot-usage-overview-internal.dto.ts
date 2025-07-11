/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { z } from 'zod'

export const SnapshotUsageOverviewSchema = z.object({
  totalSnapshotQuota: z.number(),
  currentSnapshotUsage: z.number(),
  _fetchedAt: z.date(),
})

export type SnapshotUsageOverviewInternalDto = {
  totalSnapshotQuota: number
  currentSnapshotUsage: number
  _fetchedAt: Date
}

//export type SnapshotUsageOverviewInternalDto = z.infer<typeof SnapshotUsageOverviewSchema>
