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
import { useMemo, useState } from 'react'
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
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'

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
  const [selectAllConfirmationOpen, setSelectAllConfirmationOpen] = useState(false)

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
    initialState: {
      pagination: {
        pageSize: DEFAULT_PAGE_SIZE,
      },
    },
  })

  const selectableRows = table.getRowModel().rows.filter((row) => !loadingWorkspaces[row.original.id])
  const totalSelectableCount = selectableRows.length
  const showConfirmation = totalSelectableCount > 15

  const handleSelectAll = () => {
    if (showConfirmation) {
      setSelectAllConfirmationOpen(true)
    } else {
      performSelectAll()
    }
  }

  const performSelectAll = () => {
    for (const row of table.getRowModel().rows) {
      if (!loadingWorkspaces[row.original.id]) {
        row.toggleSelected(true)
      }
    }
    setSelectAllConfirmationOpen(false)
  }

  const handleUndoSelectAll = () => {
    for (const row of table.getRowModel().rows) {
      row.toggleSelected(false)
    }
  }

  const handleHeaderCheckboxChange = (value: boolean | 'indeterminate') => {
    if (value && showConfirmation) {
      setSelectAllConfirmationOpen(true)
    } else {
      for (const row of table.getRowModel().rows) {
        if (loadingWorkspaces[row.original.id]) {
          row.toggleSelected(false)
        } else {
          row.toggleSelected(!!value)
        }
      }
    }
  }

  const [bulkDeleteConfirmationOpen, setBulkDeleteConfirmationOpen] = useState(false)

  return (
    <div>
      {/* Filter section */}
      <div className="flex items-center gap-4 mb-4">
        <DebouncedInput
          value={(table.getColumn('id')?.getFilterValue() as string) ?? ''}
          onChange={(value) => table.getColumn('id')?.setFilterValue(value)}
          placeholder="Search..."
          className="max-w-sm"
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

      {/* Selection header */}
      <div className="flex items-center gap-4 mb-4">
        <div className="flex items-center gap-2">
          {table.getSelectedRowModel().rows.length > 0 ? (
            <>
              <span className="text-sm text-muted-foreground">
                {table.getSelectedRowModel().rows.length}{' '}
                {table.getSelectedRowModel().rows.length === 1 ? 'sandbox' : 'sandboxes'} selected
              </span>
              <span
                className="text-sm text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 cursor-pointer"
                onClick={handleUndoSelectAll}
              >
                Undo
              </span>
            </>
          ) : (
            <>
              {showConfirmation ? (
                <Popover open={selectAllConfirmationOpen} onOpenChange={setSelectAllConfirmationOpen}>
                  <PopoverTrigger asChild>
                    <span
                      className="text-sm text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 cursor-pointer"
                      onClick={handleSelectAll}
                    >
                      Select all sandboxes
                    </span>
                  </PopoverTrigger>
                  <PopoverContent side="right" align="start" className="w-auto ml-2">
                    <div className="flex flex-col gap-3">
                      <p className="text-sm font-medium">Select ALL {totalSelectableCount} Sandboxes?</p>
                      <p className="text-xs text-muted-foreground">
                        This will select all selectable sandboxes in the current view.
                      </p>
                      <div className="flex items-center space-x-2">
                        <Button size="sm" onClick={performSelectAll}>
                          Select All
                        </Button>
                        <Button variant="outline" size="sm" onClick={() => setSelectAllConfirmationOpen(false)}>
                          Cancel
                        </Button>
                      </div>
                    </div>
                  </PopoverContent>
                </Popover>
              ) : (
                <span
                  className="text-sm text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 cursor-pointer"
                  onClick={handleSelectAll}
                >
                  Select all sandboxes
                </span>
              )}
            </>
          )}
        </div>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  // Special handling for the select column header
                  if (header.id === 'select') {
                    const isAllSelected = table.getIsAllPageRowsSelected()
                    const isSomeSelected = table.getIsSomePageRowsSelected()

                    return (
                      <TableHead key={header.id}>
                        {showConfirmation ? (
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Checkbox
                                  checked={isAllSelected || (isSomeSelected && 'indeterminate')}
                                  onCheckedChange={handleHeaderCheckboxChange}
                                  className="translate-y-[2px]"
                                />
                              </TooltipTrigger>
                            </Tooltip>
                          </TooltipProvider>
                        ) : (
                          <Checkbox
                            checked={isAllSelected || (isSomeSelected && 'indeterminate')}
                            onCheckedChange={handleHeaderCheckboxChange}
                            className="translate-y-[2px]"
                          />
                        )}
                      </TableHead>
                    )
                  }

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
          {table.getRowModel().rows.some((row) => row.getIsSelected()) && (
            <div className="flex items-center space-x-2">
              <Popover open={bulkDeleteConfirmationOpen} onOpenChange={setBulkDeleteConfirmationOpen}>
                <PopoverTrigger>
                  <Button variant="destructive" size="sm" className="h-8">
                    Bulk Delete
                  </Button>
                </PopoverTrigger>
                <PopoverContent side="top">
                  <div className="flex flex-col gap-4">
                    <p>Are you sure you want to delete these workspaces?</p>
                    <div className="flex items-center space-x-2">
                      <Button
                        variant="destructive"
                        onClick={() => {
                          handleBulkDelete(
                            table
                              .getRowModel()
                              .rows.filter((row) => row.getIsSelected())
                              .map((row) => row.original.id),
                          )
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
      header: ({ table }) => null,
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
            className="px-2 hover:bg-muted/50 whitespace-nowrap"
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
        return <span className="px-2 whitespace-nowrap">{row.original.id}</span>
      },
    },
    {
      id: 'state',
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
            className="px-2 hover:bg-muted/50 w-24 justify-start"
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
          <div className={`flex items-center gap-2 px-2 w-24 ${color}`}>
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
            className="px-2 hover:bg-muted/50 whitespace-nowrap"
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
        return <span className="px-2 whitespace-nowrap">{getLastEvent(row.original).relativeTimeString}</span>
      },
    },
    {
      id: 'createdAt',
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
            className="px-2 hover:bg-muted/50 whitespace-nowrap"
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
        return <span className="px-2 whitespace-nowrap">{getCreatedAt(row.original).relativeTimeString}</span>
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
