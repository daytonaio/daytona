/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SnapshotQueryParams } from './useSnapshotsQuery'
import type { AuditLogsQueryParams } from './useAuditLogsQuery'
import type { SandboxQueryParams } from '../useSandboxes'

export const queryKeys = {
  config: {
    all: ['config'] as const,
  },
  apiKeys: {
    all: ['api-keys'] as const,
    list: (organizationId: string) => [...queryKeys.apiKeys.all, organizationId, 'list'] as const,
  },
  webhooks: {
    all: ['webhooks'] as const,
    appPortalAccess: (organizationId: string) =>
      [...queryKeys.webhooks.all, organizationId, 'app-portal-access'] as const,
    initializationStatus: (organizationId: string) =>
      [...queryKeys.webhooks.all, organizationId, 'initialization-status'] as const,
  },
  organization: {
    all: ['organization'] as const,

    list: () => [...queryKeys.organization.all, 'list'] as const,
    detail: (organizationId: string) => [...queryKeys.organization.all, organizationId, 'detail'] as const,
    roles: (organizationId: string) => [...queryKeys.organization.all, organizationId, 'roles'] as const,
    invitations: (organizationId: string) => [...queryKeys.organization.all, organizationId, 'invitations'] as const,

    usage: {
      overview: (organizationId: string) =>
        [...queryKeys.organization.all, organizationId, 'usage', 'overview'] as const,
      current: (organizationId: string) => [...queryKeys.organization.all, organizationId, 'usage', 'current'] as const,
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
    checkoutUrl: (organizationId: string) => [...queryKeys.billing.all, organizationId, 'checkout-url'] as const,
    invoices: (organizationId: string, page?: number, perPage?: number) =>
      [
        ...queryKeys.billing.all,
        organizationId,
        'invoices',
        ...(page !== undefined && perPage !== undefined ? [{ page, perPage }] : []),
      ] as const,
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
  registries: {
    all: ['registries'] as const,
    list: (organizationId: string) => [...queryKeys.registries.all, organizationId, 'list'] as const,
  },
  volumes: {
    all: ['volumes'] as const,
    list: (organizationId: string) => [...queryKeys.volumes.all, organizationId, 'list'] as const,
  },
  audit: {
    all: ['audit'] as const,
    logs: (organizationId: string, params: AuditLogsQueryParams) =>
      [
        ...queryKeys.audit.all,
        organizationId,
        'logs',
        {
          page: params.page,
          pageSize: params.pageSize,
          ...(params.from && { from: params.from.toISOString() }),
          ...(params.to && { to: params.to.toISOString() }),
          ...(params.cursor && { cursor: params.cursor }),
        },
      ] as const,
  },
  sandboxes: {
    all: ['sandboxes'] as const,
    list: (organizationId: string | undefined, params?: SandboxQueryParams) => {
      const base = [...queryKeys.sandboxes.all, organizationId, 'list'] as const
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
    detail: (organizationId: string, sandboxId: string) =>
      [...queryKeys.sandboxes.all, organizationId, sandboxId, 'detail'] as const,
    terminalSession: (sandboxId: string) => [...queryKeys.sandboxes.all, sandboxId, 'terminal-session'] as const,
    vncInitialStatus: (sandboxId: string) => [...queryKeys.sandboxes.all, sandboxId, 'vnc-initial-status'] as const,
    vncPollStatus: (sandboxId: string) => [...queryKeys.sandboxes.all, sandboxId, 'vnc-poll-status'] as const,
    vncSession: (sandboxId: string) => [...queryKeys.sandboxes.all, sandboxId, 'vnc-session'] as const,
  },
  telemetry: {
    all: ['telemetry'] as const,
    logs: (sandboxId: string, params: object) => [...queryKeys.telemetry.all, sandboxId, 'logs', params] as const,
    traces: (sandboxId: string, params: object) => [...queryKeys.telemetry.all, sandboxId, 'traces', params] as const,
    metrics: (sandboxId: string, params: object) => [...queryKeys.telemetry.all, sandboxId, 'metrics', params] as const,
    traceSpans: (sandboxId: string, traceId: string) =>
      [...queryKeys.telemetry.all, sandboxId, 'traces', traceId] as const,
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
  analytics: {
    all: ['analytics'] as const,
    aggregatedUsage: (organizationId: string, params: object) =>
      [...queryKeys.analytics.all, organizationId, 'aggregated-usage', params] as const,
    usageChart: (organizationId: string, params: object) =>
      [...queryKeys.analytics.all, organizationId, 'usage-chart', params] as const,
    sandboxesUsage: (organizationId: string, params: object) =>
      [...queryKeys.analytics.all, organizationId, 'sandboxes-usage', params] as const,
    sandboxUsagePeriods: (organizationId: string, sandboxId: string, params: object) =>
      [...queryKeys.analytics.all, organizationId, sandboxId, 'usage-periods', params] as const,
  },
} as const
