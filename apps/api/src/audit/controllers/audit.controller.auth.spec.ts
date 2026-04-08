/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AuditController } from './audit.controller'
import { OrganizationAuthContextGuard } from '../../organization/guards/organization-auth-context.guard'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import {
  getAuthContextGuards,
  getAllowedAuthStrategies,
  getRequiredOrganizationMemberRole,
  getRequiredOrganizationResourcePermissions,
  expectArrayMatch,
  createCoverageTracker,
} from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] AuditController', () => {
  const trackMethod = createCoverageTracker(AuditController)

  it('getOrganizationLogs', () => {
    const methodName = trackMethod('getOrganizationLogs')
    expectArrayMatch(getAllowedAuthStrategies(AuditController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(AuditController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(AuditController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(AuditController, methodName), [
      OrganizationResourcePermission.READ_AUDIT_LOGS,
    ])
  })
})
