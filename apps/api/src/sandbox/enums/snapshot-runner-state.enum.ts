/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export enum SnapshotRunnerState {
  PULLING_SNAPSHOT = 'pulling_snapshot',
  BUILDING_SNAPSHOT = 'building_snapshot',
  READY = 'ready',
  ERROR = 'error',
  REMOVING = 'removing',
}
