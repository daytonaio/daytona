/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DebouncedInput } from '@/components/DebouncedInput'
import { Pagination } from '@/components/Pagination'
import { TableEmptyState } from '@/components/TableEmptyState'
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
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { RoutePath } from '@/enums/RoutePath'
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
    },
    meta: {
      webhookEndpoints: {
        onDisable: setDisableEndpoint,
        onDelete: setDeleteEndpoint,
        isLoadingEndpoint,
      },
    },
  })

  const handleRowClick = (endpoint: EndpointOut) => {
    navigate(RoutePath.WEBHOOK_ENDPOINT_DETAILS.replace(':endpointId', endpoint.id))
  }

  return (
    <div>
      <div className="flex items-center mb-4">
        <DebouncedInput
          value={globalFilter ?? ''}
          onChange={(value) => setGlobalFilter(String(value))}
          placeholder="Search by URL or Description"
          className="max-w-sm"
        />
      </div>
      <div className="rounded-md border">
        <Table style={{ tableLayout: 'fixed', width: '100%' }}>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead
                      className="px-2"
                      key={header.id}
                      style={{
                        width: `${header.column.getSize()}px`,
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
              <>
                {Array.from(new Array(5)).map((_, i) => (
                  <TableRow key={i}>
                    {table.getVisibleLeafColumns().map((column, colIndex, arr) =>
                      colIndex === arr.length - 1 ? null : (
                        <TableCell key={column.id} className="px-2">
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
                        style={{
                          width: `${cell.column.getSize()}px`,
                        }}
                      >
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </TableCell>
                    ))}
                  </TableRow>
                )
              })
            ) : (
              <TableEmptyState
                colSpan={columns.length}
                message="No webhook endpoints found."
                icon={<Mail className="size-8" />}
                description={
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
                }
              />
            )}
          </TableBody>
        </Table>
      </div>
      <Pagination table={table} className="mt-4" entityName="Endpoints" />

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
