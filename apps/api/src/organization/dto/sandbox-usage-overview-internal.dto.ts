/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { z } from 'zod'

export const SandboxUsageOverviewSchema = z.object({
  totalCpuQuota: z.number(),
  totalMemoryQuota: z.number(),
  totalDiskQuota: z.number(),
  currentCpuUsage: z.number(),
  currentMemoryUsage: z.number(),
  currentDiskUsage: z.number(),
  _fetchedAt: z.date(),
})

export type SandboxUsageOverviewInternalDto = {
  totalCpuQuota: number
  totalMemoryQuota: number
  totalDiskQuota: number
  currentCpuUsage: number
  currentMemoryUsage: number
  currentDiskUsage: number
  _fetchedAt: Date
}

// export type SandboxUsageOverviewInternalDto = z.infer<typeof SandboxUsageOverviewSchema>
