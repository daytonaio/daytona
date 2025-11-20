/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SnapshotState } from '../../sandbox/enums/snapshot-state.enum'

export const SNAPSHOT_STATES_CONSUMING_RESOURCES: SnapshotState[] = [
  SnapshotState.BUILDING,
  SnapshotState.PENDING,
  SnapshotState.PULLING,
  SnapshotState.PENDING_VALIDATION,
  SnapshotState.VALIDATING,
  SnapshotState.ACTIVE,
]
