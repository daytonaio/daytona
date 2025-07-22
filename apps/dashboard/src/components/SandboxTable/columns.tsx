/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
import { Sandbox, SandboxState } from '@daytonaio/api-client'
import { ColumnDef } from '@tanstack/react-table'
import { ArrowUp, ArrowDown } from 'lucide-react'
import { Checkbox } from '../ui/checkbox'
import { getRelativeTimeString } from '@/lib/utils'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '../ui/tooltip'
import { SandboxState as SandboxStateComponent } from './SandboxState'
import { SandboxTableActions } from './SandboxTableActions'
import { STATE_PRIORITY_ORDER } from './constants'
import { ResourceFilterValue } from './filters/ResourceFilter'
import { arrayIncludesFilter, arrayIntersectionFilter, resourceRangeFilter, dateRangeFilter } from './filters/utils'

interface SortableHeaderProps {
  column: any
  label: string
  dataState?: string
}

const SortableHeader: React.FC<SortableHeaderProps> = ({ column, label, dataState }) => {
  return (
    <div
      role="button"
      onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
      className="flex items-center"
      {...(dataState && { 'data-state': dataState })}
    >
      {label}
      {column.getIsSorted() === 'asc' ? (
        <ArrowUp className="ml-2 h-4 w-4" />
      ) : column.getIsSorted() === 'desc' ? (
        <ArrowDown className="ml-2 h-4 w-4" />
      ) : (
        <div className="ml-2 w-4 h-4" />
      )}
    </div>
  )
}

interface GetColumnsProps {
  handleStart: (id: string) => void
  handleStop: (id: string) => void
  handleDelete: (id: string) => void
  handleArchive: (id: string) => void
  handleVnc: (id: string) => void
  getWebTerminalUrl: (id: string) => Promise<string | null>
  loadingSandboxes: Record<string, boolean>
  writePermitted: boolean
  deletePermitted: boolean
}

