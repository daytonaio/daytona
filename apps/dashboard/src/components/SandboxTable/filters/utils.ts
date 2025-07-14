/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Row } from '@tanstack/react-table'
import { ResourceFilterValue } from './ResourceFilter'

export function arrayIncludesFilter<TData>(row: Row<TData>, id: string, value: string[]): boolean {
  return value.includes(row.getValue(id) as string)
}

export function arrayIntersectionFilter<TData>(row: Row<TData>, id: string, value: string[]): boolean {
  const cellValues = row.getValue(id) as string[]
  return value.some((filterValue) => cellValues.includes(filterValue))
}

export function resourceRangeFilter<TData>(row: Row<TData>, value: ResourceFilterValue): boolean {
  if (!value) return true

  const { cpu, memory, disk } = row.original as any

  if (value.cpu) {
    if (value.cpu.min !== undefined && cpu < value.cpu.min) return false
    if (value.cpu.max !== undefined && cpu > value.cpu.max) return false
  }

  if (value.memory) {
    if (value.memory.min !== undefined && memory < value.memory.min) return false
    if (value.memory.max !== undefined && memory > value.memory.max) return false
  }

  if (value.disk) {
    if (value.disk.min !== undefined && disk < value.disk.min) return false
    if (value.disk.max !== undefined && disk > value.disk.max) return false
  }

  return true
}

export function dateRangeFilter<TData>(row: Row<TData>, id: string, value: Date[]): boolean {
  if (!value) return true

  const date = row.getValue(id) as Date
  const [start, end] = value

  if ((start || end) && !date) return false

  if (start && !end) {
    return date.getTime() >= start.getTime()
  } else if (!start && end) {
    return date.getTime() <= end.getTime()
  } else if (start && end) {
    return date.getTime() >= start.getTime() && date.getTime() <= end.getTime()
  } else return true
}
