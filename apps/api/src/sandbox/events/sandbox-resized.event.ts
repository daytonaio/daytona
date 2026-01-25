/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Sandbox } from '../entities/sandbox.entity'

export class SandboxResizedEvent {
  constructor(
    public readonly sandbox: Sandbox,
    public readonly oldCpu: number,
    public readonly newCpu: number,
    public readonly oldMemory: number,
    public readonly newMemory: number,
  ) {}
}
