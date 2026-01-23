/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { formatTimestamp, getRelativeTimeString } from '@/lib/utils'
import { Sandbox, SandboxDesiredState, RunnerClass, SandboxBackupStateEnum } from '@daytonaio/api-client'
import { ColumnDef } from '@tanstack/react-table'
import { ArrowDown, ArrowUp, Loader2 } from 'lucide-react'
import React from 'react'
import { EllipsisWithTooltip } from '../EllipsisWithTooltip'
import { Checkbox } from '../ui/checkbox'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '../ui/tooltip'
import { SandboxState as SandboxStateComponent } from './SandboxState'
import { SandboxTableActions } from './SandboxTableActions'

const LinuxIcon: React.FC<{ className?: string }> = ({ className }) => (
  <svg viewBox="0 0 24 24" fill="currentColor" className={className}>
    <path d="M12.504 0c-.155 0-.315.008-.48.021-4.226.333-3.105 4.807-3.17 6.298-.076 1.092-.3 1.953-1.05 3.02-.885 1.051-2.127 2.75-2.716 4.521-.278.832-.41 1.684-.287 2.489a.424.424 0 00-.11.135c-.26.268-.45.6-.663.839-.199.199-.485.267-.797.4-.313.136-.658.269-.864.68-.09.189-.136.394-.132.602 0 .199.027.4.055.536.058.399.116.728.04.97-.249.68-.28 1.145-.106 1.484.174.334.535.47.94.601.81.2 1.91.135 2.774.6.926.466 1.866.67 2.616.47.526-.116.97-.464 1.208-.946.587-.003 1.23-.269 2.26-.334.699-.058 1.574.267 2.577.2.025.134.063.198.114.333l.003.003c.391.778 1.113 1.132 1.884 1.071.771-.06 1.592-.536 2.257-1.306.631-.765 1.683-1.084 2.378-1.503.348-.199.629-.469.649-.853.023-.4-.2-.811-.714-1.376v-.097l-.003-.003c-.17-.2-.25-.535-.338-.926-.085-.401-.182-.786-.492-1.046h-.003c-.059-.054-.123-.067-.188-.135a.357.357 0 00-.19-.064c.431-1.278.264-2.55-.173-3.694-.533-1.41-1.465-2.638-2.175-3.483-.796-1.005-1.576-1.957-1.56-3.368.026-2.152.236-6.133-3.544-6.139zm.529 3.405h.013c.213 0 .396.062.584.198.19.135.33.332.438.533.105.259.158.459.166.724 0-.02.006-.04.006-.06v.105a.086.086 0 01-.004-.021l-.004-.024a1.807 1.807 0 01-.15.706.953.953 0 01-.213.335.71.71 0 00-.088-.042c-.104-.045-.198-.064-.284-.133a1.312 1.312 0 00-.22-.066c.05-.06.146-.133.183-.198.053-.128.082-.264.088-.402v-.02a1.21 1.21 0 00-.061-.4c-.045-.134-.101-.2-.183-.333-.084-.066-.167-.132-.267-.132h-.016c-.093 0-.176.03-.262.132a.8.8 0 00-.205.334 1.18 1.18 0 00-.09.468v.018c0 .138.033.267.09.399.023.066.053.133.09.199a.716.716 0 01-.096.042c-.078.02-.14.04-.192.063l-.004.002a.894.894 0 01-.176.074.712.712 0 01-.256-.329 2.11 2.11 0 01-.164-.703v-.004l-.003-.025c0-.02-.005-.04-.006-.061v-.105c.006-.267.063-.533.166-.725.103-.2.244-.397.434-.533.19-.135.371-.197.584-.197zm-2.512.134c.178-.004.296.076.465.133.32.106.383.133.492.134.109 0 .326 0 .51-.066a.652.652 0 00.333-.198l.003-.003c.03.166.063.332.103.465.06.199.135.332.223.464h-.016c-.155.002-.274.066-.39.2-.109.132-.178.332-.193.535v.003c-.076-.02-.14-.042-.2-.061a.645.645 0 00-.197-.064c-.033 0-.066.003-.1.006-.069.006-.137.018-.199.033a.718.718 0 00-.398.467l-.001.003a.723.723 0 00-.027.198v.004c0 .007 0 .013.002.02 0 .007.002.014.002.021l.002.018.004.023a.86.86 0 00.027.132c.034.126.09.217.166.327l.002.003c.045.069.098.128.154.183a.75.75 0 00.183.138l-.003-.006h.002a.558.558 0 00.332.066 1.047 1.047 0 00.36-.127c.17-.1.269-.267.377-.467.11-.2.174-.4.213-.535.04.063.087.134.124.199l.003.003c.087.2.133.333.16.465.024.133.035.232.035.398 0 .135-.012.265-.027.399-.016.13-.037.265-.065.398l-.004.02c-.051.257-.123.466-.199.665l-.003.006a.727.727 0 01-.264.336c-.184.103-.398.136-.535.166-.133.027-.265.033-.398.033-.133 0-.265-.006-.398-.033-.133-.027-.332-.063-.465-.199a.795.795 0 01-.132-.265c-.043-.132-.074-.2-.116-.267-.117-.197-.299-.332-.516-.398a1.393 1.393 0 00-.531-.065c-.148.013-.298.039-.447.066-.298.058-.597.143-.863.272-.266.127-.5.292-.663.498a1.057 1.057 0 00-.197.4c-.013.066-.024.132-.027.198 0 .068.006.135.02.2l.002.01c.052.197.153.384.298.533.293.303.688.458 1.106.473.42.013.855-.103 1.24-.32l.003-.002.003-.001a2.5 2.5 0 00.352-.252l.003-.002.003-.003c.11.197.264.398.465.465.132.046.267.056.4.046.132-.006.264-.033.398-.066a3.146 3.146 0 00.795-.334l.003-.002c.3-.2.501-.461.663-.733.16-.265.28-.535.398-.798a6.797 6.797 0 00.332-.865c.052-.197.083-.4.105-.598l.002-.016.002-.022a2.472 2.472 0 00.024-.467v-.064c0-.038-.004-.076-.006-.114l-.005-.066a1.877 1.877 0 00-.115-.465 1.564 1.564 0 00-.465-.665l-.002-.002a1.556 1.556 0 00-.198-.132l.003-.002c.195.063.39.067.586.066.197 0 .397-.018.598-.066.133-.034.266-.084.399-.134.13-.053.265-.116.398-.2.132-.081.264-.181.377-.299a1.42 1.42 0 00.299-.447c.053-.132.088-.268.106-.4.013-.134.016-.267.016-.4 0-.133-.003-.265-.016-.398-.013-.133-.045-.265-.1-.398-.055-.132-.13-.264-.23-.377a1.41 1.41 0 00-.365-.299 1.59 1.59 0 00-.4-.166 2.007 2.007 0 00-.797-.065c-.133.006-.265.019-.398.05-.116.027-.23.06-.342.1h-.002l-.003.002a4.647 4.647 0 00-.4.183c-.133.066-.265.14-.377.22-.117.085-.22.168-.305.265l-.005.008c-.1.116-.165.246-.218.377a1.367 1.367 0 00-.109.533c.003.131.02.262.05.392.03.133.072.265.13.377a1.4 1.4 0 00.219.332c.023.024.047.047.072.07l.02.017a1.273 1.273 0 00.327.2l-.04.016c-.094.033-.183.07-.268.11l-.016.008a2.213 2.213 0 00-.37.227 2.39 2.39 0 00-.358.316l-.003.003a2.166 2.166 0 00-.281.366 1.96 1.96 0 00-.17.333l-.001.002a1.748 1.748 0 00-.109.401l-.002.016a1.768 1.768 0 00-.028.465c.006.135.024.27.05.4.028.132.063.264.11.396.049.133.108.257.177.377l.003.003c.05.089.105.172.165.252l-.002-.003c-.168.123-.303.265-.39.465l-.002.006a.871.871 0 00-.076.364v.024c0 .066.006.132.02.197.007.066.02.13.038.193.038.131.095.255.167.366.072.11.16.21.26.295a1.157 1.157 0 00.632.268c.079.008.158.01.238.006h.024l.022.001c.133.006.267 0 .4-.02.132-.02.264-.052.396-.1a1.605 1.605 0 00.72-.499c.112-.132.2-.28.26-.44a1.387 1.387 0 00.109-.532 1.46 1.46 0 00-.09-.533 1.506 1.506 0 00-.23-.423l-.003-.003a1.94 1.94 0 00-.334-.3 2.12 2.12 0 00-.398-.2l-.004-.002c.06-.03.12-.063.177-.1.059-.035.116-.073.17-.114l.008-.005a1.606 1.606 0 00.331-.332 1.47 1.47 0 00.199-.4c.05-.134.08-.268.09-.4a1.57 1.57 0 00-.006-.4 1.593 1.593 0 00-.1-.4 1.586 1.586 0 00-.196-.361c-.018-.025-.037-.05-.057-.074l.027-.028c.13-.132.225-.294.29-.467.064-.169.096-.348.1-.531a1.82 1.82 0 00-.065-.531 1.635 1.635 0 00-.222-.467c.066-.066.127-.14.18-.22.072-.105.13-.217.175-.336l.001-.003c.052-.132.084-.268.096-.402.013-.133.006-.266-.016-.399-.022-.132-.065-.264-.118-.377a1.175 1.175 0 00-.228-.332 1.12 1.12 0 00-.333-.226c-.066-.03-.133-.054-.2-.075l-.003-.001a1.08 1.08 0 00-.398-.05c-.133.003-.265.02-.398.05z" />
  </svg>
)

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

