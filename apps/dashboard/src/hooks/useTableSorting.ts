/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useState, useEffect } from 'react'
import { SortingState } from '@tanstack/react-table'

export const useTableSorting = (tableId: string, initialState?: SortingState) => {
  const [sorting, setSorting] = useState<SortingState>(() => {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem(`table-sorting-${tableId}`)
      return stored ? JSON.parse(stored) : initialState || []
    }
    return initialState || []
  })

  useEffect(() => {
    if (typeof window !== 'undefined') {
      localStorage.setItem(`table-sorting-${tableId}`, JSON.stringify(sorting))
    }
  }, [sorting, tableId])

  return [sorting, setSorting] as const
}
