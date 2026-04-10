/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ToolboxController } from './toolbox.deprecated.controller'
import { OrganizationAuthContextGuard } from '../../organization/guards/organization-auth-context.guard'
import { SandboxAccessGuard } from '../guards/sandbox-access.guard'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import {
  getAuthContextGuards,
  getResourceAccessGuards,
  getAllowedAuthStrategies,
  getRequiredOrganizationMemberRole,
  getRequiredOrganizationResourcePermissions,
  expectArrayMatch,
} from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] ToolboxController', () => {
  it('allows API_KEY and JWT', () => {
    expectArrayMatch(getAllowedAuthStrategies(ToolboxController), [AuthStrategyType.API_KEY, AuthStrategyType.JWT])
  })

  it('requires OrganizationAuthContextGuard', () => {
    expectArrayMatch(getAuthContextGuards(ToolboxController), [OrganizationAuthContextGuard])
  })

  it('requires SandboxAccessGuard', () => {
    expectArrayMatch(getResourceAccessGuards(ToolboxController), [SandboxAccessGuard])
  })

  it('requires WRITE_SANDBOXES permission', () => {
    expect(getRequiredOrganizationMemberRole(ToolboxController)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(ToolboxController), [
      OrganizationResourcePermission.WRITE_SANDBOXES,
    ])
  })
})
