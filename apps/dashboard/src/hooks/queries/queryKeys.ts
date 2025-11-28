/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const queryKeys = {
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

  billing: {
    all: ['billing'] as const,
    tiers: () => [...queryKeys.billing.all, 'tiers'] as const,
  },
} as const
