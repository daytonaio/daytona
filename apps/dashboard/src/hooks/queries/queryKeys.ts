/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SnapshotQueryParams } from './useSnapshotsQuery'

export const queryKeys = {
  config: {
    all: ['config'] as const,
  },
  apiKeys: {
    all: ['api-keys'] as const,
    list: (organizationId: string) => [...queryKeys.apiKeys.all, organizationId, 'list'] as const,
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

  },
  user: {
    all: ['users'] as const,
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
  sandbox: {
    all: ['sandbox'] as const,
    session: (scope: string) => [...queryKeys.sandbox.all, scope] as const,
    currentId: (scope: string) => [...queryKeys.sandbox.all, scope, 'current-id'] as const,
    instance: (scope: string, id: string) => [...queryKeys.sandbox.all, scope, id] as const,
    terminalUrl: (scope: string, id: string) => [...queryKeys.sandbox.all, scope, id, 'terminal-url'] as const,
    vncStatus: (scope: string, id: string) => [...queryKeys.sandbox.all, scope, id, 'vnc-status'] as const,
    vncUrl: (scope: string, id: string) => [...queryKeys.sandbox.all, scope, id, 'vnc-url'] as const,
  },
} as const
