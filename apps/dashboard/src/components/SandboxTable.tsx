/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationRolePermissionsEnum, Sandbox, SandboxState } from '@daytonaio/api-client'
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
  data: Sandbox[]
  loadingSandboxes: Record<string, boolean>
  loading: boolean
  handleStart: (id: string) => void
  handleStop: (id: string) => void
  handleDelete: (id: string) => void
  handleBulkDelete: (ids: string[]) => void
  handleArchive: (id: string) => void
}

export function SandboxTable({
  data,
  loadingSandboxes,
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

  const [sorting, setSorting] = useState<SortingState>([
    {
      id: 'state',
      desc: false,
    },
    {
      id: 'lastEvent',
      desc: true,
    },
  ])
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])

  // Enhanced selection state
  const [crossPageSelection, setCrossPageSelection] = useState<Set<string>>(new Set())
  const [crossPageSelectionPopoverOpen, setCrossPageSelectionPopoverOpen] = useState(false)
  const [isAllPagesSelected, setIsAllPagesSelected] = useState(false)

  const labelOptions: FacetedFilterOption[] = useMemo(() => {
    const labels = new Set<string>()
    data.forEach((sandbox) => {
      Object.entries(sandbox.labels ?? {}).forEach(([key, value]) => {
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
    loadingSandboxes,
    writePermitted,
    deletePermitted,
    data,
    crossPageSelection,
    setCrossPageSelection,
    crossPageSelectionPopoverOpen,
    setCrossPageSelectionPopoverOpen,
    isAllPagesSelected,
    setIsAllPagesSelected,
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
                  className={`${loadingSandboxes[row.original.id] || row.original.state === SandboxState.DESTROYING ? 'opacity-50 pointer-events-none' : ''}`}
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
          {(table.getRowModel().rows.some((row) => row.getIsSelected()) || isAllPagesSelected) && (
            <div className="flex items-center space-x-2">
              {/* Selection Counter with Undo Button */}
              <div className="flex items-center space-x-2 text-sm text-muted-foreground">
                <span>
                  {isAllPagesSelected
                    ? `${crossPageSelection.size} of ${data.length} row(s) selected.`
                    : `${table.getSelectedRowModel().rows.length} of ${table.getRowModel().rows.length} row(s) selected.`}
                </span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    // Clear all selections
                    table.resetRowSelection()
                    setIsAllPagesSelected(false)
                    setCrossPageSelection(new Set())
                  }}
                  aria-label="Undo selection"
                  className="h-6 px-2 text-xs"
                >
                  Undo
                </Button>
              </div>

              <Popover open={bulkDeleteConfirmationOpen} onOpenChange={setBulkDeleteConfirmationOpen}>
                <PopoverTrigger asChild>
                  <Button variant="destructive" size="sm" className="h-8">
                    Bulk Delete
                  </Button>
                </PopoverTrigger>
                <PopoverContent side="top">
                  <div className="flex flex-col gap-4">
                    <p>Are you sure you want to delete these sandboxes?</p>
                    <div className="flex items-center space-x-2">
                      <Button
                        variant="destructive"
                        onClick={() => {
                          if (isAllPagesSelected) {
                            // Delete all selected across pages
                            handleBulkDelete(Array.from(crossPageSelection))
                          } else {
                            // Delete current page selection
                            handleBulkDelete(
                              table
                                .getRowModel()
                                .rows.filter((row) => row.getIsSelected())
                                .map((row) => row.original.id),
                            )
                          }
                          setBulkDeleteConfirmationOpen(false)
                          // Clear selections after delete
                          table.resetRowSelection()
                          setIsAllPagesSelected(false)
                          setCrossPageSelection(new Set())
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
        <Pagination table={table} selectionEnabled entityName="Sandboxes" />
      </div>
    </div>
  )
}

const getStateIcon = (state?: SandboxState) => {
  switch (state) {
    case SandboxState.STARTED:
      return <CheckCircle className="w-4 h-4" />
    case SandboxState.STOPPED:
      return <Circle className="w-4 h-4" />
    case SandboxState.ERROR:
    case SandboxState.BUILD_FAILED:
      return <AlertTriangle className="w-4 h-4" />
    case SandboxState.CREATING:
    case SandboxState.STARTING:
    case SandboxState.STOPPING:
    case SandboxState.DESTROYING:
    case SandboxState.ARCHIVING:
      return <Timer className="w-4 h-4" />
    case SandboxState.ARCHIVED:
      return <Archive className="w-4 h-4" />
    default:
      return null
  }
}

const getLastEvent = (sandbox: Sandbox): { date: Date; relativeTimeString: string } => {
  return getRelativeTimeString(sandbox.updatedAt)
}

const getCreatedAt = (sandbox: Sandbox): { date: Date; relativeTimeString: string } => {
  return getRelativeTimeString(sandbox.createdAt)
}

const getStateColor = (state?: SandboxState) => {
  switch (state) {
    case SandboxState.STARTED:
      return 'text-green-500'
    case SandboxState.STOPPED:
      return 'text-gray-500'
    case SandboxState.ERROR:
    case SandboxState.BUILD_FAILED:
      return 'text-red-500'
    default:
      return 'text-gray-600 dark:text-gray-400'
  }
}

const getStateLabel = (state?: SandboxState) => {
  if (!state) {
    return 'Unknown'
  }
  // TODO: remove when destroying/destroyed is migrated to deleting/deleted
  if (state === SandboxState.DESTROYING) {
    return 'Deleting'
  }
  return state
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ')
}

const statuses: FacetedFilterOption[] = [
  { label: getStateLabel(SandboxState.STARTED), value: SandboxState.STARTED, icon: CheckCircle },
  { label: getStateLabel(SandboxState.STOPPED), value: SandboxState.STOPPED, icon: Circle },
  { label: getStateLabel(SandboxState.ERROR), value: SandboxState.ERROR, icon: AlertTriangle },
  { label: getStateLabel(SandboxState.BUILD_FAILED), value: SandboxState.BUILD_FAILED, icon: AlertTriangle },
  { label: getStateLabel(SandboxState.STARTING), value: SandboxState.STARTING, icon: Timer },
  { label: getStateLabel(SandboxState.STOPPING), value: SandboxState.STOPPING, icon: Timer },
  { label: getStateLabel(SandboxState.DESTROYING), value: SandboxState.DESTROYING, icon: Timer },
  { label: getStateLabel(SandboxState.ARCHIVING), value: SandboxState.ARCHIVING, icon: Timer },
  { label: getStateLabel(SandboxState.ARCHIVED), value: SandboxState.ARCHIVED, icon: Archive },
]

const getColumns = ({
  handleStart,
  handleStop,
  handleDelete,
  handleArchive,
  loadingSandboxes,
  writePermitted,
  deletePermitted,
  data,
  crossPageSelection,
  setCrossPageSelection,
  crossPageSelectionPopoverOpen,
  setCrossPageSelectionPopoverOpen,
  isAllPagesSelected,
  setIsAllPagesSelected,
}: {
  handleStart: (id: string) => void
  handleStop: (id: string) => void
  handleDelete: (id: string) => void
  handleArchive: (id: string) => void
  loadingSandboxes: Record<string, boolean>
  writePermitted: boolean
  deletePermitted: boolean
  data: Sandbox[]
  crossPageSelection: Set<string>
  setCrossPageSelection: (selection: Set<string>) => void
  crossPageSelectionPopoverOpen: boolean
  setCrossPageSelectionPopoverOpen: (open: boolean) => void
  isAllPagesSelected: boolean
  setIsAllPagesSelected: (selected: boolean) => void
}): ColumnDef<Sandbox>[] => {
  const columns: ColumnDef<Sandbox>[] = [
    {
      id: 'select',
      header: ({ table }) => {
        return (
          <SelectAllCheckbox
            table={table}
            data={data}
            loadingSandboxes={loadingSandboxes}
            crossPageSelection={crossPageSelection}
            setCrossPageSelection={setCrossPageSelection}
            crossPageSelectionPopoverOpen={crossPageSelectionPopoverOpen}
            setCrossPageSelectionPopoverOpen={setCrossPageSelectionPopoverOpen}
            isAllPagesSelected={isAllPagesSelected}
            setIsAllPagesSelected={setIsAllPagesSelected}
          />
        )
      },
      cell: ({ row }) => {
        if (loadingSandboxes[row.original.id]) {
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
        const sandbox = row.original
        const state = row.original.state
        const color = getStateColor(state)

        if ((state === SandboxState.ERROR || state === SandboxState.BUILD_FAILED) && !!sandbox.errorReason) {
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
                  <p className="max-w-[300px]">{sandbox.errorReason}</p>
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
          [SandboxState.STARTED]: 1,
          [SandboxState.BUILDING_SNAPSHOT]: 2,
          [SandboxState.PENDING_BUILD]: 2,
          [SandboxState.RESTORING]: 3,
          [SandboxState.ERROR]: 4,
          [SandboxState.BUILD_FAILED]: 4,
          [SandboxState.STOPPED]: 5,
          [SandboxState.ARCHIVING]: 6,
          [SandboxState.ARCHIVED]: 6,
          [SandboxState.CREATING]: 7,
          [SandboxState.STARTING]: 7,
          [SandboxState.STOPPING]: 7,
          [SandboxState.DESTROYING]: 7,
          [SandboxState.DESTROYED]: 7,
          [SandboxState.PULLING_SNAPSHOT]: 7,
          [SandboxState.UNKNOWN]: 7,
        }

        const stateA = rowA.original.state || SandboxState.UNKNOWN
        const stateB = rowB.original.state || SandboxState.UNKNOWN

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
        if (!row.original.runnerDomain || row.original.state !== SandboxState.STARTED) return ''
        return (
          <a
            href={`https://22222-${row.original.id}.${row.original.runnerDomain}`}
            target="_blank"
            rel="noopener noreferrer"
          >
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

        const sandbox = row.original

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
                  {sandbox.state === SandboxState.STARTED && (
                    <DropdownMenuItem
                      onClick={() => handleStop(sandbox.id)}
                      className="cursor-pointer"
                      disabled={loadingSandboxes[sandbox.id]}
                    >
                      Stop
                    </DropdownMenuItem>
                  )}
                  {(sandbox.state === SandboxState.STOPPED || sandbox.state === SandboxState.ARCHIVED) && (
                    <DropdownMenuItem
                      onClick={() => handleStart(sandbox.id)}
                      className="cursor-pointer"
                      disabled={loadingSandboxes[sandbox.id]}
                    >
                      Start
                    </DropdownMenuItem>
                  )}
                  {sandbox.state === SandboxState.STOPPED && (
                    <DropdownMenuItem
                      onClick={() => handleArchive(sandbox.id)}
                      className="cursor-pointer"
                      disabled={loadingSandboxes[sandbox.id]}
                    >
                      Archive
                    </DropdownMenuItem>
                  )}
                </>
              )}
              {deletePermitted && (
                <>
                  {(sandbox.state === SandboxState.STOPPED || sandbox.state === SandboxState.STARTED) && (
                    <DropdownMenuSeparator />
                  )}
                  <DropdownMenuItem
                    className="cursor-pointer text-red-600 dark:text-red-400"
                    disabled={loadingSandboxes[sandbox.id]}
                    onClick={() => handleDelete(sandbox.id)}
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

// Component for the select all checkbox with cross-page selection support
interface SelectAllCheckboxProps {
  table: any
  data: Sandbox[]
  loadingSandboxes: Record<string, boolean>
  crossPageSelection: Set<string>
  setCrossPageSelection: (selection: Set<string>) => void
  crossPageSelectionPopoverOpen: boolean
  setCrossPageSelectionPopoverOpen: (open: boolean) => void
  isAllPagesSelected: boolean
  setIsAllPagesSelected: (selected: boolean) => void
}

const SelectAllCheckbox = ({
  table,
  data,
  loadingSandboxes,
  crossPageSelection,
  setCrossPageSelection,
  crossPageSelectionPopoverOpen,
  setCrossPageSelectionPopoverOpen,
  isAllPagesSelected,
  setIsAllPagesSelected,
}: SelectAllCheckboxProps) => {
  const totalPages = table.getPageCount()
  const isMultiPage = totalPages > 1

  return (
    <Popover open={crossPageSelectionPopoverOpen} onOpenChange={setCrossPageSelectionPopoverOpen}>
      <PopoverTrigger asChild>
        <Checkbox
          checked={
            isAllPagesSelected ||
            table.getIsAllPageRowsSelected() ||
            (table.getIsSomePageRowsSelected() && 'indeterminate')
          }
          onCheckedChange={(value) => {
            if (isMultiPage && value && !isAllPagesSelected) {
              // Show popover for cross-page selection
              setCrossPageSelectionPopoverOpen(true)
              // Select current page
              for (const row of table.getRowModel().rows) {
                if (!loadingSandboxes[row.original.id]) {
                  row.toggleSelected(true)
                }
              }
            } else {
              // Normal selection behavior
              for (const row of table.getRowModel().rows) {
                if (loadingSandboxes[row.original.id]) {
                  row.toggleSelected(false)
                } else {
                  row.toggleSelected(!!value)
                }
              }

              if (!value) {
                setIsAllPagesSelected(false)
                setCrossPageSelection(new Set())
              }
            }
          }}
          aria-label="Select all"
          className="translate-y-[2px]"
        />
      </PopoverTrigger>
      <PopoverContent side="bottom" className="w-80">
        <div className="flex flex-col gap-4">
          <p className="text-sm">{table.getSelectedRowModel().rows.length} selected on the current page.</p>
          <Button
            variant="default"
            size="sm"
            onClick={() => {
              // Select all pages
              const allSelectableIds = data
                .filter((sandbox) => !loadingSandboxes[sandbox.id])
                .map((sandbox) => sandbox.id)

              setCrossPageSelection(new Set(allSelectableIds))
              setIsAllPagesSelected(true)
              setCrossPageSelectionPopoverOpen(false)

              // Also select all on current page
              for (const row of table.getRowModel().rows) {
                if (!loadingSandboxes[row.original.id]) {
                  row.toggleSelected(true)
                }
              }
            }}
          >
            Select all {data.filter((sandbox) => !loadingSandboxes[sandbox.id]).length} sandboxes?
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  )
}
