/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationMemberRole } from '../../organization/enums/organization-member-role.enum'
import { Organization } from '../../organization/entities/organization.entity'
import { OrganizationUser } from '../../organization/entities/organization-user.entity'
import { Runner } from '../../sandbox/entities/runner.entity'
import { MOCK_ORGANIZATION_ID, MOCK_RUNNER_ID, MOCK_REGION_ID, MOCK_USER_ID } from './constants'

export function createMockOrganization(overrides?: Partial<Organization>): Organization {
  return {
    id: MOCK_ORGANIZATION_ID,
    ...overrides,
  } as Organization
}

export function createMockOrganizationUser(overrides?: Partial<OrganizationUser>): OrganizationUser {
  return {
    organizationId: MOCK_ORGANIZATION_ID,
    userId: MOCK_USER_ID,
    role: OrganizationMemberRole.MEMBER,
    assignedRoles: [],
    ...overrides,
  } as OrganizationUser
}

export function createMockRunner(overrides?: Partial<Runner>): Runner {
  return {
    id: MOCK_RUNNER_ID,
    region: MOCK_REGION_ID,
    ...overrides,
  } as Runner
}
