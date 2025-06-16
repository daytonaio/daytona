/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useEffect, useState } from 'react'
import { SortingState } from '@tanstack/react-table'

export function usePersistedTableSort(tableId: string, defaultSort?: SortingState) {
  // Load initial state from localStorage or use default
  const [sorting, setSorting] = useState<SortingState>(() => {
    if (typeof window === 'undefined') return defaultSort || []
    
    const saved = localStorage.getItem(`daytona-table-sort-${tableId}`)
    if (!saved) return defaultSort || []

    try {
      return JSON.parse(saved)
    } catch {
      return defaultSort || []
    }
  })

  // Save to localStorage whenever sorting changes
  useEffect(() => {
    if (typeof window === 'undefined') return
    
    localStorage.setItem(`daytona-table-sort-${tableId}`, JSON.stringify(sorting))
  }, [sorting, tableId])

  return [sorting, setSorting] as const
}
