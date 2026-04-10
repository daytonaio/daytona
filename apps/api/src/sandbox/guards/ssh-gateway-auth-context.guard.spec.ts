/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SshGatewayAuthContextGuard } from './ssh-gateway-auth-context.guard'
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

describe('[AUTH] SshGatewayAuthContextGuard', () => {
  let guard: SshGatewayAuthContextGuard

  beforeEach(() => {
    guard = new SshGatewayAuthContextGuard()
  })

  it('allows SshGatewayAuthContext', async () => {
    const { context } = createMockExecutionContext({ user: createMockSshGatewayAuthContext() })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it('allows RegionSshGatewayAuthContext', async () => {
    const { context } = createMockExecutionContext({ user: createMockRegionSshGatewayAuthContext() })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it.each([
    ['User', createMockUserAuthContext],
    ['Organization', createMockOrganizationAuthContext],
    ['Runner', createMockRunnerAuthContext],
    ['Proxy', createMockProxyAuthContext],
    ['RegionProxy', createMockRegionProxyAuthContext],
    ['HealthCheck', createMockHealthCheckAuthContext],
    ['OtelCollector', createMockOtelCollectorAuthContext],
  ])('rejects %s', async (_name, factory) => {
    const { context } = createMockExecutionContext({ user: factory() })
    await expect(guard.canActivate(context)).rejects.toThrow(InvalidAuthenticationContextException)
  })
})
