/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * @module Errors
 */

import { AxiosHeaders } from 'axios'
import type { AxiosError } from 'axios'

export type ResponseHeaders = InstanceType<typeof AxiosHeaders>

/**
 * Wire-format `source` identifiers set by the translation layer when a
 * Daytona service stamps them on the wire envelope. `source = undefined`
 * means the response did not carry a structured envelope (treat as opaque).
 */
export const SOURCE_API = 'DAYTONA_API'
export const SOURCE_DAEMON = 'DAYTONA_DAEMON'
export const SOURCE_PROXY = 'DAYTONA_PROXY'

/**
 * Base error for Daytona SDK. `statusCode` and `code` are populated only
 * for errors translated from a server response. `source` is `undefined`
 * unless the caller (or the translation layer) sets it.
 */
export class DaytonaError extends Error {
  public statusCode?: number
  public code?: string
  public readonly source?: string
  public headers?: ResponseHeaders

  constructor(message: string, statusCode?: number, headers?: ResponseHeaders, code?: string, source?: string) {
    super(message)
    this.name = new.target.name
    this.statusCode = statusCode
    this.headers = headers
    this.code = code
    this.source = source
  }
}

// HTTP-status classes — one per status code Daytona services emit.
// Names follow HTTP terminology (BadRequest=400, Forbidden=403).

export class DaytonaBadRequestError extends DaytonaError {}
export class DaytonaAuthenticationError extends DaytonaError {}
export class DaytonaForbiddenError extends DaytonaError {}
export class DaytonaNotFoundError extends DaytonaError {}
export class DaytonaTimeoutError extends DaytonaError {}
export class DaytonaConflictError extends DaytonaError {}
export class DaytonaGoneError extends DaytonaError {}
export class DaytonaUnprocessableEntityError extends DaytonaError {}
export class DaytonaRateLimitError extends DaytonaError {}
export class DaytonaInternalServerError extends DaytonaError {}
export class DaytonaBadGatewayError extends DaytonaError {}
export class DaytonaServiceUnavailableError extends DaytonaError {}

/**
 * @deprecated Use {@link DaytonaBadRequestError} instead. Re-exported so
 * existing `catch (err) { if (err instanceof DaytonaValidationError) ... }`
 * blocks keep working.
 */
export const DaytonaValidationError = DaytonaBadRequestError
/** @deprecated Use {@link DaytonaBadRequestError} instead. */
export type DaytonaValidationError = DaytonaBadRequestError

/**
 * @deprecated Use {@link DaytonaForbiddenError} instead.
 */
export const DaytonaAuthorizationError = DaytonaForbiddenError
/** @deprecated Use {@link DaytonaForbiddenError} instead. */
export type DaytonaAuthorizationError = DaytonaForbiddenError

/** Network connection failure (can't connect or mid-request drop). */
export class DaytonaConnectionError extends DaytonaError {}
/** Transport-layer timeout (connect / read). Subclass of DaytonaConnectionError. */
export class DaytonaConnectionTimeoutError extends DaytonaConnectionError {}

// Domain-specific subclasses. Each inherits from the HTTP-status class that
// matches its server-side status, so callers can catch either level.

// --- Git (daemon) ---
export class DaytonaGitAuthFailedError extends DaytonaAuthenticationError {}
export class DaytonaGitRepoNotFoundError extends DaytonaNotFoundError {}
export class DaytonaGitBranchNotFoundError extends DaytonaNotFoundError {}
export class DaytonaGitBranchExistsError extends DaytonaConflictError {}
export class DaytonaGitPushRejectedError extends DaytonaConflictError {}
export class DaytonaGitDirtyWorktreeError extends DaytonaConflictError {}
export class DaytonaGitMergeConflictError extends DaytonaConflictError {}

// --- Filesystem (daemon) ---
export class DaytonaFileNotFoundError extends DaytonaNotFoundError {}
export class DaytonaFileAccessDeniedError extends DaytonaForbiddenError {}

// --- LSP (daemon) ---
export class DaytonaLspServerNotInitializedError extends DaytonaBadRequestError {}

// --- Process / session (daemon) ---
export class DaytonaProcessExecutionTimeoutError extends DaytonaTimeoutError {}
export class DaytonaProcessNotFoundError extends DaytonaNotFoundError {}
export class DaytonaSessionEndedError extends DaytonaGoneError {}
export class DaytonaCommandAlreadyCompletedError extends DaytonaGoneError {}

