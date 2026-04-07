/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxListContext, SandboxListContextValue } from './SandboxListContext'
import { useSandboxListMutations } from './useSandboxListMutations'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { queryKeys } from '@/hooks/queries/queryKeys'
import { useSandboxWsSync } from '@/hooks/useSandboxWsSync'
import {
  DEFAULT_SANDBOX_SORTING,
  SandboxFilters,
  SandboxQueryParams,
  SandboxSorting,
  useSandboxes,
} from '@/hooks/useSandboxes'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { Sandbox, SandboxState } from '@daytonaio/api-client'
import { QueryKey, useQueryClient } from '@tanstack/react-query'
import React, { useCallback, useEffect, useMemo, useState } from 'react'

export const SandboxListServerPaginatedProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const queryClient = useQueryClient()
  const { selectedOrganization } = useSelectedOrganization()

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

  const queryParams = useMemo<SandboxQueryParams>(
    () => ({
      page: paginationParams.pageIndex + 1,
      pageSize: paginationParams.pageSize,
      filters,
      sorting,
    }),
    [paginationParams, filters, sorting],
  )

  const baseQueryKey = useMemo<QueryKey>(
    () => queryKeys.sandboxes.list(selectedOrganization?.id),
    [selectedOrganization?.id],
  )

  const queryKey = useMemo<QueryKey>(
    () => queryKeys.sandboxes.list(selectedOrganization?.id, queryParams),
    [selectedOrganization?.id, queryParams],
  )

  const {
    data: sandboxesData,
    isLoading: sandboxesDataIsLoading,
    isFetching: sandboxesDataIsFetching,
    error: sandboxesDataError,
    refetch: refetchSandboxesData,
  } = useSandboxes(queryKey, queryParams)

  useEffect(() => {
    if (sandboxesData?.items.length === 0 && paginationParams.pageIndex > 0) {
      setPaginationParams((prev) => ({
        ...prev,
        pageIndex: prev.pageIndex - 1,
      }))
    }
  }, [sandboxesData?.items.length, paginationParams.pageIndex])

  const [isRefreshing, setIsRefreshing] = useState(false)

  const handleRefresh = useCallback(async () => {
    setIsRefreshing(true)
    try {
      const result = await refetchSandboxesData()
      if (result.error) {
        throw result.error
      }
    } catch (error) {
      handleApiError(error, 'Failed to refresh sandboxes')
    } finally {
      setIsRefreshing(false)
    }
  }, [refetchSandboxesData])

  const updateSandboxInCache = useCallback(
    (sandboxId: string, updates: Partial<Sandbox>) => {
      queryClient.setQueryData(queryKey, (oldData: any) => {
        if (!oldData) return oldData
        return {
          ...oldData,
          items: oldData.items.map((sandbox: Sandbox) =>
            sandbox.id === sandboxId ? { ...sandbox, ...updates } : sandbox,
          ),
        }
      })
    },
    [queryClient, queryKey],
  )

  const markAllQueriesAsStale = useCallback(
    (shouldRefetchActive = false) =>
      queryClient.invalidateQueries({
        queryKey: baseQueryKey,
        refetchType: shouldRefetchActive ? 'active' : 'none',
      }),
    [queryClient, baseQueryKey],
  )

  const cancelOutgoingRefetches = useCallback(() => queryClient.cancelQueries({ queryKey }), [queryClient, queryKey])

  const performSandboxStateOptimisticUpdate = useCallback(
    (sandboxId: string, newState: SandboxState) => {
      updateSandboxInCache(sandboxId, { state: newState })
    },
    [updateSandboxInCache],
  )

  const revertSandboxStateOptimisticUpdate = useCallback(
    (sandboxId: string, previousState?: SandboxState) => {
      if (!previousState) return
      updateSandboxInCache(sandboxId, { state: previousState })
    },
    [updateSandboxInCache],
  )

  const getSandboxState = useCallback(
    (sandboxId: string) => sandboxesData?.items.find((sandbox) => sandbox.id === sandboxId)?.state,
    [sandboxesData?.items],
  )

  const shouldRefetchOnCreate = useMemo(() => {
    const isFirstPage = paginationParams.pageIndex === 0
    const isDefaultFilters = Object.keys(filters).length === 0
    const isDefaultSorting =
      sorting.field === DEFAULT_SANDBOX_SORTING.field && sorting.direction === DEFAULT_SANDBOX_SORTING.direction

    return isFirstPage && isDefaultFilters && isDefaultSorting
  }, [filters, paginationParams.pageIndex, sorting.direction, sorting.field])

  useSandboxWsSync({ refetchOnCreate: shouldRefetchOnCreate })

  const { startSandbox, recoverSandbox, stopSandbox, archiveSandbox, deleteSandbox } = useSandboxListMutations({
    getSandboxState,
    performSandboxStateOptimisticUpdate,
    revertSandboxStateOptimisticUpdate,
    cancelOutgoingRefetches,
    markAllQueriesAsStale,
  })

  const value = useMemo<SandboxListContextValue>(
    () => ({
      sandboxes: sandboxesData?.items || [],
      totalItems: sandboxesData?.total || 0,
      pageCount: sandboxesData?.totalPages || 0,
      isLoading: sandboxesDataIsLoading,
      isRefetching: sandboxesDataIsFetching && Boolean(sandboxesData),
      error: sandboxesDataError ?? null,
      pagination: paginationParams,
      onPaginationChange: handlePaginationChange,
      sorting,
      onSortingChange: handleSortingChange,
      filters,
      onFiltersChange: handleFiltersChange,
      handleRefresh,
      isRefreshing,
      startSandbox,
      recoverSandbox,
      stopSandbox,
      archiveSandbox,
      deleteSandbox,
    }),
    [
      sandboxesData,
      sandboxesDataIsLoading,
      sandboxesDataIsFetching,
      sandboxesDataError,
      paginationParams,
      handlePaginationChange,
      sorting,
      handleSortingChange,
      filters,
      handleFiltersChange,
      handleRefresh,
      isRefreshing,
      startSandbox,
      recoverSandbox,
      stopSandbox,
      archiveSandbox,
      deleteSandbox,
    ],
  )

  return <SandboxListContext.Provider value={value}>{children}</SandboxListContext.Provider>
}
