/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { EntityManager } from 'typeorm'
import { OrganizationResourcePermission } from '../enums/organization-resource-permission.enum'

export class OrganizationResourcePermissionsUnassignedEvent {
  constructor(
    public readonly entityManager: EntityManager,
    public readonly organizationId: string,
    public readonly userId: string,
    public readonly unassignedPermissions: OrganizationResourcePermission[],
  ) {}
}
