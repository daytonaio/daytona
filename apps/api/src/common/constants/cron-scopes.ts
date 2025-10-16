/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const CRON_SCOPES = {
  AUDIT: 'audit',
  SANDBOXES: 'sandboxes',
  ORGANIZATIONS: 'organizations',
  SNAPSHOTS: 'snapshots',
  RUNNERS: 'runners',
  BACKUPS: 'backups',
  VOLUMES: 'volumes',
  WARM_POOLS: 'warm-pools',
  USAGE_PERIODS: 'usage-periods',
} as const
