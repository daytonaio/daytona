/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useEffect, useState } from 'react'
import { TableSortingContext } from '@/contexts/TableSortingContext'
import { LocalStorageKey } from '@/enums/LocalStorageKey'
import { SortingState } from '@tanstack/react-table'

type Props = {
  children: React.ReactNode
}

export function TableSortingProvider({ children }: Props) {
  const [sortingStates, setSortingStates] = useState<SortingState>(() => {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem(LocalStorageKey.DashboardSortingStorage)
      return stored ? JSON.parse(stored) : {}
    }
    return {}
  })

  const updateSortingState = (viewId: string, field: string, direction: 'asc' | 'desc') => {
    const newState = {
      ...sortingStates,
      [viewId]: { field, direction },
    }
    setSortingStates(newState)
    localStorage.setItem(LocalStorageKey.DashboardSortingStorage, JSON.stringify(newState))
  }

  return (
    <TableSortingContext.Provider value={{ sortingStates, updateSortingState }}>
      {children}
    </TableSortingContext.Provider>
  )
}
