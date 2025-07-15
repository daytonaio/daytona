/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiKeyList } from '@daytonaio/api-client'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { TableHeader, TableRow, TableHead, TableBody, TableCell, Table } from './ui/table'
import { Button } from './ui/button'
import { useState } from 'react'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from './ui/dialog'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from './ui/tooltip'
import { Pagination } from './Pagination'
import { Loader2, KeyRound } from 'lucide-react'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { getRelativeTimeString } from '@/lib/utils'
import { TableEmptyState } from './TableEmptyState'

interface DataTableProps {
  data: ApiKeyList[]
  loading: boolean
  loadingKeys: Record<string, boolean>
  onRevoke: (keyName: string) => void
}

export function ApiKeyTable({ data, loading, loadingKeys, onRevoke }: DataTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const columns = getColumns({ onRevoke, loadingKeys })
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    state: {
      sorting,
    },
    initialState: {
      pagination: {
        pageSize: DEFAULT_PAGE_SIZE,
      },
    },
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
                    <TableHead className="px-2" key={header.id}>
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
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && 'selected'}
                  className={`${loadingKeys[row.original.name] ? 'opacity-50 pointer-events-none' : ''}`}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell className="px-2" key={cell.id}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableEmptyState
                colSpan={columns.length}
                message="No API Keys yet."
                icon={<KeyRound className="w-8 h-8" />}
                description={
                  <div className="space-y-2">
                    <p>API Keys authenticate requests made through the Daytona SDK or CLI.</p>
                    <p>
                      Generate one and{' '}
                      <a
                        href="https://www.daytona.io/docs/api-keys"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-primary hover:underline font-medium"
                      >
                        check out the API Key setup guide
                      </a>
                      .
                    </p>
                  </div>
                }
              />
            )}
          </TableBody>
        </Table>
      </div>
      <Pagination table={table} className="mt-4" entityName="API Keys" />
    </div>
  )
}

const getExpiresAtColor = (expiresAt: Date | null) => {
  if (!expiresAt) {
    return 'text-foreground'
  }

  const MILLISECONDS_IN_MINUTE = 1000 * 60
  const MINUTES_IN_DAY = 24 * 60

  const diffInMinutes = Math.floor((new Date(expiresAt).getTime() - new Date().getTime()) / MILLISECONDS_IN_MINUTE)

  // Already expired
  if (diffInMinutes < 0) {
    return 'text-red-500'
  }

  // Expires within a day
  if (diffInMinutes < MINUTES_IN_DAY) {
    return 'text-yellow-600 dark:text-yellow-400'
  }

  // Expires in more than a day
  return 'text-foreground'
}

const getColumns = ({
  onRevoke,
  loadingKeys,
}: {
  onRevoke: (keyName: string) => void
  loadingKeys: Record<string, boolean>
}): ColumnDef<ApiKeyList>[] => {
  const columns: ColumnDef<ApiKeyList>[] = [
    {
      accessorKey: 'name',
      header: 'Name',
    },
    {
      accessorKey: 'value',
      header: 'Key',
    },
    {
      accessorKey: 'permissions',
      header: () => {
        return <div className="max-w-md px-3">Permissions</div>
      },
      cell: ({ row }) => {
        const permissions = row.original.permissions.join(', ')
        return (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger>
                <div className="truncate max-w-md px-3 cursor-text">{permissions || '-'}</div>
              </TooltipTrigger>
              {permissions && (
                <TooltipContent>
                  <p className="max-w-[300px]">{permissions}</p>
                </TooltipContent>
              )}
            </Tooltip>
          </TooltipProvider>
        )
      },
    },
    {
      accessorKey: 'createdAt',
      header: 'Created',
      cell: ({ row }) => {
        const createdAt = row.original.createdAt
        const relativeTime = getRelativeTimeString(createdAt).relativeTimeString
        const fullDate = new Date(createdAt).toLocaleString()

        return (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger>
                <span className="cursor-default">{relativeTime}</span>
              </TooltipTrigger>
              <TooltipContent>
                <p>{fullDate}</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        )
      },
    },
    {
      accessorKey: 'lastUsedAt',
      header: 'Last Used',
      cell: ({ row }) => {
        const lastUsedAt = row.original.lastUsedAt
        const relativeTime = getRelativeTimeString(lastUsedAt).relativeTimeString

        if (!lastUsedAt) {
          return relativeTime
        }

        const fullDate = new Date(lastUsedAt).toLocaleString()

        return (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger>
                <span className="cursor-default">{relativeTime}</span>
              </TooltipTrigger>
              <TooltipContent>
                <p>{fullDate}</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        )
      },
    },
    {
      accessorKey: 'expiresAt',
      header: 'Expires',
      cell: ({ row }) => {
        const expiresAt = row.original.expiresAt
        const relativeTime = getRelativeTimeString(expiresAt).relativeTimeString

        if (!expiresAt) {
          return relativeTime
        }

        const fullDate = new Date(expiresAt).toLocaleString()
        const color = getExpiresAtColor(expiresAt)

        return (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger>
                <span className={`cursor-default ${color}`}>{relativeTime}</span>
              </TooltipTrigger>
              <TooltipContent>
                <p>{fullDate}</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        )
      },
    },
    {
      id: 'actions',
      header: () => {
        return <div className="px-4">Actions</div>
      },
      cell: ({ row }) => {
        const isLoading = loadingKeys[row.original.name]

        return (
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="ghost" size="icon" disabled={isLoading} className="w-20" title="Revoke Key">
                {isLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Revoke'}
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Confirm Key Revocation</DialogTitle>
                <DialogDescription>
                  Are you absolutely sure? This action cannot be undone. This will permanently delete this API key.
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <DialogClose asChild>
                  <Button type="button" variant="secondary">
                    Close
                  </Button>
                </DialogClose>
                <DialogClose asChild>
                  <Button variant="destructive" onClick={() => onRevoke(row.original.name)}>
                    Revoke
                  </Button>
                </DialogClose>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        )
      },
    },
  ]

  return columns
}
