/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useCallback, useEffect, useState } from 'react'
import { useApi } from '@/hooks/useApi'
import { AuditLog, PaginatedAuditLogs } from '@daytonaio/api-client'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { AuditLogTable } from '@/components/AuditLogTable'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { useNotificationSocket } from '@/hooks/useNotificationSocket'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'

const AuditLogs: React.FC = () => {
  const { auditApi } = useApi()
  const { notificationSocket } = useNotificationSocket()

  const [data, setData] = useState<PaginatedAuditLogs>({
    items: [],
    total: 0,
    page: 1,
    totalPages: 0,
  })
  const [loadingData, setLoadingData] = useState(true)
  const [realTimeUpdatesEnabled, setRealTimeUpdatesEnabled] = useState(false)

  const { selectedOrganization } = useSelectedOrganization()

  const [paginationParams, setPaginationParams] = useState({
    pageIndex: 0,
    pageSize: DEFAULT_PAGE_SIZE,
  })

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
            paginationParams.pageSize,
            paginationParams.pageIndex + 1,
          )
        ).data
        setData(response)
      } catch (error) {
        handleApiError(error, 'Failed to fetch audit logs')
      } finally {
        setLoadingData(false)
      }
    },
    [auditApi, selectedOrganization, paginationParams.pageIndex, paginationParams.pageSize],
  )

  const handlePaginationChange = useCallback(({ pageIndex, pageSize }: { pageIndex: number; pageSize: number }) => {
    setPaginationParams({ pageIndex, pageSize })
  }, [])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  useEffect(() => {
    const handleAuditLogCreatedEvent = (auditLog: AuditLog) => {
      setData((prev) => {
        // if on first page, add to the top of the list
        const newItems = paginationParams.pageIndex === 0 ? [auditLog, ...prev.items] : prev.items
        const newTotal = prev.total + 1

        // make sure to respect pagination size and recalculate total pages
        return {
          ...prev,
          items: newItems.slice(0, paginationParams.pageSize),
          total: newTotal,
          totalPages: Math.ceil(newTotal / paginationParams.pageSize),
        }
      })
    }

    const handleAuditLogUpdatedEvent = (auditLog: AuditLog) => {
      setData((prev) => ({
        ...prev,
        items: prev.items.map((i) => (i.id === auditLog.id ? auditLog : i)),
      }))
    }

    if (!notificationSocket) {
      return
    }

    if (realTimeUpdatesEnabled) {
      notificationSocket.on('audit-log.created', handleAuditLogCreatedEvent)
      notificationSocket.on('audit-log.updated', handleAuditLogUpdatedEvent)
    }

    return () => {
      notificationSocket.off('audit-log.created', handleAuditLogCreatedEvent)
      notificationSocket.off('audit-log.updated', handleAuditLogUpdatedEvent)
    }
  }, [notificationSocket, paginationParams.pageIndex, paginationParams.pageSize, realTimeUpdatesEnabled])

  // handle case where there are no items on the current page, and we are not on the first page
  useEffect(() => {
    if (data.items.length === 0 && paginationParams.pageIndex > 0) {
      setPaginationParams((prev) => ({
        ...prev,
        pageIndex: prev.pageIndex - 1,
      }))
    }
  }, [data.items.length, paginationParams.pageIndex])

  const handleRealTimeUpdatesEnabledChange = useCallback(
    (enabled: boolean) => {
      setRealTimeUpdatesEnabled(enabled)
      if (enabled) {
        fetchData(false)
      }
    },
    [fetchData],
  )

  return (
    <div className="p-6 pt-2">
      <div className="mb-2 h-12 flex items-center justify-between">
        <h1 className="text-2xl font-medium">Audit Logs</h1>
        <div className="flex items-center gap-2">
          <Label htmlFor="real-time-updates">Real-Time Updates</Label>
          <Switch
            id="real-time-updates"
            checked={realTimeUpdatesEnabled}
            onCheckedChange={handleRealTimeUpdatesEnabledChange}
          />
        </div>
      </div>

      <AuditLogTable
        data={data.items}
        loading={loadingData}
        pageCount={data.totalPages}
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
