/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationAuthContextGuard } from './organization-auth-context.guard'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { InvalidAuthenticationContextException } from '../../common/exceptions/invalid-authentication-context.exception'
import { AccessDeniedException } from '../../common/exceptions/access-denied.exception'
import {
  createMockUserAuthContext,
  createMockRunnerAuthContext,
  createMockProxyAuthContext,
  createMockSshGatewayAuthContext,
  createMockRegionProxyAuthContext,
  createMockRegionSshGatewayAuthContext,
  createMockHealthCheckAuthContext,
  createMockOtelCollectorAuthContext,
} from '../../test/helpers/auth-context.factory'
import { createMockOrganization, createMockOrganizationUser } from '../../test/helpers/entity.factory'
import { MOCK_ORGANIZATION_ID, MOCK_USER_ID } from '../../test/helpers/constants'
import { createMockExecutionContext } from '../../test/helpers/execution-context.factory'
import { OrganizationResourcePermission } from '../enums/organization-resource-permission.enum'
import { RequiredOrganizationMemberRole } from '../decorators/required-organization-member-role.decorator'
import { RequiredOrganizationResourcePermissions } from '../decorators/required-organization-resource-permissions.decorator'

describe('[AUTH] OrganizationAuthContextGuard', () => {
  let guard: OrganizationAuthContextGuard
  let mockRedis: any
  let mockOrgService: any
  let mockOrgUserService: any
  let mockReflector: any

  const testOrganization = createMockOrganization({ id: MOCK_ORGANIZATION_ID })
  const testMember = createMockOrganizationUser({
    userId: MOCK_USER_ID,
    organizationId: MOCK_ORGANIZATION_ID,
    role: OrganizationMemberRole.MEMBER,
    assignedRoles: [{ permissions: [OrganizationResourcePermission.WRITE_SANDBOXES] }] as any,
  })
  const testOwner = createMockOrganizationUser({
    userId: MOCK_USER_ID,
    organizationId: MOCK_ORGANIZATION_ID,
    role: OrganizationMemberRole.OWNER,
  })

  beforeEach(() => {
    mockRedis = {
      get: jest.fn().mockResolvedValue(null),
      set: jest.fn().mockResolvedValue('OK'),
    }
    mockOrgService = { findOne: jest.fn().mockResolvedValue(testOrganization) }
    mockOrgUserService = { findOne: jest.fn().mockResolvedValue(testMember) }
    mockReflector = { getAllAndOverride: jest.fn().mockReturnValue(undefined) }
    guard = new OrganizationAuthContextGuard(
      mockRedis as any,
      mockOrgService as any,
      mockOrgUserService as any,
      mockReflector as any,
    )
  })

  it.each([
    ['Runner', createMockRunnerAuthContext],
    ['Proxy', createMockProxyAuthContext],
    ['SshGateway', createMockSshGatewayAuthContext],
    ['RegionProxy', createMockRegionProxyAuthContext],
    ['RegionSshGateway', createMockRegionSshGatewayAuthContext],
    ['HealthCheck', createMockHealthCheckAuthContext],
    ['OtelCollector', createMockOtelCollectorAuthContext],
  ])('should reject %sAuthContext', async (_name, factory) => {
    const { context } = createMockExecutionContext({ user: factory() })
    await expect(guard.canActivate(context)).rejects.toThrow(InvalidAuthenticationContextException)
  })

  it('should reject when no organizationId in params or auth context', async () => {
    const user = createMockUserAuthContext()
    const { context } = createMockExecutionContext({ user, params: {} })
    await expect(guard.canActivate(context)).rejects.toThrow(InvalidAuthenticationContextException)
  })

  it('should reject when API key org does not match requested org', async () => {
    const user = createMockUserAuthContext({
      organizationId: 'different-org',
      apiKey: { organizationId: 'different-org', permissions: [] } as any,
    })
    const { context } = createMockExecutionContext({ user, params: { organizationId: MOCK_ORGANIZATION_ID } })
    await expect(guard.canActivate(context)).rejects.toThrow(InvalidAuthenticationContextException)
  })

  it('should reject when organization is not found', async () => {
    mockOrgService.findOne.mockResolvedValue(null)
    const user = createMockUserAuthContext({ organizationId: MOCK_ORGANIZATION_ID })
    const { context } = createMockExecutionContext({ user })
    await expect(guard.canActivate(context)).rejects.toThrow(InvalidAuthenticationContextException)
  })

  it('should reject when organization user is not found', async () => {
    mockOrgUserService.findOne.mockResolvedValue(null)
    const user = createMockUserAuthContext({ organizationId: MOCK_ORGANIZATION_ID })
    const { context } = createMockExecutionContext({ user })
    await expect(guard.canActivate(context)).rejects.toThrow(InvalidAuthenticationContextException)
  })

  it('should allow member with no role or permission requirements', async () => {
    const user = createMockUserAuthContext({ organizationId: MOCK_ORGANIZATION_ID })
    const { context } = createMockExecutionContext({ user })
    const result = await guard.canActivate(context)
    expect(result).toBe(true)
  })

  it('should reject member when @RequiredOrganizationMemberRole(OWNER) is set', async () => {
    mockReflector.getAllAndOverride.mockImplementation((key: any) => {
      if (key === RequiredOrganizationMemberRole) return OrganizationMemberRole.OWNER
      return undefined
    })
    const user = createMockUserAuthContext({ organizationId: MOCK_ORGANIZATION_ID })
    const { context } = createMockExecutionContext({ user })
    await expect(guard.canActivate(context)).rejects.toThrow(AccessDeniedException)
  })

  it('should allow owner when @RequiredOrganizationMemberRole(OWNER) is set', async () => {
    mockOrgUserService.findOne.mockResolvedValue(testOwner)
    mockReflector.getAllAndOverride.mockImplementation((key: any) => {
      if (key === RequiredOrganizationMemberRole) return OrganizationMemberRole.OWNER
      return undefined
    })
    const user = createMockUserAuthContext({ organizationId: MOCK_ORGANIZATION_ID })
    const { context } = createMockExecutionContext({ user })
    const result = await guard.canActivate(context)
    expect(result).toBe(true)
  })

  it('should allow owner without API key (full access bypass)', async () => {
    mockOrgUserService.findOne.mockResolvedValue(testOwner)
    mockReflector.getAllAndOverride.mockImplementation((key: any) => {
      if (key === RequiredOrganizationResourcePermissions) {
        return [OrganizationResourcePermission.WRITE_SANDBOXES]
      }
      return undefined
    })
    const user = createMockUserAuthContext({ organizationId: MOCK_ORGANIZATION_ID })
    const { context } = createMockExecutionContext({ user })
    const result = await guard.canActivate(context)
    expect(result).toBe(true)
  })

  it('should allow member with matching permissions', async () => {
    mockReflector.getAllAndOverride.mockImplementation((key: any) => {
      if (key === RequiredOrganizationResourcePermissions) {
        return [OrganizationResourcePermission.WRITE_SANDBOXES]
      }
      return undefined
    })
    const user = createMockUserAuthContext({ organizationId: MOCK_ORGANIZATION_ID })
    const { context } = createMockExecutionContext({ user })
    const result = await guard.canActivate(context)
    expect(result).toBe(true)
  })

  it('should reject member with missing permissions', async () => {
    mockReflector.getAllAndOverride.mockImplementation((key: any) => {
      if (key === RequiredOrganizationResourcePermissions) {
        return [OrganizationResourcePermission.DELETE_SANDBOXES]
      }
      return undefined
    })
    const user = createMockUserAuthContext({ organizationId: MOCK_ORGANIZATION_ID })
    const { context } = createMockExecutionContext({ user })
    await expect(guard.canActivate(context)).rejects.toThrow(AccessDeniedException)
  })

  it('should enrich request.user with organization data', async () => {
    const user = createMockUserAuthContext({ organizationId: MOCK_ORGANIZATION_ID })
    const { context, request } = createMockExecutionContext({
      user,
      params: { organizationId: MOCK_ORGANIZATION_ID },
    })
    await guard.canActivate(context)
    expect(request.user).toMatchObject({
      userId: MOCK_USER_ID,
      organizationId: MOCK_ORGANIZATION_ID,
      organization: testOrganization,
      organizationUser: testMember,
    })
  })
})
