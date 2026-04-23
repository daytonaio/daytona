/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { AxiosError, AxiosHeaders } from 'axios'
import {
  createDaytonaError,
  createAxiosDaytonaError,
  DaytonaConnectionError,
  DaytonaError,
  DaytonaNotFoundError,
  DaytonaTimeoutError,
} from '../errors/DaytonaError'

describe('Daytona error mapping', () => {
  it('classifies Axios timeouts before generic network failures', () => {
    const error = new AxiosError('timeout of 1000ms exceeded', 'ECONNABORTED')

    const daytonaError = createAxiosDaytonaError(error)

    expect(daytonaError).toBeInstanceOf(DaytonaTimeoutError)
    expect(daytonaError.message).toBe('Operation timed out')
  })

  it('classifies Axios connection failures without a response', () => {
    const error = new AxiosError('connect ECONNREFUSED', 'ERR_NETWORK', undefined, {} as never)

    const daytonaError = createAxiosDaytonaError(error)

    expect(daytonaError).toBeInstanceOf(DaytonaConnectionError)
  })

  it('maps HTTP status codes and structured error codes from Axios responses', () => {
    const headers = new AxiosHeaders({ 'x-request-id': 'req_123' })
    const error = new AxiosError('Request failed with status code 404', 'ERR_BAD_REQUEST', undefined, {} as never, {
      config: { headers } as never,
      data: {
        message: 'missing file',
        code: 'FILE_NOT_FOUND',
      },
      headers,
      status: 404,
      statusText: 'Not Found',
    })

    const daytonaError = createAxiosDaytonaError(error)

    expect(daytonaError).toBeInstanceOf(DaytonaNotFoundError)
    expect(daytonaError.statusCode).toBe(404)
    expect(daytonaError.errorCode).toBe('FILE_NOT_FOUND')
    expect(daytonaError.headers).toBe(headers)
  })

  it('extracts alternative structured error code fields', () => {
    const error = new AxiosError('Request failed', 'ERR_BAD_REQUEST', undefined, {} as never, {
      config: { headers: new AxiosHeaders() } as never,
      data: { message: 'rate limited', error_code: 'RATE_LIMITED' },
      headers: new AxiosHeaders(),
      status: 429,
      statusText: 'Too Many Requests',
    })

    const daytonaError = createAxiosDaytonaError(error)

    expect(daytonaError.errorCode).toBe('RATE_LIMITED')
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

    expect(daytonaError).toBeInstanceOf(DaytonaError)
    expect(daytonaError.message).toBe('{"nested":{"reason":"bad request"}}')
  })

  it('creates generic DaytonaError for unknown non-network axios failures', () => {
    const error = new AxiosError('unknown failure')

    const daytonaError = createAxiosDaytonaError(error)

    expect(daytonaError).toBeInstanceOf(DaytonaError)
    expect(daytonaError).not.toBeInstanceOf(DaytonaConnectionError)
  })

  it('creates structured Daytona errors directly', () => {
    const error = createDaytonaError('conflict', 409, undefined, 'ALREADY_EXISTS')

    expect(error.message).toBe('conflict')
    expect(error.statusCode).toBe(409)
    expect(error.errorCode).toBe('ALREADY_EXISTS')
  })
})
