/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * Adding/modifying job types here requires a database migration to update the VALIDATE_JOB_TYPE check constraint in the job table
 */
export enum JobType {
  CREATE_SANDBOX = 'CREATE_SANDBOX',
  START_SANDBOX = 'START_SANDBOX',
  STOP_SANDBOX = 'STOP_SANDBOX',
  DESTROY_SANDBOX = 'DESTROY_SANDBOX',
  CREATE_BACKUP = 'CREATE_BACKUP',
  BUILD_SNAPSHOT = 'BUILD_SNAPSHOT',
  PULL_SNAPSHOT = 'PULL_SNAPSHOT',
  REMOVE_SNAPSHOT = 'REMOVE_SNAPSHOT',
  UPDATE_SANDBOX_NETWORK_SETTINGS = 'UPDATE_SANDBOX_NETWORK_SETTINGS',
}
