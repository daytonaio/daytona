/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CREATE_API_KEY_PERMISSIONS_GROUPS } from '@/constants/CreateApiKeyPermissionsGroups'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { cn, getRelativeTimeString } from '@/lib/utils'
import { getColumnSizeStyles } from '@/lib/utils/table'
import { ApiKeyList, ApiKeyListPermissionsEnum, CreateApiKeyPermissionsEnum } from '@daytona/api-client'

import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  Table as ReactTable,
  RowData,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { KeyRound, Loader2, MoreHorizontal } from 'lucide-react'
import { useMemo, useState } from 'react'
import { PageFooterPortal } from './PageLayout'
import { Pagination } from './Pagination'
import { SearchInput } from './SearchInput'
import { TimestampTooltip } from './TimestampTooltip'
import { Badge } from './ui/badge'
import { Button } from './ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from './ui/dropdown-menu'
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover'
import { Skeleton } from './ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableEmptyState,
  TableHead,
  TableHeader,
  TableRow,
} from './ui/table'

type ApiKeyTableMeta = {
  isLoadingKey: (key: ApiKeyList) => boolean
  onRevokeRequest: (key: ApiKeyList) => void
}

declare module '@tanstack/react-table' {
  interface TableMeta<TData extends RowData> {
    apiKey?: TData extends ApiKeyList ? ApiKeyTableMeta : never
  }
}

const getMeta = (table: ReactTable<ApiKeyList>) => {
  return table.options.meta?.apiKey as ApiKeyTableMeta
}

interface DataTableProps {
  data: ApiKeyList[]
  loading: boolean
  isLoadingKey: (key: ApiKeyList) => boolean
  onRevokeRequest: (key: ApiKeyList) => void
}

