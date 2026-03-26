/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { DataSource } from 'typeorm'
import {
  sandboxLookupCacheKeyByAuthToken,
  sandboxLookupCacheKeyById,
  sandboxLookupCacheKeyByName,
  sandboxOrgIdCacheKeyById,
  sandboxOrgIdCacheKeyByName,
} from '../utils/sandbox-lookup-cache.util'

type InvalidateSandboxLookupCacheArgs =
  | {
      sandboxId: string
      organizationId: string
      name: string
      previousOrganizationId?: string | null
      previousName?: string | null
    }
  | {
      authToken: string
    }

@Injectable()
export class SandboxLookupCacheInvalidationService {
  private readonly logger = new Logger(SandboxLookupCacheInvalidationService.name)

  constructor(private readonly dataSource: DataSource) {}

  invalidate(args: InvalidateSandboxLookupCacheArgs): void {
    const cache = this.dataSource.queryResultCache
    if (!cache) {
      return
    }

    if ('authToken' in args) {
      cache
        .remove([sandboxLookupCacheKeyByAuthToken({ authToken: args.authToken })])
        .then(() => this.logger.debug(`Invalidated sandbox lookup cache for authToken ${args.authToken}`))
        .catch((error) =>
          this.logger.warn(
            `Failed to invalidate sandbox lookup cache for authToken ${args.authToken}: ${error instanceof Error ? error.message : String(error)}`,
          ),
        )
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
      for (const returnDestroyed of [false, true]) {
        cacheIds.push(
          sandboxLookupCacheKeyById({
            organizationId,
            returnDestroyed,
            sandboxId: args.sandboxId,
          }),
        )
        for (const sandboxName of names) {
          cacheIds.push(
            sandboxLookupCacheKeyByName({
              organizationId,
              returnDestroyed,
              sandboxName,
            }),
          )
        }
      }
    }

    if (cacheIds.length === 0) {
      return
    }

    cache
      .remove(cacheIds)
      .then(() => this.logger.debug(`Invalidated sandbox lookup cache for ${args.sandboxId}`))
      .catch((error) =>
        this.logger.warn(
          `Failed to invalidate sandbox lookup cache for ${args.sandboxId}: ${error instanceof Error ? error.message : String(error)}`,
        ),
      )
  }

  invalidateOrgId(args: {
    sandboxId: string
    organizationId: string
    name: string
    previousOrganizationId?: string | null
    previousName?: string | null
  }): void {
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
      cacheIds.push(
        sandboxOrgIdCacheKeyById({
          organizationId,
          sandboxId: args.sandboxId,
        }),
      )
      for (const sandboxName of names) {
        cacheIds.push(
          sandboxOrgIdCacheKeyByName({
            organizationId,
            sandboxName,
          }),
        )
      }
    }

    // Also invalidate the "no org" variants (when organizationId was not provided to getOrganizationId)
    cacheIds.push(sandboxOrgIdCacheKeyById({ sandboxId: args.sandboxId }))
    for (const sandboxName of names) {
      cacheIds.push(sandboxOrgIdCacheKeyByName({ sandboxName }))
    }

    cache
      .remove(cacheIds)
      .then(() => this.logger.debug(`Invalidated sandbox orgId cache for ${args.sandboxId}`))
      .catch((error) =>
        this.logger.warn(
          `Failed to invalidate sandbox orgId cache for ${args.sandboxId}: ${error instanceof Error ? error.message : String(error)}`,
        ),
      )
  }
}
