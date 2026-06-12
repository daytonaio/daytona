/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export enum SnapshotState {
  BUILDING = 'building',
  PENDING = 'pending',
  PULLING = 'pulling',
  //  snapshot is being created from a sandbox (sandbox is in the SNAPSHOTTING state)
  SNAPSHOTTING = 'snapshotting',
  ACTIVE = 'active',
  INACTIVE = 'inactive',
  ERROR = 'error',
  BUILD_FAILED = 'build_failed',
  REMOVING = 'removing',
}
