/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/*
  IMPORTANT: When adding a new permission, make sure to update apps/dashboard/src/constants/CreateApiKeyPermissionsGroups.ts accordingly
*/
export enum OrganizationResourcePermission {
  // docker registries
  WRITE_REGISTRIES = 'write:registries',
  DELETE_REGISTRIES = 'delete:registries',

  // snapshots
  WRITE_SNAPSHOTS = 'write:snapshots',
  DELETE_SNAPSHOTS = 'delete:snapshots',

  // sandboxes
  READ_SANDBOXES = 'read:sandboxes',
  WRITE_SANDBOXES = 'write:sandboxes',
  DELETE_SANDBOXES = 'delete:sandboxes',

  // volumes
  READ_VOLUMES = 'read:volumes',
  WRITE_VOLUMES = 'write:volumes',
  DELETE_VOLUMES = 'delete:volumes',

  // audit
  READ_AUDIT_LOGS = 'read:audit_logs',
}
