/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export class SnapshotStateError extends Error {
  constructor(public readonly errorReason: string) {
    super(errorReason)
    this.name = 'SnapshotStateError'
  }
}
