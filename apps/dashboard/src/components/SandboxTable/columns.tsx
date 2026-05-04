/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { formatTimestamp, getRelativeTimeString } from '@/lib/utils'
import { Sandbox, SandboxDesiredState, SandboxState } from '@daytona/api-client'
import { ColumnDef } from '@tanstack/react-table'
import React from 'react'
import { CopyButton } from '../CopyButton'
import { EllipsisWithTooltip } from '../EllipsisWithTooltip'
import { SandboxLabel } from '../SandboxLabel'
import { SortOrderIcon } from '../SortIcon'
import { TimestampTooltip } from '../TimestampTooltip'
import { SandboxState as SandboxStateComponent } from '../sandboxes/SandboxState'
import { Badge } from '../ui/badge'
import { Checkbox } from '../ui/checkbox'
import { Tooltip, TooltipContent, TooltipTrigger } from '../ui/tooltip'
import { SandboxTableActions } from './SandboxTableActions'
import { STATE_PRIORITY_ORDER } from './constants'
import { ResourceFilterValue } from './filters/ResourceFilter'
import { arrayIncludesFilter, arrayIntersectionFilter, dateRangeFilter, resourceRangeFilter } from './filters/utils'

interface SortableHeaderProps {
  column: any
  label: string
  dataState?: string
}

const SortableHeader: React.FC<SortableHeaderProps> = ({ column, label, dataState }) => {
  const sortDirection = column.getIsSorted()

  return (
    <button
      type="button"
      onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
      className="group/sort-header flex h-full w-full items-center gap-2"
      {...(dataState && { 'data-state': dataState })}
    >
      {label}
      <SortOrderIcon sort={sortDirection || null} />
    </button>
  )
}

const SandboxStateCell = React.memo(function SandboxStateCell({
  state,
  errorReason,
  recoverable,
}: Pick<Sandbox, 'state' | 'errorReason' | 'recoverable'>) {
  return (
    <div className="w-full truncate">
      <SandboxStateComponent state={state} errorReason={errorReason} recoverable={recoverable} />
    </div>
  )
})

interface GetColumnsProps {
  handleStart: (id: string) => void
  handleStop: (id: string) => void
  handleDelete: (id: string) => void
  handleArchive: (id: string) => void
  handleVnc: (id: string) => void
  sandboxIsLoading: Record<string, boolean>
  writePermitted: boolean
  deletePermitted: boolean
  handleCreateSshAccess: (id: string) => void
  handleRevokeSshAccess: (id: string) => void
  handleRecover: (id: string) => void
  getRegionName: (regionId: string) => string | undefined
  handleScreenRecordings: (id: string) => void
  handleCreateSnapshot: (id: string) => void
  handleFork: (id: string) => void
  handleViewForks: (id: string) => void
}

