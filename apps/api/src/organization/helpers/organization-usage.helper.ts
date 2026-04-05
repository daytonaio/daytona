/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export type OrganizationUsageQuotaType =
  | 'cpu'
  | 'memory'
  | 'disk'
  | 'snapshot_count'
  | 'total_snapshot_count'
  | 'volume_count'
export type OrganizationUsageResourceType = 'sandbox' | 'snapshot' | 'total_snapshot' | 'volume'

const QUOTA_TO_RESOURCE_MAP: Record<OrganizationUsageQuotaType, OrganizationUsageResourceType> = {
  cpu: 'sandbox',
  memory: 'sandbox',
  disk: 'sandbox',
  snapshot_count: 'snapshot',
  total_snapshot_count: 'total_snapshot',
  volume_count: 'volume',
} as const

export function getResourceTypeFromQuota(quotaType: OrganizationUsageQuotaType): OrganizationUsageResourceType {
  return QUOTA_TO_RESOURCE_MAP[quotaType]
}
