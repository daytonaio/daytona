/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState } from 'react'
import { TableSortingContext } from '@/contexts/TableSortingContext'
import { LocalStorageKey } from '@/enums/LocalStorageKey'
import { TableSortingStates } from '@/types/TableSortingStates'
import { SortingState } from '@tanstack/react-table'

type Props = {
  children: React.ReactNode
}

export function TableSortingProvider({ children }: Props) {
  const [sortingStates, setSortingStates] = useState<TableSortingStates>(() => {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem(LocalStorageKey.TableSorting)
      return stored ? JSON.parse(stored) : {}
    }
    return {}
  })

  const updateSortingState = (tableId: string, sortingState: SortingState) => {
    const newState = {
      ...sortingStates,
      [tableId]: sortingState,
    }
    setSortingStates(newState)
    localStorage.setItem(LocalStorageKey.TableSorting, JSON.stringify(newState))
  }

  return (
    <TableSortingContext.Provider value={{ sortingStates, updateSortingState }}>
      {children}
    </TableSortingContext.Provider>
  )
}
