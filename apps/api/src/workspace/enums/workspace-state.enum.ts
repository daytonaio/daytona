/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export enum WorkspaceState {
  CREATING = 'creating',
  RESTORING = 'restoring',
  DESTROYED = 'destroyed',
  DESTROYING = 'destroying',
  STARTED = 'started',
  STOPPED = 'stopped',
  STARTING = 'starting',
  STOPPING = 'stopping',
  ERROR = 'error',
  PENDING_BUILD = 'pending_build',
  BUILDING_IMAGE = 'building_image',
  UNKNOWN = 'unknown',
  PULLING_IMAGE = 'pulling_image',
  ARCHIVING = 'archiving',
  ARCHIVED = 'archived',
}
