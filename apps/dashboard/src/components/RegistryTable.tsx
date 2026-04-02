/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DockerRegistry, OrganizationRolePermissionsEnum } from '@daytonaio/api-client'
import {
  ColumnDef,
  ColumnFiltersState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { DebouncedInput } from './DebouncedInput'
import { TableHeader, TableRow, TableHead, TableBody, TableCell, Table, TableContainer } from './ui/table'
import { Button } from './ui/button'
import { useMemo, useState } from 'react'
import { MoreHorizontal, Package } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from './ui/dropdown-menu'
import { Pagination } from './Pagination'
import { PageFooterPortal } from './PageLayout'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { cn } from '@/lib/utils'
import {
  getColumnPinningBorderClasses,
  getColumnPinningClasses,
  getColumnPinningStyles,
  getExplicitColumnSize,
} from '@/lib/utils/table'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { Skeleton } from './ui/skeleton'
import { TableEmptyState } from './TableEmptyState'

const FIXED_COLUMN_IDS = ['actions']

interface DataTableProps {
  data: DockerRegistry[]
  loading: boolean
  onDelete: (id: string) => void
  onEdit: (registry: DockerRegistry) => void
}

export function RegistryTable({ data, loading, onDelete, onEdit }: DataTableProps) {
  const { authenticatedUserHasPermission } = useSelectedOrganization()

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_REGISTRIES),
    [authenticatedUserHasPermission],
  )

  const deletePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_REGISTRIES),
    [authenticatedUserHasPermission],
  )

  const [sorting, setSorting] = useState<SortingState>([])
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
  const columns = getColumns({ onDelete, onEdit, loading, writePermitted, deletePermitted })
  const table = useReactTable({
    data,
    columns,
    defaultColumn: {
      minSize: 0,
    },
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    state: {
      sorting,
      columnFilters,
    },
    initialState: {
      pagination: {
        pageSize: DEFAULT_PAGE_SIZE,
      },
      columnPinning: {
        right: ['actions'],
      },
    },
  })

  const isEmpty = !loading && table.getRowModel().rows.length === 0
  const hasFilters = table.getState().columnFilters.length > 0
  const leftPinnedCount = table.getLeftLeafColumns().length

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-3">
      <div className="flex items-center gap-4">
        <DebouncedInput
          value={(table.getColumn('name')?.getFilterValue() as string) ?? ''}
          onChange={(value) => table.getColumn('name')?.setFilterValue(value)}
          placeholder="Search..."
          className="max-w-sm"
        />
      </div>
      <TableContainer
        className={isEmpty ? 'min-h-[26rem]' : undefined}
        empty={
          isEmpty ? (
            <TableEmptyState
              overlay
              colSpan={columns.length}
              message={hasFilters ? 'No matching registries found.' : 'No Container registries found.'}
              icon={<Package className="h-4 w-4" />}
              description={
                hasFilters
                  ? undefined
                  : 'Connect to external container registries (e.g., Docker Hub, GCR, ECR) to pull images for your Sandboxes.'
              }
              action={
                hasFilters ? (
                  <Button variant="outline" onClick={() => table.resetColumnFilters()}>
                    Clear filters
                  </Button>
                ) : undefined
              }
            />
          ) : undefined
        }
      >
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header, headerIndex) => {
                  return (
                    <TableHead
                      key={header.id}
                      className={cn(
                        'px-2',
                        !isEmpty && getColumnPinningBorderClasses(header.column, leftPinnedCount, headerIndex),
                        !isEmpty && getColumnPinningClasses(header.column, true),
                      )}
                      style={
                        isEmpty
                          ? undefined
                          : {
                              ...getExplicitColumnSize(header),
                              ...getColumnPinningStyles(header.column, FIXED_COLUMN_IDS),
                            }
                      }
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
              <>
                {Array.from({ length: 25 }).map((_, i) => (
                  <TableRow key={i}>
                    {table.getVisibleLeafColumns().map((column, colIndex) => (
                      <TableCell
                        key={column.id}
                        className={cn(
                          'px-2',
                          getColumnPinningBorderClasses(column, leftPinnedCount, colIndex),
                          getColumnPinningClasses(column),
                        )}
                        style={getColumnPinningStyles(column, FIXED_COLUMN_IDS)}
                      >
                        <Skeleton className="h-4 w-10/12" />
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
              </>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id} data-state={row.getIsSelected() && 'selected'}>
                  {row.getVisibleCells().map((cell, cellIndex) => (
                    <TableCell
                      className={cn(
                        'px-2',
                        getColumnPinningBorderClasses(cell.column, leftPinnedCount, cellIndex),
                        getColumnPinningClasses(cell.column),
                      )}
                      key={cell.id}
                      style={getColumnPinningStyles(cell.column, FIXED_COLUMN_IDS)}
                    >
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : null}
          </TableBody>
        </Table>
      </TableContainer>
      <PageFooterPortal>
        <Pagination table={table} entityName="Registries" />
      </PageFooterPortal>
    </div>
  )
}

const getColumns = ({
  onDelete,
  onEdit,
  loading,
  writePermitted,
  deletePermitted,
}: {
  onDelete: (id: string) => void
  onEdit: (registry: DockerRegistry) => void
  loading: boolean
  writePermitted: boolean
  deletePermitted: boolean
}): ColumnDef<DockerRegistry>[] => {
  const columns: ColumnDef<DockerRegistry>[] = [
    {
      accessorKey: 'name',
      header: 'Name',
      filterFn: (row, _id, filterValue) => {
        const searchValue = String(filterValue).toLowerCase()
        const registry = row.original

        return (
          registry.name.toLowerCase().includes(searchValue) ||
          registry.url.toLowerCase().includes(searchValue) ||
          (registry.project?.toLowerCase().includes(searchValue) ?? false) ||
          registry.username.toLowerCase().includes(searchValue)
        )
      },
    },
    {
      accessorKey: 'url',
      header: 'URL',
    },
    {
      id: 'project',
      header: 'Project',
      cell: ({ row }) => {
        return row.original.project || '-'
      },
    },
    {
      accessorKey: 'username',
      header: 'Username',
    },
    {
      id: 'actions',
      header: () => null,
      size: 48,
      minSize: 48,
      maxSize: 48,
      cell: ({ row }) => {
        if (!writePermitted && !deletePermitted) {
          return null
        }

        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="h-8 w-8 p-0">
                <span className="sr-only">Open menu</span>
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>

            <DropdownMenuContent align="end">
              {writePermitted && (
                <DropdownMenuItem onClick={() => onEdit(row.original)} className="cursor-pointer" disabled={loading}>
                  Edit
                </DropdownMenuItem>
              )}
              {deletePermitted && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    className="cursor-pointer text-red-600 dark:text-red-400"
                    disabled={loading}
                    onClick={() => onDelete(row.original.id)}
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
