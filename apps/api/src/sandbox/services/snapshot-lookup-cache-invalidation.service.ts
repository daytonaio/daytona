/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { DataSource } from 'typeorm'
import { snapshotLookupCacheKeyByName, snapshotLookupCacheKeyByNameGeneral } from '../utils/snapshot-lookup-cache.util'

type InvalidateSnapshotLookupCacheArgs = {
  snapshotId: string
  organizationId?: string
  name: string
  previousOrganizationId?: string | null
  previousName?: string | null
}

@Injectable()
export class SnapshotLookupCacheInvalidationService {
  private readonly logger = new Logger(SnapshotLookupCacheInvalidationService.name)

  constructor(private readonly dataSource: DataSource) {}

  invalidate(args: InvalidateSnapshotLookupCacheArgs): void {
    const cache = this.dataSource.queryResultCache
    if (!cache) {
      return
    }

    const organizationIds = Array.from(
      new Set(
        [args.organizationId, args.previousOrganizationId].filter((id): id is string =>
          Boolean(id && id.trim().length > 0),
        ),
      ),
    )
    const names = Array.from(
      new Set([args.name, args.previousName].filter((n): n is string => Boolean(n && n.trim().length > 0))),
    )

    const cacheIds: string[] = []
    for (const organizationId of organizationIds) {
      for (const snapshotName of names) {
        cacheIds.push(
          snapshotLookupCacheKeyByName({
            organizationId,
            snapshotName,
          }),
        )
      }
    }

    // Also invalidate general snapshot cache entries
    for (const snapshotName of names) {
      cacheIds.push(snapshotLookupCacheKeyByNameGeneral({ snapshotName }))
    }

    if (cacheIds.length === 0) {
      return
    }

    cache
      .remove(cacheIds)
      .then(() => this.logger.debug(`Invalidated snapshot lookup cache for ${args.snapshotId}`))
      .catch((error) =>
        this.logger.warn(
          `Failed to invalidate snapshot lookup cache for ${args.snapshotId}: ${error instanceof Error ? error.message : String(error)}`,
        ),
      )
  }
}
