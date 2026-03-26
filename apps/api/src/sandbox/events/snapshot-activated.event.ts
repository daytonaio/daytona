/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Snapshot } from '../entities/snapshot.entity'

export class SnapshotActivatedEvent {
  constructor(public readonly snapshot: Snapshot) {}
}
