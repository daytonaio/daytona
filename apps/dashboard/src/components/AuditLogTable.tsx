/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PageFooterPortal } from '@/components/PageLayout'
import { Pagination } from '@/components/Pagination'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableEmptyState,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { cn, getMaskedTokenFromParts, getRelativeTimeString } from '@/lib/utils'
import { getColumnSizeStyles } from '@/lib/utils/table'
import { AuditLog } from '@daytona/api-client'
import { Column, ColumnDef, flexRender, getCoreRowModel, useReactTable } from '@tanstack/react-table'
import { TextSearch } from 'lucide-react'

interface Props {
  data: AuditLog[]
  loading: boolean
  isRefetching?: boolean
  pagination: {
    pageIndex: number
    pageSize: number
  }
  pageCount: number
  totalItems: number
  onPaginationChange: (pagination: { pageIndex: number; pageSize: number }) => void
  hasFilters?: boolean
  onClearFilters?: () => void
}

export function AuditLogTable({
  data,
  loading,
  isRefetching = false,
  pagination,
  pageCount,
  onPaginationChange,
  totalItems,
  hasFilters = false,
  onClearFilters,
}: Props) {
  const table = useReactTable({
    data,
    columns: auditLogColumns,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    pageCount: pageCount || 1,
    onPaginationChange: pagination
      ? (updater) => {
          const newPagination = typeof updater === 'function' ? updater(table.getState().pagination) : updater
          onPaginationChange(newPagination)
        }
      : undefined,
    state: {
      pagination: {
        pageIndex: pagination?.pageIndex || 0,
        pageSize: pagination?.pageSize || 10,
      },
    },
    getRowId: (row) => row.id,
  })

  const isEmpty = !loading && table.getRowModel().rows.length === 0

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-3 overflow-hidden">
      <TableContainer
        className={isEmpty ? 'min-h-[26rem]' : undefined}
        empty={
          isEmpty ? (
            <TableEmptyState
              overlay
              colSpan={auditLogColumns.length}
              message={hasFilters ? 'No matching logs found.' : 'No logs yet.'}
              icon={<TextSearch />}
              description={
                hasFilters ? null : (
                  <p>Audit logs are detailed records of all actions taken by users in the organization.</p>
                )
              }
              action={
                hasFilters && onClearFilters ? (
                  <Button variant="outline" onClick={onClearFilters}>
                    Clear filters
                  </Button>
                ) : null
              }
            />
          ) : null
        }
      >
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead key={header.id} style={isEmpty ? undefined : getColumnSizeStyles(header.column)}>
                      {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                    </TableHead>
                  )
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {loading ? (
              <AuditLogTableSkeleton columns={table.getVisibleLeafColumns()} />
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id} className={cn({ 'opacity-70 transition-opacity': isRefetching })}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id} style={getColumnSizeStyles(cell.column)}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : null}
          </TableBody>
        </Table>
      </TableContainer>
      <PageFooterPortal>
        <Pagination table={table} entityName="Logs" totalItems={totalItems} />
      </PageFooterPortal>
    </div>
  )
}

function AuditLogTableSkeleton({ columns }: { columns: Column<AuditLog>[] }) {
  return (
    <>
      {Array.from({ length: DEFAULT_PAGE_SIZE }).map((_, rowIndex) => (
        <TableRow key={rowIndex}>
          {columns.map((column, columnIndex) => (
            <TableCell key={`${rowIndex}-${column.id}`} style={getColumnSizeStyles(column)}>
              {columnIndex === 0 || columnIndex === 3 || columnIndex === 4 ? (
                <div className="space-y-1">
                  <Skeleton className="h-4 w-3/4" />
                  <Skeleton className="h-4 w-1/2" />
                </div>
              ) : (
                <Skeleton className="h-4 w-10/12" />
              )}
            </TableCell>
          ))}
        </TableRow>
      ))}
    </>
  )
}

