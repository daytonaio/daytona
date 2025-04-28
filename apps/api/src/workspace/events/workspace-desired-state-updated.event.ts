/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Workspace } from '../entities/workspace.entity'
import { WorkspaceDesiredState } from '../enums/workspace-desired-state.enum'

export class WorkspaceDesiredStateUpdatedEvent {
  constructor(
    public readonly workspace: Workspace,
    public readonly oldDesiredState: WorkspaceDesiredState,
    public readonly newDesiredState: WorkspaceDesiredState,
  ) {}
}
