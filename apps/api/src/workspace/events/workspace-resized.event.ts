/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Workspace } from '../entities/workspace.entity'

export class WorkspaceResizedEvent {
  constructor(
    public readonly workspace: Workspace,
    public readonly oldCpu: number,
    public readonly newCpu: number,
    public readonly oldMem: number,
    public readonly newMem: number,
    public readonly oldGpu: number,
    public readonly newGpu: number,
  ) {}
}
