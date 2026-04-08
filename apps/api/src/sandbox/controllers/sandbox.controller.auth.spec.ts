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
} from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] SandboxController', () => {
  const trackMethod = createCoverageTracker(SandboxController)

  it('listSandboxes', () => {
    const methodName = trackMethod('listSandboxes')
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
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [RunnerAuthContextGuard])
  })

  it('getSandbox', () => {
    const methodName = trackMethod('getSandbox')
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
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [RunnerAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
  })

  it('createBackup', () => {
    const methodName = trackMethod('createBackup')
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
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [SshGatewayAuthContextGuard])
  })

  it('getToolboxProxyUrl', () => {
    const methodName = trackMethod('getToolboxProxyUrl')
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
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [ProxyAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
  })

  it('getRegionQuotaBySandboxId', () => {
    const methodName = trackMethod('getRegionQuotaBySandboxId')
    expectArrayMatch(getAllowedAuthStrategies(SandboxController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(SandboxController, methodName), [ProxyAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxController, methodName), [SandboxAccessGuard])
  })
})