export function getColumns({
  handleStart,
  handleStop,
  handleDelete,
  handleArchive,
  handleVnc,
  sandboxIsLoading,
  writePermitted,
  deletePermitted,
  handleCreateSshAccess,
  handleRevokeSshAccess,
  handleRecover,
  getRegionName,
  handleScreenRecordings,
  handleCreateSnapshot,
  handleFork,
  handleViewForks,
}: GetColumnsProps): ColumnDef<Sandbox>[] {
  const columns: ColumnDef<Sandbox>[] = [
    {
      id: 'select',
      size: 44,
      minSize: 44,
      maxSize: 44,
      header: ({ table }) => (
        <div className="flex justify-center">
          <Checkbox
            checked={
              table.getIsAllPageRowsSelected() ? true : table.getIsSomePageRowsSelected() ? 'indeterminate' : false
            }
            onCheckedChange={(value) => {
              for (const row of table.getRowModel().rows) {
                if (sandboxIsLoading[row.original.id] || row.original.state === SandboxState.DESTROYED) {
                  row.toggleSelected(false)
                } else {
                  row.toggleSelected(!!value)
                }
              }
            }}
            aria-label="Select all"
          />
        </div>
      ),
      cell: ({ row }) => {
        return (
          <div className="flex justify-center">
            <Checkbox
              checked={row.getIsSelected()}
              onCheckedChange={(value) => row.toggleSelected(!!value)}
              aria-label="Select row"
              onClick={(e) => e.stopPropagation()}
            />
          </div>
        )
      },

      enableSorting: false,
      enableHiding: false,
    },
    {
      id: 'name',
      size: 350,
      enableSorting: true,
      enableHiding: true,
      header: ({ column }) => {
        return <SortableHeader column={column} label="Name" />
      },
      accessorKey: 'name',
      cell: ({ row }) => {
        const displayName = getDisplayName(row.original)
        return (
          <div className="w-full truncate">
            <span className="truncate block">{displayName}</span>
          </div>
        )
      },
    },
    {
      id: 'id',
      size: 150,
      maxSize: 150,
      enableSorting: false,
      enableHiding: true,
      header: () => {
        return <span>UUID</span>
      },
      accessorKey: 'id',
      cell: ({ row }) => {
        const id = row.original.id
        const truncated = id.length > 12 ? `${id.slice(0, 8)}…${id.slice(-4)}` : id
        return (
          <div className="w-full truncate flex items-center gap-1 group/copy-button">
            <span className="truncate block text-muted-foreground">{truncated}</span>
            <CopyButton value={id} size="icon-xs" autoHide tooltipText="Copy UUID" />
          </div>
        )
      },
    },
    {
      id: 'state',
      size: 120,
      maxSize: 120,
      enableSorting: true,
      enableHiding: false,
      header: ({ column }) => {
        return <SortableHeader column={column} label="State" />
      },
      cell: ({ row }) => (
        <SandboxStateCell
          state={row.original.state}
          errorReason={row.original.errorReason}
          recoverable={row.original.recoverable}
        />
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
      size: 150,
      enableSorting: true,
      enableHiding: false,
      header: ({ column }) => {
        return <SortableHeader column={column} label="Snapshot" />
      },
      cell: ({ row }) => {
        return (
          <div className="w-full truncate">
            {row.original.snapshot ? (
              <EllipsisWithTooltip>{row.original.snapshot}</EllipsisWithTooltip>
            ) : (
              <div className="truncate text-muted-foreground/50">-</div>
            )}
          </div>
        )
      },
      accessorKey: 'snapshot',
      filterFn: (row, id, value) => arrayIncludesFilter(row, id, value),
    },
    {
      id: 'region',
      size: 120,
      maxSize: 120,
      enableSorting: true,
      enableHiding: false,
      header: ({ column }) => {
        return <SortableHeader column={column} label="Region" dataState="sortable" />
      },
      cell: ({ row }) => {
        return (
          <div className="w-full truncate">
            <span className="truncate block">{getRegionName(row.original.target) ?? row.original.target}</span>
          </div>
        )
      },
      accessorKey: 'target',
      filterFn: (row, id, value) => arrayIncludesFilter(row, id, value),
    },
    {
      id: 'resources',
      size: 190,
      enableSorting: false,
      enableHiding: false,
      header: () => {
        return <span>Resources</span>
      },
      cell: ({ row }) => {
        return (
          <div className="flex items-center gap-2 w-full truncate">
            <div className="whitespace-nowrap">
              {row.original.cpu} <span className="text-muted-foreground">vCPU</span>
            </div>
            <div className="w-[1px] h-6 bg-muted-foreground/20 rounded-full inline-block"></div>
            <div className="whitespace-nowrap">
              {row.original.memory} <span className="text-muted-foreground">GiB</span>
            </div>
            <div className="w-[1px] h-6 bg-muted-foreground/20 rounded-full inline-block"></div>
            <div className="whitespace-nowrap">
              {row.original.disk} <span className="text-muted-foreground">GiB</span>
            </div>
          </div>
        )
      },
      filterFn: (row, id, value: ResourceFilterValue) => resourceRangeFilter(row, value),
    },
    {
      id: 'labels',
      size: 120,
      maxSize: 120,
      enableSorting: false,
      enableHiding: true,
      header: () => {
        return <span>Labels</span>
      },
      cell: ({ row }) => {
        const labelEntries = Object.entries(row.original.labels ?? {})

        if (labelEntries.length === 0) {
          return <div className="truncate max-w-md text-muted-foreground/50">-</div>
        }

        return (
          <Tooltip>
            <TooltipTrigger asChild>
              <Badge variant="info">{labelEntries.length === 1 ? '1 label' : `${labelEntries.length} labels`}</Badge>
            </TooltipTrigger>
            <TooltipContent className="max-w-[300px] max-h-[400px] overflow-y-auto scrollbar-sm p-2">
              <div className="flex flex-wrap gap-2">
                {labelEntries.map(([key, value]) => (
                  <SandboxLabel key={key} labelKey={key} value={value} />
                ))}
              </div>
            </TooltipContent>
          </Tooltip>
        )
      },
      accessorFn: (row) => Object.entries(row.labels ?? {}).map(([key, value]) => `${key}: ${value}`),
      filterFn: (row, id, value) => arrayIntersectionFilter(row, id, value),
    },
    {
      id: 'lastEvent',
      size: 140,
      maxSize: 140,
      enableSorting: true,
      enableHiding: false,
      header: ({ column }) => {
        return <SortableHeader column={column} label="Last Event" />
      },
      filterFn: (row, id, value) => dateRangeFilter(row, id, value),
      accessorFn: (row) => getLastEvent(row).date,
      cell: ({ row }) => {
        const lastEvent = getLastEvent(row.original)
        return (
          <TimestampTooltip timestamp={row.original.lastActivityAt ?? row.original.updatedAt}>
            <div className="w-full truncate">
              <span className="truncate block">{lastEvent.relativeTimeString}</span>
            </div>
          </TimestampTooltip>
        )
      },
    },
    {
      id: 'createdAt',
      size: 180,
      maxSize: 180,
      enableSorting: true,
      enableHiding: false,
      header: ({ column }) => {
        return <SortableHeader column={column} label="Created At" />
      },
      accessorFn: (row) => (row.createdAt ? new Date(row.createdAt) : new Date()),
      cell: ({ row }) => {
        const timestamp = formatTimestamp(row.original.createdAt)
        return (
          <TimestampTooltip timestamp={row.original.createdAt}>
            <div className="w-full truncate">
              <span className="truncate block">{timestamp}</span>
            </div>
          </TimestampTooltip>
        )
      },
    },
    {
      id: 'actions',
      size: 88,
      minSize: 88,
      maxSize: 88,
      enableHiding: false,
      cell: ({ row }) => (
        <div className="w-full flex justify-end">
          <SandboxTableActions
            sandbox={row.original}
            writePermitted={writePermitted}
            deletePermitted={deletePermitted}
            isLoading={sandboxIsLoading[row.original.id]}
            onStart={handleStart}
            onStop={handleStop}
            onDelete={handleDelete}
            onArchive={handleArchive}
            onVnc={handleVnc}
            onCreateSshAccess={handleCreateSshAccess}
            onRevokeSshAccess={handleRevokeSshAccess}
            onRecover={handleRecover}
            onScreenRecordings={handleScreenRecordings}
            onCreateSnapshot={() => handleCreateSnapshot(row.original.id)}
            onFork={() => handleFork(row.original.id)}
            onViewForks={() => handleViewForks(row.original.id)}
          />
        </div>
      ),
    },
  ]

  return columns
}

function getDisplayName(sandbox: Sandbox): string {
  // If the sandbox is destroying and the name starts with "DESTROYED_", trim the prefix and timestamp
  if (sandbox.desiredState === SandboxDesiredState.DESTROYED && sandbox.name.startsWith('DESTROYED_')) {
    // Remove "DESTROYED_" prefix and everything after the last underscore (timestamp)
    const withoutPrefix = sandbox.name.substring(10) // Remove "DESTROYED_"
    const lastUnderscoreIndex = withoutPrefix.lastIndexOf('_')
    if (lastUnderscoreIndex !== -1) {
      return withoutPrefix.substring(0, lastUnderscoreIndex)
    }
    return withoutPrefix
  }
  return sandbox.name
}

function getLastEvent(sandbox: Sandbox): { date: Date; relativeTimeString: string } {
  return getRelativeTimeString(sandbox.lastActivityAt ?? sandbox.updatedAt)
}
