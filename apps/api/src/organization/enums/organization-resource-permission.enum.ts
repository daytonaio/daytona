/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export enum OrganizationResourcePermission {
  // api keys
  //   WRITE_API_KEYS = 'write:api_keys',
  //   DELETE_API_KEYS = 'delete:api_keys',

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

  // audit
  READ_AUDIT_LOGS = 'read:audit_logs',
}
