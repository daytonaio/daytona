/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DEFAULT_SNAPSHOT_SORTING, SnapshotSorting } from '@/hooks/queries/useSnapshotsQuery'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { cn, getRelativeTimeString } from '@/lib/utils'
import {
  GetAllSnapshotsOrderEnum,
  GetAllSnapshotsSortEnum,
  OrganizationRolePermissionsEnum,
  SnapshotDto,
  SnapshotState,
} from '@daytonaio/api-client'
import { ColumnDef, flexRender, getCoreRowModel, SortingState, useReactTable } from '@tanstack/react-table'
import { AlertTriangle, Box, CheckCircle, Loader2, MoreHorizontal, Pause, Timer } from 'lucide-react'
import React, { useMemo, useState } from 'react'
import { Pagination } from './Pagination'
import { SortOrderIcon } from './SortIcon'
import { TableEmptyState } from './TableEmptyState'
import { Badge, BadgeProps } from './ui/badge'
import { Button } from './ui/button'
import { Checkbox } from './ui/checkbox'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from './ui/dropdown-menu'
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from './ui/tooltip'

interface SortableHeaderProps {
  column: any
  label: string
}

const SortableHeader: React.FC<SortableHeaderProps> = ({ column, label }) => {
  const sortDirection = column.getIsSorted()

  return (
    <button
      type="button"
      onClick={() => column.toggleSorting(sortDirection === 'asc')}
      className="group/sort-button flex items-center gap-2 w-full"
    >
      {label}
      <SortOrderIcon sort={sortDirection || null} />
    </button>
  )
}

const convertApiSortingToTableSorting = (sorting: SnapshotSorting): SortingState => {
  let id: string
  switch (sorting.field) {
    case GetAllSnapshotsSortEnum.NAME:
      id = 'name'
      break
    case GetAllSnapshotsSortEnum.STATE:
      id = 'state'
      break
    case GetAllSnapshotsSortEnum.CREATED_AT:
      id = 'createdAt'
      break
    case GetAllSnapshotsSortEnum.LAST_USED_AT:
    default:
      id = 'lastUsedAt'
      break
  }

  return [{ id, desc: sorting.direction === GetAllSnapshotsOrderEnum.DESC }]
}

const convertTableSortingToApiSorting = (sorting: SortingState): SnapshotSorting => {
  if (!sorting.length) {
    return DEFAULT_SNAPSHOT_SORTING
  }

  const sort = sorting[0]
  let field: GetAllSnapshotsSortEnum

  switch (sort.id) {
    case 'name':
      field = GetAllSnapshotsSortEnum.NAME
      break
    case 'state':
      field = GetAllSnapshotsSortEnum.STATE
      break
    case 'createdAt':
      field = GetAllSnapshotsSortEnum.CREATED_AT
      break
    case 'lastUsedAt':
    default:
      field = GetAllSnapshotsSortEnum.LAST_USED_AT
      break
  }

  return {
    field,
    direction: sort.desc ? GetAllSnapshotsOrderEnum.DESC : GetAllSnapshotsOrderEnum.ASC,
  }
}

interface DataTableProps {
  data: SnapshotDto[]
  loading: boolean
  loadingSnapshots: Record<string, boolean>
  getRegionName: (regionId: string) => string | undefined
  onDelete: (snapshot: SnapshotDto) => void
  onBulkDelete?: (snapshots: SnapshotDto[]) => void
  onActivate?: (snapshot: SnapshotDto) => void
  onDeactivate?: (snapshot: SnapshotDto) => void
  pagination: {
    pageIndex: number
    pageSize: number
  }
  pageCount: number
  totalItems: number
  onPaginationChange: (pagination: { pageIndex: number; pageSize: number }) => void
  sorting: SnapshotSorting
  onSortingChange: (sorting: SnapshotSorting) => void
}

