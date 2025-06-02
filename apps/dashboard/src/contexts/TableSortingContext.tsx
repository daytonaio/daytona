/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { createContext } from 'react'
import { SortingState } from '@tanstack/react-table'

export interface ITableSortingContext {
  sortingStates: SortingState
  updateSortingState: (viewId: string, field: string, direction: 'asc' | 'desc') => void
}

export const TableSortingContext = createContext<ITableSortingContext | undefined>(undefined)
