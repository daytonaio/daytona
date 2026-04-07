/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxFilters, SandboxSorting } from '@/hooks/useSandboxes'
import { Sandbox } from '@daytonaio/api-client'
import { createContext } from 'react'

export interface SandboxListContextValue {
  sandboxes: Sandbox[]
  totalItems: number
  pageCount: number
  isLoading: boolean
  isRefetching: boolean
  error: unknown | null

  pagination: { pageIndex: number; pageSize: number }
  onPaginationChange: (pagination: { pageIndex: number; pageSize: number }) => void

  sorting: SandboxSorting
  onSortingChange: (sorting: SandboxSorting) => void

  filters: SandboxFilters
  onFiltersChange: (filters: SandboxFilters) => void

  handleRefresh: () => Promise<void>
  isRefreshing: boolean

  startSandbox: (sandboxId: string) => Promise<void>
  recoverSandbox: (sandboxId: string) => Promise<void>
  stopSandbox: (sandboxId: string) => Promise<void>
  archiveSandbox: (sandboxId: string) => Promise<void>
  deleteSandbox: (sandboxId: string) => Promise<void>
}

export const SandboxListContext = createContext<SandboxListContextValue | null>(null)