// --- Computer-use (daemon) ---
export class DaytonaA11yUnavailableError extends DaytonaServiceUnavailableError {}
export class DaytonaRecordingStillActiveError extends DaytonaConflictError {}
export class DaytonaRecordingFfmpegNotFoundError extends DaytonaServiceUnavailableError {}

// --- Proxy ---
export class DaytonaSandboxNotFoundError extends DaytonaNotFoundError {}
export class DaytonaSandboxNotStartedError extends DaytonaBadRequestError {}
export class DaytonaRunnerUnreachableError extends DaytonaBadGatewayError {}

// --- API ---
export class DaytonaApiKeyExpiredError extends DaytonaAuthenticationError {}
export class DaytonaOrganizationSuspendedError extends DaytonaForbiddenError {}
export class DaytonaSandboxDiskExpansionLimitError extends DaytonaForbiddenError {}
export class DaytonaSandboxRunnerNotFoundError extends DaytonaNotFoundError {}
export class DaytonaSandboxStateChangeInProgressError extends DaytonaConflictError {}
export class DaytonaVolumeInUseError extends DaytonaConflictError {}
export class DaytonaDefaultRegionRequiredError extends DaytonaBadRequestError {}
export class DaytonaNoAvailableRunnersError extends DaytonaBadRequestError {}
export class DaytonaOrganizationQuotaExceededError extends DaytonaBadRequestError {}
export class DaytonaSandboxBackupStateError extends DaytonaBadRequestError {}
export class DaytonaSandboxOperationNotSupportedError extends DaytonaBadRequestError {}
export class DaytonaSandboxStateError extends DaytonaBadRequestError {}
export class DaytonaSnapshotStateChangeInProgressError extends DaytonaBadRequestError {}

/**
 * (source, code) → precise DaytonaError subclass. Lookup runs before the
 * HTTP status code fallback, so a domain code wins over the status default.
 *
 * Code strings are kept inline (not imported from the generated clients) so
 * tests that virtual-mock the API client modules don't break module init.
 * Drift is caught by the cross-language code-catalog generator + CI checks.
 */
const CODE_TO_ERROR_CLASS: Record<string, typeof DaytonaError> = {
  // Daemon
  'DAYTONA_DAEMON|GIT_AUTH_FAILED': DaytonaGitAuthFailedError,
  'DAYTONA_DAEMON|GIT_REPO_NOT_FOUND': DaytonaGitRepoNotFoundError,
  'DAYTONA_DAEMON|GIT_BRANCH_NOT_FOUND': DaytonaGitBranchNotFoundError,
  'DAYTONA_DAEMON|GIT_BRANCH_EXISTS': DaytonaGitBranchExistsError,
  'DAYTONA_DAEMON|GIT_PUSH_REJECTED': DaytonaGitPushRejectedError,
  'DAYTONA_DAEMON|GIT_DIRTY_WORKTREE': DaytonaGitDirtyWorktreeError,
  'DAYTONA_DAEMON|GIT_MERGE_CONFLICT': DaytonaGitMergeConflictError,
  'DAYTONA_DAEMON|FILE_NOT_FOUND': DaytonaFileNotFoundError,
  'DAYTONA_DAEMON|FILE_ACCESS_DENIED': DaytonaFileAccessDeniedError,
  'DAYTONA_DAEMON|LSP_SERVER_NOT_INITIALIZED': DaytonaLspServerNotInitializedError,
  'DAYTONA_DAEMON|PROCESS_EXECUTION_TIMEOUT': DaytonaProcessExecutionTimeoutError,
  'DAYTONA_DAEMON|PROCESS_NOT_FOUND': DaytonaProcessNotFoundError,
  'DAYTONA_DAEMON|SESSION_ENDED': DaytonaSessionEndedError,
  'DAYTONA_DAEMON|COMMAND_ALREADY_COMPLETED': DaytonaCommandAlreadyCompletedError,
  'DAYTONA_DAEMON|A11Y_UNAVAILABLE': DaytonaA11yUnavailableError,
  'DAYTONA_DAEMON|RECORDING_STILL_ACTIVE': DaytonaRecordingStillActiveError,
  'DAYTONA_DAEMON|RECORDING_FFMPEG_NOT_FOUND': DaytonaRecordingFfmpegNotFoundError,
  // Proxy
  'DAYTONA_PROXY|SANDBOX_NOT_FOUND': DaytonaSandboxNotFoundError,
  'DAYTONA_PROXY|SANDBOX_NOT_STARTED': DaytonaSandboxNotStartedError,
  'DAYTONA_PROXY|RUNNER_UNREACHABLE': DaytonaRunnerUnreachableError,
  // API
  'DAYTONA_API|API_KEY_EXPIRED': DaytonaApiKeyExpiredError,
  'DAYTONA_API|ORGANIZATION_SUSPENDED': DaytonaOrganizationSuspendedError,
  'DAYTONA_API|SANDBOX_DISK_EXPANSION_LIMIT': DaytonaSandboxDiskExpansionLimitError,
  'DAYTONA_API|SANDBOX_RUNNER_NOT_FOUND': DaytonaSandboxRunnerNotFoundError,
  'DAYTONA_API|SANDBOX_STATE_CHANGE_IN_PROGRESS': DaytonaSandboxStateChangeInProgressError,
  'DAYTONA_API|VOLUME_IN_USE': DaytonaVolumeInUseError,
  'DAYTONA_API|DEFAULT_REGION_REQUIRED': DaytonaDefaultRegionRequiredError,
  'DAYTONA_API|NO_AVAILABLE_RUNNERS': DaytonaNoAvailableRunnersError,
  'DAYTONA_API|ORGANIZATION_QUOTA_EXCEEDED': DaytonaOrganizationQuotaExceededError,
  'DAYTONA_API|SANDBOX_BACKUP_STATE_ERROR': DaytonaSandboxBackupStateError,
  'DAYTONA_API|SANDBOX_OPERATION_NOT_SUPPORTED': DaytonaSandboxOperationNotSupportedError,
  'DAYTONA_API|SANDBOX_STATE_ERROR': DaytonaSandboxStateError,
  'DAYTONA_API|SNAPSHOT_STATE_CHANGE_IN_PROGRESS': DaytonaSnapshotStateChangeInProgressError,
}

