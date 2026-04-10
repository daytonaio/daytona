/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CanActivate, ExecutionContext } from '@nestjs/common'
import { ModuleRef } from '@nestjs/core'
import { OrGuard } from './or.guard'
import { InvalidAuthenticationContextException } from '../common/exceptions/invalid-authentication-context.exception'
import { createMockExecutionContext } from '../test/helpers/execution-context.factory'

describe('[AUTH] OrGuard', () => {
  function createTestGuard(guards: { canActivate: (ctx: ExecutionContext) => Promise<boolean> | boolean }[]) {
    const guardClasses = guards.map((_, i) => {
      class TestGuard implements CanActivate {
        canActivate = guards[i].canActivate
      }
      return TestGuard
    })

    const MixedGuard = OrGuard(guardClasses)
    const moduleRef = {
      get: jest.fn((GuardClass: unknown) => {
        const idx = guardClasses.indexOf(GuardClass as (typeof guardClasses)[number])
        return guards[idx]
      }),
    } as unknown as ModuleRef

    return new MixedGuard(moduleRef)
  }

  it('should succeed when first guard passes', async () => {
    const guard = createTestGuard([
      { canActivate: jest.fn().mockResolvedValue(true) },
      { canActivate: jest.fn().mockResolvedValue(false) },
    ])

    const { context } = createMockExecutionContext()
    const result = await guard.canActivate(context)
    expect(result).toBe(true)
  })

  it('should succeed when second guard passes after first fails', async () => {
    const guard = createTestGuard([
      { canActivate: jest.fn().mockRejectedValue(new Error('fail')) },
      { canActivate: jest.fn().mockResolvedValue(true) },
    ])

    const { context } = createMockExecutionContext()
    const result = await guard.canActivate(context)
    expect(result).toBe(true)
  })

  it('should throw InvalidAuthenticationContextException when all guards fail', async () => {
    const guard = createTestGuard([
      { canActivate: jest.fn().mockRejectedValue(new Error('fail1')) },
      { canActivate: jest.fn().mockRejectedValue(new Error('fail2')) },
    ])

    const { context } = createMockExecutionContext()
    await expect(guard.canActivate(context)).rejects.toThrow(InvalidAuthenticationContextException)
  })

  it('should throw InvalidAuthenticationContextException when all guards return false', async () => {
    const guard = createTestGuard([
      { canActivate: jest.fn().mockResolvedValue(false) },
      { canActivate: jest.fn().mockResolvedValue(false) },
    ])

    const { context } = createMockExecutionContext()
    await expect(guard.canActivate(context)).rejects.toThrow(InvalidAuthenticationContextException)
  })

  it('should try guards in order and stop at first success', async () => {
    const secondGuard = { canActivate: jest.fn().mockResolvedValue(true) }
    const thirdGuard = { canActivate: jest.fn().mockResolvedValue(true) }

    const guard = createTestGuard([{ canActivate: jest.fn().mockResolvedValue(true) }, secondGuard, thirdGuard])

    const { context } = createMockExecutionContext()
    await guard.canActivate(context)
    expect(secondGuard.canActivate).not.toHaveBeenCalled()
    expect(thirdGuard.canActivate).not.toHaveBeenCalled()
  })

  it('should handle empty guard list', async () => {
    const guard = createTestGuard([])
    const { context } = createMockExecutionContext()
    await expect(guard.canActivate(context)).rejects.toThrow(InvalidAuthenticationContextException)
  })
})
