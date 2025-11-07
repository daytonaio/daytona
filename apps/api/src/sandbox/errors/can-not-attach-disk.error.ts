/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export class CanNotAttachDiskError extends Error {
  constructor(diskId: string, reason: string) {
    const message = `Can not attach disk ${diskId} to sandbox: ${reason}`
    super(message)
    this.name = 'CanNotAttachDiskError'
  }
}
