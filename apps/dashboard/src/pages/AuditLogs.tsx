/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useCallback, useEffect, useState, useMemo, useRef } from 'react'
import { useApi } from '@/hooks/useApi'
import { PaginatedAuditLogs } from '@daytonaio/api-client'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { AuditLogTable } from '@/components/AuditLogTable'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { handleApiError } from '@/lib/error-handling'
import { DateRangePicker, QuickRangesConfig, DateRangePickerRef } from '@/components/ui/date-range-picker'
import { DateRange } from 'react-day-picker'
import { useInterval } from 'usehooks-ts'

const AuditLogs: React.FC = () => {
  const { auditApi } = useApi()

  const [data, setData] = useState<PaginatedAuditLogs>({
    items: [],
    total: 0,
    page: 1,
    totalPages: 0,
    nextToken: undefined,
  })
  const [loadingData, setLoadingData] = useState(true)
  const [autoRefresh, setAutoRefresh] = useState(false)
  const [dateRange, setDateRange] = useState<DateRange>({ from: undefined, to: undefined })
  const dateRangePickerRef = useRef<DateRangePickerRef>(null)

  const { selectedOrganization } = useSelectedOrganization()

  const [paginationParams, setPaginationParams] = useState({
    pageIndex: 0,
    pageSize: DEFAULT_PAGE_SIZE,
  })
  const [currentCursor, setCurrentCursor] = useState<string | undefined>(undefined)
  const [cursorHistory, setCursorHistory] = useState<string[]>([])

  // Quick ranges configuration
  const auditLogQuickRanges: QuickRangesConfig = useMemo(
    () => ({
      minutes: [5, 15, 30],
      hours: [1, 3, 6, 12],
      days: [1, 2, 7, 30, 90],
      months: [6],
      years: [1],
    }),
    [],
  )

  const fetchData = useCallback(
    async (showTableLoadingState = true) => {
      if (!selectedOrganization) {
        return
      }

      if (showTableLoadingState) {
        setLoadingData(true)
      }

      try {
        const response = (
          await auditApi.getOrganizationAuditLogs(
            selectedOrganization.id,
            paginationParams.pageIndex + 1,
            paginationParams.pageSize,
            dateRange.from,
            dateRange.to,
            currentCursor,
          )
        ).data

        setData(response)
      } catch (error) {
        handleApiError(error, 'Failed to fetch audit logs')
      } finally {
        setLoadingData(false)
      }
    },
    [auditApi, selectedOrganization, paginationParams.pageIndex, paginationParams.pageSize, dateRange, currentCursor],
  )

  const handlePaginationChange = useCallback(
    ({ pageIndex, pageSize }: { pageIndex: number; pageSize: number }) => {
      if (pageSize !== paginationParams.pageSize) {
        // Reset to first page when changing page size
        setPaginationParams({ pageIndex: 0, pageSize })
        setCurrentCursor(undefined)
        setCursorHistory([])
      } else if (pageIndex > paginationParams.pageIndex) {
        // Next page - use cursor if available
        if (data.nextToken) {
          // Store current cursor in history before moving to next
          if (currentCursor) {
            setCursorHistory((prev) => [...prev, currentCursor])
          }
          setCurrentCursor(data.nextToken)
          setPaginationParams((prev) => ({ ...prev, pageIndex: prev.pageIndex + 1 }))
        } else {
          // Regular offset pagination
          setCurrentCursor(undefined)
          setCursorHistory([])
          setPaginationParams({ pageIndex, pageSize })
        }
      } else if (pageIndex < paginationParams.pageIndex) {
        // Previous page - check if we can go back in cursor history
        if (currentCursor && cursorHistory.length > 0) {
          // Go back in cursor pagination
          const previousCursor = cursorHistory[cursorHistory.length - 1]
          setCursorHistory((prev) => prev.slice(0, -1))
          setCurrentCursor(previousCursor)
          setPaginationParams((prev) => ({ ...prev, pageIndex: prev.pageIndex - 1 }))
        } else {
          // Go back to offset pagination
          setCurrentCursor(undefined)
          setCursorHistory([])
          setPaginationParams({ pageIndex, pageSize })
        }
      } else {
        // Same page, just update params
        setPaginationParams({ pageIndex, pageSize })
      }
      setLoadingData(true)
    },
    [paginationParams.pageIndex, paginationParams.pageSize, data.nextToken, currentCursor, cursorHistory],
  )

  useEffect(() => {
    fetchData()
  }, [fetchData])

  // Auto-refresh
  useInterval(
    () => {
      fetchData(false)
    },
    autoRefresh ? 5000 : null,
  )

  // handle case where there are no items on the current page, and we are not on the first page
  useEffect(() => {
    if (data.items.length === 0 && paginationParams.pageIndex > 0) {
      setPaginationParams((prev) => ({
        ...prev,
        pageIndex: prev.pageIndex - 1,
      }))
    }
  }, [data.items.length, paginationParams.pageIndex])

  const handleAutoRefreshChange = useCallback(
    (enabled: boolean) => {
      setAutoRefresh(enabled)
      if (enabled) {
        // Fetch immediately when enabling auto refresh
        fetchData(false)
      }
    },
    [fetchData],
  )

  const handleDateRangeChange = useCallback((range: DateRange) => {
    setDateRange(range)
    setPaginationParams({ pageIndex: 0, pageSize: DEFAULT_PAGE_SIZE })
    setCurrentCursor(undefined)
    setCursorHistory([])
    setData((prev) => ({ ...prev, page: 1, nextToken: undefined }))
  }, [])

  return (
    <div className="p-6 pt-2">
      <div className="mb-2 h-12 flex items-center justify-between">
        <h1 className="text-2xl font-medium">Audit Logs</h1>
        <div className="flex items-center gap-2">
          <Label htmlFor="auto-refresh">Auto Refresh</Label>
          <Switch id="auto-refresh" checked={autoRefresh} onCheckedChange={handleAutoRefreshChange} />
        </div>
      </div>

      <div className="flex flex-col gap-2 sm:flex-row sm:items-center mb-4">
        <div className="flex gap-2 items-center">
          <DateRangePicker
            value={dateRange}
            onChange={handleDateRangeChange}
            quickRangesEnabled={true}
            quickRanges={auditLogQuickRanges}
            timeSelection={true}
            ref={dateRangePickerRef}
            disabled={loadingData}
          />
        </div>
      </div>

      <AuditLogTable
        data={data.items}
        loading={loadingData}
        pageCount={data.totalPages}
        totalItems={data.total}
        onPaginationChange={handlePaginationChange}
        pagination={{
          pageIndex: paginationParams.pageIndex,
          pageSize: paginationParams.pageSize,
        }}
      />
    </div>
  )
}

export default AuditLogs
