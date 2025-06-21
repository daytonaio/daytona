/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SortingState } from '@tanstack/react-table'
import { useState, useEffect } from 'react'

const STORAGE_KEY_PREFIX = 'table_sort_'

export function useTableSorting(tableId: string, defaultSorting: SortingState = []) {
  const storageKey = `${STORAGE_KEY_PREFIX}${tableId}`

  const [sorting, setSorting] = useState<SortingState>(() => {
    try {
      const saved = localStorage.getItem(storageKey)
      if (saved) {
        return JSON.parse(saved)
      }
    } catch {
      // Ignore parse error, fallback to default
    }
    return defaultSorting
  })

  useEffect(() => {
    try {
      localStorage.setItem(storageKey, JSON.stringify(sorting))
    } catch {
      // Ignore storage error
    }
  }, [sorting, storageKey])

  return [sorting, setSorting] as const
} 