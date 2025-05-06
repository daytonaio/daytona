/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationRolePermissionsEnum, Workspace, WorkspaceState } from '@daytonaio/api-client'
import {
  ColumnDef,
  ColumnFiltersState,
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import {
  Loader2,
  Terminal,
  AlertTriangle,
  MoreHorizontal,
  ArrowUp,
  ArrowDown,
  Circle,
  CheckCircle,
  Timer,
  ArrowUpDown,
  Archive,
} from 'lucide-react'
import { TableHeader, TableRow, TableHead, TableBody, TableCell, Table } from './ui/table'
import { Button } from './ui/button'
import { useEffect, useMemo, useState } from 'react'
import { Checkbox } from './ui/checkbox'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from './ui/dropdown-menu'
import { Pagination } from './Pagination'
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover'
import { getRelativeTimeString } from '@/lib/utils'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from './ui/tooltip'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { DebouncedInput } from './DebouncedInput'
import { DataTableFacetedFilter, FacetedFilterOption } from './ui/data-table-faceted-filter'

interface DataTableProps {
  data: Workspace[]
  loadingWorkspaces: Record<string, boolean>
  loading: boolean
  handleStart: (id: string) => void
  handleStop: (id: string) => void
  handleDelete: (id: string) => void
  handleBulkDelete: (ids: string[]) => void
  handleArchive: (id: string) => void
}

export function WorkspaceTable({
  data,
  loadingWorkspaces,
  loading,
  handleStart,
  handleStop,
  handleDelete,
  handleBulkDelete,
  handleArchive,
}: DataTableProps) {
  const { authenticatedUserHasPermission } = useSelectedOrganization()

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_SANDBOXES),
    [authenticatedUserHasPermission],
  )

  const deletePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_SANDBOXES),
    [authenticatedUserHasPermission],
  )

  const [sorting, setSorting] = useState<SortingState>([])
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])

  const labelOptions: FacetedFilterOption[] = useMemo(() => {
    const labels = new Set<string>()
    data.forEach((workspace) => {
      Object.entries(workspace.labels ?? {}).forEach(([key, value]) => {
        labels.add(`${key}: ${value}`)
      })
    })
    return Array.from(labels).map((label) => ({ label, value: label }))
  }, [data])

  const columns = getColumns({
    handleStart,
    handleStop,
    handleDelete,
    handleArchive,
    loadingWorkspaces,
    writePermitted,
    deletePermitted,
  })
  const table = useReactTable({
    data,
    columns,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    getFilteredRowModel: getFilteredRowModel(),
    state: {
      sorting,
      columnFilters,
    },
    enableRowSelection: true,
    getRowId: (row) => row.id,
    onRowSelectionChange: (updater) => {
      // Since table is defined after this function, we need to access it via function
      // to avoid the circular reference
      const getTableState = () => table.getState()
      const getFilteredRowModel = () => table.getFilteredRowModel()

      // Check if any rows are selected
      const currentState = getTableState().rowSelection
      const newState = typeof updater === 'function' ? updater(currentState) : updater

      // If we're selecting rows and we have a reasonable number of rows, show the banner
      const selectedCount = Object.keys(newState).length
      const totalFilteredCount = getFilteredRowModel().rows.length

      if (selectedCount > 0 && selectedCount < totalFilteredCount) {
        setShowSelectAllBanner(true)
      } else {
        setShowSelectAllBanner(false)
      }

      // Don't call table.setRowSelection here, as the table will handle it internally
      return newState
    },
  })
  const [bulkDeleteConfirmationOpen, setBulkDeleteConfirmationOpen] = useState(false)
  const [showSelectAllBanner, setShowSelectAllBanner] = useState(false)

  const handleSelectAll = () => {
    const allFilteredRows = table.getFilteredRowModel().rows
    const newSelection: Record<string, boolean> = {}

    allFilteredRows.forEach((row) => {
      newSelection[row.id] = true
    })

    table.setRowSelection(newSelection)
    setShowSelectAllBanner(false)
  }

  const selectedCount = Object.keys(table.getState().rowSelection).length
  const totalFilteredCount = table.getFilteredRowModel().rows.length

  return (
    <div>
      <div className="flex items-center mb-4">
        <DebouncedInput
          value={(table.getColumn('id')?.getFilterValue() as string) ?? ''}
          onChange={(value) => table.getColumn('id')?.setFilterValue(value)}
          placeholder="Search..."
          className="max-w-sm mr-4"
        />
        {table.getColumn('state') && (
          <DataTableFacetedFilter column={table.getColumn('state')} title="State" options={statuses} />
        )}
        {table.getColumn('labels') && (
          <DataTableFacetedFilter
            className="ml-4"
            column={table.getColumn('labels')}
            title="Labels"
            options={labelOptions}
          />
        )}
      </div>

      {/* Selection banner */}
      {showSelectAllBanner && (
        <div className="flex items-center justify-between bg-muted p-2 rounded-t-md border border-b-0">
          <div className="flex items-center">
            <span className="font-medium">
              {selectedCount} {selectedCount === 1 ? 'Sandbox' : 'Sandboxes'} selected
            </span>
            <button className="ml-4 text-primary hover:underline text-sm font-medium" onClick={handleSelectAll}>
              Select all {totalFilteredCount} Sandboxes
            </button>
          </div>
          <Button variant="ghost" size="sm" onClick={() => table.resetRowSelection()} className="h-8">
            Clear selection
          </Button>
        </div>
      )}

      <div className={`rounded-md border ${showSelectAllBanner ? 'rounded-t-none' : ''}`}>
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead key={header.id}>
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
                  className={`${loadingWorkspaces[row.original.id] || row.original.state === WorkspaceState.DESTROYING ? 'opacity-50 pointer-events-none' : ''}`}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              !loading && (
                <TableRow>
                  <TableCell colSpan={columns.length} className="h-24 text-center">
                    No results.
                  </TableCell>
                </TableRow>
              )
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-between space-x-2 py-4">
        <div className="flex items-center space-x-2">
          {selectedCount > 0 && (
            <div className="flex items-center space-x-2">
              <Popover open={bulkDeleteConfirmationOpen} onOpenChange={setBulkDeleteConfirmationOpen}>
                <PopoverTrigger>
                  <Button variant="destructive" size="sm" className="h-8">
                    Bulk Delete ({selectedCount})
                  </Button>
                </PopoverTrigger>
                <PopoverContent side="top">
                  <div className="flex flex-col gap-4">
                    <p>Are you sure you want to delete these workspaces?</p>
                    <div className="flex items-center space-x-2">
                      <Button
                        variant="destructive"
                        onClick={() => {
                          handleBulkDelete(Object.keys(table.getState().rowSelection))
                          setBulkDeleteConfirmationOpen(false)
                          table.resetRowSelection()
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
            </div>
          )}
        </div>
        <Pagination table={table} selectionEnabled />
      </div>
    </div>
  )
}

const getStateIcon = (state?: WorkspaceState) => {
  switch (state) {
    case WorkspaceState.STARTED:
      return <CheckCircle className="w-4 h-4" />
    case WorkspaceState.STOPPED:
      return <Circle className="w-4 h-4" />
    case WorkspaceState.ERROR:
      return <AlertTriangle className="w-4 h-4" />
    case WorkspaceState.CREATING:
    case WorkspaceState.STARTING:
    case WorkspaceState.STOPPING:
    case WorkspaceState.DESTROYING:
    case WorkspaceState.ARCHIVING:
      return <Timer className="w-4 h-4" />
    case WorkspaceState.ARCHIVED:
      return <Archive className="w-4 h-4" />
    default:
      return null
  }
}

const getLastEvent = (workspace: Workspace): { date: Date; relativeTimeString: string } => {
  const parsed = getProviderMetadata(workspace.info?.providerMetadata)
  return getRelativeTimeString(parsed?.updatedAt)
}

const getCreatedAt = (workspace: Workspace): { date: Date; relativeTimeString: string } => {
  return getRelativeTimeString(workspace.info?.created)
}

const getProviderMetadata = (metadata: string | undefined) => {
  if (!metadata) return null
  try {
    return JSON.parse(metadata)
  } catch (e) {
    console.error('Error parsing provider metadata:', e)
    return null
  }
}

const getProviderClass = (workspace: Workspace): string => {
  const parsed = getProviderMetadata(workspace.info?.providerMetadata)
  return parsed?.class || 'unknown'
}

const getNodeDomain = (metadata: string | undefined): string | null => {
  const parsed = getProviderMetadata(metadata)
  return parsed?.nodeDomain || null
}

const getStateColor = (state?: WorkspaceState) => {
  switch (state) {
    case WorkspaceState.STARTED:
      return 'text-green-500'
    case WorkspaceState.STOPPED:
      return 'text-gray-500'
    case WorkspaceState.ERROR:
      return 'text-red-500'
    default:
      return 'text-gray-600 dark:text-gray-400'
  }
}

const getStateLabel = (state?: WorkspaceState) => {
  if (!state) {
    return 'Unknown'
  }
  // TODO: remove when destroying/destroyed is migrated to deleting/deleted
  if (state === WorkspaceState.DESTROYING) {
    return 'Deleting'
  }
  return state.charAt(0).toUpperCase() + state.slice(1)
}

const statuses: FacetedFilterOption[] = [
  { label: getStateLabel(WorkspaceState.STARTED), value: WorkspaceState.STARTED, icon: CheckCircle },
  { label: getStateLabel(WorkspaceState.STOPPED), value: WorkspaceState.STOPPED, icon: Circle },
  { label: getStateLabel(WorkspaceState.ERROR), value: WorkspaceState.ERROR, icon: AlertTriangle },
  { label: getStateLabel(WorkspaceState.STARTING), value: WorkspaceState.STARTING, icon: Timer },
  { label: getStateLabel(WorkspaceState.STOPPING), value: WorkspaceState.STOPPING, icon: Timer },
  { label: getStateLabel(WorkspaceState.DESTROYING), value: WorkspaceState.DESTROYING, icon: Timer },
  { label: getStateLabel(WorkspaceState.ARCHIVING), value: WorkspaceState.ARCHIVING, icon: Timer },
  { label: getStateLabel(WorkspaceState.ARCHIVED), value: WorkspaceState.ARCHIVED, icon: Archive },
]

const getColumns = ({
  handleStart,
  handleStop,
  handleDelete,
  handleArchive,
  loadingWorkspaces,
  writePermitted,
  deletePermitted,
}: {
  handleStart: (id: string) => void
  handleStop: (id: string) => void
  handleDelete: (id: string) => void
  handleArchive: (id: string) => void
  loadingWorkspaces: Record<string, boolean>
  writePermitted: boolean
  deletePermitted: boolean
}): ColumnDef<Workspace>[] => {
  const columns: ColumnDef<Workspace>[] = [
    {
      id: 'select',
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected() || (table.getIsSomePageRowsSelected() && 'indeterminate')}
          onCheckedChange={(value) => {
            for (const row of table.getRowModel().rows) {
              if (loadingWorkspaces[row.original.id]) {
                row.toggleSelected(false)
              } else {
                row.toggleSelected(!!value)
              }
            }
          }}
          aria-label="Select all"
          className="translate-y-[2px]"
        />
      ),
      cell: ({ row }) => {
        if (loadingWorkspaces[row.original.id]) {
          return <Loader2 className="w-4 h-4 animate-spin" />
        }
        return (
          <Checkbox
            checked={row.getIsSelected()}
            onCheckedChange={(value) => row.toggleSelected(!!value)}
            aria-label="Select row"
            className="translate-y-[2px]"
          />
        )
      },
      enableSorting: false,
      enableHiding: false,
    },
    {
      id: 'id',
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
            className="px-2 hover:bg-muted/50"
          >
            ID
            {column.getIsSorted() === 'asc' ? (
              <ArrowUp className="ml-2 h-4 w-4" />
            ) : column.getIsSorted() === 'desc' ? (
              <ArrowDown className="ml-2 h-4 w-4" />
            ) : (
              <ArrowUpDown className="ml-2 h-4 w-4" />
            )}
          </Button>
        )
      },
      accessorKey: 'id',
      cell: ({ row }) => {
        return <span className="px-2">{row.original.id}</span>
      },
    },
    {
      id: 'state',
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
            className="px-2 hover:bg-muted/50"
          >
            State
            {column.getIsSorted() === 'asc' ? (
              <ArrowUp className="ml-2 h-4 w-4" />
            ) : column.getIsSorted() === 'desc' ? (
              <ArrowDown className="ml-2 h-4 w-4" />
            ) : (
              <ArrowUpDown className="ml-2 h-4 w-4" />
            )}
          </Button>
        )
      },
      cell: ({ row }) => {
        const workspace = row.original
        const state = row.original.state
        const color = getStateColor(state)

        if (state === WorkspaceState.ERROR && !!workspace.errorReason) {
          return (
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger>
                  <div className={`flex items-center gap-2 px-2 ${color}`}>
                    {getStateIcon(state)}
                    {getStateLabel(state)}
                  </div>
                </TooltipTrigger>
                <TooltipContent>
                  <p className="max-w-[300px]">{workspace.errorReason}</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          )
        }

        return (
          <div className={`flex items-center gap-2 px-2 ${color}`}>
            {getStateIcon(state)}
            <span>{getStateLabel(state)}</span>
          </div>
        )
      },
      accessorKey: 'state',
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id))
      },
    },
    {
      id: 'region',
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
            className="px-2 hover:bg-muted/50"
          >
            Region
            {column.getIsSorted() === 'asc' ? (
              <ArrowUp className="ml-2 h-4 w-4" />
            ) : column.getIsSorted() === 'desc' ? (
              <ArrowDown className="ml-2 h-4 w-4" />
            ) : (
              <ArrowUpDown className="ml-2 h-4 w-4" />
            )}
          </Button>
        )
      },
      cell: ({ row }) => {
        return <span className="px-2">{row.original.target}</span>
      },
      accessorKey: 'target',
    },
    // {
    //   id: 'class',
    //   header: ({ column }) => {
    //     return (
    //       <Button
    //         variant="ghost"
    //         onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
    //         className="px-2 hover:bg-muted/50"
    //       >
    //         Class
    //         {column.getIsSorted() === 'asc' ? (
    //           <ArrowUp className="ml-2 h-4 w-4" />
    //         ) : column.getIsSorted() === 'desc' ? (
    //           <ArrowDown className="ml-2 h-4 w-4" />
    //         ) : (
    //           <ArrowUpDown className="ml-2 h-4 w-4" />
    //         )}
    //       </Button>
    //     )
    //   },
    //   cell: ({ row }) => {
    //     return <span className="px-2">{getProviderClass(row.original)}</span>
    //   },
    //   accessorFn: (row) => getProviderClass(row),
    // },
    {
      id: 'labels',
      header: () => {
        return <span className="px-2">Labels</span>
      },
      cell: ({ row }) => {
        const labels = Object.entries(row.original.labels ?? {})
          .map(([key, value]) => `${key}: ${value}`)
          .join(', ')
        return (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger>
                <div className="truncate max-w-md px-2 cursor-text">{labels || '-'}</div>
              </TooltipTrigger>
              {labels && (
                <TooltipContent>
                  <p className="max-w-[300px]">{labels}</p>
                </TooltipContent>
              )}
            </Tooltip>
          </TooltipProvider>
        )
      },
      accessorFn: (row) => Object.entries(row.labels ?? {}).map(([key, value]) => `${key}: ${value}`),
      filterFn: (row, id, value) => {
        return value.some((label: string) => (row.getValue(id) as string).includes(label))
      },
    },
    {
      id: 'lastEvent',
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
            className="px-2 hover:bg-muted/50"
          >
            Last Event
            {column.getIsSorted() === 'asc' ? (
              <ArrowUp className="ml-2 h-4 w-4" />
            ) : column.getIsSorted() === 'desc' ? (
              <ArrowDown className="ml-2 h-4 w-4" />
            ) : (
              <ArrowUpDown className="ml-2 h-4 w-4" />
            )}
          </Button>
        )
      },
      accessorFn: (row) => getLastEvent(row).date,
      cell: ({ row }) => {
        return <span className="px-2">{getLastEvent(row.original).relativeTimeString}</span>
      },
    },
    {
      id: 'createdAt',
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
            className="px-2 hover:bg-muted/50"
          >
            Created
            {column.getIsSorted() === 'asc' ? (
              <ArrowUp className="ml-2 h-4 w-4" />
            ) : column.getIsSorted() === 'desc' ? (
              <ArrowDown className="ml-2 h-4 w-4" />
            ) : (
              <ArrowUpDown className="ml-2 h-4 w-4" />
            )}
          </Button>
        )
      },
      accessorFn: (row) => getCreatedAt(row).date,
      cell: ({ row }) => {
        return <span className="px-2">{getCreatedAt(row.original).relativeTimeString}</span>
      },
    },
    {
      id: 'access',
      header: 'Access',
      cell: ({ row }) => {
        const nodeDomain = getNodeDomain(row.original.info?.providerMetadata)
        if (!nodeDomain || row.original.state !== WorkspaceState.STARTED) return ''
        return (
          <a href={`https://22222-${row.original.id}.${nodeDomain}`} target="_blank" rel="noopener noreferrer">
            <Terminal className="w-4 h-4" />
          </a>
        )
      },
    },
    {
      id: 'actions',
      enableHiding: false,
      cell: ({ row }) => {
        if (!writePermitted && !deletePermitted) {
          return null
        }

        const workspace = row.original

        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="h-8 w-8 p-0">
                <span className="sr-only">Open menu</span>
                <MoreHorizontal />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {writePermitted && (
                <>
                  {workspace.state === WorkspaceState.STARTED && (
                    <DropdownMenuItem
                      onClick={() => handleStop(workspace.id)}
                      className="cursor-pointer"
                      disabled={loadingWorkspaces[workspace.id]}
                    >
                      Stop
                    </DropdownMenuItem>
                  )}
                  {(workspace.state === WorkspaceState.STOPPED || workspace.state === WorkspaceState.ARCHIVED) && (
                    <DropdownMenuItem
                      onClick={() => handleStart(workspace.id)}
                      className="cursor-pointer"
                      disabled={loadingWorkspaces[workspace.id]}
                    >
                      Start
                    </DropdownMenuItem>
                  )}
                  {workspace.state === WorkspaceState.STOPPED && (
                    <DropdownMenuItem
                      onClick={() => handleArchive(workspace.id)}
                      className="cursor-pointer"
                      disabled={loadingWorkspaces[workspace.id]}
                    >
                      Archive
                    </DropdownMenuItem>
                  )}
                </>
              )}
              {deletePermitted && (
                <>
                  {(workspace.state === WorkspaceState.STOPPED || workspace.state === WorkspaceState.STARTED) && (
                    <DropdownMenuSeparator />
                  )}
                  <DropdownMenuItem
                    className="cursor-pointer text-red-600 dark:text-red-400"
                    disabled={loadingWorkspaces[workspace.id]}
                    onClick={() => handleDelete(workspace.id)}
                  >
                    Delete
                  </DropdownMenuItem>
                </>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        )
      },
    },
  ]

  return columns
}
