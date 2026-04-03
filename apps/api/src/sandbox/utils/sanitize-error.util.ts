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

export function sanitizeSandboxError(error: any): { recoverable: boolean; errorReason: string } {
  if (typeof error === 'string') {
    try {
      const errObj = JSON.parse(error) as { recoverable: boolean; errorReason: string }
      return { recoverable: errObj.recoverable, errorReason: redactString(errObj.errorReason) }
    } catch {
      return { recoverable: false, errorReason: redactString(error) }
    }
  } else if (typeof error === 'object' && error !== null && 'recoverable' in error && 'errorReason' in error) {
    return { recoverable: error.recoverable, errorReason: redactString(error.errorReason) }
  } else if (typeof error === 'object' && error.message) {
    return sanitizeSandboxError(error.message)
  }

  return { recoverable: false, errorReason: redactString(String(error)) }
}

function redactString(input: string): string {
  let redacted = input
  for (const pattern of secretPatterns) {
    redacted = redacted.replace(pattern, '$1[REDACTED]')
  }
  return redacted
}
