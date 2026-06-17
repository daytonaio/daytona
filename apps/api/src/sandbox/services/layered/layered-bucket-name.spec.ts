/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { awsRegionFromStorageRegion, layeredBucketNameFor } from './layered-bucket-name'

describe('layeredBucketNameFor', () => {
  const orgId = '123e4567-e89b-12d3-a456-426614174000'

  it('returns the legacy single-bucket name when storageRegion is null', () => {
    expect(layeredBucketNameFor(orgId, null)).toBe(`daytona-volume-layered-${orgId}`)
  })

  it('returns the per-region name when storageRegion is set', () => {
    expect(layeredBucketNameFor(orgId, 'aws-us-east-1')).toBe(`dt-vl-${orgId}-aws-us-east-1`)
  })

  it('throws when orgId is empty', () => {
    expect(() => layeredBucketNameFor('', 'aws-us-east-1')).toThrow(/orgId is required/)
  })

  it('produces a name within the AWS 63-char bucket-name limit for the worst-case slug', () => {
    // Worst case: `dt-vl-` (6) + UUID (36) + `-` (1) + longest `aws-` region (18) = 61.
    const worstCase = layeredBucketNameFor(orgId, 'aws-ap-southeast-2')
    expect(worstCase.length).toBeLessThanOrEqual(63)
    expect(worstCase.length).toBe(61)
  })

  it('throws when the resulting bucket name exceeds 63 characters', () => {
    // Oversized region slug to trip the length assertion (real AWS slugs never reach this).
    const overlong = 'aws-' + 'x'.repeat(40)
    expect(() => layeredBucketNameFor(orgId, overlong)).toThrow(/length \d+ is outside/)
  })

  it('throws when the resulting bucket name contains an invalid character', () => {
    expect(() => layeredBucketNameFor(orgId, 'aws_us_east_1')).toThrow(/must be lowercase/)
  })
})

describe('awsRegionFromStorageRegion', () => {
  it('strips the aws- prefix', () => {
    expect(awsRegionFromStorageRegion('aws-us-east-1')).toBe('us-east-1')
    expect(awsRegionFromStorageRegion('aws-ap-southeast-2')).toBe('ap-southeast-2')
  })

  it('throws on non-aws providers', () => {
    expect(() => awsRegionFromStorageRegion('gcp-us-central1')).toThrow(/only 'aws-\*' is supported/)
  })

  it('throws on a missing region body', () => {
    expect(() => awsRegionFromStorageRegion('aws-')).toThrow(/missing AWS region/)
  })
})
