/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { HttpException, HttpStatus, ServiceUnavailableException, UnauthorizedException } from '@nestjs/common'
import { Reflector } from '@nestjs/core'
import { GlobalAuthGuard } from './global-auth.guard'
import { AuthStrategyType } from './enums/auth-strategy-type.enum'
import { AuthStrategy } from './decorators/auth-strategy.decorator'
import { IS_PUBLIC_KEY } from './decorators/public.decorator'
import { createMockExecutionContext } from '../test/helpers/execution-context.factory'

describe('[AUTH] GlobalAuthGuard', () => {
  let guard: GlobalAuthGuard
  let reflector: Reflector

  beforeEach(() => {
    reflector = new Reflector()
    guard = new GlobalAuthGuard(reflector)
  })

  describe('[AUTH] canActivate', () => {
    it('should bypass authentication for @Public() endpoints', () => {
      const { context } = createMockExecutionContext()
      jest.spyOn(reflector, 'getAllAndOverride').mockImplementation((key) => {
        if (key === IS_PUBLIC_KEY) return true
        return undefined
      })

      const result = guard.canActivate(context)
      expect(result).toBe(true)
    })

    it('should set authMetadata with both JWT/API_KEY as allowed strategies on the request', () => {
      const { context, request } = createMockExecutionContext()
      jest.spyOn(reflector, 'getAllAndOverride').mockImplementation((key) => {
        if (key === IS_PUBLIC_KEY) return false
        if (key === AuthStrategy) return [AuthStrategyType.API_KEY, AuthStrategyType.JWT]
        return undefined
      })

      // Mocks canActivate on Passport's base class to return true, bypassing the actual JWT/API-key verification.
      // We're testing the guard's strategy resolution logic, not Passport's auth flow.
      jest.spyOn(Object.getPrototypeOf(Object.getPrototypeOf(guard)), 'canActivate').mockReturnValue(true)
      guard.canActivate(context)

      expect(request.authMetadata).toBeDefined()
      expect(request.authMetadata!.isStrategyAllowed(AuthStrategyType.API_KEY)).toBe(true)
      expect(request.authMetadata!.isStrategyAllowed(AuthStrategyType.JWT)).toBe(true)
    })

    it('should default to JWT-only when no @AuthStrategy() is present', () => {
      const { context, request } = createMockExecutionContext()
      jest.spyOn(reflector, 'getAllAndOverride').mockImplementation((key) => {
        if (key === IS_PUBLIC_KEY) return false
        if (key === AuthStrategy) return undefined
        return undefined
      })

      // Mocks canActivate on Passport's base class to return true, bypassing the actual JWT/API-key verification.
      // We're testing the guard's strategy resolution logic, not Passport's auth flow.
      jest.spyOn(Object.getPrototypeOf(Object.getPrototypeOf(guard)), 'canActivate').mockReturnValue(true)
      guard.canActivate(context)

      expect(request.authMetadata).toBeDefined()
      expect(request.authMetadata!.isStrategyAllowed(AuthStrategyType.JWT)).toBe(true)
      expect(request.authMetadata!.isStrategyAllowed(AuthStrategyType.API_KEY)).toBe(false)
    })

    it('should handle single strategy value (not array)', () => {
      const { context, request } = createMockExecutionContext()
      jest.spyOn(reflector, 'getAllAndOverride').mockImplementation((key, targets) => {
        if (key === IS_PUBLIC_KEY) return false
        if (key === AuthStrategy) return AuthStrategyType.API_KEY
        return undefined
      })

      // Mocks canActivate on Passport's base class to return true, bypassing the actual JWT/API-key verification.
      // We're testing the guard's strategy resolution logic, not Passport's auth flow.
      jest.spyOn(Object.getPrototypeOf(Object.getPrototypeOf(guard)), 'canActivate').mockReturnValue(true)
      guard.canActivate(context)

      expect(request.authMetadata!.isStrategyAllowed(AuthStrategyType.API_KEY)).toBe(true)
      expect(request.authMetadata!.isStrategyAllowed(AuthStrategyType.JWT)).toBe(false)
    })
  })

  describe('[AUTH] handleRequest', () => {
    it('should return user when authentication succeeds', () => {
      const user = { role: 'user', userId: 'test' }
      const result = guard.handleRequest(null, user, null)
      expect(result).toBe(user)
    })

    it('should throw UnauthorizedException when user is falsy', () => {
      expect(() => guard.handleRequest(null, null, null)).toThrow(UnauthorizedException)
    })

    it('should throw ServiceUnavailableException when error is present and not a HttpException', () => {
      expect(() => guard.handleRequest(new Error('fail'), null, null)).toThrow(ServiceUnavailableException)
    })

    it('should throw ServiceUnavailableException when error is present and a HttpException with status code >= 500', () => {
      expect(() => guard.handleRequest(new HttpException('fail', HttpStatus.GATEWAY_TIMEOUT), null, null)).toThrow(
        ServiceUnavailableException,
      )
    })
  })
})
