/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * Approach A — one enum per component.
 * All API error codes live here; the code field on every error response references this enum.
 * Generators emit a typed enum class in each language (ApiErrorCode in Python, Go, TS, Ruby, Java).
 */
export enum ApiErrorCode {
  // Generic
  BAD_REQUEST = 'BAD_REQUEST',
  UNAUTHORIZED = 'UNAUTHORIZED',
  FORBIDDEN = 'FORBIDDEN',
  NOT_FOUND = 'NOT_FOUND',
  CONFLICT = 'CONFLICT',
  RATE_LIMIT_EXCEEDED = 'RATE_LIMIT_EXCEEDED',
  INTERNAL_SERVER_ERROR = 'INTERNAL_SERVER_ERROR',

  // Snapshot
  SNAPSHOT_NOT_FOUND = 'SNAPSHOT_NOT_FOUND',
  SNAPSHOT_ACCESS_DENIED = 'SNAPSHOT_ACCESS_DENIED',

  // Sandbox
  SANDBOX_NOT_FOUND = 'SANDBOX_NOT_FOUND',
}

export const HTTP_STATUS_TO_API_CODE: Record<number, ApiErrorCode> = {
  400: ApiErrorCode.BAD_REQUEST,
  401: ApiErrorCode.UNAUTHORIZED,
  403: ApiErrorCode.FORBIDDEN,
  404: ApiErrorCode.NOT_FOUND,
  409: ApiErrorCode.CONFLICT,
  429: ApiErrorCode.RATE_LIMIT_EXCEEDED,
  500: ApiErrorCode.INTERNAL_SERVER_ERROR,
}