const auditLogColumns: ColumnDef<AuditLog>[] = [
  {
    header: 'Time',
    size: 200,
    cell: ({ row }) => {
      const createdAt = new Date(row.original.createdAt)
      const localeString = createdAt.toLocaleString()
      const relativeTimeString = getRelativeTimeString(row.original.createdAt).relativeTimeString

      return (
        <div className="space-y-1">
          <div className="font-medium truncate">{relativeTimeString}</div>
          <div className="text-sm text-muted-foreground truncate">{localeString}</div>
        </div>
      )
    },
  },
  {
    header: 'User',
    size: 240,
    cell: ({ row }) => {
      const actorEmail = row.original.actorEmail
      const actorId = row.original.actorId
      const label = actorEmail || actorId
      const apiKeyPrefix = row.original.actorApiKeyPrefix
      const apiKeySuffix = row.original.actorApiKeySuffix
      const maskedApiKey =
        apiKeyPrefix && apiKeySuffix ? getMaskedTokenFromParts(apiKeyPrefix, apiKeySuffix) : undefined

      return (
        <div className="space-y-1">
          <Tooltip>
            <TooltipTrigger asChild>
              <div className="font-medium truncate w-fit max-w-full">{label}</div>
            </TooltipTrigger>
            <TooltipContent>
              <p>{label}</p>
            </TooltipContent>
          </Tooltip>
          {maskedApiKey && (
            <Tooltip>
              <TooltipTrigger asChild>
                <div className="text-sm text-muted-foreground truncate w-fit max-w-full">{maskedApiKey}</div>
              </TooltipTrigger>
              <TooltipContent>
                <p>{maskedApiKey}</p>
              </TooltipContent>
            </Tooltip>
          )}
        </div>
      )
    },
  },
  {
    header: 'Action',
    size: 240,
    cell: ({ row }) => {
      const action = row.original.action

      return (
        <Tooltip>
          <TooltipTrigger asChild>
            <div className="font-medium truncate w-fit max-w-full">{action}</div>
          </TooltipTrigger>
          <TooltipContent>
            <p>{action}</p>
          </TooltipContent>
        </Tooltip>
      )
    },
  },
  {
    header: 'Target',
    size: 360,
    cell: ({ row }) => {
      const targetType = row.original.targetType
      const targetId = row.original.targetId

      if (!targetType && !targetId) {
        return '-'
      }

      return (
        <div className="space-y-1">
          {targetType && <div className="font-medium truncate">{targetType}</div>}
          {targetId && <div className="text-sm text-muted-foreground truncate">{targetId}</div>}
        </div>
      )
    },
  },
  {
    header: 'Outcome',
    size: 320,
    cell: ({ row }) => {
      const statusCode = row.original.statusCode
      const errorMessage = row.original.errorMessage
      const outcomeInfo = getOutcomeInfo(statusCode)

      return (
        <div className="space-y-1">
          <div className={`font-medium ${outcomeInfo.colorClass}`}>{outcomeInfo.label}</div>
          {!errorMessage ? (
            <div className="text-sm text-muted-foreground truncate">{statusCode || '204'}</div>
          ) : (
            <Tooltip>
              <TooltipTrigger asChild>
                <div className="text-sm text-muted-foreground truncate">
                  {statusCode || '500'}
                  {` - ${errorMessage}`}
                </div>
              </TooltipTrigger>
              <TooltipContent>
                <p>{errorMessage}</p>
              </TooltipContent>
            </Tooltip>
          )}
        </div>
      )
    },
  },
]

type OutcomeCategory = 'informational' | 'success' | 'redirect' | 'client-error' | 'server-error' | 'unknown'

interface OutcomeInfo {
  label: string
  colorClass: string
}

const getOutcomeCategory = (statusCode: number | null | undefined): OutcomeCategory => {
  if (!statusCode) return 'unknown'

  if (statusCode >= 100 && statusCode < 200) return 'informational'
  if (statusCode >= 200 && statusCode < 300) return 'success'
  if (statusCode >= 300 && statusCode < 400) return 'redirect'
  if (statusCode >= 400 && statusCode < 500) return 'client-error'
  if (statusCode >= 500 && statusCode < 600) return 'server-error'

  return 'unknown'
}

const getOutcomeInfo = (statusCode: number | null | undefined): OutcomeInfo => {
  const category = getOutcomeCategory(statusCode)

  switch (category) {
    case 'informational':
      return {
        label: 'Info',
        colorClass: 'text-blue-500 dark:text-blue-300',
      }
    case 'success':
      return {
        label: 'Success',
        colorClass: 'text-green-600 dark:text-green-400',
      }
    case 'redirect':
      return {
        label: 'Redirect',
        colorClass: 'text-blue-600 dark:text-blue-400',
      }
    case 'client-error':
    case 'server-error':
      return {
        label: 'Error',
        colorClass: 'text-red-600 dark:text-red-400',
      }
    case 'unknown':
    default:
      return {
        label: 'Unknown',
        colorClass: 'text-gray-600 dark:text-gray-400',
      }
  }
}
