/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { getColumnSizeStyles } from '@/lib/utils/table'
import { DockerRegistry, OrganizationRolePermissionsEnum } from '@daytona/api-client'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  Table as ReactTable,
  RowData,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { MoreHorizontal, PackageOpen } from 'lucide-react'
import { useMemo, useState } from 'react'
import { PageFooterPortal } from './PageLayout'
import { Pagination } from './Pagination'
import { SearchInput } from './SearchInput'
import { Button } from './ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from './ui/dropdown-menu'
import { Skeleton } from './ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableEmptyState,
  TableHead,
  TableHeader,
  TableRow,
} from './ui/table'

type RegistryTableMeta = {
  onDelete: (id: string) => void
  onEdit: (registry: DockerRegistry) => void
  loading: boolean
  writePermitted: boolean
  deletePermitted: boolean
}

declare module '@tanstack/react-table' {
  interface TableMeta<TData extends RowData> {
    registry?: TData extends DockerRegistry ? RegistryTableMeta : never
  }
}

const getMeta = (table: ReactTable<DockerRegistry>) => {
  return table.options.meta?.registry as RegistryTableMeta
}

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
  const [globalFilter, setGlobalFilter] = useState('')
  const table = useReactTable({
    data,
    columns,
    meta: {
      registry: { onDelete, onEdit, loading, writePermitted, deletePermitted },
    },
    defaultColumn: {
      minSize: 0,
    },
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    onGlobalFilterChange: setGlobalFilter,
    globalFilterFn: (row, _columnId, filterValue) => {
      const searchValue = String(filterValue).toLowerCase()
      const registry = row.original

      return (
        registry.name.toLowerCase().includes(searchValue) ||
        registry.url.toLowerCase().includes(searchValue) ||
        (registry.project?.toLowerCase().includes(searchValue) ?? false) ||
        registry.username.toLowerCase().includes(searchValue)
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
      columnPinning: {
        right: ['actions'],
      },
    },
  })

  const isEmpty = !loading && table.getRowModel().rows.length === 0
  const hasFilters = globalFilter.trim().length > 0

  const handleChangeFilter = (value: string) => {
    setGlobalFilter(value)
    table.setPageIndex(0)
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-3">
      <div className="flex items-center gap-2">
        <SearchInput
          debounced
          value={globalFilter}
          onValueChange={handleChangeFilter}
          placeholder="Search by Name, URL, Project, or Username"
          containerClassName="max-w-sm"
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
              icon={<PackageOpen />}
              description={
                hasFilters ? null : (
                  <p>
                    Connect to external container registries (e.g., Docker Hub, GCR, ECR) to pull images for your
                    Sandboxes.
                  </p>
                )
              }
              action={
                hasFilters ? (
                  <Button variant="outline" onClick={() => handleChangeFilter('')}>
                    Clear filters
                  </Button>
                ) : null
              }
            />
          ) : null
        }
      >
        <Table className="table-fixed" style={{ minWidth: table.getTotalSize() }}>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead
                    key={header.id}
                    sticky={header.column.getIsPinned()}
                    style={getColumnSizeStyles(header.column)}
                  >
                    {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {loading ? (
              <>
                {Array.from({ length: DEFAULT_PAGE_SIZE }).map((_, i) => (
                  <TableRow key={i}>
                    {table.getVisibleLeafColumns().map((column) => (
                      <TableCell key={column.id} sticky={column.getIsPinned()} style={getColumnSizeStyles(column)}>
                        <Skeleton className="h-4 w-10/12" />
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
              </>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id} data-state={row.getIsSelected() && 'selected'}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell
                      key={cell.id}
                      sticky={cell.column.getIsPinned()}
                      style={getColumnSizeStyles(cell.column)}
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

const columns: ColumnDef<DockerRegistry>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
    size: 200,
  },
  {
    accessorKey: 'url',
    header: 'URL',
    size: 300,
  },
  {
    id: 'project',
    header: 'Project',
    size: 150,
    cell: ({ row }) => {
      return row.original.project || '-'
    },
  },
  {
    accessorKey: 'username',
    header: 'Username',
    size: 150,
  },
  {
    id: 'actions',
    header: () => null,
    size: 48,
    minSize: 48,
    maxSize: 48,
    cell: ({ row, table }) => {
      const { writePermitted, deletePermitted, loading, onEdit, onDelete } = getMeta(table)

      if (!writePermitted && !deletePermitted) {
        return null
      }

      return (
        <div className="flex justify-end">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon-sm" aria-label="Open menu">
                <MoreHorizontal />
              </Button>
            </DropdownMenuTrigger>

            <DropdownMenuContent align="end">
              {writePermitted && (
                <DropdownMenuItem onClick={() => onEdit(row.original)} disabled={loading}>
                  Edit
                </DropdownMenuItem>
              )}
              {deletePermitted && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem variant="destructive" disabled={loading} onClick={() => onDelete(row.original.id)}>
                    Delete
                  </DropdownMenuItem>
                </>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      )
    },
  },
]
