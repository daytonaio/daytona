/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxFilters, SandboxSorting } from '@/hooks/useSandboxes'
import { Sandbox, SandboxState } from '@daytonaio/api-client'
import { createContext } from 'react'

export interface SandboxListContextValue {
  sandboxes: Sandbox[]
  totalItems: number
  pageCount: number
  isLoading: boolean

  pagination: { pageIndex: number; pageSize: number }
  onPaginationChange: (pagination: { pageIndex: number; pageSize: number }) => void

  sorting: SandboxSorting
  onSortingChange: (sorting: SandboxSorting) => void

  filters: SandboxFilters
  onFiltersChange: (filters: SandboxFilters) => void

  handleRefresh: () => void
  isRefreshing: boolean

  performSandboxStateOptimisticUpdate: (sandboxId: string, newState: SandboxState) => void
  revertSandboxStateOptimisticUpdate: (sandboxId: string, previousState?: SandboxState) => void

  cancelOutgoingRefetches: () => Promise<void>
  markAllQueriesAsStale: (shouldRefetchActive?: boolean) => Promise<void>
}

export const SandboxListContext = createContext<SandboxListContextValue | null>(null)
