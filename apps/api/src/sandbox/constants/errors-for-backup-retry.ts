/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

// Substrings in a backup error reason that indicate a transient failure eligible for automatic retry
export const BACKUP_RETRY_ERROR_SUBSTRINGS: string[] = [
  'connect ECONNREFUSED',
  'received unexpected HTTP status',
  'read: connection reset by peer',
  'Backup timed out after 2 hours',
]
