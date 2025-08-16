/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const WebhookEvents = {
  // Sandbox events (matching notification.service.ts)
  SANDBOX_CREATED: 'sandbox.created',
  SANDBOX_STATE_UPDATED: 'sandbox.state.updated',
  SANDBOX_DESIRED_STATE_UPDATED: 'sandbox.desired-state.updated',

  // Snapshot events (matching notification.service.ts)
  SNAPSHOT_CREATED: 'snapshot.created',
  SNAPSHOT_STATE_UPDATED: 'snapshot.state.updated',
  SNAPSHOT_REMOVED: 'snapshot.removed',

  // Volume events (matching notification.service.ts)
  VOLUME_CREATED: 'volume.created',
  VOLUME_STATE_UPDATED: 'volume.state.updated',
  VOLUME_LAST_USED_AT_UPDATED: 'volume.lastUsedAt.updated',

  // Audit events (matching notification.service.ts)
  AUDIT_LOG_CREATED: 'audit-log.created',
  AUDIT_LOG_UPDATED: 'audit-log.updated',

  // Organization events
  ORGANIZATION_CREATED: 'organization.created',
  ORGANIZATION_UPDATED: 'organization.updated',
  ORGANIZATION_SUSPENDED: 'organization.suspended',
  ORGANIZATION_UNSUSPENDED: 'organization.unsuspended',

  // User events
  USER_CREATED: 'user.created',
  USER_UPDATED: 'user.updated',
  USER_DELETED: 'user.deleted',
  USER_EMAIL_VERIFIED: 'user.email_verified',

  // Custom events
  CUSTOM: 'custom',
} as const

export type WebhookEventType = (typeof WebhookEvents)[keyof typeof WebhookEvents]