function lookupErrorClass(source: string | undefined, code: string | undefined): typeof DaytonaError | undefined {
  if (!code || !source) return undefined
  return CODE_TO_ERROR_CLASS[`${source}|${code}`]
}

const STATUS_CODE_TO_ERROR: Record<number, typeof DaytonaError> = {
  400: DaytonaBadRequestError,
  401: DaytonaAuthenticationError,
  403: DaytonaForbiddenError,
  404: DaytonaNotFoundError,
  408: DaytonaTimeoutError,
  409: DaytonaConflictError,
  410: DaytonaGoneError,
  422: DaytonaUnprocessableEntityError,
  429: DaytonaRateLimitError,
  500: DaytonaInternalServerError,
  502: DaytonaBadGatewayError,
  503: DaytonaServiceUnavailableError,
  504: DaytonaTimeoutError,
}

/** Maps an HTTP status code to the corresponding Daytona error class. */
export function errorClassFromStatusCode(statusCode?: number): typeof DaytonaError {
  if (statusCode === undefined) {
    return DaytonaError
  }

  return STATUS_CODE_TO_ERROR[statusCode] || DaytonaError
}

/**
 * Creates the appropriate Daytona error subclass from structured error metadata.
 *
 * Resolution order: (source, code) override -> HTTP status code -> DaytonaError.
 */
export function createDaytonaError(
  message: string,
  statusCode?: number,
  headers?: ResponseHeaders,
  code?: string,
  source?: string,
): DaytonaError {
  const ErrorClass = lookupErrorClass(source, code) || errorClassFromStatusCode(statusCode)
  return new ErrorClass(message, statusCode, headers, code, source)
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
  return typeof responseData?.code === 'string' ? responseData.code : undefined
}

function extractAxiosErrorSource(responseData?: Record<string, unknown>): string | undefined {
  return typeof responseData?.source === 'string' ? responseData.source : undefined
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
 * Creates the appropriate Daytona error subclass from an Axios error. Maps
 * client-side timeouts to DaytonaConnectionTimeoutError, networking failures
 * (no response received) to DaytonaConnectionError, and HTTP responses to
 * the most specific subclass via `createDaytonaError`.
 */
export function createAxiosDaytonaError(error: AxiosError): DaytonaError {
  const message = extractAxiosErrorMessage(error)
  const statusCode = error.response?.status
  const headers = error.response?.headers as ResponseHeaders | undefined
  const responseData = getAxiosResponseDataObject(error)
  const code = extractAxiosErrorCode(responseData)
  const source = extractAxiosErrorSource(responseData)

  if (isAxiosTimeoutError(error)) {
    return new DaytonaConnectionTimeoutError(message, statusCode, headers, code, source)
  }

  if (!error.response && (error.request || error.code)) {
    return new DaytonaConnectionError(message, statusCode, headers, code, source)
  }

  return createDaytonaError(message, statusCode, headers, code, source)
}
