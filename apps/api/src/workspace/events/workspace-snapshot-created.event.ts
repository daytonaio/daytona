/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Workspace } from '../entities/workspace.entity'

export class WorkspaceSnapshotCreatedEvent {
  constructor(public readonly workspace: Workspace) {}
}
