/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

const DEFAULT_MAX_LENGTH = 10_000
const TRUNCATION_SEPARATOR = '\n\n... [truncated] ...\n\n'

/**
 * Truncates a long error message keeping the first line (to preserve the
 * primary error) and as much of the tail as fits within {@link maxLength}.
 */
export function truncateErrorMessage(message: string, maxLength = DEFAULT_MAX_LENGTH): string {
  if (!message || message.length <= maxLength) {
    return message
  }

  const firstNewline = message.indexOf('\n')
  const firstLine = firstNewline === -1 ? message : message.slice(0, firstNewline)

  if (firstLine.length + TRUNCATION_SEPARATOR.length >= maxLength) {
    return firstLine.slice(0, maxLength)
  }

  const tailBudget = maxLength - firstLine.length - TRUNCATION_SEPARATOR.length
  const tail = message.slice(-tailBudget)

  return firstLine + TRUNCATION_SEPARATOR + tail
}