export function getColumns({
  handleStart,
  handleStop,
  handleDelete,
  handleArchive,
  handleVnc,
  getWebTerminalUrl,
  loadingSandboxes,
  writePermitted,
  deletePermitted,
}: GetColumnsProps): ColumnDef<Sandbox>[] {
  const handleOpenWebTerminal = async (sandboxId: string) => {
    const url = await getWebTerminalUrl(sandboxId)
    if (url) {
      window.open(url, '_blank')
    }
  }

  const columns: ColumnDef<Sandbox>[] = [
    {
      id: 'select',
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected() || (table.getIsSomePageRowsSelected() && 'indeterminate')}
          onCheckedChange={(value) => {
            for (const row of table.getRowModel().rows) {
              if (loadingSandboxes[row.original.id]) {
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
        return (
          <div>
            <Checkbox
              checked={row.getIsSelected()}
              onCheckedChange={(value) => row.toggleSelected(!!value)}
              aria-label="Select row"
              onClick={(e) => e.stopPropagation()}
              className="translate-y-[1px]"
            />
          </div>
        )
      },

      enableSorting: false,
      enableHiding: false,
    },
    {
      id: 'id',
      header: ({ column }) => {
        return <SortableHeader column={column} label="ID" />
      },
      accessorKey: 'id',
      cell: ({ row }) => {
        return (
          <div className=" w-full truncate">
            <span>{row.original.id}</span>
          </div>
        )
      },
    },
    {
      id: 'state',
      size: 140,
      header: ({ column }) => {
        return <SortableHeader column={column} label="State" />
      },
      cell: ({ row }) => (
        <div className=" w-full truncate">
          <SandboxStateComponent state={row.original.state} errorReason={row.original.errorReason} />
        </div>
      ),
      accessorKey: 'state',
      sortingFn: (rowA, rowB) => {
        const stateA = rowA.original.state || SandboxState.UNKNOWN
        const stateB = rowB.original.state || SandboxState.UNKNOWN

        if (stateA === stateB) {
          return 0
        }

        return STATE_PRIORITY_ORDER[stateA] - STATE_PRIORITY_ORDER[stateB]
      },
      filterFn: (row, id, value) => arrayIncludesFilter(row, id, value),
    },
    {
      id: 'snapshot',
      header: ({ column }) => {
        return <SortableHeader column={column} label="Snapshot" />
      },
      cell: ({ row }) => {
        return (
          <div className=" w-full truncate">
            {row.original.snapshot ? (
              <div className="truncate max-w-md">{row.original.snapshot}</div>
            ) : (
              <div className="truncate max-w-md text-muted-foreground/50">-</div>
            )}
          </div>
        )
      },
      accessorKey: 'snapshot',
      filterFn: (row, id, value) => arrayIncludesFilter(row, id, value),
    },
    {
      id: 'region',
      size: 80,
      header: ({ column }) => {
        return <SortableHeader column={column} label="Region" dataState="sortable" />
      },
      cell: ({ row }) => {
        return <span>{row.original.target}</span>
      },
      accessorKey: 'target',
      filterFn: (row, id, value) => arrayIncludesFilter(row, id, value),
    },
    {
      id: 'resources',
      size: 10,
      minSize: 200,
      header: () => {
        return <span>Resources</span>
      },
      cell: ({ row }) => {
        return (
          <div className="flex items-center gap-2">
            <div>
              {row.original.cpu} <span className="text-muted-foreground">vCPU</span>
            </div>
            <div className="w-[1px] h-6 bg-muted-foreground/20 rounded-full inline-block"></div>
            <div>
              {row.original.memory} <span className=" text-muted-foreground">GiB</span>
            </div>
            <div className="w-[1px] h-6 bg-muted-foreground/20 rounded-full inline-block"></div>
            <div>
              {row.original.disk} <span className=" text-muted-foreground">GiB</span>
            </div>
          </div>
        )
      },
      filterFn: (row, id, value: ResourceFilterValue) => resourceRangeFilter(row, value),
    },
    {
      id: 'labels',
      size: 110,
      enableSorting: false,
      header: () => {
        return <span>Labels</span>
      },
      cell: ({ row }) => {
        const labels = Object.entries(row.original.labels ?? {})
          .map(([key, value]) => `${key}: ${value}`)
          .join(', ')

        const labelCount = Object.keys(row.original.labels ?? {}).length
        return (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                {labelCount > 0 ? (
                  <div className="truncate w-fit bg-blue-100 rounded-sm text-blue-800 dark:bg-blue-950 dark:text-blue-200 px-1">
                    {labelCount > 0 ? (labelCount === 1 ? '1 label' : `${labelCount} labels`) : '/'}
                  </div>
                ) : (
                  <div className="truncate max-w-md text-muted-foreground/50">-</div>
                )}
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
      filterFn: (row, id, value) => arrayIntersectionFilter(row, id, value),
    },
    {
      id: 'lastEvent',
      size: 140,
      header: ({ column }) => {
        return <SortableHeader column={column} label="Last Event" />
      },
      filterFn: (row, id, value) => dateRangeFilter(row, id, value),
      accessorFn: (row) => getLastEvent(row).date,
      cell: ({ row }) => {
        return <span>{getLastEvent(row.original).relativeTimeString}</span>
      },
    },
    {
      id: 'actions',
      size: 100,
      enableHiding: false,
      cell: ({ row }) => (
        <div>
          <SandboxTableActions
            sandbox={row.original}
            writePermitted={writePermitted}
            deletePermitted={deletePermitted}
            isLoading={loadingSandboxes[row.original.id]}
            onStart={handleStart}
            onStop={handleStop}
            onDelete={handleDelete}
            onArchive={handleArchive}
            onVnc={handleVnc}
            onOpenWebTerminal={handleOpenWebTerminal}
          />
        </div>
      ),
    },
  ]

  return columns
}

function getLastEvent(sandbox: Sandbox): { date: Date; relativeTimeString: string } {
  return getRelativeTimeString(sandbox.updatedAt)
}
