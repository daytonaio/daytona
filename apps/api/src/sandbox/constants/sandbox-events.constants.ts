/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const SandboxEvents = {
  ARCHIVED: 'sandbox.archived',
  STATE_UPDATED: 'sandbox.state.updated',
  DESIRED_STATE_UPDATED: 'sandbox.desired-state.updated',
  CREATED: 'sandbox.created',
  STARTED: 'sandbox.started',
  STOPPED: 'sandbox.stopped',
  DESTROYED: 'sandbox.destroyed',
  PUBLIC_STATUS_UPDATED: 'sandbox.public-status.updated',
  ORGANIZATION_UPDATED: 'sandbox.organization.updated',
  BACKUP_CREATED: 'sandbox.backup.created',
  AUTO_STOPPED: 'sandbox.auto-stopped',
  AUTO_ARCHIVED: 'sandbox.auto-archived',
  AUTO_DELETED: 'sandbox.auto-deleted',
} as const
