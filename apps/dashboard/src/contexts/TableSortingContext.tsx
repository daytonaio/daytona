/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { createContext } from 'react'
import { TableSortingStates } from '@/types/TableSortingStates'
import { SortingState } from '@tanstack/react-table'

export interface ITableSortingContext {
  sortingStates: TableSortingStates
  updateSortingState: (tableId: string, sortingState: SortingState) => void
}

export const TableSortingContext = createContext<ITableSortingContext | undefined>(undefined)
