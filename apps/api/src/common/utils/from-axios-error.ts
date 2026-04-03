/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

const secretPatterns = [
  /(api[_-]?key|access[_-]?key|secret[_-]?key|secret|token|password|passwd|authorization)[:=]\s*([^\s,;]+)/gi,
  /(bearer\s+)[A-Za-z0-9._\-~+/=]+/gi,
  /(AKIA[0-9A-Z]{16})/g,
  /(gh[pousr]_[A-Za-z0-9]{20,})/g,
  /(sk_live_[A-Za-z0-9]{16,})/g,
]

export function fromAxiosError(error: any): Error {
  const message = error.response?.data?.message || error.response?.data || error.message || error
  return new Error(redactString(String(message)))
}

function redactString(input: string): string {
  let redacted = input
  for (const pattern of secretPatterns) {
    redacted = redacted.replace(pattern, '$1[REDACTED]')
  }
  return redacted
}
