/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export enum SnapshotState {
  BUILDING = 'building',
  PENDING = 'pending',
  PULLING = 'pulling',
  ACTIVE = 'active',
  INACTIVE = 'inactive',
  ERROR = 'error',
  BUILD_FAILED = 'build_failed',
  REMOVING = 'removing',
}
