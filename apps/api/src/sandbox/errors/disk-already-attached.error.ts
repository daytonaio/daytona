/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export class DiskAlreadyAttachedError extends Error {
  constructor(diskId: string, sandboxId: string) {
    const message = `Disk ${diskId} is already attached to sandbox ${sandboxId}`
    super(message)
    this.name = 'DiskAlreadyAttachedError'
  }
}
