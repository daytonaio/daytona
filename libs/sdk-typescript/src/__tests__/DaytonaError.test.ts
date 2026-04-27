// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import {
  createDaytonaError,
  DaytonaAuthenticationError,
  DaytonaAuthorizationError,
  DaytonaConflictError,
  DaytonaError,
  DaytonaNotFoundError,
  DaytonaRateLimitError,
  DaytonaTimeoutError,
  DaytonaValidationError,
  errorClassFromStatusCode,
} from '../errors/DaytonaError'

describe('Daytona errors', () => {
  it('constructs DaytonaError with properties', () => {
    const err = new DaytonaError('boom', 500)
    expect(err).toBeInstanceOf(Error)
    expect(err.name).toBe('DaytonaError')
    expect(err.message).toBe('boom')
    expect(err.statusCode).toBe(500)
  })

  test.each([
    [DaytonaNotFoundError, 'DaytonaNotFoundError'],
    [DaytonaRateLimitError, 'DaytonaRateLimitError'],
    [DaytonaTimeoutError, 'DaytonaTimeoutError'],
  ])('constructs %s', (ErrCtor, expectedName) => {
    const err = new ErrCtor('x', 404)
    expect(err).toBeInstanceOf(DaytonaError)
    expect(err.name).toBe(expectedName)
    expect(err.statusCode).toBe(404)
  })

  test.each([
    [400, DaytonaValidationError],
    [401, DaytonaAuthenticationError],
    [403, DaytonaAuthorizationError],
    [404, DaytonaNotFoundError],
    [409, DaytonaConflictError],
    [429, DaytonaRateLimitError],
    [500, DaytonaError],
    [undefined, DaytonaError],
  ])('maps status %s to the correct error class', (statusCode, ErrCtor) => {
    expect(errorClassFromStatusCode(statusCode)).toBe(ErrCtor)
  })

  it('creates subclassed errors from structured metadata', () => {
    const err = createDaytonaError('missing', 404, undefined, 'FILE_NOT_FOUND')

    expect(err).toBeInstanceOf(DaytonaNotFoundError)
    expect(err.errorCode).toBe('FILE_NOT_FOUND')
    expect(err.message).toBe('missing')
  })
})