export function ApiKeyTable({ data, loading, isLoadingKey, onRevokeRequest }: DataTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const [globalFilter, setGlobalFilter] = useState('')
  const table = useReactTable({
    data,
    columns,
    meta: {
      apiKey: { isLoadingKey, onRevokeRequest },
    },
    defaultColumn: {
      minSize: 0,
    },
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    onGlobalFilterChange: setGlobalFilter,
    globalFilterFn: (row, _columnId, filterValue) => {
      const apiKey = row.original
      const searchValue = String(filterValue).toLowerCase()

      return (
        apiKey.name.toLowerCase().includes(searchValue) ||
        apiKey.permissions.some((permission) => permission.toLowerCase().includes(searchValue))
      )
    },
    state: {
      globalFilter,
      sorting,
    },
    initialState: {
      columnPinning: {
        left: ['name'],
        right: ['actions'],
      },
      pagination: {
        pageSize: DEFAULT_PAGE_SIZE,
      },
    },
  })

  const isEmpty = !loading && table.getRowModel().rows.length === 0
  const hasSearch = globalFilter.trim().length > 0

  const handleChangeFilter = (value: string) => {
    setGlobalFilter(value)
    table.setPageIndex(0)
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-3">
      <div>
        <SearchInput
          debounced
          value={globalFilter}
          onValueChange={handleChangeFilter}
          placeholder="Search by Name or Permission"
          containerClassName="max-w-sm"
        />
      </div>
      <TableContainer
        className={isEmpty ? 'min-h-[26rem]' : undefined}
        empty={
          isEmpty ? (
            <TableEmptyState
              overlay
              colSpan={columns.length}
              message={hasSearch ? 'No matching API Keys found.' : 'No API Keys yet.'}
              icon={<KeyRound />}
              description={
                hasSearch ? null : (
                  <div className="space-y-2">
                    <p>API Keys authenticate requests made through the Daytona SDK or CLI.</p>
                    <p>
                      Generate one and{' '}
                      <a
                        href="https://www.daytona.io/docs/api-keys"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-foreground hover:underline"
                      >
                        check out the API Key setup guide
                      </a>
                      .
                    </p>
                  </div>
                )
              }
              action={
                hasSearch ? (
                  <Button
                    variant="outline"
                    onClick={() => {
                      handleChangeFilter('')
                    }}
                  >
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
                {headerGroup.headers.map((header) => (
                  <TableHead
                    key={header.id}
                    sticky={header.column.getIsPinned()}
                    style={getColumnSizeStyles(header.column)}
                  >
                    {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {loading ? (
              <>
                {Array.from({ length: DEFAULT_PAGE_SIZE }).map((_, i) => (
                  <TableRow key={i}>
                    {table.getVisibleLeafColumns().map((column) => (
                      <TableCell key={column.id} sticky={column.getIsPinned()} style={getColumnSizeStyles(column)}>
                        <Skeleton className="h-4 w-10/12" />
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
              </>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && 'selected'}
                  className={cn({ 'opacity-50 pointer-events-none': isLoadingKey(row.original) })}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell
                      key={cell.id}
                      sticky={cell.column.getIsPinned()}
                      style={getColumnSizeStyles(cell.column)}
                    >
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
        <Pagination table={table} entityName="API Keys" />
      </PageFooterPortal>
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

const columns: ColumnDef<ApiKeyList>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
    size: 200,
  },
  {
    accessorKey: 'value',
    header: 'Key',
    size: 220,
    cell: ({ row }) => {
      return <div className="truncate">{row.original.value}</div>
    },
  },
  {
    accessorKey: 'permissions',
    size: 170,
    header: () => {
      return <div className="px-3">Permissions</div>
    },
    cell: ({ row }) => {
      return (
        <div className="flex min-w-0">
          <PermissionsTooltip permissions={row.original.permissions} availablePermissions={allPermissions} />
        </div>
      )
    },
  },
  {
    accessorKey: 'createdAt',
    header: 'Created',
    size: 120,
    cell: ({ row }) => {
      const createdAt = row.original.createdAt
      const relativeTime = getRelativeTimeString(createdAt).relativeTimeString

      return (
        <TimestampTooltip timestamp={createdAt?.toString()}>
          <span className="cursor-default">{relativeTime}</span>
        </TimestampTooltip>
      )
    },
  },
  {
    accessorKey: 'lastUsedAt',
    header: 'Last Used',
    size: 140,
    cell: ({ row }) => {
      const lastUsedAt = row.original.lastUsedAt
      const relativeTime = getRelativeTimeString(lastUsedAt).relativeTimeString

      if (!lastUsedAt) {
        return <span className="text-muted-foreground">{relativeTime}</span>
      }

      return (
        <TimestampTooltip timestamp={lastUsedAt?.toString()}>
          <span className="cursor-default">{relativeTime}</span>
        </TimestampTooltip>
      )
    },
  },
  {
    accessorKey: 'expiresAt',
    header: 'Expires',
    size: 150,
    cell: ({ row }) => {
      const expiresAt = row.original.expiresAt
      const relativeTime = getRelativeTimeString(expiresAt).relativeTimeString

      if (!expiresAt) {
        return <span className="text-muted-foreground">{relativeTime}</span>
      }

      const color = getExpiresAtColor(expiresAt)

      return (
        <TimestampTooltip timestamp={expiresAt?.toString()}>
          <span className={`cursor-default ${color}`}>{relativeTime}</span>
        </TimestampTooltip>
      )
    },
  },
  {
    id: 'actions',
    header: () => null,
    size: 48,
    minSize: 48,
    maxSize: 48,
    cell: ({ row, table }) => {
      const { isLoadingKey, onRevokeRequest } = getMeta(table)
      const isLoading = isLoadingKey(row.original)

      return (
        <div className="flex justify-end">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon-sm" aria-label="Open menu" disabled={isLoading}>
                {isLoading ? <Loader2 className="size-4 animate-spin" /> : <MoreHorizontal />}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                variant="destructive"
                onClick={() => onRevokeRequest(row.original)}
                disabled={isLoading}
              >
                Revoke
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      )
    },
  },
]

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
