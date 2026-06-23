/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const mutationKeys = {
  sandboxes: {
    all: ['sandboxes'] as const,
    start: () => [...mutationKeys.sandboxes.all, 'start'] as const,
    stop: () => [...mutationKeys.sandboxes.all, 'stop'] as const,
    archive: () => [...mutationKeys.sandboxes.all, 'archive'] as const,
    recover: () => [...mutationKeys.sandboxes.all, 'recover'] as const,
    remove: () => [...mutationKeys.sandboxes.all, 'remove'] as const,
    fork: () => [...mutationKeys.sandboxes.all, 'fork'] as const,
    createSnapshot: () => [...mutationKeys.sandboxes.all, 'create-snapshot'] as const,
    vnc: () => [...mutationKeys.sandboxes.all, 'vnc'] as const,
    screenRecordings: () => [...mutationKeys.sandboxes.all, 'screen-recordings'] as const,
  },
  regions: {
    all: ['regions'] as const,
    create: () => [...mutationKeys.regions.all, 'create'] as const,
    update: () => [...mutationKeys.regions.all, 'update'] as const,
    remove: () => [...mutationKeys.regions.all, 'remove'] as const,
    regenerateProxyApiKey: () => [...mutationKeys.regions.all, 'regenerate-proxy-api-key'] as const,
    regenerateSshGatewayApiKey: () => [...mutationKeys.regions.all, 'regenerate-ssh-gateway-api-key'] as const,
    regenerateSnapshotManagerCredentials: () =>
      [...mutationKeys.regions.all, 'regenerate-snapshot-manager-credentials'] as const,
  },
  runners: {
    all: ['runners'] as const,
    create: () => [...mutationKeys.runners.all, 'create'] as const,
    updateScheduling: () => [...mutationKeys.runners.all, 'update-scheduling'] as const,
    remove: () => [...mutationKeys.runners.all, 'remove'] as const,
  },
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
