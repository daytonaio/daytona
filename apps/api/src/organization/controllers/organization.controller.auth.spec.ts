/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationController } from './organization.controller'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { OrganizationAuthContextGuard } from '../guards/organization-auth-context.guard'
import { OtelCollectorAuthContextGuard } from '../guards/otel-collector-auth-context.guard'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import {
  getAuthContextGuards,
  getAllowedAuthStrategies,
  getRequiredOrganizationMemberRole,
  getRequiredOrganizationResourcePermissions,
  expectArrayMatch,
  getRequiredSystemRole,
  createCoverageTracker,
} from '../../test/helpers/controller-metadata.helper'
import { SystemRole } from '../../user/enums/system-role.enum'
import { UserAuthContextGuard } from '../../user/guards/user-auth-context.guard'

describe('[AUTH] OrganizationController', () => {
  const trackMethod = createCoverageTracker(OrganizationController)

  it('findInvitationsByUser', () => {
    const methodName = trackMethod('findInvitationsByUser')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [UserAuthContextGuard])
  })

  it('getInvitationsCountByUser', () => {
    const methodName = trackMethod('getInvitationsCountByUser')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [UserAuthContextGuard])
  })

  it('acceptInvitation', () => {
    const methodName = trackMethod('acceptInvitation')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [UserAuthContextGuard])
  })

  it('declineInvitation', () => {
    const methodName = trackMethod('declineInvitation')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [UserAuthContextGuard])
  })

  it('create', () => {
    const methodName = trackMethod('create')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [UserAuthContextGuard])
  })

  it('setDefaultRegion', () => {
    const methodName = trackMethod('setDefaultRegion')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationController, methodName)).toBe(OrganizationMemberRole.OWNER)
  })

  it('findAll', () => {
    const methodName = trackMethod('findAll')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [UserAuthContextGuard])
  })

  it('findOne', () => {
    const methodName = trackMethod('findOne')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(OrganizationController, methodName)).toBeUndefined()
  })

  it('delete', () => {
    const methodName = trackMethod('delete')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationController, methodName)).toBe(OrganizationMemberRole.OWNER)
  })

  it('getUsageOverview', () => {
    const methodName = trackMethod('getUsageOverview')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(OrganizationController, methodName)).toBeUndefined()
  })

  it('leave', () => {
    const methodName = trackMethod('leave')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(OrganizationController, methodName)).toBeUndefined()
  })

  it('updateOrganizationQuota', () => {
    const methodName = trackMethod('updateOrganizationQuota')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [])
    expect(getRequiredSystemRole(OrganizationController, methodName)).toBe(SystemRole.ADMIN)
  })

  it('updateOrganizationRegionQuota', () => {
    const methodName = trackMethod('updateOrganizationRegionQuota')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [])
    expect(getRequiredSystemRole(OrganizationController, methodName)).toBe(SystemRole.ADMIN)
  })

  it('suspend', () => {
    const methodName = trackMethod('suspend')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [])
    expect(getRequiredSystemRole(OrganizationController, methodName)).toBe(SystemRole.ADMIN)
  })

  it('unsuspend', () => {
    const methodName = trackMethod('unsuspend')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [])
    expect(getRequiredSystemRole(OrganizationController, methodName)).toBe(SystemRole.ADMIN)
  })

  it('getOtelConfigBySandboxAuthToken', () => {
    const methodName = trackMethod('getOtelConfigBySandboxAuthToken')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [OtelCollectorAuthContextGuard])
  })

  it('updateSandboxDefaultLimitedNetworkEgress', () => {
    const methodName = trackMethod('updateSandboxDefaultLimitedNetworkEgress')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [])
    expect(getRequiredSystemRole(OrganizationController, methodName)).toBe(SystemRole.ADMIN)
  })

  it('updateExperimentalConfig', () => {
    const methodName = trackMethod('updateExperimentalConfig')
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationController, methodName)).toBe(OrganizationMemberRole.OWNER)
  })
})
