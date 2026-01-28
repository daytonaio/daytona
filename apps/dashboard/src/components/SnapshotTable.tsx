/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SnapshotDto, SnapshotState, OrganizationRolePermissionsEnum, RunnerClass } from '@daytonaio/api-client'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { TableHeader, TableRow, TableHead, TableBody, TableCell, Table } from './ui/table'
import { Button } from './ui/button'
import { useMemo, useState } from 'react'
import { AlertTriangle, CheckCircle, MoreHorizontal, Timer, Pause, Box } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from './ui/dropdown-menu'
import { Pagination } from './Pagination'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from './ui/tooltip'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { Checkbox } from './ui/checkbox'
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover'
import { getRelativeTimeString } from '@/lib/utils'
import { TableEmptyState } from './TableEmptyState'
import { Loader2 } from 'lucide-react'
import { Badge } from './ui/badge'

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

interface DataTableProps {
  data: SnapshotDto[]
  loading: boolean
  loadingSnapshots: Record<string, boolean>
  getRegionName: (regionId: string) => string | undefined
  onDelete: (snapshot: SnapshotDto) => void
  onBulkDelete?: (snapshots: SnapshotDto[]) => void
  onActivate?: (snapshot: SnapshotDto) => void
  onDeactivate?: (snapshot: SnapshotDto) => void
  pagination: {
    pageIndex: number
    pageSize: number
  }
  pageCount: number
  totalItems: number
  onPaginationChange: (pagination: { pageIndex: number; pageSize: number }) => void
}

export function SnapshotTable({
  data,
  loading,
  loadingSnapshots,
  getRegionName,
  onDelete,
  onActivate,
  onDeactivate,
  pagination,
  pageCount,
  totalItems,
  onBulkDelete,
  onPaginationChange,
}: DataTableProps) {
  const { authenticatedUserHasPermission } = useSelectedOrganization()

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_SNAPSHOTS),
    [authenticatedUserHasPermission],
  )

  const deletePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_SNAPSHOTS),
    [authenticatedUserHasPermission],
  )

  const [sorting, setSorting] = useState<SortingState>([])

  const columns = useMemo(
    () =>
      getColumns({
        onDelete,
        onActivate,
        onDeactivate,
        loadingSnapshots,
        getRegionName,
        writePermitted,
        deletePermitted,
      }),
    [onDelete, onActivate, onDeactivate, loadingSnapshots, getRegionName, writePermitted, deletePermitted],
  )

  const columnsWithSelection = useMemo(() => {
    const selectionColumn: ColumnDef<SnapshotDto> = {
      id: 'select',
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected() || (table.getIsSomePageRowsSelected() && 'indeterminate')}
          onCheckedChange={(value) => {
            table.getRowModel().rows.forEach((row) => {
              if (!row.original.general) {
                row.toggleSelected()
              }
            })
          }}
          aria-label="Select all"
          disabled={!deletePermitted || loading}
          className="translate-y-[2px]"
        />
      ),
      cell: ({ row }) => {
        if (loadingSnapshots[row.original.id]) {
          return <Loader2 className="w-4 h-4 animate-spin" />
        }

        if (row.original.general) {
          return null
        }

        return (
          <Checkbox
            checked={row.getIsSelected()}
            onCheckedChange={(value) => row.toggleSelected(!!value)}
            aria-label="Select row"
            disabled={!deletePermitted || loadingSnapshots[row.original.id] || loading}
            className="translate-y-[2px]"
          />
        )
      },
      enableSorting: false,
      enableHiding: false,
    }

    return deletePermitted ? [selectionColumn, ...columns] : columns
  }, [deletePermitted, columns, loading, loadingSnapshots])

  const table = useReactTable({
    data,
    columns: columnsWithSelection,
    getCoreRowModel: getCoreRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    manualPagination: true,
    pageCount: pageCount || 1,
    onPaginationChange: pagination
      ? (updater) => {
          const newPagination = typeof updater === 'function' ? updater(table.getState().pagination) : updater
          onPaginationChange(newPagination)
        }
      : undefined,
    state: {
      sorting,
      pagination: {
        pageIndex: pagination?.pageIndex || 0,
        pageSize: pagination?.pageSize || 10,
      },
    },
    getRowId: (row) => row.id,
    enableRowSelection: deletePermitted,
  })

  const selectedRows = table.getSelectedRowModel().rows
  const [bulkDeleteConfirmationOpen, setBulkDeleteConfirmationOpen] = useState(false)
  const selectedImages = selectedRows.map((row) => row.original)

  const handleBulkDelete = () => {
    if (onBulkDelete && selectedImages.length > 0) {
      onBulkDelete(selectedImages)
    }
  }

  return (
    <div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead className="px-2" key={header.id}>
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
                <TableCell colSpan={columnsWithSelection.length} className="h-24 text-center">
                  Loading...
                </TableCell>
              </TableRow>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() ? 'selected' : undefined}
                  className={`${
                    loadingSnapshots[row.original.id] || row.original.state === SnapshotState.REMOVING
                      ? 'opacity-50 pointer-events-none'
                      : ''
                  } ${row.original.general ? 'pointer-events-none' : ''}`}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell className="px-2" key={cell.id}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableEmptyState
                colSpan={columns.length}
                message="No Snapshots yet."
                icon={<Box className="w-8 h-8" />}
                description={
                  <div className="space-y-2">
                    <p>
                      Snapshots are reproducible, pre-configured environments based on any Docker-compatible image. Use
                      them to define language runtimes, dependencies, and tools for your sandboxes.
                    </p>
                    <p>
                      Create one from the Dashboard, CLI, or SDK to get started. <br />
                      <a
                        href="https://www.daytona.io/docs/snapshots"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-primary hover:underline font-medium"
                      >
                        Read the Snapshots guide
                      </a>{' '}
                      to learn more.
                    </p>
                  </div>
                }
              />
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-between space-x-2 py-4">
        {deletePermitted && selectedRows.length > 0 && (
          <Popover open={bulkDeleteConfirmationOpen} onOpenChange={setBulkDeleteConfirmationOpen}>
            <PopoverTrigger>
              <Button variant="destructive" size="sm" className="h-8">
                Bulk Delete
              </Button>
            </PopoverTrigger>
            <PopoverContent side="top">
              <div className="flex flex-col gap-4">
                <p>Are you sure you want to delete these Snapshots?</p>
                <div className="flex items-center space-x-2">
                  <Button
                    variant="destructive"
                    onClick={() => {
                      handleBulkDelete()
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
        )}
        <Pagination table={table} selectionEnabled={deletePermitted} entityName="Snapshots" totalItems={totalItems} />
      </div>
    </div>
  )
}

