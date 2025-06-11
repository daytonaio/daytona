/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useState, useEffect, useContext } from 'react'
import { SortingState } from '@tanstack/react-table'
import { TableSortingContext } from '@/contexts/TableSortingContext'

export const useTableSorting = (tableId: string, initialState?: SortingState) => {
  const context = useContext(TableSortingContext)

  if (!context) {
    throw new Error('useTableSorting must be used within a TableSortingProvider')
  }

  const [sorting, setSorting] = useState<SortingState>(() => {
    // Try to get persisted state from context first
    return context.sortingStates[tableId] || initialState || []
  })

  useEffect(() => {
    // Update context when sorting changes using the proper method name
    context.updateSortingState(tableId, sorting)
  }, [sorting, tableId, context])

  return [sorting, setSorting] as const
}
