/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Runner, RunnerState, Region } from '@daytonaio/api-client'
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
import { TableHeader, TableRow, TableHead, TableBody, TableCell, Table } from './ui/table'
import { Button } from './ui/button'
import { Switch } from './ui/switch'
import { useState } from 'react'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from './ui/dropdown-menu'
import { Pagination } from './Pagination'
import { Server, MoreHorizontal, Copy, AlertTriangle, CheckCircle, Timer, Pause } from 'lucide-react'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { TableEmptyState } from './TableEmptyState'
import { toast } from 'sonner'
import { DebouncedInput } from './DebouncedInput'

interface RunnerTableProps {
  data: Runner[]
  regions: Region[]
  loading: boolean
  isLoadingRunner: (runner: Runner) => boolean
  writePermitted: boolean
  deletePermitted: boolean
  onToggleEnabled: (runner: Runner) => void
  onDelete: (runner: Runner) => void
  getRegionName: (regionId: string) => string | undefined
}

export function RunnerTable({
  data,
  regions,
  loading,
  isLoadingRunner,
  writePermitted,
  deletePermitted,
  onToggleEnabled,
  onDelete,
  getRegionName,
}: RunnerTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const [globalFilter, setGlobalFilter] = useState('')

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      toast.success('Copied to clipboard')
    } catch (err) {
      console.error('Failed to copy text:', err)
      toast.error('Failed to copy to clipboard')
    }
  }

  const columns = getColumns({
    onToggleEnabled,
    onDelete,
    isLoadingRunner,
    writePermitted,
    deletePermitted,
    copyToClipboard,
    getRegionName,
  })

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
      const runner = row.original as Runner
      const searchValue = filterValue.toLowerCase()
      const regionName = getRegionName(runner.region) ?? runner.region
      return (
        runner.id.toLowerCase().includes(searchValue) ||
        runner.name.toLowerCase().includes(searchValue) ||
        regionName.toLowerCase().includes(searchValue)
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

  return (
    <div>
      <div className="flex items-center mb-4">
        <DebouncedInput
          value={globalFilter ?? ''}
          onChange={(value) => setGlobalFilter(String(value))}
          placeholder="Search by ID or Region"
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
                  className={`${isLoadingRunner(row.original) ? 'opacity-50 pointer-events-none' : ''}`}
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
              ))
            ) : (
              <TableEmptyState
                colSpan={columns.length}
                message="No runners found."
                icon={<Server className="w-8 h-8" />}
                description={
                  <div className="space-y-2">
                    <p>Runners are the machines that run your sandboxes.</p>
                    {regions.length === 0 && (
                      <p>There must be at least one region in your organization before runners can be created.</p>
                    )}
                  </div>
                }
              />
            )}
          </TableBody>
        </Table>
      </div>
      <Pagination table={table} className="mt-4" entityName="Runners" />
    </div>
  )
}

