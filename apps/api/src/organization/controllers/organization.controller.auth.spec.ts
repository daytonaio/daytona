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
  isPublicEndpoint,
} from '../../test/helpers/controller-metadata.helper'
import { SystemRole } from '../../user/enums/system-role.enum'
import { UserAuthContextGuard } from '../../user/guards/user-auth-context.guard'

describe('[AUTH] OrganizationController', () => {
  const trackMethod = createCoverageTracker(OrganizationController)

  it('findInvitationsByUser', () => {
    const methodName = trackMethod('findInvitationsByUser')
    expect(isPublicEndpoint(OrganizationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [UserAuthContextGuard])
  })

  it('getInvitationsCountByUser', () => {
    const methodName = trackMethod('getInvitationsCountByUser')
    expect(isPublicEndpoint(OrganizationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [UserAuthContextGuard])
  })

  it('acceptInvitation', () => {
    const methodName = trackMethod('acceptInvitation')
    expect(isPublicEndpoint(OrganizationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [UserAuthContextGuard])
  })

  it('declineInvitation', () => {
    const methodName = trackMethod('declineInvitation')
    expect(isPublicEndpoint(OrganizationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [UserAuthContextGuard])
  })

  it('create', () => {
    const methodName = trackMethod('create')
    expect(isPublicEndpoint(OrganizationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [UserAuthContextGuard])
  })

  it('setDefaultRegion', () => {
    const methodName = trackMethod('setDefaultRegion')
    expect(isPublicEndpoint(OrganizationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationController, methodName)).toBe(OrganizationMemberRole.OWNER)
  })

  it('findAll', () => {
    const methodName = trackMethod('findAll')
    expect(isPublicEndpoint(OrganizationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [UserAuthContextGuard])
  })

  it('findOne', () => {
    const methodName = trackMethod('findOne')
    expect(isPublicEndpoint(OrganizationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(OrganizationController, methodName)).toBeUndefined()
  })

  it('delete', () => {
    const methodName = trackMethod('delete')
    expect(isPublicEndpoint(OrganizationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationController, methodName)).toBe(OrganizationMemberRole.OWNER)
  })

  it('getUsageOverview', () => {
    const methodName = trackMethod('getUsageOverview')
    expect(isPublicEndpoint(OrganizationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(OrganizationController, methodName)).toBeUndefined()
  })

  it('leave', () => {
    const methodName = trackMethod('leave')
    expect(isPublicEndpoint(OrganizationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(OrganizationController, methodName)).toBeUndefined()
  })

  it('updateOrganizationQuota', () => {
    const methodName = trackMethod('updateOrganizationQuota')
    expect(isPublicEndpoint(OrganizationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [])
    expect(getRequiredSystemRole(OrganizationController, methodName)).toBe(SystemRole.ADMIN)
  })

  it('updateOrganizationRegionQuota', () => {
    const methodName = trackMethod('updateOrganizationRegionQuota')
    expect(isPublicEndpoint(OrganizationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [])
    expect(getRequiredSystemRole(OrganizationController, methodName)).toBe(SystemRole.ADMIN)
  })

  it('suspend', () => {
    const methodName = trackMethod('suspend')
    expect(isPublicEndpoint(OrganizationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [])
    expect(getRequiredSystemRole(OrganizationController, methodName)).toBe(SystemRole.ADMIN)
  })

  it('unsuspend', () => {
    const methodName = trackMethod('unsuspend')
    expect(isPublicEndpoint(OrganizationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [])
    expect(getRequiredSystemRole(OrganizationController, methodName)).toBe(SystemRole.ADMIN)
  })

  it('getOtelConfigBySandboxAuthToken', () => {
    const methodName = trackMethod('getOtelConfigBySandboxAuthToken')
    expect(isPublicEndpoint(OrganizationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [OtelCollectorAuthContextGuard])
  })

  it('getOtelConfig', () => {
    const methodName = trackMethod('getOtelConfig')
    expect(isPublicEndpoint(OrganizationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [OtelCollectorAuthContextGuard])
  })

  it('updateSandboxDefaultLimitedNetworkEgress', () => {
    const methodName = trackMethod('updateSandboxDefaultLimitedNetworkEgress')
    expect(isPublicEndpoint(OrganizationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [])
    expect(getRequiredSystemRole(OrganizationController, methodName)).toBe(SystemRole.ADMIN)
  })

  it('updateExperimentalConfig', () => {
    const methodName = trackMethod('updateExperimentalConfig')
    expect(isPublicEndpoint(OrganizationController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(OrganizationController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationController, methodName)).toBe(OrganizationMemberRole.OWNER)
  })
})
