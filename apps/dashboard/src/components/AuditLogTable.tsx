/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AuditLog } from '@daytonaio/api-client'
import { ColumnDef, flexRender, getCoreRowModel, useReactTable } from '@tanstack/react-table'
import { TextSearch } from 'lucide-react'
import { TableHeader, TableRow, TableHead, TableBody, TableCell, Table } from '@/components/ui/table'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { TableEmptyState } from '@/components/TableEmptyState'
import { Pagination } from '@/components/Pagination'
import { getRelativeTimeString } from '@/lib/utils'

interface Props {
  data: AuditLog[]
  loading: boolean
  pagination: {
    pageIndex: number
    pageSize: number
  }
  pageCount: number
  onPaginationChange: (pagination: { pageIndex: number; pageSize: number }) => void
}

export function AuditLogTable({ data, loading, pagination, pageCount, onPaginationChange }: Props) {
  const columns = getColumns()

  const table = useReactTable({
    data,
    columns,
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

  return (
    <div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead
                      key={header.id}
                      style={{
                        minWidth: header.column.columnDef.size,
                        maxWidth: header.column.columnDef.size,
                      }}
                    >
                      {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                    </TableHead>
                  )
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  Loading...
                </TableCell>
              </TableRow>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell
                      key={cell.id}
                      style={{
                        minWidth: cell.column.columnDef.size,
                        maxWidth: cell.column.columnDef.size,
                      }}
                    >
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableEmptyState
                colSpan={columns.length}
                message="No logs yet."
                icon={<TextSearch className="w-8 h-8" />}
                description={
                  <div className="space-y-2">
                    <p>Audit logs are detailed records of all actions taken by users in the organization.</p>
                  </div>
                }
              />
            )}
          </TableBody>
        </Table>
      </div>
      <Pagination table={table} className="mt-4" entityName="Logs" />
    </div>
  )
}

const getColumns = (): ColumnDef<AuditLog>[] => {
  const columns: ColumnDef<AuditLog>[] = [
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

        return (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <div className="font-medium truncate w-fit max-w-full">{label}</div>
              </TooltipTrigger>
              <TooltipContent>
                <p>{label}</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        )
      },
    },
    {
      header: 'Action',
      size: 240,
      cell: ({ row }) => {
        const action = row.original.action

        return (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <div className="font-medium truncate w-fit max-w-full">{action}</div>
              </TooltipTrigger>
              <TooltipContent>
                <p>{action}</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
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
              <TooltipProvider>
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
              </TooltipProvider>
            )}
          </div>
        )
      },
    },
  ]

  return columns
}

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
