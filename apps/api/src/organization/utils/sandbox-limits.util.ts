/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Organization } from '../entities/organization.entity'
import { RegionQuotaDto } from '../dto/region-quota.dto'

/**
 * Get the effective per-sandbox limits.
 * If the region quota limit is set, it overrides the organization limit.
 *
 * @param organization - The organization to get the limits for.
 * @param regionQuota - The region quota to get the limits for.
 * @returns The effective per-sandbox limits.
 */
export function getEffectivePerSandboxLimits(
  organization: Organization,
  regionQuota: RegionQuotaDto | null | undefined,
): {
  maxCpuPerSandbox: number
  maxMemoryPerSandbox: number
  maxDiskPerSandbox: number
  maxDiskPerNonEphemeralSandbox: number | null
} {
  return {
    maxCpuPerSandbox: regionQuota?.maxCpuPerSandbox ?? organization.maxCpuPerSandbox,
    maxMemoryPerSandbox: regionQuota?.maxMemoryPerSandbox ?? organization.maxMemoryPerSandbox,
    maxDiskPerSandbox: regionQuota?.maxDiskPerSandbox ?? organization.maxDiskPerSandbox,
    maxDiskPerNonEphemeralSandbox: regionQuota?.maxDiskPerNonEphemeralSandbox ?? null,
  }
}
