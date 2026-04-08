/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/// <reference types="jest" />

import { ExecutionContext } from '@nestjs/common'
import { BaseAuthContext } from '../../common/interfaces/base-auth-context.interface'

interface MockRequest {
  user?: BaseAuthContext | Record<string, unknown>
  params?: Record<string, string>
  headers?: Record<string, string>
  query?: Record<string, string>
  authMetadata?: {
    isStrategyAllowed: (type: string) => boolean
  }
}

/**
 * Creates a mock ExecutionContext for testing guards.
 * @param options - Optional request properties to override.
 * @returns An object containing the ExecutionContext and request object.
 */
export function createMockExecutionContext(options?: Omit<MockRequest, 'authMetadata'>): {
  context: ExecutionContext
  request: MockRequest
} {
  const request: MockRequest = {
    user: options?.user,
    params: options?.params ?? {},
    headers: options?.headers ?? {},
    query: options?.query ?? {},
  }

  const handler = jest.fn()
  const classRef = jest.fn()

  const context = {
    switchToHttp: jest.fn().mockReturnValue({
      getRequest: jest.fn().mockReturnValue(request),
      getResponse: jest.fn().mockReturnValue({}),
      getNext: jest.fn(),
    }),
    getHandler: jest.fn().mockReturnValue(handler),
    getClass: jest.fn().mockReturnValue(classRef),
    getType: jest.fn().mockReturnValue('http'),
    getArgs: jest.fn().mockReturnValue([request, {}, jest.fn()]),
    getArgByIndex: jest.fn(),
    switchToRpc: jest.fn(),
    switchToWs: jest.fn(),
  } as unknown as ExecutionContext

  return { context, request }
}
