/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CREATE_API_KEY_PERMISSIONS_GROUPS } from '@/constants/CreateApiKeyPermissionsGroups'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { getRelativeTimeString } from '@/lib/utils'
import { ApiKeyList, ApiKeyListPermissionsEnum, CreateApiKeyPermissionsEnum } from '@daytonaio/api-client'

import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { KeyRound, Loader2 } from 'lucide-react'
import { useMemo, useState } from 'react'
import { Pagination } from './Pagination'
import { TableEmptyState } from './TableEmptyState'
import { Badge } from './ui/badge'
import { Button } from './ui/button'
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
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover'
import { Skeleton } from './ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table'
import { Tooltip, TooltipContent, TooltipTrigger } from './ui/tooltip'

interface DataTableProps {
  data: ApiKeyList[]
  loading: boolean
  isLoadingKey: (key: ApiKeyList) => boolean
  onRevoke: (key: ApiKeyList) => void
}

export function ApiKeyTable({ data, loading, isLoadingKey, onRevoke }: DataTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const columns = getColumns({ onRevoke, isLoadingKey })
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
              <>
                {Array.from(new Array(5)).map((_, i) => (
                  <TableRow key={i}>
                    {table.getVisibleLeafColumns().map((column, i, arr) =>
                      i === arr.length - 1 ? null : (
                        <TableCell key={column.id}>
                          <Skeleton className="h-4 w-10/12" />
                        </TableCell>
                      ),
                    )}
                  </TableRow>
                ))}
              </>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && 'selected'}
                  className={`${isLoadingKey(row.original) ? 'opacity-50 pointer-events-none' : ''}`}
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
  isLoadingKey,
}: {
  onRevoke: (key: ApiKeyList) => void
  isLoadingKey: (key: ApiKeyList) => boolean
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
        return <PermissionsTooltip permissions={row.original.permissions} availablePermissions={allPermissions} />
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
          <Tooltip>
            <TooltipTrigger>
              <span className="cursor-default">{relativeTime}</span>
            </TooltipTrigger>
            <TooltipContent>
              <p>{fullDate}</p>
            </TooltipContent>
          </Tooltip>
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
          return <span className="text-muted-foreground">{relativeTime}</span>
        }

        const fullDate = new Date(lastUsedAt).toLocaleString()

        return (
          <Tooltip>
            <TooltipTrigger>
              <span className="cursor-default">{relativeTime}</span>
            </TooltipTrigger>
            <TooltipContent>
              <p>{fullDate}</p>
            </TooltipContent>
          </Tooltip>
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
          return <span className="text-muted-foreground">{relativeTime}</span>
        }

        const fullDate = new Date(expiresAt).toLocaleString()
        const color = getExpiresAtColor(expiresAt)

        return (
          <Tooltip>
            <TooltipTrigger>
              <span className={`cursor-default ${color}`}>{relativeTime}</span>
            </TooltipTrigger>
            <TooltipContent>
              <p>{fullDate}</p>
            </TooltipContent>
          </Tooltip>
        )
      },
    },
    {
      id: 'actions',
      size: 80,
      cell: ({ row }) => {
        const isLoading = isLoadingKey(row.original)

        return (
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="ghost" size={isLoading ? 'icon-sm' : 'sm'} disabled={isLoading} title="Revoke Key">
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
                  <Button variant="destructive" onClick={() => onRevoke(row.original)}>
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

const allPermissions = Object.values(CreateApiKeyPermissionsEnum)

const IMPLICIT_READ_RESOURCES = ['Sandboxes', 'Snapshots', 'Registries', 'Regions']

function PermissionsTooltip({
  permissions,
  availablePermissions,
}: {
  permissions: ApiKeyListPermissionsEnum[]
  availablePermissions: CreateApiKeyPermissionsEnum[]
}) {
  const isFullAccess = allPermissions.length === permissions.length
  const isSingleResourceAccess = CREATE_API_KEY_PERMISSIONS_GROUPS.find(
    (group) =>
      group.permissions.length === permissions.length && group.permissions.every((p) => permissions.includes(p)),
  )

  const availableGroups = useMemo(() => {
    return CREATE_API_KEY_PERMISSIONS_GROUPS.map((group) => ({
      ...group,
      permissions: group.permissions.filter((p) => availablePermissions.includes(p)),
    })).filter((group) => group.permissions.length > 0)
  }, [availablePermissions])

  const badgeVariant = isFullAccess ? 'warning' : 'outline'
  const badgeText = isFullAccess ? 'Full' : isSingleResourceAccess ? isSingleResourceAccess.name : 'Restricted'

  return (
    <Popover>
      <PopoverTrigger>
        <Badge variant={badgeVariant} className="whitespace-nowrap">
          {badgeText} <span className="hidden xs:inline ml-1">Access</span>
        </Badge>
      </PopoverTrigger>
      <PopoverContent className="p-0">
        <p className="p-2 text-muted-foreground text-xs font-medium border-b">Permissions</p>
        <div className="flex flex-col">
          {availableGroups.map((group) => {
            const selectedPermissions = group.permissions.filter((p) => permissions.includes(p))
            const hasImplicitRead = IMPLICIT_READ_RESOURCES.includes(group.name)

            if (selectedPermissions.length === 0 && !hasImplicitRead) {
              return null
            }

            return (
              <div key={group.name} className="flex justify-between gap-3 border-b last:border-b-0 p-2">
                <h3 className="text-sm">{group.name}</h3>
                <div className="flex gap-2 flex-wrap justify-end">
                  {hasImplicitRead && (
                    <Badge variant="outline" className="capitalize rounded-sm">
                      Read
                    </Badge>
                  )}
                  {selectedPermissions.map((p) => (
                    <Badge key={p} variant="outline" className="capitalize rounded-sm">
                      {p.split(':')[0]}
                    </Badge>
                  ))}
                </div>
              </div>
            )
          })}
        </div>
      </PopoverContent>
    </Popover>
  )
}
