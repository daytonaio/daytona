/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { EntityManager } from 'typeorm'
import { CreateOrganizationQuotaDto } from '../../organization/dto/create-organization-quota.dto'

export class UserCreatedEvent {
  constructor(
    public readonly entityManager: EntityManager,
    public readonly userId: string,
    public readonly email?: string,
    public readonly personalOrganizationQuota?: CreateOrganizationQuotaDto,
    public readonly emailVerified?: boolean,
  ) {}
}
