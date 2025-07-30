/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const OrganizationEvents = {
  INVITATION_CREATED: 'invitation.created',
  INVITATION_ACCEPTED: 'invitation.accepted',
  INVITATION_DECLINED: 'invitation.declined',
  INVITATION_CANCELLED: 'invitation.cancelled',
  CREATED: 'organization.created',
  SUSPENDED_SANDBOX_STOPPED: 'organization.suspended-sandbox-stopped',
  SUSPENDED_SNAPSHOT_DEACTIVATED: 'organization.suspended-snapshot-deactivated',
} as const
