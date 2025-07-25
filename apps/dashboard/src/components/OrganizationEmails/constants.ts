/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const ORGANIZATION_EMAILS_TABLE_CONSTANTS = {
  DEFAULT_PAGE_SIZE: 10,
  MAX_EMAIL_LENGTH: 254,
  EMAIL_REGEX: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
} as const

export const EMAIL_STATUS = {
  VERIFIED: 'verified',
  PENDING: 'pending',
} as const

export type EmailStatus = (typeof EMAIL_STATUS)[keyof typeof EMAIL_STATUS]
