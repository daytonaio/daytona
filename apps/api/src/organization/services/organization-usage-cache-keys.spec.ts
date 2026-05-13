/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  getCacheStalenessKey,
  getCurrentQuotaUsageCacheKey,
  getPendingQuotaUsageCacheKey,
} from './organization-usage-cache-keys'

describe('organization-usage-cache-keys', () => {
  const organizationId = 'org-1'
  const regionId = 'us-east-1'

  const extractHashTag = (key: string): string => {
    const match = key.match(/^\{[^}]+\}/)

    if (!match) {
      throw new Error(`Expected hash tag in key: ${key}`)
    }

    return match[0]
  }

  describe('getCurrentQuotaUsageCacheKey', () => {
    it('produces hash-tagged current quota keys with regionId for all regional quota types', () => {
      expect(getCurrentQuotaUsageCacheKey(organizationId, 'cpu', regionId)).toBe(
        '{org:org-1:region:us-east-1}:quota:cpu:usage',
      )
      expect(getCurrentQuotaUsageCacheKey(organizationId, 'memory', regionId)).toBe(
        '{org:org-1:region:us-east-1}:quota:memory:usage',
      )
      expect(getCurrentQuotaUsageCacheKey(organizationId, 'disk', regionId)).toBe(
        '{org:org-1:region:us-east-1}:quota:disk:usage',
      )
    })

    it('produces hash-tagged current quota keys without regionId for global quota types', () => {
      expect(getCurrentQuotaUsageCacheKey(organizationId, 'snapshot_count')).toBe(
        '{org:org-1}:quota:snapshot_count:usage',
      )
      expect(getCurrentQuotaUsageCacheKey(organizationId, 'volume_count')).toBe('{org:org-1}:quota:volume_count:usage')
    })
  })

  describe('getPendingQuotaUsageCacheKey', () => {
    it('produces hash-tagged pending quota keys with regionId for all regional quota types', () => {
      expect(getPendingQuotaUsageCacheKey(organizationId, 'cpu', regionId)).toBe(
        '{org:org-1:region:us-east-1}:quota:cpu:pending',
      )
      expect(getPendingQuotaUsageCacheKey(organizationId, 'memory', regionId)).toBe(
        '{org:org-1:region:us-east-1}:quota:memory:pending',
      )
      expect(getPendingQuotaUsageCacheKey(organizationId, 'disk', regionId)).toBe(
        '{org:org-1:region:us-east-1}:quota:disk:pending',
      )
    })

    it('produces hash-tagged pending quota keys without regionId for global quota types', () => {
      expect(getPendingQuotaUsageCacheKey(organizationId, 'snapshot_count')).toBe(
        '{org:org-1}:quota:snapshot_count:pending',
      )
      expect(getPendingQuotaUsageCacheKey(organizationId, 'volume_count')).toBe(
        '{org:org-1}:quota:volume_count:pending',
      )
    })
  })

  describe('hash tag behavior', () => {
    it('uses the identical hash tag prefix for all current and pending regional quota keys in the same org and region', () => {
      const keys = [
        getCurrentQuotaUsageCacheKey(organizationId, 'cpu', regionId),
        getCurrentQuotaUsageCacheKey(organizationId, 'memory', regionId),
        getCurrentQuotaUsageCacheKey(organizationId, 'disk', regionId),
        getPendingQuotaUsageCacheKey(organizationId, 'cpu', regionId),
        getPendingQuotaUsageCacheKey(organizationId, 'memory', regionId),
        getPendingQuotaUsageCacheKey(organizationId, 'disk', regionId),
      ]

      const hashTags = keys.map(extractHashTag)

      expect(new Set(hashTags)).toEqual(new Set(['{org:org-1:region:us-east-1}']))
    })

    it('produces different hash tags for different organizations', () => {
      const keyA = getCurrentQuotaUsageCacheKey('org-1', 'cpu', regionId)
      const keyB = getCurrentQuotaUsageCacheKey('org-2', 'cpu', regionId)

      expect(extractHashTag(keyA)).not.toBe(extractHashTag(keyB))
    })

    it('produces different hash tags for different regions', () => {
      const keyA = getCurrentQuotaUsageCacheKey(organizationId, 'cpu', 'us-east-1')
      const keyB = getCurrentQuotaUsageCacheKey(organizationId, 'cpu', 'eu-west-1')

      expect(extractHashTag(keyA)).not.toBe(extractHashTag(keyB))
    })
  })

  describe('getCacheStalenessKey', () => {
    it('produces correct staleness keys with regionId for sandbox resources', () => {
      expect(getCacheStalenessKey(organizationId, 'sandbox', regionId)).toBe(
        '{org:org-1:region:us-east-1}:resource:sandbox:usage:fetched_at',
      )
    })

    it('produces correct staleness keys without regionId for snapshot and volume resources', () => {
      expect(getCacheStalenessKey(organizationId, 'snapshot')).toBe('{org:org-1}:resource:snapshot:usage:fetched_at')
      expect(getCacheStalenessKey(organizationId, 'volume')).toBe('{org:org-1}:resource:volume:usage:fetched_at')
    })
  })
})
