/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { parseAsArrayOf, parseAsInteger, parseAsIsoDateTime, parseAsString, parseAsStringLiteral } from 'nuqs'

export const TAB_VALUES = ['overview', 'logs', 'traces', 'metrics', 'terminal', 'vnc'] as const
export type TabValue = (typeof TAB_VALUES)[number]

export const SEVERITY_OPTIONS = ['DEBUG', 'INFO', 'WARN', 'ERROR'] as const

export const tabParser = parseAsStringLiteral(TAB_VALUES).withDefault('overview')

export const logsSearchParams = {
  logsPage: parseAsInteger.withDefault(1),
  search: parseAsString.withDefault(''),
  severity: parseAsArrayOf(parseAsStringLiteral(SEVERITY_OPTIONS)).withDefault([]),
}

export const tracesSearchParams = {
  tracesPage: parseAsInteger.withDefault(1),
}

export const timeRangeSearchParams = {
  from: parseAsIsoDateTime,
  to: parseAsIsoDateTime,
}
