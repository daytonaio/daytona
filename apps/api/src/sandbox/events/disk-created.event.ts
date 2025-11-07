/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Disk } from '../entities/disk.entity'

export class DiskCreatedEvent {
  constructor(public readonly disk: Disk) {}
}
