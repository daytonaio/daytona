/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Disk } from '../entities/disk.entity'
import { DiskState } from '../enums/disk-state.enum'

export class DiskStateUpdatedEvent {
  constructor(
    public readonly disk: Disk,
    public readonly oldState: DiskState,
    public readonly newState: DiskState,
  ) {}
}
