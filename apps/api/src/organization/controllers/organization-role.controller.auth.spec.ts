/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationRoleController } from './organization-role.controller'
import { OrganizationAuthContextGuard } from '../guards/organization-auth-context.guard'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import {
  getAuthContextGuards,
  getAllowedAuthStrategies,
  getRequiredOrganizationMemberRole,
  getRequiredOrganizationResourcePermissions,
  expectArrayMatch,
  createCoverageTracker,
} from '../../test/helpers/controller-metadata.helper'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'

describe('[AUTH] OrganizationRoleController', () => {
  const trackMethod = createCoverageTracker(OrganizationRoleController)

  it('create', () => {
    const methodName = trackMethod('create')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationRoleController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationRoleController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationRoleController, methodName)).toBe(OrganizationMemberRole.OWNER)
  })

  it('findAll', () => {
    const methodName = trackMethod('findAll')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationRoleController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationRoleController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationRoleController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(OrganizationRoleController, methodName)).toBeUndefined()
  })

  it('updateRole', () => {
    const methodName = trackMethod('updateRole')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationRoleController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationRoleController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationRoleController, methodName)).toBe(OrganizationMemberRole.OWNER)
  })

  it('delete', () => {
    const methodName = trackMethod('delete')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationRoleController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationRoleController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationRoleController, methodName)).toBe(OrganizationMemberRole.OWNER)
  })
})
