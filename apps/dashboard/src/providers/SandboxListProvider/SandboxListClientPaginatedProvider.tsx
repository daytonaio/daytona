/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxListContext, SandboxListContextValue } from './SandboxListContext'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { DEFAULT_SANDBOX_SORTING, SandboxFilters, SandboxSorting } from '@/hooks/useSandboxes'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { compareSandboxesBySorting, matchesSandboxFilters } from './SandboxListClientUtils'
import { Sandbox, SandboxState } from '@daytonaio/api-client'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import React, { useCallback, useEffect, useMemo, useState } from 'react'

const QUERY_KEY_PREFIX = 'sandboxes-client'

/**
 * @deprecated Temporary provider using client-side pagination and filtering while the
 * server-side paginated endpoint is not yet deployed to production.
 * Use SandboxListServerPaginatedProvider once the backend supports it.
 */
export const SandboxListClientPaginatedProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { sandboxApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()
  const queryClient = useQueryClient()

  const queryKey = useMemo(() => [QUERY_KEY_PREFIX, selectedOrganization?.id] as const, [selectedOrganization?.id])

  const {
    data: allSandboxes = [],
    isLoading,
    isRefetching,
    refetch,
  } = useQuery<Sandbox[]>({
    queryKey,
    queryFn: async () => {
      if (!selectedOrganization) return []
      const response = await sandboxApi.listSandboxes(selectedOrganization.id)
      return response.data
    },
    enabled: !!selectedOrganization,
  })

  const [paginationParams, setPaginationParams] = useState({
    pageIndex: 0,
    pageSize: DEFAULT_PAGE_SIZE,
  })

  const handlePaginationChange = useCallback(({ pageIndex, pageSize }: { pageIndex: number; pageSize: number }) => {
    setPaginationParams({ pageIndex, pageSize })
  }, [])

  const [filters, setFilters] = useState<SandboxFilters>({})

  const handleFiltersChange = useCallback((filters: SandboxFilters) => {
    setFilters(filters)
    setPaginationParams((prev) => ({ ...prev, pageIndex: 0 }))
  }, [])

  const [sorting, setSorting] = useState<SandboxSorting>(DEFAULT_SANDBOX_SORTING)

  const handleSortingChange = useCallback((sorting: SandboxSorting) => {
    setSorting(sorting)
    setPaginationParams((prev) => ({ ...prev, pageIndex: 0 }))
  }, [])

  const processedData = useMemo(() => {
    const filtered = allSandboxes.filter((sandbox) => matchesSandboxFilters(sandbox, filters))
    filtered.sort((a, b) => compareSandboxesBySorting(a, b, sorting))

    const totalItems = filtered.length
    const pageCount = Math.max(1, Math.ceil(totalItems / paginationParams.pageSize))
    const start = paginationParams.pageIndex * paginationParams.pageSize
    const items = filtered.slice(start, start + paginationParams.pageSize)

    return { items, totalItems, pageCount }
  }, [allSandboxes, filters, sorting, paginationParams])

  useEffect(() => {
    if (processedData.items.length === 0 && paginationParams.pageIndex > 0) {
      setPaginationParams((prev) => ({
        ...prev,
        pageIndex: prev.pageIndex - 1,
      }))
    }
  }, [processedData.items.length, paginationParams.pageIndex])

  const handleRefresh = useCallback(async () => {
    await refetch()
  }, [refetch])

  const performSandboxStateOptimisticUpdate = useCallback(
    (sandboxId: string, newState: SandboxState) => {
      queryClient.setQueryData<Sandbox[]>(queryKey, (prev) =>
        prev?.map((s) => (s.id === sandboxId ? { ...s, state: newState } : s)),
      )
    },
    [queryClient, queryKey],
  )

  const revertSandboxStateOptimisticUpdate = useCallback(
    (sandboxId: string, previousState?: SandboxState) => {
      if (!previousState) return
      queryClient.setQueryData<Sandbox[]>(queryKey, (prev) =>
        prev?.map((s) => (s.id === sandboxId ? { ...s, state: previousState } : s)),
      )
    },
    [queryClient, queryKey],
  )

  const markAllQueriesAsStale = useCallback(
    async (_shouldRefetchActive?: boolean) => {
      await queryClient.invalidateQueries({ queryKey })
    },
    [queryClient, queryKey],
  )

  const cancelOutgoingRefetches = useCallback(async () => {
    await queryClient.cancelQueries({ queryKey })
  }, [queryClient, queryKey])

  const value = useMemo<SandboxListContextValue>(
    () => ({
      sandboxes: processedData.items,
      totalItems: processedData.totalItems,
      pageCount: processedData.pageCount,
      isLoading,
      pagination: paginationParams,
      onPaginationChange: handlePaginationChange,
      sorting,
      onSortingChange: handleSortingChange,
      filters,
      onFiltersChange: handleFiltersChange,
      handleRefresh,
      isRefreshing: isRefetching,
      performSandboxStateOptimisticUpdate,
      revertSandboxStateOptimisticUpdate,
      cancelOutgoingRefetches,
      markAllQueriesAsStale,
    }),
    [
      processedData,
      isLoading,
      paginationParams,
      handlePaginationChange,
      sorting,
      handleSortingChange,
      filters,
      handleFiltersChange,
      handleRefresh,
      isRefetching,
      performSandboxStateOptimisticUpdate,
      revertSandboxStateOptimisticUpdate,
      cancelOutgoingRefetches,
      markAllQueriesAsStale,
    ],
  )

  return <SandboxListContext.Provider value={value}>{children}</SandboxListContext.Provider>
}
