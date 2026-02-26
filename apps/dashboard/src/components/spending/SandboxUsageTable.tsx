/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useMemo, useState } from 'react'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react'
import { ModelsSandboxUsage } from '@daytonaio/analytics-api-client'
import { CopyButton } from '@/components/CopyButton'
import { PAGE_SIZE_OPTIONS } from '@/constants/Pagination'

interface SandboxUsageTableProps {
  data: ModelsSandboxUsage[] | undefined
  isLoading: boolean
}

type SortField = 'totalPrice' | 'totalCPUSeconds' | 'totalRAMGBSeconds' | 'totalDiskGBSeconds'
type SortDirection = 'asc' | 'desc'

export const SandboxUsageTable: React.FC<SandboxUsageTableProps> = ({ data, isLoading }) => {
  const [sortField, setSortField] = useState<SortField>('totalPrice')
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc')
  const [pageIndex, setPageIndex] = useState(0)
  const [pageSize, setPageSize] = useState(25)

  const sortedData = useMemo(() => {
    if (!data) return []
    return [...data].sort((a, b) => {
      const aVal = (a[sortField] ?? 0) as number
      const bVal = (b[sortField] ?? 0) as number
      return sortDirection === 'desc' ? bVal - aVal : aVal - bVal
    })
  }, [data, sortField, sortDirection])

  const pageCount = Math.max(1, Math.ceil(sortedData.length / pageSize))
  const paginatedData = useMemo(() => {
    const start = pageIndex * pageSize
    return sortedData.slice(start, start + pageSize)
  }, [sortedData, pageIndex, pageSize])

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection((prev) => (prev === 'desc' ? 'asc' : 'desc'))
    } else {
      setSortField(field)
      setSortDirection('desc')
    }
    setPageIndex(0)
  }

  const getSortIndicator = (field: SortField) => {
    if (sortField !== field) return ''
    return sortDirection === 'desc' ? ' \u2193' : ' \u2191'
  }

  const handlePageSizeChange = (value: string) => {
    setPageSize(Number(value))
    setPageIndex(0)
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
          <>
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
                {paginatedData.map((sandbox, index) => (
                  <TableRow key={sandbox.sandboxId ?? `row-${index}`}>
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

            <div className="flex flex-col sm:flex-row gap-2 sm:items-center justify-between w-full px-4 pb-2 pt-4">
              <div className="flex items-center gap-4">
                <Select value={`${pageSize}`} onValueChange={handlePageSizeChange}>
                  <SelectTrigger className="h-8 w-[164px]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent side="top">
                    {PAGE_SIZE_OPTIONS.map((size) => (
                      <SelectItem key={size} value={`${size}`}>
                        {size} per page
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <div className="text-sm text-muted-foreground">{sortedData.length} total item(s)</div>
              </div>
              <div className="flex items-center gap-4">
                <div className="text-sm font-medium text-muted-foreground">
                  Page {pageIndex + 1} of {pageCount}
                </div>
                <div className="flex items-center space-x-2">
                  <Button
                    variant="outline"
                    className="hidden h-8 w-8 p-0 lg:flex"
                    onClick={() => setPageIndex(0)}
                    disabled={pageIndex === 0}
                  >
                    <span className="sr-only">Go to first page</span>
                    <ChevronsLeft />
                  </Button>
                  <Button
                    variant="outline"
                    className="h-8 w-8 p-0"
                    onClick={() => setPageIndex((p) => p - 1)}
                    disabled={pageIndex === 0}
                  >
                    <span className="sr-only">Go to previous page</span>
                    <ChevronLeft />
                  </Button>
                  <Button
                    variant="outline"
                    className="h-8 w-8 p-0"
                    onClick={() => setPageIndex((p) => p + 1)}
                    disabled={pageIndex >= pageCount - 1}
                  >
                    <span className="sr-only">Go to next page</span>
                    <ChevronRight />
                  </Button>
                  <Button
                    variant="outline"
                    className="hidden h-8 w-8 p-0 lg:flex"
                    onClick={() => setPageIndex(pageCount - 1)}
                    disabled={pageIndex >= pageCount - 1}
                  >
                    <span className="sr-only">Go to last page</span>
                    <ChevronsRight />
                  </Button>
                </div>
              </div>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  )
}