export function SnapshotTable({
  data,
  loading,
  loadingSnapshots,
  getRegionName,
  onDelete,
  onActivate,
  onDeactivate,
  pagination,
  pageCount,
  totalItems,
  onBulkDelete,
  onPaginationChange,
  sorting,
  onSortingChange,
}: DataTableProps) {
  const { authenticatedUserHasPermission } = useSelectedOrganization()

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_SNAPSHOTS),
    [authenticatedUserHasPermission],
  )

  const deletePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_SNAPSHOTS),
    [authenticatedUserHasPermission],
  )

  const tableSorting = useMemo(() => convertApiSortingToTableSorting(sorting), [sorting])

  const columns = useMemo(
    () =>
      getColumns({
        onDelete,
        onActivate,
        onDeactivate,
        loadingSnapshots,
        getRegionName,
        writePermitted,
        deletePermitted,
      }),
    [onDelete, onActivate, onDeactivate, loadingSnapshots, getRegionName, writePermitted, deletePermitted],
  )

  const columnsWithSelection = useMemo(() => {
    const selectionColumn: ColumnDef<SnapshotDto> = {
      id: 'select',
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected() || (table.getIsSomePageRowsSelected() && 'indeterminate')}
          onCheckedChange={(value) => {
            table.getRowModel().rows.forEach((row) => {
              if (!row.original.general) {
                row.toggleSelected()
              }
            })
          }}
          aria-label="Select all"
          disabled={!deletePermitted || loading}
          className="translate-y-[2px]"
        />
      ),
      cell: ({ row }) => {
        if (loadingSnapshots[row.original.id]) {
          return <Loader2 className="w-4 h-4 animate-spin" />
        }

        if (row.original.general) {
          return null
        }

        return (
          <Checkbox
            checked={row.getIsSelected()}
            onCheckedChange={(value) => row.toggleSelected(!!value)}
            aria-label="Select row"
            disabled={!deletePermitted || loadingSnapshots[row.original.id] || loading}
            className="translate-y-[2px]"
          />
        )
      },
      enableSorting: false,
      enableHiding: false,
    }

    return deletePermitted ? [selectionColumn, ...columns] : columns
  }, [deletePermitted, columns, loading, loadingSnapshots])

  const table = useReactTable({
    data,
    columns: columnsWithSelection,
    getCoreRowModel: getCoreRowModel(),
    manualSorting: true,
    onSortingChange: (updater) => {
      const newTableSorting = typeof updater === 'function' ? updater(table.getState().sorting) : updater
      const newApiSorting = convertTableSortingToApiSorting(newTableSorting)
      onSortingChange(newApiSorting)
    },
    manualPagination: true,
    pageCount: pageCount || 1,
    onPaginationChange: pagination
      ? (updater) => {
          const newPagination = typeof updater === 'function' ? updater(table.getState().pagination) : updater
          onPaginationChange(newPagination)
        }
      : undefined,
    state: {
      sorting: tableSorting,
      pagination: {
        pageIndex: pagination?.pageIndex || 0,
        pageSize: pagination?.pageSize || 10,
      },
    },
    getRowId: (row) => row.id,
    enableRowSelection: deletePermitted,
  })

  const selectedRows = table.getSelectedRowModel().rows
  const [bulkDeleteConfirmationOpen, setBulkDeleteConfirmationOpen] = useState(false)
  const selectedImages = selectedRows.map((row) => row.original)

  const handleBulkDelete = () => {
    if (onBulkDelete && selectedImages.length > 0) {
      onBulkDelete(selectedImages)
    }
  }

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
                      className={cn('px-2', header.column.getCanSort() && 'hover:bg-muted cursor-pointer')}
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
                <TableCell colSpan={columnsWithSelection.length} className="h-24 text-center">
                  Loading...
                </TableCell>
              </TableRow>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() ? 'selected' : undefined}
                  className={`${
                    loadingSnapshots[row.original.id] || row.original.state === SnapshotState.REMOVING
                      ? 'opacity-50 pointer-events-none'
                      : ''
                  } ${row.original.general ? 'pointer-events-none' : ''}`}
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
                message="No Snapshots yet."
                icon={<Box className="w-8 h-8" />}
                description={
                  <div className="space-y-2">
                    <p>
                      Snapshots are reproducible, pre-configured environments based on any Docker-compatible image. Use
                      them to define language runtimes, dependencies, and tools for your sandboxes.
                    </p>
                    <p>
                      Create one from the Dashboard, CLI, or SDK to get started. <br />
                      <a
                        href="https://www.daytona.io/docs/snapshots"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-primary hover:underline font-medium"
                      >
                        Read the Snapshots guide
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
      <div className="flex items-center justify-between space-x-2 py-4">
        {deletePermitted && selectedRows.length > 0 && (
          <Popover open={bulkDeleteConfirmationOpen} onOpenChange={setBulkDeleteConfirmationOpen}>
            <PopoverTrigger>
              <Button variant="destructive" size="sm" className="h-8">
                Bulk Delete
              </Button>
            </PopoverTrigger>
            <PopoverContent side="top">
              <div className="flex flex-col gap-4">
                <p>Are you sure you want to delete these Snapshots?</p>
                <div className="flex items-center space-x-2">
                  <Button
                    variant="destructive"
                    onClick={() => {
                      handleBulkDelete()
                      setBulkDeleteConfirmationOpen(false)
                    }}
                  >
                    Delete
                  </Button>
                  <Button variant="outline" onClick={() => setBulkDeleteConfirmationOpen(false)}>
                    Cancel
                  </Button>
                </div>
              </div>
            </PopoverContent>
          </Popover>
        )}
        <Pagination table={table} selectionEnabled={deletePermitted} entityName="Snapshots" totalItems={totalItems} />
      </div>
    </div>
  )
}

