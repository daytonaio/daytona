/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CopyButton } from '@/components/CopyButton'
import { EllipsisWithTooltip } from '@/components/EllipsisWithTooltip'
import { SandboxLabel } from '@/components/SandboxLabel'
import { SortOrderIcon } from '@/components/SortIcon'
import { TimestampTooltip } from '@/components/TimestampTooltip'
import { SandboxState as SandboxStateComponent } from '@/components/sandboxes/SandboxState'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { MiddleTruncate } from '@/components/ui/middle-truncate'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { getRelativeTimeString } from '@/lib/utils'
import { getTableColumnMaxResizeSize } from '@/lib/utils/table'
import { SandboxDesiredState, SandboxListItem } from '@daytona/api-client'
import { Column, ColumnDef, RowData, Table } from '@tanstack/react-table'
import { Loader2 } from 'lucide-react'
import React from 'react'
import { SandboxTableActions } from './SandboxTableActions'
import { getSandboxClassIcon, getSandboxClassLabel } from './constants'

type SandboxTableMeta = {
  sandboxIsLoading: Record<string, boolean>
  writePermitted: boolean
  deletePermitted: boolean
  selectableCount: number
  handleStart: (id: string) => void
  handleStop: (id: string) => void
  handleDelete: (id: string) => void
  handleArchive: (id: string) => void
  handleVnc: (id: string) => void
  handleCreateSshAccess: (id: string) => void
  handleRevokeSshAccess: (id: string) => void
  handleRecover: (id: string) => void
  getRegionName: (regionId: string) => string | undefined
  handleScreenRecordings: (id: string) => void
  handleCreateSnapshot: (id: string) => void
  handleFork: (id: string) => void
  handlePause: (id: string) => void
  handleViewForks: (id: string) => void
  handleOpenTerminal: (sandbox: SandboxListItem) => void
}

declare module '@tanstack/react-table' {
  interface TableMeta<TData extends RowData> {
    sandbox?: TData extends SandboxListItem ? SandboxTableMeta : never
  }
}

interface SortableHeaderProps {
  column: Column<SandboxListItem, unknown>
  label: string
  dataState?: string
}

const getMeta = (table: Table<SandboxListItem>) => {
  return table.options.meta?.sandbox as SandboxTableMeta
}

const SortableHeader: React.FC<SortableHeaderProps> = ({ column, label, dataState }) => {
  const sortDirection = column.getIsSorted()

  return (
    <button
      type="button"
      onClick={() => column.toggleSorting(sortDirection === 'asc')}
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
}: Pick<SandboxListItem, 'state' | 'errorReason' | 'recoverable'>) {
  return (
    <div className="w-full truncate">
      <SandboxStateComponent state={state} errorReason={errorReason} recoverable={recoverable} animate />
    </div>
  )
})

