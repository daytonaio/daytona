/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PageFooterPortal } from '@/components/PageLayout'
import { Pagination } from '@/components/Pagination'
import { SearchInput } from '@/components/SearchInput'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
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
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { RoutePath } from '@/enums/RoutePath'
import { cn } from '@/lib/utils'
import { getColumnSizeStyles } from '@/lib/utils/table'
import {
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { Mail } from 'lucide-react'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { EndpointOut } from 'svix'
import { columns } from './columns'

interface WebhooksEndpointTableProps {
  data: EndpointOut[]
  loading: boolean
  onDisable: (endpoint: EndpointOut) => void
  onDelete: (endpoint: EndpointOut) => void
  isLoadingEndpoint: (endpoint: EndpointOut) => boolean
}

export function WebhooksEndpointTable({
  data,
  loading,
  onDisable,
  onDelete,
  isLoadingEndpoint,
}: WebhooksEndpointTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const [globalFilter, setGlobalFilter] = useState('')
  const [deleteEndpoint, setDeleteEndpoint] = useState<EndpointOut | null>(null)
  const [disableEndpoint, setDisableEndpoint] = useState<EndpointOut | null>(null)
  const navigate = useNavigate()

  const handleConfirmDelete = () => {
    if (deleteEndpoint) {
      onDelete(deleteEndpoint)
      setDeleteEndpoint(null)
    }
  }

  const handleConfirmDisable = () => {
    if (disableEndpoint) {
      onDisable(disableEndpoint)
      setDisableEndpoint(null)
    }
  }

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    onGlobalFilterChange: setGlobalFilter,
    globalFilterFn: (row, _columnId, filterValue) => {
      const endpoint = row.original
      const searchValue = filterValue.toLowerCase()
      return (
        endpoint.url.toLowerCase().includes(searchValue) ||
        (endpoint.description?.toLowerCase().includes(searchValue) ?? false) ||
        endpoint.id.toLowerCase().includes(searchValue)
      )
    },
    state: {
      sorting,
      globalFilter,
    },
    initialState: {
      pagination: {
        pageSize: DEFAULT_PAGE_SIZE,
      },
      columnPinning: {
        right: ['actions'],
      },
    },
    meta: {
      webhookEndpoints: {
        onDisable: setDisableEndpoint,
        onDelete: setDeleteEndpoint,
        isLoadingEndpoint,
      },
    },
  })

  const isEmpty = !loading && table.getRowModel().rows.length === 0
  const hasSearch = globalFilter.trim().length > 0

  const handleRowClick = (endpoint: EndpointOut) => {
    navigate(RoutePath.WEBHOOK_ENDPOINT_DETAILS.replace(':endpointId', endpoint.id))
  }

  const handleChangeFilter = (value: string) => {
    setGlobalFilter(value)
    table.setPageIndex(0)
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-3">
      <div className="flex items-center gap-4">
        <SearchInput
          debounced
          value={globalFilter ?? ''}
          onValueChange={handleChangeFilter}
          placeholder="Search by URL or Description"
          containerClassName="max-w-sm"
        />
      </div>
      <TableContainer
        className={cn({
          'min-h-[26rem]': isEmpty,
        })}
        empty={
          isEmpty ? (
            <TableEmptyState
              overlay
              colSpan={columns.length}
              message={hasSearch ? 'No matching webhook endpoints found.' : 'No webhook endpoints found.'}
              icon={<Mail />}
              description={
                hasSearch ? null : (
                  <div className="space-y-2">
                    <p>Create an endpoint to start receiving webhook events.</p>
                    <p>
                      <a
                        href="https://www.daytona.io/docs/en/tools/api/#daytona/webhook/undefined/"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-primary hover:underline font-medium"
                      >
                        Check out the Docs
                      </a>{' '}
                      to learn more.
                    </p>
                  </div>
                )
              }
              action={
                hasSearch ? (
                  <Button variant="outline" onClick={() => handleChangeFilter('')}>
                    Clear filters
                  </Button>
                ) : null
              }
            />
          ) : null
        }
      >
        <Table className="table-fixed" style={{ minWidth: table.getTotalSize() }}>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead
                      className="px-2"
                      key={header.id}
                      style={getColumnSizeStyles(header.column)}
                      sticky={header.column.getIsPinned()}
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
              <>
                {Array.from({ length: DEFAULT_PAGE_SIZE }).map((_, i) => (
                  <TableRow key={i}>
                    {table.getVisibleLeafColumns().map((column, colIndex, arr) =>
                      colIndex === arr.length - 1 ? null : (
                        <TableCell key={column.id} className="px-2" style={getColumnSizeStyles(column)}>
                          <Skeleton className="h-4 w-10/12" />
                        </TableCell>
                      ),
                    )}
                  </TableRow>
                ))}
              </>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => {
                const isLoading = isLoadingEndpoint(row.original)
                return (
                  <TableRow
                    key={row.id}
                    data-state={row.getIsSelected() && 'selected'}
                    className={`${isLoading ? 'opacity-50 pointer-events-none' : 'cursor-pointer hover:bg-muted/50 focus-visible:bg-muted/50 focus-visible:outline-none'}`}
                    tabIndex={isLoading ? undefined : 0}
                    role={isLoading ? undefined : 'link'}
                    onClick={() => {
                      if (!isLoading) {
                        handleRowClick(row.original)
                      }
                    }}
                    onKeyDown={(e) => {
                      if (!isLoading && (e.key === 'Enter' || e.key === ' ')) {
                        e.preventDefault()
                        handleRowClick(row.original)
                      }
                    }}
                  >
                    {row.getVisibleCells().map((cell) => (
                      <TableCell
                        className="px-2"
                        key={cell.id}
                        style={getColumnSizeStyles(cell.column)}
                        sticky={cell.column.getIsPinned()}
                      >
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </TableCell>
                    ))}
                  </TableRow>
                )
              })
            ) : null}
          </TableBody>
        </Table>
      </TableContainer>
      <PageFooterPortal>
        <Pagination table={table} entityName="Endpoints" />
      </PageFooterPortal>

      <AlertDialog open={!!deleteEndpoint} onOpenChange={(open) => !open && setDeleteEndpoint(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Webhook Endpoint</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete this webhook endpoint? This action cannot be undone.
              {deleteEndpoint && (
                <div className="mt-2 text-sm">
                  <strong>URL:</strong> {deleteEndpoint.url}
                </div>
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={handleConfirmDelete}>
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog open={!!disableEndpoint} onOpenChange={(open) => !open && setDisableEndpoint(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{disableEndpoint?.disabled ? 'Enable' : 'Disable'} Webhook Endpoint</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to {disableEndpoint?.disabled ? 'enable' : 'disable'} this webhook endpoint?
              {disableEndpoint && (
                <div className="mt-2 text-sm">
                  <strong>URL:</strong> {disableEndpoint.url}
                </div>
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleConfirmDisable}>
              {disableEndpoint?.disabled ? 'Enable' : 'Disable'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
