/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxController } from './sandbox.controller'
import { OrganizationAuthContextGuard } from '../../organization/guards/organization-auth-context.guard'
import { SandboxAccessGuard } from '../guards/sandbox-access.guard'
import { RunnerAuthContextGuard } from '../guards/runner-auth-context.guard'
import { SshGatewayAuthContextGuard } from '../guards/ssh-gateway-auth-context.guard'
import { ProxyAuthContextGuard } from '../guards/proxy-auth-context.guard'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import {
  getAuthContextGuards,
  getResourceAccessGuards,
  getAllowedAuthStrategies,
  getRequiredOrganizationMemberRole,
  getRequiredOrganizationResourcePermissions,
  expectArrayMatch,
  createCoverageTracker,
  isPublicEndpoint,
} from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] SandboxController', () => {
  const trackMethod = createCoverageTracker(SandboxController)

  it('listSandboxes', () => {
    const methodName = trackMethod('listSandboxes')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(SandboxController, methodName)).toBeUndefined()
  })

  it('listSandboxesPaginated', () => {
    const methodName = trackMethod('listSandboxesPaginated')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(SandboxController, methodName)).toBeUndefined()
  })

  it('createSandbox', () => {
    const methodName = trackMethod('createSandbox')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SandboxController, methodName), [
      OrganizationResourcePermission.WRITE_SANDBOXES,
    ])
  })

  it('getSandboxesForRunner', () => {
    const methodName = trackMethod('getSandboxesForRunner')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [RunnerAuthContextGuard])
  })

  it('getSandbox', () => {
    const methodName = trackMethod('getSandbox')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [
      OrganizationAuthContextGuard,
      ProxyAuthContextGuard,
      SshGatewayAuthContextGuard,
    ])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(SandboxController, methodName)).toBeUndefined()
  })

  it('deleteSandbox', () => {
    const methodName = trackMethod('deleteSandbox')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SandboxController, methodName), [
      OrganizationResourcePermission.DELETE_SANDBOXES,
    ])
  })

  it('recoverSandbox', () => {
    const methodName = trackMethod('recoverSandbox')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SandboxController, methodName), [
      OrganizationResourcePermission.WRITE_SANDBOXES,
    ])
  })

  it('startSandbox', () => {
    const methodName = trackMethod('startSandbox')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SandboxController, methodName), [
      OrganizationResourcePermission.WRITE_SANDBOXES,
    ])
  })

  it('stopSandbox', () => {
    const methodName = trackMethod('stopSandbox')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SandboxController, methodName), [
      OrganizationResourcePermission.WRITE_SANDBOXES,
    ])
  })

  it('resizeSandbox', () => {
    const methodName = trackMethod('resizeSandbox')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SandboxController, methodName), [
      OrganizationResourcePermission.WRITE_SANDBOXES,
    ])
  })

  it('replaceLabels', () => {
    const methodName = trackMethod('replaceLabels')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SandboxController, methodName), [
      OrganizationResourcePermission.WRITE_SANDBOXES,
    ])
  })

  it('updateSandboxState', () => {
    const methodName = trackMethod('updateSandboxState')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [RunnerAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
  })

  it('createBackup', () => {
    const methodName = trackMethod('createBackup')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SandboxController, methodName), [
      OrganizationResourcePermission.WRITE_SANDBOXES,
    ])
  })

  it('updatePublicStatus', () => {
    const methodName = trackMethod('updatePublicStatus')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SandboxController, methodName), [
      OrganizationResourcePermission.WRITE_SANDBOXES,
    ])
  })

  it('updateLastActivity', () => {
    const methodName = trackMethod('updateLastActivity')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [
      OrganizationAuthContextGuard,
      ProxyAuthContextGuard,
      SshGatewayAuthContextGuard,
    ])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(SandboxController, methodName)).toBeUndefined()
  })

  it('setAutostopInterval', () => {
    const methodName = trackMethod('setAutostopInterval')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SandboxController, methodName), [
      OrganizationResourcePermission.WRITE_SANDBOXES,
    ])
  })

  it('setAutoArchiveInterval', () => {
    const methodName = trackMethod('setAutoArchiveInterval')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SandboxController, methodName), [
      OrganizationResourcePermission.WRITE_SANDBOXES,
    ])
  })

  it('setAutoDeleteInterval', () => {
    const methodName = trackMethod('setAutoDeleteInterval')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SandboxController, methodName), [
      OrganizationResourcePermission.WRITE_SANDBOXES,
    ])
  })

  it('updateNetworkSettings', () => {
    const methodName = trackMethod('updateNetworkSettings')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SandboxController, methodName), [
      OrganizationResourcePermission.WRITE_SANDBOXES,
    ])
  })

  it('archiveSandbox', () => {
    const methodName = trackMethod('archiveSandbox')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SandboxController, methodName), [
      OrganizationResourcePermission.WRITE_SANDBOXES,
    ])
  })

  it('getPortPreviewUrl', () => {
    const methodName = trackMethod('getPortPreviewUrl')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(SandboxController, methodName)).toBeUndefined()
  })

  it('getSignedPortPreviewUrl', () => {
    const methodName = trackMethod('getSignedPortPreviewUrl')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(SandboxController, methodName)).toBeUndefined()
  })

  it('expireSignedPortPreviewUrl', () => {
    const methodName = trackMethod('expireSignedPortPreviewUrl')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(SandboxController, methodName)).toBeUndefined()
  })

  it('getBuildLogs', () => {
    const methodName = trackMethod('getBuildLogs')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(SandboxController, methodName)).toBeUndefined()
  })

  it('getBuildLogsUrl', () => {
    const methodName = trackMethod('getBuildLogsUrl')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(SandboxController, methodName)).toBeUndefined()
  })

  it('createSshAccess', () => {
    const methodName = trackMethod('createSshAccess')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SandboxController, methodName), [
      OrganizationResourcePermission.WRITE_SANDBOXES,
    ])
  })

  it('revokeSshAccess', () => {
    const methodName = trackMethod('revokeSshAccess')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SandboxController, methodName), [
      OrganizationResourcePermission.WRITE_SANDBOXES,
    ])
  })

  it('validateSshAccess', () => {
    const methodName = trackMethod('validateSshAccess')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [SshGatewayAuthContextGuard])
  })

  it('getToolboxProxyUrl', () => {
    const methodName = trackMethod('getToolboxProxyUrl')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(SandboxController, methodName)).toBeUndefined()
  })

  it('getOrganizationBySandboxId', () => {
    const methodName = trackMethod('getOrganizationBySandboxId')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [ProxyAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
  })

  it('getRegionQuotaBySandboxId', () => {
    const methodName = trackMethod('getRegionQuotaBySandboxId')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [ProxyAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
  })

  it('createSandboxSnapshot', () => {
    const methodName = trackMethod('createSandboxSnapshot')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SandboxController, methodName), [
      OrganizationResourcePermission.WRITE_SANDBOXES,
    ])
  })

  it('forkSandbox', () => {
    const methodName = trackMethod('forkSandbox')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SandboxController, methodName), [
      OrganizationResourcePermission.WRITE_SANDBOXES,
    ])
  })

  it('getSandboxForks', () => {
    const methodName = trackMethod('getSandboxForks')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(SandboxController, methodName)).toBeUndefined()
  })

  it('getSandboxParent', () => {
    const methodName = trackMethod('getSandboxParent')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(SandboxController, methodName)).toBeUndefined()
  })

  it('getSandboxAncestors', () => {
    const methodName = trackMethod('getSandboxAncestors')
    expect(isPublicEndpoint(SandboxController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
    expect(getRequiredOrganizationMemberRole(SandboxController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(SandboxController, methodName)).toBeUndefined()
  })
})
