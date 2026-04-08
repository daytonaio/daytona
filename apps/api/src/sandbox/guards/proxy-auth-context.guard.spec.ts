/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ProxyAuthContextGuard } from './proxy-auth-context.guard'
import { InvalidAuthenticationContextException } from '../../common/exceptions/invalid-authentication-context.exception'
import {
  createMockHealthCheckAuthContext,
  createMockOrganizationAuthContext,
  createMockOtelCollectorAuthContext,
  createMockProxyAuthContext,
  createMockRegionProxyAuthContext,
  createMockRegionSshGatewayAuthContext,
  createMockRunnerAuthContext,
  createMockSshGatewayAuthContext,
  createMockUserAuthContext,
} from '../../test/helpers/auth-context.factory'
import { createMockExecutionContext } from '../../test/helpers/execution-context.factory'

describe('[AUTH] ProxyAuthContextGuard', () => {
  let guard: ProxyAuthContextGuard

  beforeEach(() => {
    guard = new ProxyAuthContextGuard()
  })

  it('allows RegionProxyAuthContext', async () => {
    const { context } = createMockExecutionContext({ user: createMockRegionProxyAuthContext() })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it('allows ProxyAuthContext', async () => {
    const { context } = createMockExecutionContext({ user: createMockProxyAuthContext() })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it.each([
    ['User', createMockUserAuthContext],
    ['Organization', createMockOrganizationAuthContext],
    ['Runner', createMockRunnerAuthContext],
    ['SshGateway', createMockSshGatewayAuthContext],
    ['RegionSshGateway', createMockRegionSshGatewayAuthContext],
    ['HealthCheck', createMockHealthCheckAuthContext],
    ['OtelCollector', createMockOtelCollectorAuthContext],
  ])('rejects %s', async (_name, factory) => {
    const { context } = createMockExecutionContext({ user: factory() })
    await expect(guard.canActivate(context)).rejects.toThrow(InvalidAuthenticationContextException)
  })
})
