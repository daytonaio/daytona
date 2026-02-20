/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const SNAPSHOT_LOOKUP_CACHE_TTL_MS = 60_000

type SnapshotLookupCacheKeyArgs = {
  organizationId: string
}

export function snapshotLookupCacheKeyByName(args: SnapshotLookupCacheKeyArgs & { snapshotName: string }): string {
  return `snapshot:lookup:by-name:org:${args.organizationId}:value:${args.snapshotName}`
}

export function snapshotLookupCacheKeyByNameGeneral(args: { snapshotName: string }): string {
  return `snapshot:lookup:by-name:general:value:${args.snapshotName}`
}
