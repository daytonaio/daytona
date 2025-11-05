/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export enum DiskState {
  FRESH = 'fresh',
  PULLING = 'pulling',
  READY = 'ready',
  ATTACHED = 'attached',
  DETACHED = 'detached',
  PENDING_PUSH = 'pending_push',
  PUSHING = 'pushing',
  STORED = 'stored',
  PENDING_DELETE = 'pending_delete',
  DELETING = 'deleting',
  DELETED = 'deleted',
  ERROR = 'error',
  PENDING_FORK = 'pending_fork',
  FORKING = 'forking',
  LOCKED = 'locked',
}
