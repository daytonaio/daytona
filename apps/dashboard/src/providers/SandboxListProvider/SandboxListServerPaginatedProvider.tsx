/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxListContext, SandboxListContextValue } from './SandboxListContext'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import {
  DEFAULT_SANDBOX_SORTING,
  getSandboxesQueryKey,
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
    () => getSandboxesQueryKey(selectedOrganization?.id),
    [selectedOrganization?.id],
  )

  const queryKey = useMemo<QueryKey>(
    () => getSandboxesQueryKey(selectedOrganization?.id, queryParams),
    [selectedOrganization?.id, queryParams],
  )

  const {
    data: sandboxesData,
    isLoading: sandboxesDataIsLoading,
    error: sandboxesDataError,
    refetch: refetchSandboxesData,
  } = useSandboxes(queryKey, queryParams)

  useEffect(() => {
    if (sandboxesDataError) {
      handleApiError(sandboxesDataError, 'Failed to fetch sandboxes')
    }
  }, [sandboxesDataError])

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
      await refetchSandboxesData()
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
    async (shouldRefetchActive = false) => {
      queryClient.invalidateQueries({
        queryKey: baseQueryKey,
        refetchType: shouldRefetchActive ? 'active' : 'none',
      })
    },
    [queryClient, baseQueryKey],
  )

  const cancelOutgoingRefetches = useCallback(async () => {
    queryClient.cancelQueries({ queryKey })
  }, [queryClient, queryKey])

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

  const value = useMemo<SandboxListContextValue>(
    () => ({
      sandboxes: sandboxesData?.items || [],
      totalItems: sandboxesData?.total || 0,
      pageCount: sandboxesData?.totalPages || 0,
      isLoading: sandboxesDataIsLoading,
      pagination: paginationParams,
      onPaginationChange: handlePaginationChange,
      sorting,
      onSortingChange: handleSortingChange,
      filters,
      onFiltersChange: handleFiltersChange,
      handleRefresh,
      isRefreshing,
      performSandboxStateOptimisticUpdate,
      revertSandboxStateOptimisticUpdate,
      cancelOutgoingRefetches,
      markAllQueriesAsStale,
    }),
    [
      sandboxesData,
      sandboxesDataIsLoading,
      paginationParams,
      handlePaginationChange,
      sorting,
      handleSortingChange,
      filters,
      handleFiltersChange,
      handleRefresh,
      isRefreshing,
      performSandboxStateOptimisticUpdate,
      revertSandboxStateOptimisticUpdate,
      cancelOutgoingRefetches,
      markAllQueriesAsStale,
    ],
  )

  return <SandboxListContext.Provider value={value}>{children}</SandboxListContext.Provider>
}
