/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SnapshotState } from '../../sandbox/enums/snapshot-state.enum'

export const SNAPSHOT_USAGE_IGNORED_STATES: SnapshotState[] = [
  SnapshotState.ERROR,
  SnapshotState.BUILD_FAILED,
  SnapshotState.INACTIVE,
  SnapshotState.REMOVING,
]
