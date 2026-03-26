/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Sandbox } from '../entities/sandbox.entity'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'

export class SandboxDesiredStateUpdatedEvent {
  constructor(
    public readonly sandbox: Sandbox,
    public readonly oldDesiredState: SandboxDesiredState,
    public readonly newDesiredState: SandboxDesiredState,
  ) {}
}
