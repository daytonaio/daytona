/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export enum SessionLanguage {
  PYTHON = 'python',
  TYPESCRIPT = 'typescript',
  JAVASCRIPT = 'javascript',
  BASH = 'bash',
}

export const SESSION_LANGUAGES: readonly SessionLanguage[] = [
  SessionLanguage.PYTHON,
  SessionLanguage.TYPESCRIPT,
  SessionLanguage.JAVASCRIPT,
  SessionLanguage.BASH,
] as const
