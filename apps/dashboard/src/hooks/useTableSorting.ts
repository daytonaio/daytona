/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useState, useEffect } from 'react'
import { SortingState } from '@tanstack/react-table'

export const useTableSorting = (tableId: string) => {
  const [sorting, setSorting] = useState<SortingState>(() => {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem(`table-sorting-${tableId}`)
      return stored ? JSON.parse(stored) : []
    }
    return []
  })

  useEffect(() => {
    if (typeof window !== 'undefined') {
      localStorage.setItem(`table-sorting-${tableId}`, JSON.stringify(sorting))
    }
  }, [sorting, tableId])

  // Return as an array to match the useState pattern that TanStack Table expects
  return [sorting, setSorting] as const
}
