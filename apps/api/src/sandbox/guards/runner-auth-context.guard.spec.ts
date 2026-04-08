/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RunnerAuthContextGuard } from './runner-auth-context.guard'
import { InvalidAuthenticationContextException } from '../../common/exceptions/invalid-authentication-context.exception'
import {
  createMockHealthCheckAuthContext,
  createMockOrganizationAuthContext,
  createMockOtelCollectorAuthContext,
  createMockOwnerOrganizationAuthContext,
  createMockProxyAuthContext,
  createMockRegionProxyAuthContext,
  createMockRegionSshGatewayAuthContext,
  createMockRunnerAuthContext,
  createMockSshGatewayAuthContext,
  createMockUserAuthContext,
} from '../../test/helpers/auth-context.factory'
import { createMockExecutionContext } from '../../test/helpers/execution-context.factory'

describe('[AUTH] RunnerAuthContextGuard', () => {
  let guard: RunnerAuthContextGuard

  beforeEach(() => {
    guard = new RunnerAuthContextGuard()
  })

  it('allows RunnerAuthContext', async () => {
    const { context } = createMockExecutionContext({ user: createMockRunnerAuthContext() })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it.each([
    ['User', createMockUserAuthContext],
    ['Organization', createMockOrganizationAuthContext],
    ['OwnerOrganization', createMockOwnerOrganizationAuthContext],
    ['Proxy', createMockProxyAuthContext],
    ['SshGateway', createMockSshGatewayAuthContext],
    ['RegionProxy', createMockRegionProxyAuthContext],
    ['RegionSshGateway', createMockRegionSshGatewayAuthContext],
    ['HealthCheck', createMockHealthCheckAuthContext],
    ['OtelCollector', createMockOtelCollectorAuthContext],
  ])('rejects %s', async (_name, factory) => {
    const { context } = createMockExecutionContext({ user: factory() })
    await expect(guard.canActivate(context)).rejects.toThrow(InvalidAuthenticationContextException)
  })
})
