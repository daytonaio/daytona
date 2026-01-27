/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const WebhookEvents = {
  SANDBOX_CREATED: 'sandbox.created',
  SANDBOX_STATE_UPDATED: 'sandbox.state.updated',
  SNAPSHOT_CREATED: 'snapshot.created',
  SNAPSHOT_STATE_UPDATED: 'snapshot.state.updated',
  SNAPSHOT_REMOVED: 'snapshot.removed',
  VOLUME_CREATED: 'volume.created',
  VOLUME_STATE_UPDATED: 'volume.state.updated',
} as const
