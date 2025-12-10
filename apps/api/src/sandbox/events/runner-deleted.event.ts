/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { EntityManager } from 'typeorm/entity-manager/EntityManager.js'

export class RunnerDeletedEvent {
  constructor(
    public readonly entityManager: EntityManager,
    public readonly runnerId: string,
  ) {}
}
