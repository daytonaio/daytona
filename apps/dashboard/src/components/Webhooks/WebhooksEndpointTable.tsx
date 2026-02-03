/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DebouncedInput } from '@/components/DebouncedInput'
import { Pagination } from '@/components/Pagination'
import { TableEmptyState } from '@/components/TableEmptyState'
import { TimestampTooltip } from '@/components/TimestampTooltip'
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
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { RoutePath } from '@/enums/RoutePath'
import { getRelativeTimeString } from '@/lib/utils'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { Mail, MoreHorizontal } from 'lucide-react'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { EndpointOut } from 'svix'
import { CopyButton } from '../CopyButton'

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

  const columns = getColumns({
    onDisable: setDisableEndpoint,
    onDelete: setDeleteEndpoint,
    isLoadingEndpoint,
  })

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
    globalFilterFn: (row, columnId, filterValue) => {
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
                    className={`${isLoading ? 'opacity-50 pointer-events-none' : 'cursor-pointer hover:bg-muted/50'}`}
                    onClick={() => {
                      if (!isLoading) {
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
                icon={<Mail className="w-8 h-8" />}
                description={
                  <div className="space-y-2">
                    <p>Create an endpoint to start receiving webhook events.</p>
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

const getColumns = ({
  onDisable,
  onDelete,
  isLoadingEndpoint,
}: {
  onDisable: (endpoint: EndpointOut) => void
  onDelete: (endpoint: EndpointOut) => void
  isLoadingEndpoint: (endpoint: EndpointOut) => boolean
}): ColumnDef<EndpointOut>[] => {
  const columns: ColumnDef<EndpointOut>[] = [
    {
      accessorKey: 'description',
      header: 'Name',
      size: 200,
      cell: ({ row }) => (
        <div className="w-full truncate flex items-center gap-2">
          <span className="truncate block">{row.original.description || 'Unnamed Endpoint'}</span>
        </div>
      ),
    },
    {
      accessorKey: 'url',
      header: 'URL',
      size: 300,
      cell: ({ row }) => (
        <div
          className="w-full truncate flex items-center gap-2 group/copy-button"
          onClick={(e) => {
            e.stopPropagation()
          }}
        >
          <span className="truncate block">{row.original.url}</span>
          <CopyButton value={row.original.url} size="icon-xs" autoHide />
        </div>
      ),
    },
    {
      accessorKey: 'disabled',
      header: 'Status',
      size: 100,
      cell: ({ row }) => (
        <Badge variant={row.original.disabled ? 'secondary' : 'success'}>
          {row.original.disabled ? 'Disabled' : 'Active'}
        </Badge>
      ),
    },
    {
      accessorKey: 'createdAt',
      header: 'Created',
      size: 150,
      cell: ({ row }) => {
        const createdAt = row.original.createdAt
        const relativeTime = getRelativeTimeString(createdAt).relativeTimeString

        return (
          <TimestampTooltip timestamp={typeof createdAt === 'string' ? createdAt : createdAt.toISOString()}>
            <span className="cursor-default">{relativeTime}</span>
          </TimestampTooltip>
        )
      },
    },
    {
      id: 'options',
      header: () => null,
      size: 50,
      cell: ({ row }) => {
        const isLoading = isLoadingEndpoint(row.original)

        return (
          <div className="flex justify-end" onClick={(e) => e.stopPropagation()}>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" className="h-8 w-8 p-0" disabled={isLoading}>
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem
                  onClick={() => onDisable(row.original)}
                  className="cursor-pointer"
                  disabled={isLoading}
                >
                  {row.original.disabled ? 'Enable' : 'Disable'}
                </DropdownMenuItem>
                <DropdownMenuItem
                  variant="destructive"
                  onClick={() => onDelete(row.original)}
                  className="cursor-pointer"
                  disabled={isLoading}
                >
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        )
      },
    },
  ]

  return columns
}
