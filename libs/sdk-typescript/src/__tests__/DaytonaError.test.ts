// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import { AxiosError, AxiosHeaders } from 'axios'

import {
  createAxiosDaytonaError,
  createDaytonaError,
  DaytonaA11yUnavailableError,
  DaytonaApiKeyExpiredError,
  DaytonaAuthenticationError,
  DaytonaAuthorizationError,
  DaytonaBadGatewayError,
  DaytonaConflictError,
  DaytonaConnectionError,
  DaytonaConnectionTimeoutError,
  DaytonaError,
  DaytonaFileNotFoundError,
  DaytonaGitAuthFailedError,
  DaytonaGoneError,
  DaytonaInternalServerError,
  DaytonaNotFoundError,
  DaytonaRateLimitError,
  DaytonaRunnerUnreachableError,
  DaytonaSandboxNotFoundError,
  DaytonaServiceUnavailableError,
  DaytonaSessionEndedError,
  DaytonaTimeoutError,
  DaytonaUnprocessableEntityError,
  DaytonaValidationError,
  errorClassFromStatusCode,
} from '../errors/DaytonaError'

describe('DaytonaError construction', () => {
  it('constructs DaytonaError with properties', () => {
    const err = new DaytonaError('boom', 500, undefined, 'INTERNAL', 'DAYTONA_RUNNER')
    expect(err).toBeInstanceOf(Error)
    expect(err.name).toBe('DaytonaError')
    expect(err.message).toBe('boom')
    expect(err.statusCode).toBe(500)
    expect(err.code).toBe('INTERNAL')
    expect(err.source).toBe('DAYTONA_RUNNER')
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
})

describe('HTTP status code classification', () => {
  test.each([
    [400, DaytonaValidationError],
    [401, DaytonaAuthenticationError],
    [403, DaytonaAuthorizationError],
    [404, DaytonaNotFoundError],
    [408, DaytonaTimeoutError],
    [409, DaytonaConflictError],
    [410, DaytonaGoneError],
    [422, DaytonaUnprocessableEntityError],
    [429, DaytonaRateLimitError],
    [500, DaytonaInternalServerError],
    [502, DaytonaBadGatewayError],
    [503, DaytonaServiceUnavailableError],
    [504, DaytonaTimeoutError],
  ])('maps status %s to its typed class', (statusCode, ErrCtor) => {
    expect(errorClassFromStatusCode(statusCode)).toBe(ErrCtor)
  })

  it('falls back to DaytonaError for unknown status codes', () => {
    expect(errorClassFromStatusCode(418)).toBe(DaytonaError)
    expect(errorClassFromStatusCode(undefined)).toBe(DaytonaError)
  })
})

describe('Domain code classification with status-class inheritance', () => {
  it('daemon GIT_AUTH_FAILED inherits from DaytonaAuthenticationError', () => {
    const err = createDaytonaError('git auth bad', 401, undefined, 'GIT_AUTH_FAILED', 'DAYTONA_DAEMON')
    expect(err).toBeInstanceOf(DaytonaGitAuthFailedError)
    expect(err).toBeInstanceOf(DaytonaAuthenticationError)
  })

  it('daemon FILE_NOT_FOUND inherits from DaytonaNotFoundError', () => {
    const err = createDaytonaError('missing', 404, undefined, 'FILE_NOT_FOUND', 'DAYTONA_DAEMON')
    expect(err).toBeInstanceOf(DaytonaFileNotFoundError)
    expect(err).toBeInstanceOf(DaytonaNotFoundError)
    expect(err.code).toBe('FILE_NOT_FOUND')
  })

  it('daemon SESSION_ENDED inherits from DaytonaGoneError', () => {
    const err = createDaytonaError('session ended', 410, undefined, 'SESSION_ENDED', 'DAYTONA_DAEMON')
    expect(err).toBeInstanceOf(DaytonaSessionEndedError)
    expect(err).toBeInstanceOf(DaytonaGoneError)
  })

  it('daemon A11Y_UNAVAILABLE inherits from DaytonaServiceUnavailableError', () => {
    const err = createDaytonaError('a11y bus down', 503, undefined, 'A11Y_UNAVAILABLE', 'DAYTONA_DAEMON')
    expect(err).toBeInstanceOf(DaytonaA11yUnavailableError)
    expect(err).toBeInstanceOf(DaytonaServiceUnavailableError)
  })

  it('proxy SANDBOX_NOT_FOUND inherits from DaytonaNotFoundError', () => {
    const err = createDaytonaError('sandbox gone', 404, undefined, 'SANDBOX_NOT_FOUND', 'DAYTONA_PROXY')
    expect(err).toBeInstanceOf(DaytonaSandboxNotFoundError)
    expect(err).toBeInstanceOf(DaytonaNotFoundError)
  })

  it('proxy RUNNER_UNREACHABLE inherits from DaytonaBadGatewayError', () => {
    const err = createDaytonaError('dial tcp: refused', 502, undefined, 'RUNNER_UNREACHABLE', 'DAYTONA_PROXY')
    expect(err).toBeInstanceOf(DaytonaRunnerUnreachableError)
    expect(err).toBeInstanceOf(DaytonaBadGatewayError)
  })

  it('api API_KEY_EXPIRED inherits from DaytonaAuthenticationError', () => {
    const err = createDaytonaError('key expired', 401, undefined, 'API_KEY_EXPIRED', 'DAYTONA_API')
    expect(err).toBeInstanceOf(DaytonaApiKeyExpiredError)
    expect(err).toBeInstanceOf(DaytonaAuthenticationError)
  })

  it('falls back to status class when (source, code) is unknown', () => {
    const err = createDaytonaError('mystery 404', 404, undefined, 'UNKNOWN_CODE', 'DAYTONA_DAEMON')
    expect(err).toBeInstanceOf(DaytonaNotFoundError)
    expect(err).not.toBeInstanceOf(DaytonaFileNotFoundError)
  })

  it('code without source falls back to status class', () => {
    const err = createDaytonaError('no source', 401, undefined, 'GIT_AUTH_FAILED')
    expect(err).toBeInstanceOf(DaytonaAuthenticationError)
    expect(err).not.toBeInstanceOf(DaytonaGitAuthFailedError)
  })
})

describe('Axios error mapping', () => {
  it('classifies Axios timeouts as DaytonaConnectionTimeoutError', () => {
    const error = new AxiosError('timeout of 1000ms exceeded', 'ECONNABORTED')

    const daytonaError = createAxiosDaytonaError(error)

    expect(daytonaError).toBeInstanceOf(DaytonaConnectionTimeoutError)
    expect(daytonaError).toBeInstanceOf(DaytonaConnectionError)
    expect(daytonaError.message).toBe('Operation timed out')
  })

  it('classifies network failures without a response as DaytonaConnectionError', () => {
    const error = new AxiosError('connect ECONNREFUSED', 'ERR_NETWORK', undefined, {} as never)

    const daytonaError = createAxiosDaytonaError(error)

    expect(daytonaError).toBeInstanceOf(DaytonaConnectionError)
    expect(daytonaError).not.toBeInstanceOf(DaytonaConnectionTimeoutError)
  })

  it('maps HTTP status + domain code to the precise subclass', () => {
    const headers = new AxiosHeaders({ 'x-request-id': 'req_123' })
    const error = new AxiosError('Request failed with status code 404', 'ERR_BAD_REQUEST', undefined, {} as never, {
      config: { headers } as never,
      data: { message: 'missing file', code: 'FILE_NOT_FOUND', source: 'DAYTONA_DAEMON' },
      headers,
      status: 404,
      statusText: 'Not Found',
    })

    const daytonaError = createAxiosDaytonaError(error)

    expect(daytonaError).toBeInstanceOf(DaytonaFileNotFoundError)
    expect(daytonaError).toBeInstanceOf(DaytonaNotFoundError)
    expect(daytonaError.statusCode).toBe(404)
    expect(daytonaError.code).toBe('FILE_NOT_FOUND')
    expect(daytonaError.source).toBe('DAYTONA_DAEMON')
    expect(daytonaError.headers).toBe(headers)
  })

  it('falls back to status-code class when no domain code is present', () => {
    const error = new AxiosError('Not found', 'ERR_BAD_REQUEST', undefined, {} as never, {
      config: { headers: new AxiosHeaders() } as never,
      data: { message: 'missing thing' },
      headers: new AxiosHeaders(),
      status: 404,
      statusText: 'Not Found',
    })

    const daytonaError = createAxiosDaytonaError(error)

    expect(daytonaError).toBeInstanceOf(DaytonaNotFoundError)
    expect(daytonaError).not.toBeInstanceOf(DaytonaFileNotFoundError)
  })

  it('stringifies object payloads when mapping axios errors', () => {
    const error = new AxiosError('Request failed', 'ERR_BAD_REQUEST', undefined, {} as never, {
      config: { headers: new AxiosHeaders() } as never,
      data: { nested: { reason: 'bad request' } },
      headers: new AxiosHeaders(),
      status: 500,
      statusText: 'Server Error',
    })

    const daytonaError = createAxiosDaytonaError(error)

    expect(daytonaError).toBeInstanceOf(DaytonaInternalServerError)
    expect(daytonaError.message).toBe('{"nested":{"reason":"bad request"}}')
  })

  it('does not use the deprecated "error" field as a fallback code', () => {
    const error = new AxiosError('Request failed', 'ERR_BAD_REQUEST', undefined, {} as never, {
      config: { headers: new AxiosHeaders() } as never,
      data: { message: 'missing file', error: 'Not Found' },
      headers: new AxiosHeaders(),
      status: 404,
      statusText: 'Not Found',
    })

    const daytonaError = createAxiosDaytonaError(error)

    expect(daytonaError).toBeInstanceOf(DaytonaNotFoundError)
    expect(daytonaError.code).toBeUndefined()
  })

  it('creates a generic DaytonaError for unknown non-network failures', () => {
    const error = new AxiosError('unknown failure')

    const daytonaError = createAxiosDaytonaError(error)

    expect(daytonaError).toBeInstanceOf(DaytonaError)
    expect(daytonaError).not.toBeInstanceOf(DaytonaConnectionError)
  })

  it('preserves DaytonaAuthorizationError mapping for 403 responses', () => {
    const error = new AxiosError('forbidden', 'ERR_BAD_REQUEST', undefined, {} as never, {
      config: { headers: new AxiosHeaders() } as never,
      data: { message: 'forbidden' },
      headers: new AxiosHeaders(),
      status: 403,
      statusText: 'Forbidden',
    })

    const daytonaError = createAxiosDaytonaError(error)

    expect(daytonaError).toBeInstanceOf(DaytonaAuthorizationError)
  })
})
