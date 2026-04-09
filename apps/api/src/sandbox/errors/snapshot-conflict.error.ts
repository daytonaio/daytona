/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ConflictException } from '@nestjs/common'

export class SnapshotConflictError extends ConflictException {
  constructor() {
    super('Snapshot was modified by another operation')
  }
}
