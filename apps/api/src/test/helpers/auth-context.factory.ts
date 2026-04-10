/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SystemRole } from '../../user/enums/system-role.enum'
import { OrganizationMemberRole } from '../../organization/enums/organization-member-role.enum'
import { UserAuthContext } from '../../common/interfaces/user-auth-context.interface'
import { OrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { RunnerAuthContext } from '../../common/interfaces/runner-auth-context.interface'
import { ProxyAuthContext } from '../../common/interfaces/proxy-auth-context.interface'
import { SshGatewayAuthContext } from '../../common/interfaces/ssh-gateway-auth-context.interface'
import { RegionAuthContext } from '../../common/interfaces/region-auth-context.interface'
import { RegionProxyAuthContext } from '../../common/interfaces/region-proxy-auth-context.interface'
import { RegionSSHGatewayAuthContext } from '../../common/interfaces/region-ssh-gateway-auth-context.interface'
import { HealthCheckAuthContext } from '../../common/interfaces/health-check-auth-context.interface'
import { OtelCollectorAuthContext } from '../../common/interfaces/otel-collector-auth-context.interface'
import { MOCK_USER_ID, MOCK_USER_EMAIL, MOCK_ORGANIZATION_ID, MOCK_RUNNER_ID, MOCK_REGION_ID } from './constants'
import { createMockOrganization, createMockOrganizationUser, createMockRunner } from './entity.factory'

export function createMockUserAuthContext(overrides?: Partial<Omit<UserAuthContext, 'role'>>): UserAuthContext {
  return {
    role: SystemRole.USER,
    userId: MOCK_USER_ID,
    email: MOCK_USER_EMAIL,
    ...overrides,
  }
}

export function createMockAdminUserAuthContext(overrides?: Partial<Omit<UserAuthContext, 'role'>>): UserAuthContext {
  return {
    role: SystemRole.ADMIN,
    userId: MOCK_USER_ID,
    email: MOCK_USER_EMAIL,
    ...overrides,
  }
}

export function createMockOrganizationAuthContext(
  overrides?: Partial<Omit<OrganizationAuthContext, 'role'>>,
): OrganizationAuthContext {
  return {
    role: SystemRole.USER,
    userId: MOCK_USER_ID,
    email: MOCK_USER_EMAIL,
    organizationId: MOCK_ORGANIZATION_ID,
    organization: createMockOrganization(),
    organizationUser: createMockOrganizationUser(),
    ...overrides,
  }
}

export function createMockOwnerOrganizationAuthContext(
  overrides?: Partial<Omit<OrganizationAuthContext, 'role' | 'organizationUser'>>,
): OrganizationAuthContext {
  return {
    role: SystemRole.USER,
    userId: MOCK_USER_ID,
    email: MOCK_USER_EMAIL,
    organizationId: MOCK_ORGANIZATION_ID,
    organization: createMockOrganization(),
    organizationUser: createMockOrganizationUser({ role: OrganizationMemberRole.OWNER }),
    ...overrides,
  }
}
export function createMockRunnerAuthContext(overrides?: Partial<Omit<RunnerAuthContext, 'role'>>): RunnerAuthContext {
  return {
    role: 'runner',
    runnerId: MOCK_RUNNER_ID,
    runner: createMockRunner(),
    ...overrides,
  }
}

export function createMockProxyAuthContext(overrides?: Partial<Omit<ProxyAuthContext, 'role'>>): ProxyAuthContext {
  return {
    role: 'proxy',
    ...overrides,
  }
}

export function createMockSshGatewayAuthContext(
  overrides?: Partial<Omit<SshGatewayAuthContext, 'role'>>,
): SshGatewayAuthContext {
  return {
    role: 'ssh-gateway',
    ...overrides,
  }
}

export function createMockRegionAuthContext(overrides?: Partial<Omit<RegionAuthContext, 'role'>>): RegionAuthContext {
  return {
    role: 'proxy',
    regionId: MOCK_REGION_ID,
    ...overrides,
  }
}

export function createMockRegionProxyAuthContext(
  overrides?: Partial<Omit<RegionProxyAuthContext, 'role'>>,
): RegionProxyAuthContext {
  return {
    role: 'region-proxy',
    regionId: MOCK_REGION_ID,
    ...overrides,
  }
}

export function createMockRegionSshGatewayAuthContext(
  overrides?: Partial<Omit<RegionSSHGatewayAuthContext, 'role'>>,
): RegionSSHGatewayAuthContext {
  return {
    role: 'region-ssh-gateway',
    regionId: MOCK_REGION_ID,
    ...overrides,
  }
}

export function createMockHealthCheckAuthContext(
  overrides?: Partial<Omit<HealthCheckAuthContext, 'role'>>,
): HealthCheckAuthContext {
  return {
    role: 'health-check',
    ...overrides,
  }
}

export function createMockOtelCollectorAuthContext(
  overrides?: Partial<Omit<OtelCollectorAuthContext, 'role'>>,
): OtelCollectorAuthContext {
  return {
    role: 'otel-collector',
    ...overrides,
  }
}
