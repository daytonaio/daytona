/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Snapshot } from '../entities/snapshot.entity'
import { SnapshotState } from '../enums/snapshot-state.enum'

export class SnapshotStateUpdatedEvent {
  constructor(
    public readonly snapshot: Snapshot,
    public readonly oldState: SnapshotState,
    public readonly newState: SnapshotState,
  ) {}
}
