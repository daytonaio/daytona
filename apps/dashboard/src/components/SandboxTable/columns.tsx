/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { formatTimestamp, getRelativeTimeString } from '@/lib/utils'
import { Sandbox, SandboxDesiredState, RunnerClass } from '@daytonaio/api-client'
import { ColumnDef } from '@tanstack/react-table'
import { ArrowDown, ArrowUp, Box } from 'lucide-react'
import React from 'react'
import { EllipsisWithTooltip } from '../EllipsisWithTooltip'
import { Checkbox } from '../ui/checkbox'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '../ui/tooltip'
import { SandboxState as SandboxStateComponent } from './SandboxState'
import { SandboxTableActions } from './SandboxTableActions'

const WindowsIcon: React.FC<{ className?: string }> = ({ className }) => (
  <svg viewBox="0 0 24 24" fill="currentColor" className={className}>
    <path d="M0 3.449L9.75 2.1v9.451H0m10.949-9.602L24 0v11.4H10.949M0 12.6h9.75v9.451L0 20.699M10.949 12.6H24V24l-12.9-1.801" />
  </svg>
)

const UbuntuIcon: React.FC<{ className?: string }> = ({ className }) => (
  <svg viewBox="0 0 24 24" fill="currentColor" className={className}>
    <path d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm-1.243 2.398c2.456-.153 4.89.618 6.756 2.14l-1.63 2.252c-2.612-2.026-6.39-1.558-8.418 1.055-.464.598-.81 1.279-1.02 1.999l-2.637-.861c.754-2.571 2.637-4.716 5.008-5.899a9.02 9.02 0 0 1 1.941-.686zm-5.39 9.376c.008-1.074.228-2.134.642-3.12l2.638.861a5.99 5.99 0 0 0 2.013 6.497l-1.63 2.252a9.096 9.096 0 0 1-3.663-6.49zm11.458 5.785a9.04 9.04 0 0 1-6.702 2.17l.304-2.77a6.02 6.02 0 0 0 4.767-1.652 6.02 6.02 0 0 0 .392-8.116l1.632-2.251a9.076 9.076 0 0 1 2.219 6.545 9.074 9.074 0 0 1-2.612 6.074zM3.6 12a1.8 1.8 0 1 1 3.6 0 1.8 1.8 0 0 1-3.6 0zm5.4 6.6a1.8 1.8 0 1 1 3.6 0 1.8 1.8 0 0 1-3.6 0zm4.2-10.8a1.8 1.8 0 1 1 3.6 0 1.8 1.8 0 0 1-3.6 0z" />
  </svg>
)

const AndroidIcon: React.FC<{ className?: string }> = ({ className }) => (
  <svg viewBox="0 0 16 16" fill="currentColor" className={className}>
    <path
      fillRule="evenodd"
      d="M15.48 9.83c-.39-2.392-1.768-4.268-3.653-5.338l1.106-2.432a.75.75 0 1 0-1.366-.62l-1.112 2.446A7.9 7.9 0 0 0 8 3.5a7.9 7.9 0 0 0-2.455.386L4.433 1.44a.75.75 0 1 0-1.366.62l1.106 2.432C2.288 5.562.909 7.438.52 9.83c-.13.798-.178 1.655.107 2.433.325.89.989 1.441 1.768 1.75.701.28 1.54.383 2.404.433.887.052 1.963.054 3.201.054s2.314-.002 3.2-.054c.864-.05 1.704-.154 2.405-.432.78-.31 1.443-.86 1.768-1.75.285-.78.237-1.636.107-2.434M2 10.071C1.53 12.961 3 13 8 13s6.47-.038 6-2.929C13.5 7 11 5 8 5s-5.5 2-6 5.071m8.5 1.179a.75.75 0 0 1-.75-.75V9a.75.75 0 0 1 1.5 0v1.5a.75.75 0 0 1-.75.75m-5.75-.75a.75.75 0 0 0 1.5 0V9a.75.75 0 0 0-1.5 0z"
      clipRule="evenodd"
    />
  </svg>
)