const columns: ColumnDef<SandboxListItem>[] = [
  {
    id: 'select',
    size: 44,
    minSize: 44,
    maxSize: 44,
    header: ({ table }) => {
      const { writePermitted, deletePermitted, selectableCount } = getMeta(table)
      const selectedCount = table.getSelectedRowModel().rows.length
      const anySelectable = selectableCount > 0
      const allSelected = selectedCount > 0 && selectedCount === selectableCount
      const partiallySelected = selectedCount > 0 && selectedCount < selectableCount

      if ((!writePermitted && !deletePermitted) || !anySelectable) {
        return null
      }

      return (
        <div className="flex justify-center">
          <Checkbox
            checked={allSelected || (partiallySelected && 'indeterminate')}
            onCheckedChange={() => {
              table.getRowModel().rows.forEach((row) => {
                if (!row.getCanSelect()) {
                  row.toggleSelected(false)
                  return
                }

                row.toggleSelected(!allSelected)
              })
            }}
            aria-label="Select all"
          />
        </div>
      )
    },
    cell: ({ row, table }) => {
      const { writePermitted, deletePermitted, sandboxIsLoading } = getMeta(table)
      const isLoading = Boolean(sandboxIsLoading[row.original.id])

      if ((!writePermitted && !deletePermitted) || !row.getCanSelect()) {
        return null
      }

      if (isLoading) {
        return (
          <div className="flex justify-center">
            <Loader2 className="h-4 w-4 animate-spin" />
          </div>
        )
      }

      return (
        <div className="flex justify-center">
          <Checkbox
            checked={row.getIsSelected()}
            onCheckedChange={(value) => row.toggleSelected(!!value)}
            aria-label="Select row"
            onClick={(event) => event.stopPropagation()}
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
    header: ({ column }) => <SortableHeader column={column} label="Name" />,
    accessorKey: 'name',
    cell: ({ row }) => {
      const displayName = getDisplayName(row.original)
      return (
        <div className="w-full truncate">
          <span className="block truncate">{displayName}</span>
        </div>
      )
    },
  },
  {
    id: 'id',
    size: 240,
    minSize: 150,
    maxSize: getTableColumnMaxResizeSize(360),
    enableSorting: false,
    enableHiding: true,
    header: () => <span>UUID</span>,
    accessorKey: 'id',
    cell: ({ row }) => {
      const id = row.original.id
      return (
        <div className="w-full min-w-0 flex items-center gap-1 group/copy-button">
          <MiddleTruncate value={id} start={8} end={4} className="font-mono text-muted-foreground" />
          <CopyButton value={id} size="icon-xs" autoHide tooltipText="Copy UUID" />
        </div>
      )
    },
  },
  {
    id: 'state',
    size: 110,
    minSize: 110,
    enableSorting: false,
    enableHiding: true,
    header: () => <span>State</span>,
    cell: ({ row }) => (
      <SandboxStateCell
        state={row.original.state}
        errorReason={row.original.errorReason}
        recoverable={row.original.recoverable}
      />
    ),
    accessorKey: 'state',
  },
  {
    id: 'sandboxClass',
    size: 64,
    minSize: 64,
    maxSize: 64,
    enableSorting: false,
    enableHiding: true,
    header: () => <span>Class</span>,
    cell: ({ row }) => {
      const sandboxClass = row.original.sandboxClass
      const Icon = getSandboxClassIcon(sandboxClass)
      const label = getSandboxClassLabel(sandboxClass)
      return (
        <Tooltip>
          <TooltipTrigger asChild>
            <span className="inline-flex items-center" aria-label={label}>
              <Icon className="size-4 text-muted-foreground shrink-0" />
            </span>
          </TooltipTrigger>
          <TooltipContent>{label}</TooltipContent>
        </Tooltip>
      )
    },
    accessorKey: 'sandboxClass',
  },
  {
    id: 'snapshot',
    size: 150,
    enableSorting: false,
    enableHiding: true,
    header: () => <span>Snapshot</span>,
    cell: ({ row }) => (
      <div className="w-full truncate">
        {row.original.snapshot ? (
          <EllipsisWithTooltip>{row.original.snapshot}</EllipsisWithTooltip>
        ) : (
          <div className="truncate text-muted-foreground/50">-</div>
        )}
      </div>
    ),
    accessorKey: 'snapshot',
  },
  {
    id: 'region',
    size: 120,
    enableSorting: false,
    enableHiding: true,
    header: () => <span>Region</span>,
    cell: ({ row, table }) => {
      const { getRegionName } = getMeta(table)
      return (
        <div className="w-full truncate">
          <span className="block truncate">{getRegionName(row.original.target) ?? row.original.target}</span>
        </div>
      )
    },
    accessorKey: 'target',
  },
  {
    id: 'resources',
    size: 190,
    enableSorting: false,
    enableHiding: true,
    header: () => <span>Resources</span>,
    cell: ({ row }) => (
      <div className="flex w-full items-center gap-2 truncate">
        <div className="whitespace-nowrap">
          {row.original.cpu} <span className="text-muted-foreground">vCPU</span>
        </div>
        <div className="inline-block h-6 w-[1px] rounded-full bg-muted-foreground/20" />
        <div className="whitespace-nowrap">
          {row.original.memory} <span className="text-muted-foreground">GiB</span>
        </div>
        <div className="inline-block h-6 w-[1px] rounded-full bg-muted-foreground/20" />
        <div className="whitespace-nowrap">
          {row.original.disk} <span className="text-muted-foreground">GiB</span>
        </div>
        {row.original.gpu > 0 && (
          <>
            <div className="inline-block h-6 w-[1px] rounded-full bg-muted-foreground/20" />
            <div className="whitespace-nowrap">
              {row.original.gpu} <span className="text-muted-foreground">GPU</span>
            </div>
          </>
        )}
      </div>
    ),
  },
  {
    id: 'labels',
    size: 120,
    maxSize: getTableColumnMaxResizeSize(120),
    enableSorting: false,
    enableHiding: true,
    header: () => <span>Labels</span>,
    cell: ({ row }) => {
      const labelEntries = Object.entries(row.original.labels ?? {})

      if (labelEntries.length === 0) {
        return <div className="max-w-md truncate text-muted-foreground/50">-</div>
      }

      return (
        <Tooltip>
          <TooltipTrigger asChild>
            <Badge variant="info">{labelEntries.length === 1 ? '1 label' : `${labelEntries.length} labels`}</Badge>
          </TooltipTrigger>
          <TooltipContent className="scrollbar-sm max-h-[400px] max-w-[300px] overflow-y-auto p-2">
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
  },
  {
    id: 'lastEvent',
    size: 140,
    maxSize: getTableColumnMaxResizeSize(140),
    enableSorting: true,
    enableHiding: true,
    header: ({ column }) => <SortableHeader column={column} label="Last Event" />,
    accessorFn: (row) => getLastEvent(row).date,
    cell: ({ row }) => {
      const lastEvent = getLastEvent(row.original)
      return (
        <TimestampTooltip timestamp={row.original.lastActivityAt ?? row.original.updatedAt}>
          <div className="w-full truncate">
            <span className="block truncate">{lastEvent.relativeTimeString}</span>
          </div>
        </TimestampTooltip>
      )
    },
  },
  {
    id: 'createdAt',
    size: 180,
    maxSize: getTableColumnMaxResizeSize(180),
    enableSorting: true,
    enableHiding: true,
    header: ({ column }) => <SortableHeader column={column} label="Created" />,
    accessorFn: (row) => (row.createdAt ? new Date(row.createdAt) : new Date()),
    cell: ({ row }) => {
      const timestamp = getRelativeTimeString(row.original.createdAt)
      return (
        <TimestampTooltip timestamp={row.original.createdAt}>
          <div className="w-full truncate">
            <span className="block truncate">{timestamp.relativeTimeString}</span>
          </div>
        </TimestampTooltip>
      )
    },
  },
  {
    id: 'actions',
    size: 136,
    maxSize: 136,
    minSize: 136,
    enableHiding: false,
    cell: ({ row, table }) => {
      const {
        writePermitted,
        deletePermitted,
        sandboxIsLoading,
        handleStart,
        handleStop,
        handleDelete,
        handleArchive,
        handleVnc,
        handleCreateSshAccess,
        handleRevokeSshAccess,
        handleRecover,
        handleScreenRecordings,
        handleCreateSnapshot,
        handleFork,
        handlePause,
        handleViewForks,
        handleOpenTerminal,
      } = getMeta(table)

      return (
        <div className="flex w-full justify-end">
          <SandboxTableActions
            sandbox={row.original}
            writePermitted={writePermitted}
            deletePermitted={deletePermitted}
            isLoading={Boolean(sandboxIsLoading[row.original.id])}
            onStart={handleStart}
            onStop={handleStop}
            onDelete={handleDelete}
            onArchive={handleArchive}
            onVnc={handleVnc}
            onCreateSshAccess={handleCreateSshAccess}
            onRevokeSshAccess={handleRevokeSshAccess}
            onPause={handlePause}
            onRecover={handleRecover}
            onScreenRecordings={handleScreenRecordings}
            onCreateSnapshot={() => handleCreateSnapshot(row.original.id)}
            onFork={() => handleFork(row.original.id)}
            onViewForks={() => handleViewForks(row.original.id)}
            onOpenTerminal={() => handleOpenTerminal(row.original)}
          />
        </div>
      )
    },
  },
  {
    id: 'isPublic',
    enableHiding: false,
    enableSorting: false,
  },
  {
    id: 'isRecoverable',
    enableHiding: false,
    enableSorting: false,
  },
]

function getDisplayName(sandbox: SandboxListItem): string {
  if (sandbox.desiredState === SandboxDesiredState.DESTROYED && sandbox.name.startsWith('DESTROYED_')) {
    const withoutPrefix = sandbox.name.substring(10)
    const lastUnderscoreIndex = withoutPrefix.lastIndexOf('_')
    if (lastUnderscoreIndex !== -1) {
      return withoutPrefix.substring(0, lastUnderscoreIndex)
    }
    return withoutPrefix
  }
  return sandbox.name
}

function getLastEvent(sandbox: SandboxListItem): { date: Date; relativeTimeString: string } {
  return getRelativeTimeString(sandbox.lastActivityAt)
}

export { columns }
export type { SandboxTableMeta }
