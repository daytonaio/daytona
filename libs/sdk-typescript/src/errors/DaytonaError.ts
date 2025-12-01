/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { AxiosHeaders } from 'axios'

/**
 * @module Errors
 */

type ResponseHeaders = InstanceType<typeof AxiosHeaders>

/**
 * Base error for Daytona SDK.
 */
export class DaytonaError extends Error {
  /** HTTP status code if available */
  public statusCode?: number
  /** Response headers if available */
  public headers?: ResponseHeaders

  constructor(message: string, statusCode?: number, headers?: ResponseHeaders) {
    super(message)
    this.name = 'DaytonaError'
    this.statusCode = statusCode
    this.headers = headers
  }
}

export class DaytonaNotFoundError extends DaytonaError {
  constructor(message: string, statusCode?: number, headers?: ResponseHeaders) {
    super(message, statusCode, headers)
    this.name = 'DaytonaNotFoundError'
  }
}

/**
 * Error thrown when rate limit is exceeded.
 */
export class DaytonaRateLimitError extends DaytonaError {
  constructor(message: string, statusCode?: number, headers?: ResponseHeaders) {
    super(message, statusCode, headers)
    this.name = 'DaytonaRateLimitError'
  }
}

/**
 * Error thrown when a timeout occurs.
 */
export class DaytonaTimeoutError extends DaytonaError {
  constructor(message: string, statusCode?: number, headers?: Record<string, string>) {
    super(message, statusCode, headers)
    this.name = 'DaytonaTimeoutError'
  }
}
