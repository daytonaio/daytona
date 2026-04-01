/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * @module Errors
 */

import type { AxiosHeaders } from 'axios'

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

/**
 * Error thrown for malformed requests or invalid parameters (HTTP 400).
 *
 * @example
 * try {
 *   await daytona.create(params)
 * } catch (e) {
 *   if (e instanceof DaytonaBadRequestError) {
 *     console.error('Invalid request parameters:', e.message)
 *   }
 * }
 */
export class DaytonaBadRequestError extends DaytonaError {
  constructor(message: string, statusCode?: number, headers?: ResponseHeaders) {
    super(message, statusCode, headers)
    this.name = 'DaytonaBadRequestError'
  }
}

/**
 * Error thrown when API credentials are missing or invalid (HTTP 401).
 *
 * @example
 * try {
 *   const daytona = new Daytona({ apiKey: 'invalid' })
 *   await daytona.create()
 * } catch (e) {
 *   if (e instanceof DaytonaAuthenticationError) {
 *     console.error('Invalid or missing API key')
 *   }
 * }
 */
export class DaytonaAuthenticationError extends DaytonaError {
  constructor(message: string, statusCode?: number, headers?: ResponseHeaders) {
    super(message, statusCode, headers)
    this.name = 'DaytonaAuthenticationError'
  }
}

/**
 * Error thrown when the authenticated user lacks permission (HTTP 403).
 *
 * @example
 * try {
 *   await daytona.sandbox.delete(sandboxId)
 * } catch (e) {
 *   if (e instanceof DaytonaForbiddenError) {
 *     console.error('Not authorized to delete this sandbox')
 *   }
 * }
 */
export class DaytonaForbiddenError extends DaytonaError {
  constructor(message: string, statusCode?: number, headers?: ResponseHeaders) {
    super(message, statusCode, headers)
    this.name = 'DaytonaForbiddenError'
  }
}

/**
 * Error thrown when a requested resource is not found (HTTP 404).
 *
 * @example
 * try {
 *   const sandbox = await daytona.get('nonexistent-id')
 * } catch (e) {
 *   if (e instanceof DaytonaNotFoundError) {
 *     console.error('Sandbox does not exist')
 *   }
 * }
 */
export class DaytonaNotFoundError extends DaytonaError {
  constructor(message: string, statusCode?: number, headers?: ResponseHeaders) {
    super(message, statusCode, headers)
    this.name = 'DaytonaNotFoundError'
  }
}

/**
 * Error thrown when an operation conflicts with existing state (HTTP 409).
 *
 * Raised when creating a resource with a name that already exists,
 * or performing an operation that is incompatible with the current resource state.
 *
 * @example
 * try {
 *   await daytona.snapshot.create({ name: 'my-snapshot' })
 * } catch (e) {
 *   if (e instanceof DaytonaConflictError) {
 *     console.error('A snapshot with this name already exists')
 *   }
 * }
 */
export class DaytonaConflictError extends DaytonaError {
  constructor(message: string, statusCode?: number, headers?: ResponseHeaders) {
    super(message, statusCode, headers)
    this.name = 'DaytonaConflictError'
  }
}

/**
 * Error thrown for semantic validation failures (HTTP 422).
 *
 * Raised when the request is well-formed but values fail business logic
 * validation (e.g., unsupported resource class, invalid configuration).
 *
 * @example
 * try {
 *   await daytona.create({ resources: { ... } })
 * } catch (e) {
 *   if (e instanceof DaytonaValidationError) {
 *     console.error('Validation failed:', e.message)
 *   }
 * }
 */
export class DaytonaValidationError extends DaytonaError {
  constructor(message: string, statusCode?: number, headers?: ResponseHeaders) {
    super(message, statusCode, headers)
    this.name = 'DaytonaValidationError'
  }
}

/**
 * Error thrown when rate limit is exceeded (HTTP 429).
 *
 * @example
 * try {
 *   await daytona.create()
 * } catch (e) {
 *   if (e instanceof DaytonaRateLimitError) {
 *     console.error('Rate limit exceeded, back off and retry')
 *   }
 * }
 */
export class DaytonaRateLimitError extends DaytonaError {
  constructor(message: string, statusCode?: number, headers?: ResponseHeaders) {
    super(message, statusCode, headers)
    this.name = 'DaytonaRateLimitError'
  }
}

/**
 * Error thrown for unexpected server-side failures (HTTP 5xx).
 *
 * These are typically transient and safe to retry with exponential backoff.
 *
 * @example
 * try {
 *   await daytona.create()
 * } catch (e) {
 *   if (e instanceof DaytonaServerError) {
 *     console.error('Server error, retry later')
 *   }
 * }
 */
export class DaytonaServerError extends DaytonaError {
  constructor(message: string, statusCode?: number, headers?: ResponseHeaders) {
    super(message, statusCode, headers)
    this.name = 'DaytonaServerError'
  }
}

/**
 * Error thrown when a timeout occurs.
 *
 * Raised when a polling operation (e.g., waiting for a sandbox to start)
 * exceeds the configured timeout.
 *
 * @example
 * try {
 *   await sandbox.start(10)
 * } catch (e) {
 *   if (e instanceof DaytonaTimeoutError) {
 *     console.error('Sandbox did not start within 10 seconds')
 *   }
 * }
 */
export class DaytonaTimeoutError extends DaytonaError {
  constructor(message: string, statusCode?: number, headers?: ResponseHeaders) {
    super(message, statusCode, headers)
    this.name = 'DaytonaTimeoutError'
  }
}

/**
 * Error thrown for network-level connection failures.
 *
 * Raised when the SDK cannot reach the Daytona API due to network issues
 * (DNS failure, connection refused, TLS error, etc.) with no HTTP response.
 *
 * @example
 * try {
 *   await daytona.create()
 * } catch (e) {
 *   if (e instanceof DaytonaConnectionError) {
 *     console.error('Cannot reach Daytona API, check network connectivity')
 *   }
 * }
 */
export class DaytonaConnectionError extends DaytonaError {
  constructor(message: string, statusCode?: number, headers?: ResponseHeaders) {
    super(message, statusCode, headers)
    this.name = 'DaytonaConnectionError'
  }
}
