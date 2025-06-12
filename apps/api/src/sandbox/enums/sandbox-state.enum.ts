/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export enum SandboxState {
  CREATING = 'creating',
  RESTORING = 'restoring',
  DESTROYED = 'destroyed',
  DESTROYING = 'destroying',
  STARTED = 'started',
  STOPPED = 'stopped',
  STARTING = 'starting',
  STOPPING = 'stopping',
  ERROR = 'error',
  BUILD_FAILED = 'build_failed',
  PENDING_BUILD = 'pending_build',
  BUILDING_SNAPSHOT = 'building_snapshot',
  UNKNOWN = 'unknown',
  PULLING_SNAPSHOT = 'pulling_snapshot',
  ARCHIVING = 'archiving',
  ARCHIVED = 'archived',
}
