/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Workspace } from '../entities/workspace.entity'
import { WorkspaceState } from '../enums/workspace-state.enum'

export class WorkspaceStateUpdatedEvent {
  constructor(
    public readonly workspace: Workspace,
    public readonly oldState: WorkspaceState,
    public readonly newState: WorkspaceState,
  ) {}
}
