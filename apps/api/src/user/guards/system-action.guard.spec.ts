/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SystemActionGuard } from './system-action.guard'
import { SystemRole } from '../enums/system-role.enum'
import { AccessDeniedException } from '../../common/exceptions/access-denied.exception'
import { createMockExecutionContext } from '../../test/helpers/execution-context.factory'
import {
  createMockAdminUserAuthContext,
  createMockUserAuthContext,
  createMockRunnerAuthContext,
  createMockProxyAuthContext,
  createMockSshGatewayAuthContext,
  createMockRegionProxyAuthContext,
  createMockRegionSshGatewayAuthContext,
  createMockHealthCheckAuthContext,
  createMockOtelCollectorAuthContext,
} from '../../test/helpers/auth-context.factory'
import { IS_PUBLIC_KEY } from '../../auth/decorators/public.decorator'
import { RequiredSystemRole } from '../decorators/required-system-role.decorator'
import { Reflector } from '@nestjs/core'

describe('[AUTH] SystemActionGuard', () => {
  let guard: SystemActionGuard
  let reflector: Reflector

  beforeEach(() => {
    reflector = new Reflector()
    guard = new SystemActionGuard(reflector)
  })

  it('should allow access when no @RequiredSystemRole is set', async () => {
    const { context } = createMockExecutionContext({ user: createMockUserAuthContext() })
    jest.spyOn(reflector, 'getAllAndOverride').mockReturnValue(false)
    jest.spyOn(reflector, 'get').mockReturnValue(undefined)

    const result = await guard.canActivate(context)
    expect(result).toBe(true)
  })

  it('should allow access for @Public() endpoints', async () => {
    const { context } = createMockExecutionContext()
    jest.spyOn(reflector, 'getAllAndOverride').mockImplementation((key) => {
      if (key === IS_PUBLIC_KEY) return true
      return undefined
    })

    const result = await guard.canActivate(context)
    expect(result).toBe(true)
  })

  it('should allow admin when @RequiredSystemRole(ADMIN) is set', async () => {
    const { context } = createMockExecutionContext({ user: createMockAdminUserAuthContext() })
    jest.spyOn(reflector, 'getAllAndOverride').mockImplementation((key) => {
      if (key === IS_PUBLIC_KEY) return false
      return undefined
    })
    jest.spyOn(reflector, 'get').mockImplementation((key) => {
      if (key === RequiredSystemRole) return SystemRole.ADMIN
      return undefined
    })

    const result = await guard.canActivate(context)
    expect(result).toBe(true)
  })

  it('should reject non-admin user when @RequiredSystemRole(ADMIN) is set', async () => {
    const { context } = createMockExecutionContext({ user: createMockUserAuthContext() })
    jest.spyOn(reflector, 'getAllAndOverride').mockImplementation((key) => {
      if (key === IS_PUBLIC_KEY) return false
      return undefined
    })
    jest.spyOn(reflector, 'get').mockImplementation((key) => {
      if (key === RequiredSystemRole) return SystemRole.ADMIN
      return undefined
    })

    await expect(guard.canActivate(context)).rejects.toThrow(AccessDeniedException)
  })

  it('should handle array of required roles', async () => {
    const { context } = createMockExecutionContext({ user: createMockUserAuthContext() })
    jest.spyOn(reflector, 'getAllAndOverride').mockImplementation((key) => {
      if (key === IS_PUBLIC_KEY) return false
      return undefined
    })
    jest.spyOn(reflector, 'get').mockImplementation((key) => {
      if (key === RequiredSystemRole) return [SystemRole.ADMIN, SystemRole.USER]
      return undefined
    })

    const result = await guard.canActivate(context)
    expect(result).toBe(true)
  })

  it.each([
    ['Runner', createMockRunnerAuthContext],
    ['Proxy', createMockProxyAuthContext],
    ['SshGateway', createMockSshGatewayAuthContext],
    ['RegionProxy', createMockRegionProxyAuthContext],
    ['RegionSshGateway', createMockRegionSshGatewayAuthContext],
    ['HealthCheck', createMockHealthCheckAuthContext],
    ['OtelCollector', createMockOtelCollectorAuthContext],
  ])('rejects %sAuthContext', async (_name, factory) => {
    const { context } = createMockExecutionContext({ user: factory() })
    jest.spyOn(reflector, 'getAllAndOverride').mockImplementation((key) => {
      if (key === IS_PUBLIC_KEY) return false
      return undefined
    })
    jest.spyOn(reflector, 'get').mockImplementation((key) => {
      if (key === RequiredSystemRole) return SystemRole.ADMIN
      return undefined
    })
    await expect(guard.canActivate(context)).rejects.toThrow(AccessDeniedException)
  })
})
