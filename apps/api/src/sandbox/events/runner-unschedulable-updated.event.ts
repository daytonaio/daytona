/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Runner } from '../entities/runner.entity'

export class RunnerUnschedulableUpdatedEvent {
  constructor(
    public readonly runner: Runner,
    public readonly oldUnschedulable: boolean,
    public readonly newUnschedulable: boolean,
  ) {}
}