const RunnerClassIcon: React.FC<{ runnerClass: RunnerClass }> = ({ runnerClass }) => {
  const iconClass = 'h-4 w-4'

  switch (runnerClass) {
    case 'linux':
      return (
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <span className="inline-flex">
                <LinuxIcon className={iconClass} />
              </span>
            </TooltipTrigger>
            <TooltipContent>
              <p>Linux</p>
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
        // Show "Snapshotting" when backup state is in progress for experimental runners (linux-exp, windows-exp)
        // These runners support snapshot/fork operations that set backupState to IN_PROGRESS
        const runnerClass = row.original.snapshot ? runnerClassMap[row.original.snapshot] : undefined
        const isExperimentalRunner = runnerClass === 'linux-exp' || runnerClass === 'windows-exp'

        if (isExperimentalRunner && row.original.backupState === SandboxBackupStateEnum.IN_PROGRESS) {
          return (
            <div className="w-full truncate">
              <div className="flex items-center gap-1">
                <Loader2 className="w-4 h-4 animate-spin" />
                <span className="truncate">Snapshotting</span>
              </div>
            </div>
          )
        }
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
              isLoading={
                sandboxIsLoading[row.original.id] || row.original.backupState === SandboxBackupStateEnum.IN_PROGRESS
              }
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
