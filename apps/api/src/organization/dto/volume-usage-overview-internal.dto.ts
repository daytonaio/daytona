/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */
import { z } from 'zod'

export const VolumeUsageOverviewSchema = z.object({
  totalVolumeQuota: z.number(),
  currentVolumeUsage: z.number(),
  _fetchedAt: z.date(),
})

export type VolumeUsageOverviewInternalDto = {
  totalVolumeQuota: number
  currentVolumeUsage: number
  _fetchedAt: Date
}

// export type VolumeUsageOverviewInternalDto = z.infer<typeof VolumeUsageOverviewSchema>
