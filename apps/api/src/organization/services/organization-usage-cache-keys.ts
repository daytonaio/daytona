/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationUsageQuotaType, OrganizationUsageResourceType } from '../helpers/organization-usage.helper'

/**
 * Build the hash-tag prefix that ensures all keys for the same org+region
 * land on the same Redis Cluster slot.
 *
 * In single-node Redis this is harmless -- the braces are just part of the key string.
 */
function hashTag(organizationId: string, regionId?: string): string {
  return regionId ? `{org:${organizationId}:region:${regionId}}` : `{org:${organizationId}}`
}

export function getCurrentQuotaUsageCacheKey(
  organizationId: string,
  quotaType: 'cpu' | 'memory' | 'disk',
  regionId: string,
): string
export function getCurrentQuotaUsageCacheKey(
  organizationId: string,
  quotaType: 'snapshot_count' | 'volume_count',
): string
export function getCurrentQuotaUsageCacheKey(
  organizationId: string,
  quotaType: OrganizationUsageQuotaType,
  regionId?: string,
): string {
  return `${hashTag(organizationId, regionId)}:quota:${quotaType}:usage`
}

export function getPendingQuotaUsageCacheKey(
  organizationId: string,
  quotaType: 'cpu' | 'memory' | 'disk',
  regionId: string,
): string
export function getPendingQuotaUsageCacheKey(
  organizationId: string,
  quotaType: 'snapshot_count' | 'volume_count',
): string
export function getPendingQuotaUsageCacheKey(
  organizationId: string,
  quotaType: OrganizationUsageQuotaType,
  regionId?: string,
): string {
  return `${hashTag(organizationId, regionId)}:quota:${quotaType}:pending`
}

export function getCacheStalenessKey(
  organizationId: string,
  resourceType: OrganizationUsageResourceType,
  regionId?: string,
): string {
  return `${hashTag(organizationId, regionId)}:resource:${resourceType}:usage:fetched_at`
}
