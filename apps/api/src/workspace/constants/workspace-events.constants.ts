/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const WorkspaceEvents = {
  STATE_UPDATED: 'workspace.state.updated',
  DESIRED_STATE_UPDATED: 'workspace.desired-state.updated',
  CREATED: 'workspace.created',
  STARTED: 'workspace.started',
  STOPPED: 'workspace.stopped',
  DESTROYED: 'workspace.destroyed',
  RESIZED: 'workspace.resized',
  PUBLIC_STATUS_UPDATED: 'workspace.public-status.updated',
  ORGANIZATION_UPDATED: 'workspace.organization.updated',
} as const
