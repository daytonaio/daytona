/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export enum SnapshotState {
  PENDING = 'pending',
  PULLING = 'pulling',
  PENDING_VALIDATION = 'pending_validation',
  VALIDATING = 'validating',
  ACTIVE = 'active',
  INACTIVE = 'inactive',
  BUILDING = 'building',
  WARMING_UP = 'warming_up',
  ERROR = 'error',
  BUILD_FAILED = 'build_failed',
  REMOVING = 'removing',
}
