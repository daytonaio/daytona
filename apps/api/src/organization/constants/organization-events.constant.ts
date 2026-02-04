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
  DELETED: 'organization.deleted',
  SUSPENDED_SANDBOX_STOPPED: 'organization.suspended-sandbox-stopped',
  SUSPENDED_SNAPSHOT_DEACTIVATED: 'organization.suspended-snapshot-deactivated',
  PERMISSIONS_UNASSIGNED: 'permissions.unassigned',
  ASSERT_NO_USERS: 'organization.assert-no-users',
  ASSERT_NO_SANDBOXES: 'organization.assert-no-sandboxes',
  ASSERT_NO_SNAPSHOTS: 'organization.assert-no-snapshots',
  ASSERT_NO_VOLUMES: 'organization.assert-no-volumes',
  ASSERT_NO_RUNNERS: 'organization.assert-no-runners',
} as const
