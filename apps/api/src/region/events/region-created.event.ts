/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { EntityManager } from 'typeorm'
import { Region } from '../entities/region.entity'

export class RegionCreatedEvent {
  constructor(
    public readonly entityManager: EntityManager,
    public readonly region: Region,
    public readonly organizationId: string | null,
  ) {}
}
