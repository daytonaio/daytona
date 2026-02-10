/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * Joins grouped code snippet sections with consistent spacing.
 * Each non-empty section gets a `\n\n` prefix, producing a blank line between sections.
 */
export function joinGroupedSections(sections: string[]): string {
  const nonEmpty = sections.filter(Boolean)
  if (nonEmpty.length === 0) return ''
  return nonEmpty.map((section) => '\n\n' + section).join('')
}
