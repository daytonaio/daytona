/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationUserController } from './organization-user.controller'
import { OrganizationAuthContextGuard } from '../guards/organization-auth-context.guard'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import {
  getAuthContextGuards,
  getAllowedAuthStrategies,
  getRequiredOrganizationMemberRole,
  getRequiredOrganizationResourcePermissions,
  expectArrayMatch,
  createCoverageTracker,
  isPublicEndpoint,
} from '../../test/helpers/controller-metadata.helper'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'

describe('[AUTH] OrganizationUserController', () => {
  const trackMethod = createCoverageTracker(OrganizationUserController)

  it('findAll', () => {
    const methodName = trackMethod('findAll')
    expect(isPublicEndpoint(OrganizationUserController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationUserController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationUserController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationUserController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(OrganizationUserController, methodName)).toBeUndefined()
  })

  it('updateAccess', () => {
    const methodName = trackMethod('updateAccess')
    expect(isPublicEndpoint(OrganizationUserController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationUserController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationUserController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationUserController, methodName)).toBe(OrganizationMemberRole.OWNER)
  })

  it('delete', () => {
    const methodName = trackMethod('delete')
    expect(isPublicEndpoint(OrganizationUserController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationUserController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationUserController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationUserController, methodName)).toBe(OrganizationMemberRole.OWNER)
  })
})
