/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column } from '@tanstack/react-table'

export interface FilterOption {
  label: string
  value: string
  icon?: React.ComponentType<{ className?: string }>
}

export interface FilterValue {
  columnId: string
  value: string | string[]
}

export interface FilterProps<TData, TValue> {
  column: Column<TData, TValue>
  title: string
  options: FilterOption[]
  onFilterChange: (value: FilterValue) => void
  currentFilters: FilterValue[]
}

export interface FilterGroupProps {
  filters: FilterValue[]
  onClearFilter: (filter: FilterValue) => void
  onClearAll: () => void
}