const getColumns = ({
  onToggleEnabled,
  onDelete,
  isLoadingRunner,
  writePermitted,
  deletePermitted,
  copyToClipboard,
  getRegionName,
}: {
  onToggleEnabled: (runner: Runner) => void
  onDelete: (runner: Runner) => void
  isLoadingRunner: (runner: Runner) => boolean
  writePermitted: boolean
  deletePermitted: boolean
  copyToClipboard: (text: string) => Promise<void>
  getRegionName: (regionId: string) => string | undefined
}): ColumnDef<Runner>[] => {
  const getStateIcon = (state: RunnerState) => {
    switch (state) {
      case RunnerState.READY:
        return <CheckCircle className="w-4 h-4 flex-shrink-0" />
      case RunnerState.DISABLED:
      case RunnerState.DECOMMISSIONED:
        return <Pause className="w-4 h-4 flex-shrink-0" />
      case RunnerState.UNRESPONSIVE:
        return <AlertTriangle className="w-4 h-4 flex-shrink-0" />
      default:
        return <Timer className="w-4 h-4 flex-shrink-0" />
    }
  }

  const getStateColor = (state: RunnerState) => {
    switch (state) {
      case RunnerState.READY:
        return 'text-green-500'
      case RunnerState.DISABLED:
      case RunnerState.DECOMMISSIONED:
        return 'text-gray-500 dark:text-gray-400'
      case RunnerState.UNRESPONSIVE:
        return 'text-red-500'
      default:
        return 'text-gray-600 dark:text-gray-400'
    }
  }

  const getStateLabel = (state: RunnerState) => {
    return state
      .split('_')
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join(' ')
  }

  const isRunnerSchedulable = (runner: Runner) => {
    return !runner.unschedulable
  }

  const columns: ColumnDef<Runner>[] = [
    {
      accessorKey: 'id',
      header: 'ID',
      size: 240,
      cell: ({ row }) => (
        <div className="w-full truncate flex items-center gap-2">
          <span className="truncate block text-sm">{row.original.id}</span>
          <button
            onClick={(e) => {
              e.stopPropagation()
              copyToClipboard(row.original.id)
            }}
            className="text-muted-foreground hover:text-foreground transition-colors"
            aria-label="Copy ID"
          >
            <Copy className="w-3 h-3" />
          </button>
        </div>
      ),
    },
    {
      accessorKey: 'name',
      header: 'Name',
      size: 240,
      cell: ({ row }) => (
        <div className="w-full truncate flex items-center gap-2">
          <span className="truncate block text-sm">{row.original.name}</span>
          <button
            onClick={(e) => {
              e.stopPropagation()
              copyToClipboard(row.original.name)
            }}
            className="text-muted-foreground hover:text-foreground transition-colors"
            aria-label="Copy Name"
          >
            <Copy className="w-3 h-3" />
          </button>
        </div>
      ),
    },
    {
      accessorKey: 'regionId',
      header: 'Region',
      size: 180,
      cell: ({ row }) => (
        <div className="w-full truncate flex items-center gap-2">
          <span className="truncate block text-sm">{getRegionName(row.original.region) ?? row.original.region}</span>
          <button
            onClick={(e) => {
              e.stopPropagation()
              copyToClipboard(getRegionName(row.original.region) ?? row.original.region)
            }}
            className="text-muted-foreground hover:text-foreground transition-colors"
            aria-label="Copy Region"
          >
            <Copy className="w-3 h-3" />
          </button>
        </div>
      ),
    },
    // {
    //   accessorKey: 'domain',
    //   header: 'Domain',
    //   size: 180,
    //   cell: ({ row }) => (
    //     <div className="w-full truncate flex items-center gap-2">
    //       <span className="truncate block text-sm">{row.original.domain || '/'}</span>
    //       {row.original.domain && (
    //         <button
    //           onClick={(e) => {
    //             e.stopPropagation()
    //             copyToClipboard(row.original.domain!)
    //           }}
    //           className="text-muted-foreground hover:text-foreground transition-colors"
    //           aria-label="Copy Domain"
    //         >
    //           <Copy className="w-3 h-3" />
    //         </button>
    //       )}
    //     </div>
    //   ),
    // },
    {
      accessorKey: 'state',
      header: 'State',
      size: 200,
      cell: ({ row }) => (
        <div className={`flex items-center gap-2 ${getStateColor(row.original.state)}`}>
          {getStateIcon(row.original.state)}
          {getStateLabel(row.original.state)}
        </div>
      ),
    },
    {
      accessorKey: 'unschedulable',
      header: 'Schedulable',
      size: 60,
      cell: ({ row }) => {
        const isLoading = isLoadingRunner(row.original)
        return (
          <Switch
            checked={isRunnerSchedulable(row.original)}
            onCheckedChange={() => writePermitted && !isLoading && onToggleEnabled(row.original)}
            disabled={!writePermitted || isLoading}
          />
        )
      },
    },
  ]

  columns.push({
    id: 'options',
    header: () => {
      return null
    },
    cell: ({ row }) => {
      if (!deletePermitted) {
        return null
      }

      const isLoading = isLoadingRunner(row.original)

      return (
        <div className="flex justify-end">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="sm" className="h-8 w-8 p-0" disabled={isLoading}>
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                onClick={() => onDelete(row.original)}
                className="cursor-pointer text-red-600 dark:text-red-400"
                disabled={isLoading}
              >
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      )
    },
  })

  return columns
}
