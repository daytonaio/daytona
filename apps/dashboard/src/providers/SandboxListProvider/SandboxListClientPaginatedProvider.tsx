/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxListContext, SandboxListContextValue } from './SandboxListContext'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { DEFAULT_SANDBOX_SORTING, SandboxFilters, SandboxSorting } from '@/hooks/useSandboxes'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { compareSandboxesBySorting, matchesSandboxFilters } from './SandboxListClientUtils'
import { useSandboxListMutations } from './useSandboxListMutations'
import { useSandboxListWsSync } from './useSandboxListWsSync'
import { Sandbox, SandboxDesiredState, SandboxState } from '@daytonaio/api-client'
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
    error,
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

  const [isRefreshing, setIsRefreshing] = useState(false)

  const handleRefresh = useCallback(async () => {
    setIsRefreshing(true)
    try {
      const result = await refetch()
      if (result.error) {
        throw result.error
      }
    } catch (refreshError) {
      handleApiError(refreshError, 'Failed to refresh sandboxes')
    } finally {
      setIsRefreshing(false)
    }
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

  const getSandboxState = useCallback(
    (sandboxId: string) => allSandboxes.find((sandbox) => sandbox.id === sandboxId)?.state,
    [allSandboxes],
  )

  const upsertSandboxInCache = useCallback(
    (sandbox: Sandbox) => {
      queryClient.setQueryData<Sandbox[]>(queryKey, (prev = []) => {
        const sandboxExists = prev.some((item) => item.id === sandbox.id)
        if (!sandboxExists) {
          return [sandbox, ...prev]
        }

        return prev.map((item) => (item.id === sandbox.id ? sandbox : item))
      })
    },
    [queryClient, queryKey],
  )

  const onSandboxCreated = useCallback(
    (sandbox: Sandbox) => {
      upsertSandboxInCache(sandbox)
    },
    [upsertSandboxInCache],
  )

  const onSandboxStateUpdated = useCallback(
    (data: { sandbox: Sandbox; oldState: SandboxState; newState: SandboxState }) => {
      if (data.oldState === data.newState && data.newState === SandboxState.STARTED) {
        onSandboxCreated(data.sandbox)
        return
      }

      if (data.newState === SandboxState.DESTROYED) {
        upsertSandboxInCache({ ...data.sandbox, state: SandboxState.DESTROYED })
        return
      }

      if (
        data.sandbox.desiredState === SandboxDesiredState.DESTROYED &&
        (data.newState === SandboxState.ERROR || data.newState === SandboxState.BUILD_FAILED)
      ) {
        upsertSandboxInCache({ ...data.sandbox, state: SandboxState.DESTROYED })
        return
      }

      upsertSandboxInCache(data.sandbox)
    },
    [onSandboxCreated, upsertSandboxInCache],
  )

  const onSandboxDesiredStateUpdated = useCallback(
    (data: { sandbox: Sandbox; oldDesiredState: SandboxDesiredState; newDesiredState: SandboxDesiredState }) => {
      if (data.newDesiredState !== SandboxDesiredState.DESTROYED) {
        return
      }

      if (data.sandbox.state !== SandboxState.ERROR && data.sandbox.state !== SandboxState.BUILD_FAILED) {
        return
      }

      upsertSandboxInCache({ ...data.sandbox, state: SandboxState.DESTROYED })
    },
    [upsertSandboxInCache],
  )

  useSandboxListWsSync({
    onSandboxCreated,
    onSandboxStateUpdated,
    onSandboxDesiredStateUpdated,
  })

  const markAllQueriesAsStale = useCallback(
    async (shouldRefetchActive = false) => {
      await queryClient.invalidateQueries({
        queryKey,
        refetchType: shouldRefetchActive ? 'active' : 'none',
      })
    },
    [queryClient, queryKey],
  )

  const cancelOutgoingRefetches = useCallback(async () => {
    await queryClient.cancelQueries({ queryKey })
  }, [queryClient, queryKey])

  const { startSandbox, recoverSandbox, stopSandbox, archiveSandbox, deleteSandbox } = useSandboxListMutations({
    getSandboxState,
    performSandboxStateOptimisticUpdate,
    revertSandboxStateOptimisticUpdate,
    cancelOutgoingRefetches,
    markAllQueriesAsStale,
  })

  const value = useMemo<SandboxListContextValue>(
    () => ({
      sandboxes: processedData.items,
      totalItems: processedData.totalItems,
      pageCount: processedData.pageCount,
      isLoading,
      error: error ?? null,
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
      processedData,
      isLoading,
      error,
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