const getColumns = ({
  onDelete,
  onActivate,
  onDeactivate,
  loadingSnapshots,
  getRegionName,
  writePermitted,
  deletePermitted,
}: {
  onDelete: (snapshot: SnapshotDto) => void
  onActivate?: (snapshot: SnapshotDto) => void
  onDeactivate?: (snapshot: SnapshotDto) => void
  loadingSnapshots: Record<string, boolean>
  getRegionName: (regionId: string) => string | undefined
  writePermitted: boolean
  deletePermitted: boolean
}): ColumnDef<SnapshotDto>[] => {
  const columns: ColumnDef<SnapshotDto>[] = [
    {
      accessorKey: 'name',
      enableSorting: true,
      header: ({ column }) => <SortableHeader column={column} label="Name" />,
      cell: ({ row }) => {
        const snapshot = row.original
        return (
          <div className="flex items-center gap-2">
            {snapshot.name}
            {snapshot.general && <Badge variant="secondary">System</Badge>}
          </div>
        )
      },
    },
    {
      accessorKey: 'imageName',
      enableSorting: false,
      header: 'Image',
      cell: ({ row }) => {
        const snapshot = row.original
        if (!snapshot.imageName && snapshot.buildInfo) {
          return (
            <Badge variant="secondary" className="rounded-sm px-1 font-medium">
              DECLARATIVE BUILD
            </Badge>
          )
        }
        return snapshot.imageName
      },
    },
    {
      accessorKey: 'regionIds',
      enableSorting: false,
      header: 'Region',
      cell: ({ row }) => {
        const snapshot = row.original
        if (!snapshot.regionIds?.length) {
          return '-'
        }

        const regionNames = snapshot.regionIds.map((id) => getRegionName(id) ?? id)
        const firstRegion = regionNames[0]
        const remainingCount = regionNames.length - 1

        if (remainingCount === 0) {
          return (
            <span className="truncate max-w-[150px] block" title={firstRegion}>
              {firstRegion}
            </span>
          )
        }

        return (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <div className="flex items-center gap-1.5">
                  <span className="truncate max-w-[150px]">{firstRegion}</span>
                  <Badge variant="secondary" className="text-xs px-1.5 py-0 h-5">
                    +{remainingCount}
                  </Badge>
                </div>
              </TooltipTrigger>
              <TooltipContent>
                <div className="flex flex-col gap-1">
                  {regionNames.map((name, idx) => (
                    <span key={idx}>{name}</span>
                  ))}
                </div>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        )
      },
    },
    {
      id: 'resources',
      enableSorting: false,
      header: 'Resources',
      cell: ({ row }) => {
        const snapshot = row.original

        return (
          <div className="flex items-center gap-2 w-full truncate">
            <div className="whitespace-nowrap">
              {snapshot.cpu} <span className="text-muted-foreground">vCPU</span>
            </div>
            <div className="w-[1px] h-6 bg-muted-foreground/20 rounded-full inline-block"></div>
            <div className="whitespace-nowrap">
              {snapshot.mem} <span className="text-muted-foreground">GiB</span>
            </div>
            <div className="w-[1px] h-6 bg-muted-foreground/20 rounded-full inline-block"></div>
            <div className="whitespace-nowrap">
              {snapshot.disk} <span className="text-muted-foreground">GiB</span>
            </div>
          </div>
        )
      },
    },
    {
      accessorKey: 'state',
      enableSorting: true,
      header: ({ column }) => <SortableHeader column={column} label="State" />,
      cell: ({ row }) => {
        const snapshot = row.original
        const variant = getStateBadgeVariant(snapshot.state)

        if (
          (snapshot.state === SnapshotState.ERROR || snapshot.state === SnapshotState.BUILD_FAILED) &&
          !!snapshot.errorReason
        ) {
          return (
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger>
                  <Badge variant={variant}>{getStateLabel(snapshot.state)}</Badge>
                </TooltipTrigger>
                <TooltipContent>
                  <p className="max-w-[300px]">{snapshot.errorReason}</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          )
        }

        return <Badge variant={variant}>{getStateLabel(snapshot.state)}</Badge>
      },
    },
    {
      accessorKey: 'createdAt',
      enableSorting: true,
      header: ({ column }) => <SortableHeader column={column} label="Created" />,
      cell: ({ row }) => {
        const snapshot = row.original
        return snapshot.general ? '' : getRelativeTimeString(snapshot.createdAt).relativeTimeString
      },
    },
    {
      accessorKey: 'lastUsedAt',
      enableSorting: true,
      header: ({ column }) => <SortableHeader column={column} label="Last Used" />,
      cell: ({ row }) => {
        const snapshot = row.original
        return snapshot.general ? '' : getRelativeTimeString(snapshot.lastUsedAt).relativeTimeString
      },
    },
    {
      id: 'actions',
      cell: ({ row }) => {
        if ((!writePermitted && !deletePermitted) || row.original.general) {
          return null
        }

        const showActivate = writePermitted && onActivate && row.original.state === SnapshotState.INACTIVE
        const showDeactivate = writePermitted && onDeactivate && row.original.state === SnapshotState.ACTIVE
        const showDelete = deletePermitted

        const showSeparator = (showActivate || showDeactivate) && showDelete

        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="h-8 w-8 p-0">
                <span className="sr-only">Open menu</span>
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {showActivate && (
                <DropdownMenuItem
                  onClick={() => onActivate(row.original)}
                  className="cursor-pointer"
                  disabled={loadingSnapshots[row.original.id]}
                >
                  Activate
                </DropdownMenuItem>
              )}
              {showDeactivate && (
                <DropdownMenuItem
                  onClick={() => onDeactivate(row.original)}
                  className="cursor-pointer"
                  disabled={loadingSnapshots[row.original.id]}
                >
                  Deactivate
                </DropdownMenuItem>
              )}
              {showSeparator && <DropdownMenuSeparator />}
              {showDelete && (
                <DropdownMenuItem
                  onClick={() => onDelete(row.original)}
                  className="cursor-pointer text-red-600 dark:text-red-400"
                  disabled={loadingSnapshots[row.original.id]}
                >
                  Delete
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        )
      },
    },
  ]

  return columns
}

const getStateIcon = (state: SnapshotState) => {
  switch (state) {
    case SnapshotState.ACTIVE:
      return <CheckCircle className="w-4 h-4 flex-shrink-0" />
    case SnapshotState.INACTIVE:
      return <Pause className="w-4 h-4 flex-shrink-0" />
    case SnapshotState.ERROR:
    case SnapshotState.BUILD_FAILED:
      return <AlertTriangle className="w-4 h-4 flex-shrink-0" />
    default:
      return <Timer className="w-4 h-4 flex-shrink-0" />
  }
}

const getStateBadgeVariant = (state: SnapshotState): BadgeProps['variant'] => {
  switch (state) {
    case SnapshotState.ACTIVE:
      return 'success'
    case SnapshotState.INACTIVE:
      return 'secondary'
    case SnapshotState.ERROR:
    case SnapshotState.BUILD_FAILED:
      return 'destructive'
    default:
      return 'secondary'
  }
}

const getStateLabel = (state: SnapshotState) => {
  // TODO: remove when removing is migrated to deleted
  if (state === SnapshotState.REMOVING) {
    return 'Deleting'
  }
  return state
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ')
}
