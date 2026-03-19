/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DEFAULT_PAGE_SIZE, PAGE_SIZE_OPTIONS } from '@/constants/Pagination'
import { LocalStorageKey } from '@/enums/LocalStorageKey'
import { getLocalStorageItem, setLocalStorageItem } from '@/lib/local-storage'
import { PaginationState, Updater } from '@tanstack/react-table'
import { useCallback, useMemo, useRef, useState } from 'react'

type PageSizeStorageKey =
  | LocalStorageKey.PaginationPageSize_Sandboxes
  | LocalStorageKey.PaginationPageSize_Snapshots
  | LocalStorageKey.PaginationPageSize_AuditLogs
  | LocalStorageKey.PaginationPageSize_Registries
  | LocalStorageKey.PaginationPageSize_Volumes
  | LocalStorageKey.PaginationPageSize_ApiKeys
  | LocalStorageKey.PaginationPageSize_Members
  | LocalStorageKey.PaginationPageSize_MemberInvitations

function getPersistedPageSize(key: PageSizeStorageKey): number {
  const stored = getLocalStorageItem(key)
  if (stored) {
    const parsed = parseInt(stored, 10)
    if (!isNaN(parsed) && PAGE_SIZE_OPTIONS.includes(parsed as (typeof PAGE_SIZE_OPTIONS)[number])) {
      return parsed
    }
  }
  return DEFAULT_PAGE_SIZE
}

export function usePersistedPageSize(key: PageSizeStorageKey) {
  const [paginationParams, setInternalPaginationParams] = useState(() => ({
    pageIndex: 0,
    pageSize: getPersistedPageSize(key),
  }))

  // Use ref to track previous page size to avoid stale closures
  const prevPageSizeRef = useRef(paginationParams.pageSize)

  const handlePaginationChange = useCallback(
    (updaterOrValue: Updater<PaginationState>) => {
      setInternalPaginationParams((prev) => {
        const newValue = typeof updaterOrValue === 'function' ? updaterOrValue(prev) : updaterOrValue

        if (newValue.pageSize !== prevPageSizeRef.current) {
          setLocalStorageItem(key, newValue.pageSize.toString())
          prevPageSizeRef.current = newValue.pageSize
        }

        return newValue
      })
    },
    [key],
  )

  const setPaginationParams = useCallback(
    (
      value:
        | { pageIndex: number; pageSize: number }
        | ((prev: { pageIndex: number; pageSize: number }) => { pageIndex: number; pageSize: number }),
    ) => {
      setInternalPaginationParams((prev) => {
        const newValue = typeof value === 'function' ? value(prev) : value
        if (newValue.pageSize !== prevPageSizeRef.current) {
          setLocalStorageItem(key, newValue.pageSize.toString())
          prevPageSizeRef.current = newValue.pageSize
        }
        return newValue
      })
    },
    [key],
  )

  return useMemo(
    () => ({
      paginationParams,
      setPaginationParams,
      handlePaginationChange,
    }),
    [paginationParams, setPaginationParams, handlePaginationChange],
  )
}
