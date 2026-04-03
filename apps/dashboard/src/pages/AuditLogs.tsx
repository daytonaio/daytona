/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AuditLogTable } from '@/components/AuditLogTable'
import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { RefreshSegmentedButton } from '@/components/RefreshSegmentedButton'
import { DateRangePicker, QuickRangesConfig } from '@/components/ui/date-range-picker'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { useAuditLogsQuery, type AuditLogsQueryParams } from '@/hooks/queries/useAuditLogsQuery'
import { handleApiError } from '@/lib/error-handling'
import { PaginatedAuditLogs } from '@daytona/api-client'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { DateRange } from 'react-day-picker'

const EMPTY_AUDIT_LOGS: PaginatedAuditLogs = {
  items: [],
  total: 0,
  page: 1,
  totalPages: 0,
  nextToken: undefined,
}

const AUDIT_LOG_QUICK_RANGES: QuickRangesConfig = {
  minutes: [5, 15, 30],
  hours: [1, 3, 6, 12],
  days: [1, 2, 7, 30, 90],
  months: [6],
  years: [1],
}

interface AuditLogsPaginationState {
  pageIndex: number
  pageSize: number
  cursors: Record<number, string>
}

function useAuditLogsPagination(initialPageSize: number) {
  const [pagination, setPagination] = useState<AuditLogsPaginationState>({
    pageIndex: 0,
    pageSize: initialPageSize,
    cursors: {},
  })

  const currentCursor = pagination.cursors[pagination.pageIndex]

  const resetPagination = useCallback(() => {
    setPagination({
      pageIndex: 0,
      pageSize: initialPageSize,
      cursors: {},
    })
  }, [initialPageSize])

  const setPageSize = useCallback((pageSize: number) => {
    setPagination({
      pageIndex: 0,
      pageSize,
      cursors: {},
    })
  }, [])

  const setOffsetPage = useCallback((pageIndex: number, pageSize: number) => {
    setPagination({
      pageIndex,
      pageSize,
      cursors: {},
    })
  }, [])

  const goNextWithCursor = useCallback((nextCursor: string) => {
    setPagination((prev) => {
      const nextPageIndex = prev.pageIndex + 1
      return {
        ...prev,
        pageIndex: nextPageIndex,
        cursors: {
          ...prev.cursors,
          [nextPageIndex]: nextCursor,
        },
      }
    })
  }, [])

  const goPreviousPage = useCallback(() => {
    setPagination((prev) => {
      if (prev.pageIndex === 0) {
        return prev
      }

      const nextPageIndex = prev.pageIndex - 1
      const nextCursors = { ...prev.cursors }
      delete nextCursors[prev.pageIndex]

      return {
        ...prev,
        pageIndex: nextPageIndex,
        cursors: nextCursors,
      }
    })
  }, [])

  return {
    pagination,
    currentCursor,
    resetPagination,
    setPageSize,
    setOffsetPage,
    goNextWithCursor,
    goPreviousPage,
  }
}

const AuditLogs: React.FC = () => {
  const [refreshInterval, setRefreshInterval] = useState<number | false>(false)
  const [dateRange, setDateRange] = useState<DateRange>({ from: undefined, to: undefined })
  const { pagination, currentCursor, resetPagination, setPageSize, setOffsetPage, goNextWithCursor, goPreviousPage } =
    useAuditLogsPagination(DEFAULT_PAGE_SIZE)
  const scrollToTableTop = useCallback(() => {
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }, [])

  const queryParams = useMemo<AuditLogsQueryParams>(
    () => ({
      page: pagination.pageIndex + 1,
      pageSize: pagination.pageSize,
      from: dateRange.from,
      to: dateRange.to,
      cursor: currentCursor,
    }),
    [pagination.pageIndex, pagination.pageSize, dateRange.from, dateRange.to, currentCursor],
  )

  const {
    data = EMPTY_AUDIT_LOGS,
    isLoading,
    isRefetching,
    isPlaceholderData,
    error,
    refetch,
    dataUpdatedAt,
  } = useAuditLogsQuery(queryParams, {
    refetchInterval: refreshInterval,
  })

  const handlePaginationChange = useCallback(
    ({ pageIndex, pageSize }: { pageIndex: number; pageSize: number }) => {
      if (isPlaceholderData) {
        return
      }

      if (pageSize !== pagination.pageSize) {
        scrollToTableTop()
        setPageSize(pageSize)
        return
      }

      const pageDelta = pageIndex - pagination.pageIndex

      if (pageDelta === 0) {
        return
      }

      if (Math.abs(pageDelta) > 1) {
        scrollToTableTop()
        setOffsetPage(pageIndex, pageSize)
        return
      }

      if (pageDelta > 0) {
        scrollToTableTop()
        if (data.nextToken) {
          goNextWithCursor(data.nextToken)
        } else {
          setOffsetPage(pageIndex, pageSize)
        }
        return
      }

      scrollToTableTop()
      if (currentCursor !== undefined) {
        goPreviousPage()
      } else {
        setOffsetPage(pageIndex, pageSize)
      }
    },
    [
      isPlaceholderData,
      pagination.pageIndex,
      pagination.pageSize,
      currentCursor,
      goNextWithCursor,
      goPreviousPage,
      setOffsetPage,
      setPageSize,
      scrollToTableTop,
      data.nextToken,
    ],
  )

  useEffect(() => {
    if (error) {
      handleApiError(error, 'Failed to fetch audit logs', { toastId: 'audit-logs-fetch' })
    }
  }, [error])

  useEffect(() => {
    if (!isLoading && data.items.length === 0 && pagination.pageIndex > 0) {
      goPreviousPage()
    }
  }, [isLoading, data.items.length, pagination.pageIndex, goPreviousPage])

  const handleDateRangeChange = useCallback(
    (range: DateRange) => {
      setDateRange(range)
      resetPagination()
    },
    [resetPagination],
  )

  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>Audit Logs</PageTitle>
      </PageHeader>

      <PageContent size="full">
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
          <div className="flex gap-2 items-center">
            <DateRangePicker
              value={dateRange}
              onChange={handleDateRangeChange}
              quickRangesEnabled
              quickRanges={AUDIT_LOG_QUICK_RANGES}
              timeSelection
              disabled={isLoading}
            />
          </div>
          <RefreshSegmentedButton
            className="ml-auto"
            value={refreshInterval}
            onChange={setRefreshInterval}
            onRefresh={refetch}
            isRefreshing={isRefetching}
            lastUpdatedAt={dataUpdatedAt}
          />
        </div>

        <AuditLogTable
          data={data.items}
          loading={isLoading}
          isRefetching={isRefetching}
          pageCount={data.totalPages}
          totalItems={data.total}
          onPaginationChange={handlePaginationChange}
          pagination={{
            pageIndex: pagination.pageIndex,
            pageSize: pagination.pageSize,
          }}
        />
      </PageContent>
    </PageLayout>
  )
}

export default AuditLogs
