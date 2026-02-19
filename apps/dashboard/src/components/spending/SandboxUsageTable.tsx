/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useMemo, useState } from 'react'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ModelsSandboxUsage } from '@daytonaio/analytics-api-client'
import { CopyButton } from '@/components/CopyButton'

interface SandboxUsageTableProps {
  data: ModelsSandboxUsage[] | undefined
  isLoading: boolean
}

type SortField = 'totalPrice' | 'totalCPUSeconds' | 'totalRAMGBSeconds' | 'totalDiskGBSeconds'
type SortDirection = 'asc' | 'desc'

export const SandboxUsageTable: React.FC<SandboxUsageTableProps> = ({ data, isLoading }) => {
  const [sortField, setSortField] = useState<SortField>('totalPrice')
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc')

  const sortedData = useMemo(() => {
    if (!data) return []
    return [...data].sort((a, b) => {
      const aVal = (a[sortField] ?? 0) as number
      const bVal = (b[sortField] ?? 0) as number
      return sortDirection === 'desc' ? bVal - aVal : aVal - bVal
    })
  }, [data, sortField, sortDirection])

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection((prev) => (prev === 'desc' ? 'asc' : 'desc'))
    } else {
      setSortField(field)
      setSortDirection('desc')
    }
  }

  const getSortIndicator = (field: SortField) => {
    if (sortField !== field) return ''
    return sortDirection === 'desc' ? ' \u2193' : ' \u2191'
  }

  if (isLoading || !data) {
    return null
  }

  return (
    <Card>
      <CardHeader className="border-b p-4">
        <CardTitle>Per-Sandbox Usage</CardTitle>
      </CardHeader>
      <CardContent className="p-0">
        {sortedData.length === 0 ? (
          <div className="flex items-center justify-center py-12 text-muted-foreground text-sm">
            No sandbox usage data
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Sandbox ID</TableHead>
                <TableHead className="cursor-pointer select-none text-right" onClick={() => handleSort('totalPrice')}>
                  Total Price{getSortIndicator('totalPrice')}
                </TableHead>
                <TableHead
                  className="cursor-pointer select-none text-right"
                  onClick={() => handleSort('totalCPUSeconds')}
                >
                  CPU (seconds){getSortIndicator('totalCPUSeconds')}
                </TableHead>
                <TableHead
                  className="cursor-pointer select-none text-right"
                  onClick={() => handleSort('totalRAMGBSeconds')}
                >
                  RAM (GB-seconds){getSortIndicator('totalRAMGBSeconds')}
                </TableHead>
                <TableHead
                  className="cursor-pointer select-none text-right"
                  onClick={() => handleSort('totalDiskGBSeconds')}
                >
                  Disk (GB-seconds){getSortIndicator('totalDiskGBSeconds')}
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {sortedData.map((sandbox) => (
                <TableRow key={sandbox.sandboxId}>
                  <TableCell className="font-mono text-sm">
                    <div className="flex items-center gap-1">
                      <span className="truncate max-w-[200px]">{sandbox.sandboxId}</span>
                      {sandbox.sandboxId && (
                        <CopyButton value={sandbox.sandboxId} tooltipText="Copy sandbox ID" size="icon-xs" />
                      )}
                    </div>
                  </TableCell>
                  <TableCell className="text-right">${(sandbox.totalPrice ?? 0).toFixed(2)}</TableCell>
                  <TableCell className="text-right">{(sandbox.totalCPUSeconds ?? 0).toFixed(1)}</TableCell>
                  <TableCell className="text-right">{(sandbox.totalRAMGBSeconds ?? 0).toFixed(1)}</TableCell>
                  <TableCell className="text-right">{(sandbox.totalDiskGBSeconds ?? 0).toFixed(1)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  )
}
