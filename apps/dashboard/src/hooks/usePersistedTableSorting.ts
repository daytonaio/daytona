/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useState, useCallback, useEffect } from 'react'
import { SortingState } from '@tanstack/react-table'
import { LocalStorageKey } from '@/enums/LocalStorageKey'
import { getLocalStorageItem, setLocalStorageItem } from '@/lib/local-storage'

interface UsePersistedTableSortingProps {
  /**
   * Unique identifier for the table (e.g., 'snapshot-table', 'volume-table')
   */
  tableId: string
  /**
   * Default sorting state to use if no persisted state exists
   */
  defaultSorting?: SortingState
}

interface UsePersistedTableSortingReturn {
  /**
   * Current sorting state
   */
  sorting: SortingState
  /**
   * Function to update sorting state (to be passed to onSortingChange)
   */
  setSorting: (sorting: SortingState | ((prev: SortingState) => SortingState)) => void
  /**
   * Function to clear persisted sorting state and reset to default
   */
  clearPersistedSorting: () => void
}

export type { UsePersistedTableSortingReturn }

/**
 * Custom hook to manage persistent table sorting state using localStorage
 *
 * @param tableId - Unique identifier for the table
 * @param defaultSorting - Default sorting state to use if no persisted state exists
 * @returns Object containing sorting state and setter functions
 *
 * @example
 * ```tsx
 * const { sorting, setSorting, clearPersistedSorting } = usePersistedTableSorting({
 *   tableId: 'snapshot-table',
 *   defaultSorting: [{ id: 'createdAt', desc: true }]
 * })
 *
 * const table = useReactTable({
 *   // ... other config
 *   onSortingChange: setSorting,
 *   state: { sorting }
 * })
 * ```
 */
export function usePersistedTableSorting({
  tableId,
  defaultSorting = [],
}: UsePersistedTableSortingProps): UsePersistedTableSortingReturn {
  const storageKey = `${LocalStorageKey.TableSortingStatePrefix}${tableId}`

  // Initialize sorting state from localStorage or default
  const [sorting, setSortingState] = useState<SortingState>(() => {
    try {
      const persistedSorting = getLocalStorageItem(storageKey)
      if (persistedSorting) {
        const parsed = JSON.parse(persistedSorting) as SortingState
        // Validate that parsed data is a valid SortingState
        if (
          Array.isArray(parsed) &&
          parsed.every(
            (item) => typeof item === 'object' && typeof item.id === 'string' && typeof item.desc === 'boolean',
          )
        ) {
          return parsed
        }
      }
    } catch (error) {
      console.warn(`Failed to parse persisted sorting state for table ${tableId}:`, error)
    }
    return defaultSorting
  })

  // Function to persist to localStorage immediately
  const persistSorting = useCallback(
    (sortingState: SortingState) => {
      try {
        if (sortingState.length === 0) {
          // If sorting is cleared, remove from localStorage
          localStorage.removeItem(storageKey)
        } else {
          setLocalStorageItem(storageKey, JSON.stringify(sortingState))
        }
      } catch (error) {
        console.error(`Failed to persist sorting state for table ${tableId}:`, error)
      }
    },
    [storageKey, tableId],
  )

  // Update sorting state and persist to localStorage
  const setSorting = useCallback(
    (updater: SortingState | ((prev: SortingState) => SortingState)) => {
      setSortingState((prevSorting) => {
        const newSorting = typeof updater === 'function' ? updater(prevSorting) : updater

        // Persist to localStorage immediately
        persistSorting(newSorting)

        return newSorting
      })
    },
    [persistSorting],
  )

  // Function to clear persisted sorting state
  const clearPersistedSorting = useCallback(() => {
    try {
      localStorage.removeItem(storageKey)
      setSortingState(defaultSorting)
    } catch (error) {
      console.error(`Failed to clear persisted sorting state for table ${tableId}:`, error)
    }
  }, [storageKey, tableId, defaultSorting])

  // Clean up invalid localStorage entries on mount
  useEffect(() => {
    try {
      const persistedSorting = getLocalStorageItem(storageKey)
      if (persistedSorting) {
        const parsed = JSON.parse(persistedSorting)
        if (!Array.isArray(parsed)) {
          // Invalid format, remove it
          localStorage.removeItem(storageKey)
        }
      }
    } catch (error) {
      // Invalid JSON, remove it
      localStorage.removeItem(storageKey)
    }
  }, [storageKey])

  return {
    sorting,
    setSorting,
    clearPersistedSorting,
  }
}
