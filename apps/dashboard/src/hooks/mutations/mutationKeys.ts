/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const mutationKeys = {
  organization: {
    members: {
      all: ['organization-members'] as const,
      updateAccess: () => [...mutationKeys.organization.members.all, 'update-access'] as const,
      remove: () => [...mutationKeys.organization.members.all, 'remove'] as const,
    },
    invitations: {
      all: ['organization-invitations'] as const,
      create: () => [...mutationKeys.organization.invitations.all, 'create'] as const,
      update: () => [...mutationKeys.organization.invitations.all, 'update'] as const,
      cancel: () => [...mutationKeys.organization.invitations.all, 'cancel'] as const,
    },
  },
} as const
