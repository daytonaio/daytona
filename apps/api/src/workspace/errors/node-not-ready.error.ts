/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export class NodeNotReadyError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'NodeNotReadyError'
  }
}
