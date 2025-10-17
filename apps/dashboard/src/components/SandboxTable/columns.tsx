/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
import { Sandbox, SandboxDesiredState } from '@daytonaio/api-client'
import { ColumnDef } from '@tanstack/react-table'
import { ArrowUp, ArrowDown } from 'lucide-react'
import { Checkbox } from '../ui/checkbox'
import { getRelativeTimeString } from '@/lib/utils'
import { SandboxState as SandboxStateComponent } from './SandboxState'
import { SandboxTableActions } from './SandboxTableActions'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '../ui/tooltip'

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
  sandboxIsLoading: Record<string, boolean>
  writePermitted: boolean
  deletePermitted: boolean
  handleCreateSshAccess: (id: string) => void
  handleRevokeSshAccess: (id: string) => void
}

export function getColumns({
  handleStart,
  handleStop,
  handleDelete,
  handleArchive,
  handleVnc,
  getWebTerminalUrl,
  sandboxIsLoading,
  writePermitted,
  deletePermitted,
  handleCreateSshAccess,
  handleRevokeSshAccess,
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
      size: 30,
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected() || (table.getIsSomePageRowsSelected() && 'indeterminate')}
          onCheckedChange={(value) => {
            for (const row of table.getRowModel().rows) {
              if (sandboxIsLoading[row.original.id]) {
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
      id: 'name',
      size: 320,
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
      size: 320,
      enableSorting: false,
      enableHiding: true,
      header: () => {
        return <span>UUID</span>
      },
      accessorKey: 'id',
      cell: ({ row }) => {
        return (
          <div className="w-full truncate">
            <span className="truncate block">{row.original.id}</span>
          </div>
        )
      },
    },
    {
      id: 'state',
      size: 140,
      enableSorting: true,
      enableHiding: false,
      header: ({ column }) => {
        return <SortableHeader column={column} label="State" />
      },
      cell: ({ row }) => (
        <div className="w-full truncate">
          <SandboxStateComponent state={row.original.state} errorReason={row.original.errorReason} />
        </div>
      ),
      accessorKey: 'state',
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
              <div className="truncate">{row.original.snapshot}</div>
            ) : (
              <div className="truncate text-muted-foreground/50">-</div>
            )}
          </div>
        )
      },
      accessorKey: 'snapshot',
    },
    {
      id: 'region',
      size: 100,
      enableSorting: true,
      enableHiding: false,
      header: ({ column }) => {
        return <SortableHeader column={column} label="Region" dataState="sortable" />
      },
      cell: ({ row }) => {
        return (
          <div className="w-full truncate">
            <span className="truncate block">{row.original.target}</span>
          </div>
        )
      },
      accessorKey: 'target',
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
    },
    {
      id: 'labels',
      size: 110,
      enableSorting: false,
      enableHiding: true,
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
    },
    {
      id: 'lastEvent',
      size: 120,
      enableSorting: true,
      enableHiding: false,
      header: ({ column }) => {
        return <SortableHeader column={column} label="Last Event" />
      },
      accessorFn: (row) => getLastEvent(row).date,
      cell: ({ row }) => {
        const lastEvent = getLastEvent(row.original)
        return (
          <div className="w-full truncate">
            <span className="truncate block">{lastEvent.relativeTimeString}</span>
          </div>
        )
      },
    },
    {
      id: 'actions',
      size: 100,
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
            onOpenWebTerminal={handleOpenWebTerminal}
            onCreateSshAccess={handleCreateSshAccess}
            onRevokeSshAccess={handleRevokeSshAccess}
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
  return getRelativeTimeString(sandbox.updatedAt)
}
