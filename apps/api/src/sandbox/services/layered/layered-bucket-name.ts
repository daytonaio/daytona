/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

// Deterministic per-org layered bucket name. `storageRegion === null` returns the
// legacy pre-region name (read-only; no longer created). The `dt-vl-` prefix keeps
// the name under AWS's 63-char limit (worst-case: 6 + uuid + 1 + 18 = 61).
export function layeredBucketNameFor(orgId: string, storageRegion: string | null): string {
  if (!orgId) {
    throw new Error('layeredBucketNameFor: orgId is required')
  }
  if (storageRegion === null) {
    return `daytona-volume-layered-${orgId}`
  }
  const name = `dt-vl-${orgId}-${storageRegion}`
  assertValidS3BucketName(name)
  return name
}

const BUCKET_NAME_REGEX = /^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$/

function assertValidS3BucketName(name: string): void {
  if (name.length < 3 || name.length > 63) {
    throw new Error(`Invalid S3 bucket name '${name}': length ${name.length} is outside [3, 63]`)
  }
  if (!BUCKET_NAME_REGEX.test(name)) {
    throw new Error(
      `Invalid S3 bucket name '${name}': must be lowercase a-z, digits, or hyphens and start/end with alphanumeric`,
    )
  }
}

// "aws-us-east-1" → "us-east-1". Only the aws-* provider is supported today.
export function awsRegionFromStorageRegion(storageRegion: string): string {
  const prefix = 'aws-'
  if (!storageRegion.startsWith(prefix)) {
    throw new Error(`Unsupported storageRegion '${storageRegion}': only 'aws-*' is supported`)
  }
  const rest = storageRegion.slice(prefix.length)
  if (!rest) {
    throw new Error(`Invalid storageRegion '${storageRegion}': missing AWS region after 'aws-' prefix`)
  }
  return rest
}
