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
  WRITE_SANDBOXES = 'write:sandboxes',
  DELETE_SANDBOXES = 'delete:sandboxes',

  // volumes
  READ_VOLUMES = 'read:volumes',
  WRITE_VOLUMES = 'write:volumes',
  DELETE_VOLUMES = 'delete:volumes',

  // regions
  WRITE_REGIONS = 'write:regions',
  DELETE_REGIONS = 'delete:regions',

  // runners
  READ_RUNNERS = 'read:runners',
  WRITE_RUNNERS = 'write:runners',
  DELETE_RUNNERS = 'delete:runners',

  // audit
  READ_AUDIT_LOGS = 'read:audit_logs',

  // checkpoints
  WRITE_CHECKPOINTS = 'write:checkpoints',
  DELETE_CHECKPOINTS = 'delete:checkpoints',
}
