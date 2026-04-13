/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationInvitationController } from './organization-invitation.controller'
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

describe('[AUTH] OrganizationInvitationController', () => {
  const trackMethod = createCoverageTracker(OrganizationInvitationController)

  it('create', () => {
    const methodName = trackMethod('create')
    expect(isPublicEndpoint(OrganizationInvitationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationInvitationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationInvitationController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationInvitationController, methodName)).toBe(
      OrganizationMemberRole.OWNER,
    )
  })

  it('update', () => {
    const methodName = trackMethod('update')
    expect(isPublicEndpoint(OrganizationInvitationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationInvitationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationInvitationController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationInvitationController, methodName)).toBe(
      OrganizationMemberRole.OWNER,
    )
  })

  it('findPending', () => {
    const methodName = trackMethod('findPending')
    expect(isPublicEndpoint(OrganizationInvitationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationInvitationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationInvitationController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationInvitationController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(OrganizationInvitationController, methodName)).toBeUndefined()
  })

  it('cancel', () => {
    const methodName = trackMethod('cancel')
    expect(isPublicEndpoint(OrganizationInvitationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationInvitationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationInvitationController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationInvitationController, methodName)).toBe(
      OrganizationMemberRole.OWNER,
    )
  })
})
