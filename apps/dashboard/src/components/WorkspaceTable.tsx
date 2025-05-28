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
import { TableEmptyState } from './TableEmptyState'

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
  // a prompt to select all confirmation
  const [selectAllConfirmation, setSelectAllConfirmation] = useState(false)
  const [globalSelection, setGlobalSelection] = useState(false)
  // a prompt to deselect all confirmation
  const [deselectAllConfirmation, setDeselectAllConfirmation] = useState(new Set<string>())

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
    globalSelection,
    deselectAllConfirmation,
    setDeselectAllConfirmation,
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
  const [bulkDeleteConfirmationOpen, setBulkDeleteConfirmationOpen] = useState(false)

  // if all workspaces are selected, show a confirmation prompt
  const totalWorkspace = useMemo(() => {
    return data.filter((workspace) => !loadingWorkspaces[workspace.id] && workspace.state !== WorkspaceState.DESTROYING)
  }, [data, loadingWorkspaces])
  // get the total count of workspaces
  const totalSelectCount = totalWorkspace.length

  // make a confirmation if the selected workspaces is as the current
  const currentPageSize = useMemo(() => {
    return table.getState().pagination.pageSize
  }, [table.getState().pagination.pageSize])
  const promptConfirmation = useMemo(() => {
    // if the total count of workspaces is less than the current page size, don't show confirmation
    if (totalSelectCount <= currentPageSize) {
      return false
    }
    // if the total count of workspaces is greater than the current page size, show confirmation
    return table.getFilteredSelectedRowModel().rows.length >= currentPageSize
  }, [totalSelectCount, currentPageSize, table.getFilteredSelectedRowModel().rows.length])

  // if all workspaces are selected, show a confirmation prompt
  const selectAll = () => {
    setGlobalSelection(true)
    setDeselectAllConfirmation(new Set())
    for (const row of table.getRowModel().rows) {
      if (loadingWorkspaces[row.original.id]) {
        row.toggleSelected(false)
      } else {
        row.toggleSelected(true)
      }
    }
    setSelectAllConfirmation(false)
  }

  // handle for undoSelection
  const undoSelectionAll = () => {
    setGlobalSelection(false)
    setDeselectAllConfirmation(new Set())
    for (const row of table.getRowModel().rows) {
      if (loadingWorkspaces[row.original.id]) {
        row.toggleSelected(false)
      }
    }
  }

  // make a handler for check box
  const handleCheckboxChange = (value: boolean | 'indeterminate') => {
    if (value) {
      for (const row of table.getRowModel().rows) {
        if (loadingWorkspaces[row.original.id]) {
          row.toggleSelected(false)
        } else {
          row.toggleSelected(true)
        }
      }

      // show confirmation if all workspaces are selected
      const selectedRows = table.getRowModel().rows.filter((row) => !loadingWorkspaces[row.original.id])
      if (promptConfirmation && selectedRows.length < totalSelectCount) {
        setSelectAllConfirmation(true)
      } else {
        setGlobalSelection(true)
        setDeselectAllConfirmation(new Set())
        for (const row of table.getRowModel().rows) {
          if (loadingWorkspaces[row.original.id]) {
            row.toggleSelected(false)
          } else {
            row.toggleSelected(true)
          }
        }
      }
    }
  }
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
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  if (header.id === 'select') {
                    const allSelect = table.getIsAllPageRowsSelected() || globalSelection
                    const someSelect = table.getIsSomePageRowsSelected() && !globalSelection
                    return (
                      <TableHead key={header.id}>
                        <Popover open={selectAllConfirmation} onOpenChange={setSelectAllConfirmation}>
                          <PopoverTrigger asChild>
                            <Checkbox
                              checked={allSelect || (someSelect && 'indeterminate')}
                              onCheckedChange={handleCheckboxChange}
                              className="translate-y-[2px] cursor-pointer"
                              aria-label="Select all workspaces"
                            />
                          </PopoverTrigger>
                          <PopoverContent side="bottom" className="w-72">
                            <div className="flex flex-col gap-2">
                              <p className="text-sm">
                                {globalSelection
                                  ? `You have selected all ${totalSelectCount} workspaces.`
                                  : `You have selected ${table.getFilteredSelectedRowModel().rows.length} of ${totalSelectCount} workspaces.`}
                              </p>
                              {promptConfirmation && (
                                <p className="text-sm text-red-500">
                                  Are you sure you want to select all workspaces? This will select {totalSelectCount}{' '}
                                  workspaces.
                                </p>
                              )}
                              <div className="flex items-center justify-end space-x-2">
                                <Button
                                  variant="outline"
                                  size="sm"
                                  onClick={() => {
                                    setSelectAllConfirmation(false)
                                    undoSelectionAll()
                                  }}
                                >
                                  Cancel
                                </Button>
                                <Button
                                  variant="secondary"
                                  size="sm"
                                  onClick={() => {
                                    selectAll()
                                    setSelectAllConfirmation(false)
                                  }}
                                >
                                  Select All
                                </Button>
                              </div>
                            </div>
                          </PopoverContent>
                        </Popover>
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
              <TableEmptyState colSpan={columns.length} message="No Sandboxes found." />
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-between space-x-2 py-4">
        <div className="flex items-center space-x-2">
          {table.getRowModel().rows.some((row) => row.getIsSelected() || globalSelection) && (
            <div className="flex items-center space-x-2">
              <Popover open={bulkDeleteConfirmationOpen} onOpenChange={setBulkDeleteConfirmationOpen}>
                <PopoverTrigger>
                  <Button variant="destructive" size="sm" className="h-8">
                    Bulk Delete
                  </Button>
                </PopoverTrigger>
                <PopoverContent side="top">
                  <div className="flex flex-col gap-4">
                    <p>
                      Are you sure you want to delete these workspaces?
                      {globalSelection
                        ? ` You are about to delete all ${totalSelectCount} selected workspaces.`
                        : ` You are about to delete ${table.getFilteredSelectedRowModel().rows.length} workspaces.`}
                    </p>
                    <div className="flex items-center space-x-2">
                      <Button
                        variant="destructive"
                        onClick={() => {
                          const selectedIds = table.getFilteredSelectedRowModel().rows.map((row) => row.original.id)
                          if (globalSelection) {
                            const allIds = data
                              .filter(
                                (workspace) =>
                                  !loadingWorkspaces[workspace.id] && workspace.state !== WorkspaceState.DESTROYING,
                              )
                              .map((workspace) => workspace.id)
                            handleBulkDelete(allIds)
                          } else {
                            handleBulkDelete(selectedIds)
                          }
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
        <Pagination
          table={table}
          selectionEnabled
          onUndoDelete={globalSelection ? undoSelectionAll : undefined}
          entityName="Sandboxes"
          onCustomSelect={
            globalSelection
              ? `Select all ${totalSelectCount - deselectAllConfirmation.size} Sandboxes of ${totalSelectCount}`
              : undefined
          }
        />
        <Pagination table={table} selectionEnabled entityName="Sandboxes" />
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
  globalSelection,
  deselectAllConfirmation,
  setDeselectAllConfirmation,
}: {
  handleStart: (id: string) => void
  handleStop: (id: string) => void
  handleDelete: (id: string) => void
  handleArchive: (id: string) => void
  loadingWorkspaces: Record<string, boolean>
  writePermitted: boolean
  deletePermitted: boolean
  globalSelection: boolean
  deselectAllConfirmation: Set<string>
  setDeselectAllConfirmation: (value: Set<string>) => void
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
            checked={row.getIsSelected() || (globalSelection && !deselectAllConfirmation.has(row.original.id))}
            onCheckedChange={(value) => {
              if (globalSelection) {
                const newDeselect = new Set(deselectAllConfirmation)
                if (value) {
                  newDeselect.delete(row.original.id)
                } else {
                  newDeselect.add(row.original.id)
                }
                setDeselectAllConfirmation(newDeselect)
              }
            }}
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
      sortingFn: (rowA, rowB) => {
        const statePriorityOrder = {
          [WorkspaceState.STARTED]: 1,
          [WorkspaceState.BUILDING_IMAGE]: 2,
          [WorkspaceState.PENDING_BUILD]: 2,
          [WorkspaceState.RESTORING]: 3,
          [WorkspaceState.ERROR]: 4,
          [WorkspaceState.STOPPED]: 5,
          [WorkspaceState.ARCHIVING]: 6,
          [WorkspaceState.ARCHIVED]: 6,
          [WorkspaceState.CREATING]: 7,
          [WorkspaceState.STARTING]: 7,
          [WorkspaceState.STOPPING]: 7,
          [WorkspaceState.DESTROYING]: 7,
          [WorkspaceState.DESTROYED]: 7,
          [WorkspaceState.PULLING_IMAGE]: 7,
          [WorkspaceState.UNKNOWN]: 7,
        }

        const stateA = rowA.original.state || WorkspaceState.UNKNOWN
        const stateB = rowB.original.state || WorkspaceState.UNKNOWN

        if (stateA === stateB) {
          return 0
        }

        return statePriorityOrder[stateA] - statePriorityOrder[stateB]
      },
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
