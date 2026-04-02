/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CREATE_API_KEY_PERMISSIONS_GROUPS } from '@/constants/CreateApiKeyPermissionsGroups'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { cn, getRelativeTimeString } from '@/lib/utils'
import {
  getColumnPinningBorderClasses,
  getColumnPinningClasses,
  getColumnPinningStyles,
  getExplicitColumnSize,
} from '@/lib/utils/table'
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
import { KeyRound, Loader2, MoreHorizontal } from 'lucide-react'
import { useMemo, useState } from 'react'
import { PageFooterPortal } from './PageLayout'
import { Pagination } from './Pagination'
import { TableEmptyState } from './TableEmptyState'
import { Badge } from './ui/badge'
import { Button } from './ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from './ui/dropdown-menu'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from './ui/alert-dialog'
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover'
import { Skeleton } from './ui/skeleton'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow } from './ui/table'
import { Tooltip, TooltipContent, TooltipTrigger } from './ui/tooltip'

interface DataTableProps {
  data: ApiKeyList[]
  loading: boolean
  isLoadingKey: (key: ApiKeyList) => boolean
  onRevoke: (key: ApiKeyList) => void
}

const FIXED_COLUMN_IDS = ['actions']

export function ApiKeyTable({ data, loading, isLoadingKey, onRevoke }: DataTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const [apiKeyToRevoke, setApiKeyToRevoke] = useState<ApiKeyList | null>(null)
  const columns = getColumns({ isLoadingKey, onRevokeRequest: setApiKeyToRevoke })
  const table = useReactTable({
    data,
    columns,
    defaultColumn: {
      minSize: 0,
    },
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    state: {
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
  const leftPinnedCount = table.getLeftLeafColumns().length
  const revokeLoading = apiKeyToRevoke ? isLoadingKey(apiKeyToRevoke) : false

  const handleConfirmRevoke = async () => {
    if (!apiKeyToRevoke) {
      return
    }

    await onRevoke(apiKeyToRevoke)
    setApiKeyToRevoke(null)
  }

  return (
    <>
      <div className="flex min-h-0 flex-1 flex-col pt-2">
        <TableContainer
          className={isEmpty ? 'min-h-[26rem]' : undefined}
          empty={
            isEmpty ? (
              <TableEmptyState
                overlay
                colSpan={columns.length}
                message="No API Keys yet."
                icon={<KeyRound className="h-4 w-4" />}
                description={
                  <div className="space-y-2">
                    <p>API Keys authenticate requests made through the Daytona SDK or CLI.</p>
                    <p>
                      Generate one and{' '}
                      <a href="https://www.daytona.io/docs/api-keys" target="_blank" rel="noopener noreferrer">
                        check out the API Key setup guide
                      </a>
                      .
                    </p>
                  </div>
                }
              />
            ) : undefined
          }
        >
          <Table>
            <TableHeader>
              {table.getHeaderGroups().map((headerGroup) => (
                <TableRow key={headerGroup.id}>
                  {headerGroup.headers.map((header, headerIndex) => {
                    return (
                      <TableHead
                        className={cn(
                          {
                            '!px-2': header.column.id === 'actions',
                          },
                          !isEmpty && getColumnPinningBorderClasses(header.column, leftPinnedCount, headerIndex),
                          !isEmpty && getColumnPinningClasses(header.column, true),
                        )}
                        key={header.id}
                        style={
                          isEmpty
                            ? undefined
                            : {
                                ...getExplicitColumnSize(header),
                                ...getColumnPinningStyles(header.column, FIXED_COLUMN_IDS),
                              }
                        }
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
                  {Array.from({ length: 25 }).map((_, i) => (
                    <TableRow key={i}>
                      {table.getVisibleLeafColumns().map((column, colIndex) => (
                        <TableCell
                          key={column.id}
                          className={cn(
                            {
                              '!px-2': column.id === 'actions',
                            },
                            getColumnPinningBorderClasses(column, leftPinnedCount, colIndex),
                            getColumnPinningClasses(column),
                          )}
                          style={getColumnPinningStyles(column, FIXED_COLUMN_IDS)}
                        >
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
                    className={`${isLoadingKey(row.original) ? 'opacity-50 pointer-events-none' : ''}`}
                  >
                    {row.getVisibleCells().map((cell, cellIndex) => (
                      <TableCell
                        className={cn(
                          {
                            '!px-2': cell.column.id === 'actions',
                          },
                          getColumnPinningBorderClasses(cell.column, leftPinnedCount, cellIndex),
                          getColumnPinningClasses(cell.column),
                        )}
                        key={cell.id}
                        style={getColumnPinningStyles(cell.column, FIXED_COLUMN_IDS)}
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

      <AlertDialog open={!!apiKeyToRevoke} onOpenChange={(open) => !open && setApiKeyToRevoke(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Revoke API Key</AlertDialogTitle>
            <AlertDialogDescription>
              {apiKeyToRevoke
                ? `Are you sure you want to revoke the API key "${apiKeyToRevoke.name}"? This action cannot be undone.`
                : 'Are you sure you want to revoke this API key? This action cannot be undone.'}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={revokeLoading}>Cancel</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={handleConfirmRevoke} disabled={revokeLoading}>
              Revoke
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
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
  isLoadingKey,
  onRevokeRequest,
}: {
  isLoadingKey: (key: ApiKeyList) => boolean
  onRevokeRequest: (key: ApiKeyList) => void
}): ColumnDef<ApiKeyList>[] => {
  const columns: ColumnDef<ApiKeyList>[] = [
    {
      accessorKey: 'name',
      header: 'Name',
      size: 220,
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
      header: () => null,
      size: 48,
      minSize: 48,
      maxSize: 48,
      cell: ({ row }) => {
        const isLoading = isLoadingKey(row.original)

        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="h-8 w-8 p-0" disabled={isLoading}>
                {isLoading ? <Loader2 className="size-4 animate-spin" /> : <MoreHorizontal className="size-4" />}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                variant="destructive"
                onClick={() => onRevokeRequest(row.original)}
                className="cursor-pointer"
                disabled={isLoading}
              >
                Revoke
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
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
