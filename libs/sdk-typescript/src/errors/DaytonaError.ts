/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * @module Errors
 */

import { AxiosError, AxiosHeaders } from 'axios'

export type ResponseHeaders = InstanceType<typeof AxiosHeaders>

/**
 * Base error for Daytona SDK.
 *
 * @example
 * ```ts
 * try {
 *   await daytona.get('missing-sandbox')
 * } catch (error) {
 *   if (error instanceof DaytonaError) {
 *     console.log(error.statusCode)
 *     console.log(error.errorCode)
 *     console.log(error.message)
 *   }
 * }
 * ```
 */
export class DaytonaError extends Error {
  /** HTTP status code if available */
  public statusCode?: number
  /** Machine-readable error code if available */
  public errorCode?: string
  /** Response headers if available */
  public headers?: ResponseHeaders

  constructor(message: string, statusCode?: number, headers?: ResponseHeaders, errorCode?: string) {
    super(message)
    this.name = new.target.name
    this.statusCode = statusCode
    this.headers = headers
    this.errorCode = errorCode
  }
}

/**
 * Error thrown when a resource is not found (HTTP 404).
 *
 * @example
 * ```ts
 * try {
 *   await sandbox.fs.downloadFile('/workspace/missing.txt')
 * } catch (error) {
 *   if (error instanceof DaytonaNotFoundError) {
 *     console.log(error.statusCode)
 *   }
 * }
 * ```
 */
export class DaytonaNotFoundError extends DaytonaError {
}

/**
 * Error thrown when rate limit is exceeded.
 *
 * @example
 * ```ts
 * try {
 *   await daytona.list()
 * } catch (error) {
 *   if (error instanceof DaytonaRateLimitError) {
 *     console.log(error.errorCode)
 *   }
 * }
 * ```
 */
export class DaytonaRateLimitError extends DaytonaError {
}

/**
 * Error thrown when authentication fails (HTTP 401).
 *
 * @example
 * ```ts
 * try {
 *   await daytona.list()
 * } catch (error) {
 *   if (error instanceof DaytonaAuthenticationError) {
 *     console.log(error.statusCode)
 *   }
 * }
 * ```
 */
export class DaytonaAuthenticationError extends DaytonaError {
}

/**
 * Error thrown when the request is forbidden (HTTP 403).
 *
 * @example
 * ```ts
 * try {
 *   await daytona.get('sandbox-without-access')
 * } catch (error) {
 *   if (error instanceof DaytonaAuthorizationError) {
 *     console.log(error.message)
 *   }
 * }
 * ```
 */
export class DaytonaAuthorizationError extends DaytonaError {
}

/**
 * Error thrown when a resource conflict occurs (HTTP 409).
 *
 * @example
 * ```ts
 * try {
 *   await daytona.create({ name: 'existing-sandbox' })
 * } catch (error) {
 *   if (error instanceof DaytonaConflictError) {
 *     console.log(error.errorCode)
 *   }
 * }
 * ```
 */
export class DaytonaConflictError extends DaytonaError {
}

/**
 * Error thrown when input validation fails (HTTP 400 or client-side validation).
 *
 * @example
 * ```ts
 * try {
 *   Image.debianSlim('3.8' as never)
 * } catch (error) {
 *   if (error instanceof DaytonaValidationError) {
 *     console.log(error.message)
 *   }
 * }
 * ```
 */
export class DaytonaValidationError extends DaytonaError {
}

/**
 * Error thrown when a timeout occurs.
 *
 * @example
 * ```ts
 * try {
 *   await sandbox.waitUntilStarted(1)
 * } catch (error) {
 *   if (error instanceof DaytonaTimeoutError) {
 *     console.log(error.message)
 *   }
 * }
 * ```
 */
export class DaytonaTimeoutError extends DaytonaError {
}

/**
 * Error thrown when a network connection fails.
 *
 * @example
 * ```ts
 * try {
 *   await ptyHandle.waitForConnection()
 * } catch (error) {
 *   if (error instanceof DaytonaConnectionError) {
 *     console.log(error.message)
 *   }
 * }
 * ```
 */
export class DaytonaConnectionError extends DaytonaError {
}

const STATUS_CODE_TO_ERROR: Record<number, typeof DaytonaError> = {
  400: DaytonaValidationError,
  401: DaytonaAuthenticationError,
  403: DaytonaAuthorizationError,
  404: DaytonaNotFoundError,
  409: DaytonaConflictError,
  429: DaytonaRateLimitError,
}

/**
 * Maps an HTTP status code to the corresponding Daytona error class.
 */
export function errorClassFromStatusCode(statusCode?: number): typeof DaytonaError {
  if (statusCode === undefined) {
    return DaytonaError
  }

  return STATUS_CODE_TO_ERROR[statusCode] || DaytonaError
}

/**
 * Creates the appropriate Daytona error subclass from structured error metadata.
 */
export function createDaytonaError(
  message: string,
  statusCode?: number,
  headers?: ResponseHeaders,
  errorCode?: string,
): DaytonaError {
  const ErrorClass = errorClassFromStatusCode(statusCode)
  return new ErrorClass(message, statusCode, headers, errorCode)
}

function isAxiosTimeoutError(error: AxiosError): boolean {
  return error.code === 'ECONNABORTED' || error.code === 'ETIMEDOUT' || error.message.includes('timeout of')
}

function getAxiosResponseDataObject(error: AxiosError): Record<string, unknown> | undefined {
  if (!error.response?.data || typeof error.response.data !== 'object') {
    return undefined
  }

  return error.response.data as Record<string, unknown>
}

function extractAxiosErrorCode(responseData?: Record<string, unknown>): string | undefined {
  if (typeof responseData?.code === 'string') {
    return responseData.code
  }

  if (typeof responseData?.error_code === 'string') {
    return responseData.error_code
  }

  if (typeof responseData?.error === 'string') {
    return responseData.error
  }

  return undefined
}

function extractAxiosErrorMessage(error: AxiosError): string {
  if (isAxiosTimeoutError(error)) {
    return 'Operation timed out'
  }

  const responseData = getAxiosResponseDataObject(error)
  const responseMessage: unknown = responseData?.message || error.response?.data
  const message: unknown = responseMessage || error.message || String(error)

  if (typeof message === 'object') {
    try {
      return JSON.stringify(message)
    } catch {
      return String(message)
    }
  }

  return String(message)
}

/**
 * Creates the appropriate Daytona error subclass from an Axios error.
 */
export function createAxiosDaytonaError(error: AxiosError): DaytonaError {
  const message = extractAxiosErrorMessage(error)
  const statusCode = error.response?.status
  const headers = error.response?.headers as ResponseHeaders | undefined
  const responseData = getAxiosResponseDataObject(error)
  const errorCode = extractAxiosErrorCode(responseData)

  if (isAxiosTimeoutError(error)) {
    return new DaytonaTimeoutError(message, statusCode, headers, errorCode)
  }

  if (!error.response && (error.request || error.code)) {
    return new DaytonaConnectionError(message, statusCode, headers, errorCode)
  }

  return createDaytonaError(message, statusCode, headers, errorCode)
}
