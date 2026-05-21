/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Organization } from '../entities/organization.entity'
import { RegionQuotaDto } from '../dto/region-quota.dto'

/**
 * Get the effective per-sandbox limits.
 *
 * @param organization - The organization to get the limits for.
 * @param regionQuota - The region quota to get the limits for.
 * @param gpuEnabled - Whether the sandbox uses GPU; selects GPU-aware overrides when true.
 * @returns The effective per-sandbox limits.
 */
export function getEffectivePerSandboxLimits(
  organization: Organization,
  regionQuota: RegionQuotaDto | null | undefined,
  gpuEnabled: boolean,
): {
  maxCpuPerSandbox: number
  maxMemoryPerSandbox: number
  maxDiskPerSandbox: number
  maxDiskPerNonEphemeralSandbox: number | null
} {
  if (gpuEnabled) {
    return {
      maxCpuPerSandbox:
        regionQuota?.maxCpuPerGpuSandbox ?? regionQuota?.maxCpuPerSandbox ?? organization.maxCpuPerSandbox,
      maxMemoryPerSandbox:
        regionQuota?.maxMemoryPerGpuSandbox ?? regionQuota?.maxMemoryPerSandbox ?? organization.maxMemoryPerSandbox,
      maxDiskPerSandbox:
        regionQuota?.maxDiskPerGpuSandbox ?? regionQuota?.maxDiskPerSandbox ?? organization.maxDiskPerSandbox,
      maxDiskPerNonEphemeralSandbox: null,
    }
  }

  return {
    maxCpuPerSandbox: regionQuota?.maxCpuPerSandbox ?? organization.maxCpuPerSandbox,
    maxMemoryPerSandbox: regionQuota?.maxMemoryPerSandbox ?? organization.maxMemoryPerSandbox,
    maxDiskPerSandbox: regionQuota?.maxDiskPerSandbox ?? organization.maxDiskPerSandbox,
    maxDiskPerNonEphemeralSandbox: regionQuota?.maxDiskPerNonEphemeralSandbox ?? null,
  }
}
