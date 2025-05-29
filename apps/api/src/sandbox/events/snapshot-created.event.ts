/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Snapshot } from '../entities/snapshot.entity'

export class SnapshotCreatedEvent {
  constructor(public readonly snapshot: Snapshot) {}
}
