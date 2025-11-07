/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export class CanNotDettachDiskError extends Error {
  constructor(diskId: string, reason: string) {
    const message = `Can not dettach disk ${diskId} from sandbox: ${reason}`
    super(message)
    this.name = 'CanNotDettachDiskError'
  }
}
