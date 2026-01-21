/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { EntityManager } from 'typeorm'
import { Region } from '../entities/region.entity'

export class RegionSnapshotManagerCredsRegeneratedEvent {
  constructor(
    public readonly regionId: string,
    public readonly snapshotManagerUrl: string,
    public readonly username: string,
    public readonly password: string,
    public readonly entityManager?: EntityManager,
  ) {}
}

export class RegionSnapshotManagerUpdatedEvent {
  constructor(
    public readonly region: Region,
    public readonly organizationId: string,
    public readonly snapshotManagerUrl: string | null,
    public readonly prevSnapshotManagerUrl: string | null,
    public readonly newUsername?: string,
    public readonly newPassword?: string,
    public readonly entityManager?: EntityManager,
  ) {}
}
