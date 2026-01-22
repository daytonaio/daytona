/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SnapshotQueryParams } from './useSnapshotsQuery'

export const queryKeys = {
  config: {
    all: ['config'] as const,
  },
  organization: {
    all: ['organization'] as const,

    list: () => [...queryKeys.organization.all, 'list'] as const,
    detail: (organizationId: string) => [...queryKeys.organization.all, organizationId, 'detail'] as const,

    usage: {
      overview: (organizationId: string) =>
        [...queryKeys.organization.all, organizationId, 'usage', 'overview'] as const,
      past: (organizationId: string) => [...queryKeys.organization.all, organizationId, 'usage', 'past'] as const,
    },

    tier: (organizationId: string) => [...queryKeys.organization.all, organizationId, 'tier'] as const,
    wallet: (organizationId: string) => [...queryKeys.organization.all, organizationId, 'wallet'] as const,
  },
  user: {
    all: ['users'] as const,
    accountProviders: () => [...queryKeys.user.all, 'account-providers'] as const,
  },
  billing: {
    all: ['billing'] as const,
    tiers: () => [...queryKeys.billing.all, 'tiers'] as const,
    emails: (organizationId: string) => [...queryKeys.billing.all, organizationId, 'emails'] as const,
    portalUrl: (organizationId: string) => [...queryKeys.billing.all, organizationId, 'portal-url'] as const,
  },
  snapshots: {
    all: ['snapshots'] as const,
    list: (organizationId: string, params?: SnapshotQueryParams) => {
      const base = [...queryKeys.snapshots.all, organizationId, 'list'] as const
      if (!params) return base
      return [
        ...base,
        {
          page: params.page,
          pageSize: params.pageSize,
          ...(params.filters && { filters: params.filters }),
          ...(params.sorting && { sorting: params.sorting }),
        },
      ] as const
    },
  },
} as const