const RunnerClassIcon: React.FC<{ runnerClass: RunnerClass }> = ({ runnerClass }) => {
  const iconClass = 'h-4 w-4'

  switch (runnerClass) {
    case 'linux':
      return (
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <span className="inline-flex">
                <Box className={iconClass} />
              </span>
            </TooltipTrigger>
            <TooltipContent>
              <p>Container</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      )
    case 'linux-exp':
      return (
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <span className="inline-flex">
                <UbuntuIcon className={iconClass} />
              </span>
            </TooltipTrigger>
            <TooltipContent>
              <p>Ubuntu (Experimental)</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      )
    case 'windows-exp':
      return (
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <span className="inline-flex">
                <WindowsIcon className={iconClass} />
              </span>
            </TooltipTrigger>
            <TooltipContent>
              <p>Windows (Experimental)</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      )
    case 'android-exp':
      return (
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <span className="inline-flex">
                <AndroidIcon className={iconClass} />
              </span>
            </TooltipTrigger>
            <TooltipContent>
              <p>Android (Experimental)</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      )
    default:
      return <span>{runnerClass}</span>
  }
}

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
  handleCreateSnapshot: (id: string) => void
  handleScreenRecordings: (id: string) => void
  handleFork: (id: string) => void
  handleViewForks: (id: string) => void
  handleClone: (id: string) => void
  getRegionName: (regionId: string) => string | undefined
  runnerClassMap: Record<string, RunnerClass>
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
  handleCreateSnapshot,
  handleScreenRecordings,
  handleFork,
  handleViewForks,
  handleClone,
  getRegionName,
  runnerClassMap,
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
      cell: ({ row }) => {
        return (
          <div className="w-full truncate">
            <SandboxStateComponent state={row.original.state} errorReason={row.original.errorReason} />
          </div>
        )
      },
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
              <EllipsisWithTooltip>{row.original.snapshot}</EllipsisWithTooltip>
            ) : (
              <div className="truncate text-muted-foreground/50">-</div>
            )}
          </div>
        )
      },
      accessorKey: 'snapshot',
    },
    {
      id: 'runnerClass',
      size: 60,
      enableSorting: false,
      enableHiding: true,
      header: () => {
        return <span>OS</span>
      },
      cell: ({ row }) => {
        const runnerClass = row.original.snapshot ? runnerClassMap[row.original.snapshot] : undefined
        return (
          <div className="w-full flex items-center">
            {runnerClass ? (
              <RunnerClassIcon runnerClass={runnerClass} />
            ) : (
              <span className="text-muted-foreground/50">-</span>
            )}
          </div>
        )
      },
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
            <span className="truncate block">{getRegionName(row.original.target) ?? row.original.target}</span>
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
      id: 'createdAt',
      size: 200,
      enableSorting: true,
      enableHiding: false,
      header: ({ column }) => {
        return <SortableHeader column={column} label="Created At" />
      },
      cell: ({ row }) => {
        const timestamp = formatTimestamp(row.original.createdAt)
        return (
          <div className="w-full truncate">
            <span className="truncate block">{timestamp}</span>
          </div>
        )
      },
    },
    {
      id: 'actions',
      size: 100,
      enableHiding: false,
      cell: ({ row }) => {
        const runnerClass = row.original.snapshot ? runnerClassMap[row.original.snapshot] : undefined
        return (
          <div className="w-full flex justify-end">
            <SandboxTableActions
              sandbox={row.original}
              writePermitted={writePermitted}
              deletePermitted={deletePermitted}
              isLoading={sandboxIsLoading[row.original.id]}
              runnerClass={runnerClass}
              onStart={handleStart}
              onStop={handleStop}
              onDelete={handleDelete}
              onArchive={handleArchive}
              onVnc={handleVnc}
              onOpenWebTerminal={handleOpenWebTerminal}
              onCreateSshAccess={handleCreateSshAccess}
              onRevokeSshAccess={handleRevokeSshAccess}
              onCreateSnapshot={handleCreateSnapshot}
              onScreenRecordings={handleScreenRecordings}
              onFork={handleFork}
              onViewForks={handleViewForks}
              onClone={handleClone}
            />
          </div>
        )
      },
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