const getColumns = ({
  onDelete,
  onActivate,
  onDeactivate,
  loadingSnapshots,
  getRegionName,
  writePermitted,
  deletePermitted,
}: {
  onDelete: (snapshot: SnapshotDto) => void
  onActivate?: (snapshot: SnapshotDto) => void
  onDeactivate?: (snapshot: SnapshotDto) => void
  loadingSnapshots: Record<string, boolean>
  getRegionName: (regionId: string) => string | undefined
  writePermitted: boolean
  deletePermitted: boolean
}): ColumnDef<SnapshotDto>[] => {
  const columns: ColumnDef<SnapshotDto>[] = [
    {
      accessorKey: 'name',
      header: 'Name',
      cell: ({ row }) => {
        const snapshot = row.original
        return (
          <div className="flex items-center gap-2">
            {snapshot.name}
            {snapshot.general && (
              <span className="px-2 py-0.5 text-xs rounded-full bg-green-100 text-blue-800 dark:bg-green-900 dark:text-green-300">
                System
              </span>
            )}
          </div>
        )
      },
    },
    {
      id: 'runnerClass',
      header: 'OS',
      cell: ({ row }) => {
        const snapshot = row.original
        if (!snapshot.runnerClass) {
          return '-'
        }
        return <RunnerClassIcon runnerClass={snapshot.runnerClass} />
      },
    },
    {
      accessorKey: 'imageName',
      header: 'Image',
      cell: ({ row }) => {
        const snapshot = row.original
        // Don't show image for Windows-based snapshots (they use qcow2 disk images, not Docker images)
        if (snapshot.runnerClass === 'windows-exp' || snapshot.runnerClass === 'linux-exp') {
          return '-'
        }
        if (!snapshot.imageName && snapshot.buildInfo) {
          return (
            <Badge variant="secondary" className="rounded-sm px-1 font-medium">
              DECLARATIVE BUILD
            </Badge>
          )
        }
        return snapshot.imageName
      },
    },
    {
      accessorKey: 'regionIds',
      header: 'Region',
      cell: ({ row }) => {
        const snapshot = row.original
        if (!snapshot.regionIds?.length) {
          return '-'
        }

        const regionNames = snapshot.regionIds.map((id) => getRegionName(id) ?? id)
        const firstRegion = regionNames[0]
        const remainingCount = regionNames.length - 1

        if (remainingCount === 0) {
          return (
            <span className="truncate max-w-[150px] block" title={firstRegion}>
              {firstRegion}
            </span>
          )
        }

        return (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <div className="flex items-center gap-1.5">
                  <span className="truncate max-w-[150px]">{firstRegion}</span>
                  <Badge variant="secondary" className="text-xs px-1.5 py-0 h-5">
                    +{remainingCount}
                  </Badge>
                </div>
              </TooltipTrigger>
              <TooltipContent>
                <div className="flex flex-col gap-1">
                  {regionNames.map((name, idx) => (
                    <span key={idx}>{name}</span>
                  ))}
                </div>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        )
      },
    },
    {
      id: 'resources',
      header: 'Resources',
      cell: ({ row }) => {
        const snapshot = row.original
        return `${snapshot.cpu}vCPU / ${snapshot.mem}GiB / ${snapshot.disk}GiB`
      },
    },
    {
      accessorKey: 'state',
      header: 'State',
      cell: ({ row }) => {
        const snapshot = row.original
        const color = getStateColor(snapshot.state)

        if (
          (snapshot.state === SnapshotState.ERROR || snapshot.state === SnapshotState.BUILD_FAILED) &&
          !!snapshot.errorReason
        ) {
          return (
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger>
                  <div className={`flex items-center gap-2 ${color}`}>
                    {getStateIcon(snapshot.state)}
                    {getStateLabel(snapshot.state)}
                  </div>
                </TooltipTrigger>
                <TooltipContent>
                  <p className="max-w-[300px]">{snapshot.errorReason}</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          )
        }

        return (
          <div className={`flex items-center gap-2 ${color}`}>
            {getStateIcon(snapshot.state)}
            {getStateLabel(snapshot.state)}
          </div>
        )
      },
    },
    {
      accessorKey: 'createdAt',
      header: 'Created',
      cell: ({ row }) => {
        const snapshot = row.original
        return snapshot.general ? '' : getRelativeTimeString(snapshot.createdAt).relativeTimeString
      },
    },
    {
      accessorKey: 'lastUsedAt',
      header: 'Last Used',
      cell: ({ row }) => {
        const snapshot = row.original
        return snapshot.general ? '' : getRelativeTimeString(snapshot.lastUsedAt).relativeTimeString
      },
    },
    {
      id: 'actions',
      cell: ({ row }) => {
        if ((!writePermitted && !deletePermitted) || row.original.general) {
          return null
        }

        const showActivate = writePermitted && onActivate && row.original.state === SnapshotState.INACTIVE
        const showDeactivate = writePermitted && onDeactivate && row.original.state === SnapshotState.ACTIVE
        const showDelete = deletePermitted

        const showSeparator = (showActivate || showDeactivate) && showDelete

        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="h-8 w-8 p-0">
                <span className="sr-only">Open menu</span>
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {showActivate && (
                <DropdownMenuItem
                  onClick={() => onActivate(row.original)}
                  className="cursor-pointer"
                  disabled={loadingSnapshots[row.original.id]}
                >
                  Activate
                </DropdownMenuItem>
              )}
              {showDeactivate && (
                <DropdownMenuItem
                  onClick={() => onDeactivate(row.original)}
                  className="cursor-pointer"
                  disabled={loadingSnapshots[row.original.id]}
                >
                  Deactivate
                </DropdownMenuItem>
              )}
              {showSeparator && <DropdownMenuSeparator />}
              {showDelete && (
                <DropdownMenuItem
                  onClick={() => onDelete(row.original)}
                  className="cursor-pointer text-red-600 dark:text-red-400"
                  disabled={loadingSnapshots[row.original.id]}
                >
                  Delete
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        )
      },
    },
  ]

  return columns
}

const getStateIcon = (state: SnapshotState) => {
  switch (state) {
    case SnapshotState.ACTIVE:
      return <CheckCircle className="w-4 h-4 flex-shrink-0" />
    case SnapshotState.INACTIVE:
      return <Pause className="w-4 h-4 flex-shrink-0" />
    case SnapshotState.ERROR:
    case SnapshotState.BUILD_FAILED:
      return <AlertTriangle className="w-4 h-4 flex-shrink-0" />
    default:
      return <Timer className="w-4 h-4 flex-shrink-0" />
  }
}

const getStateColor = (state: SnapshotState) => {
  switch (state) {
    case SnapshotState.ACTIVE:
      return 'text-green-500'
    case SnapshotState.INACTIVE:
      return 'text-gray-500 dark:text-gray-400'
    case SnapshotState.ERROR:
    case SnapshotState.BUILD_FAILED:
      return 'text-red-500'
    default:
      return 'text-gray-600 dark:text-gray-400'
  }
}

const getStateLabel = (state: SnapshotState) => {
  // TODO: remove when removing is migrated to deleted
  if (state === SnapshotState.REMOVING) {
    return 'Deleting'
  }
  return state
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ')
}
