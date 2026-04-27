/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DeleteOrganizationRoleDialog } from '@/components/OrganizationRoles/DeleteOrganizationRoleDialog'
import { UpdateOrganizationRoleDialog } from '@/components/OrganizationRoles/UpdateOrganizationRoleDialog'
import { PageFooterPortal } from '@/components/PageLayout'
import { Pagination } from '@/components/Pagination'
import { SearchInput } from '@/components/SearchInput'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableEmptyState,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { DEFAULT_PAGE_SIZE } from '@/constants/Pagination'
import { cn } from '@/lib/utils'
import { getColumnSizeStyles } from '@/lib/utils/table'
import { OrganizationRole, OrganizationRolePermissionsEnum } from '@daytona/api-client'
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
import { MoreHorizontal, Shield } from 'lucide-react'
import { useState } from 'react'

type OrganizationRoleTableMeta = {
  onDelete: (roleId: string) => void
  onUpdate: (role: OrganizationRole) => void
}

declare module '@tanstack/react-table' {
  interface TableMeta<TData extends RowData> {
    organizationRole?: TData extends OrganizationRole ? OrganizationRoleTableMeta : never
  }
}

const getMeta = (table: ReactTable<OrganizationRole>) => {
  return table.options.meta?.organizationRole as OrganizationRoleTableMeta
}

interface DataTableProps {
  data: OrganizationRole[]
  loadingData: boolean
  onUpdateRole: (
    roleId: string,
    name: string,
    description: string,
    permissions: OrganizationRolePermissionsEnum[],
  ) => Promise<boolean>
  onDeleteRole: (roleId: string) => Promise<boolean>
  loadingRoleAction: Record<string, boolean>
}

export function OrganizationRoleTable({
  data,
  loadingData,
  onUpdateRole,
  onDeleteRole,
  loadingRoleAction,
}: DataTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const [globalFilter, setGlobalFilter] = useState('')
  const [roleToDelete, setRoleToDelete] = useState<string | null>(null)
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
  const [roleToUpdate, setRoleToUpdate] = useState<OrganizationRole | null>(null)
  const [isUpdateDialogOpen, setIsUpdateDialogOpen] = useState(false)

  const table = useReactTable({
    data,
    columns: organizationRoleColumns,
    meta: {
      organizationRole: {
        onUpdate: (role) => {
          setRoleToUpdate(role)
          setIsUpdateDialogOpen(true)
        },
        onDelete: (roleId: string) => {
          setRoleToDelete(roleId)
          setIsDeleteDialogOpen(true)
        },
      },
    },
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    onGlobalFilterChange: setGlobalFilter,
    globalFilterFn: (row, _columnId, filterValue) => {
      const role = row.original
      const searchValue = String(filterValue).toLowerCase()

      return (
        role.name.toLowerCase().includes(searchValue) ||
        role.description.toLowerCase().includes(searchValue) ||
        role.permissions.some((permission) => permission.toLowerCase().includes(searchValue))
      )
    },
    state: {
      globalFilter,
      sorting,
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

  const isEmpty = !loadingData && table.getRowModel().rows.length === 0
  const hasSearch = globalFilter.trim().length > 0

  const handleChangeFilter = (value: string) => {
    setGlobalFilter(value)
    table.setPageIndex(0)
  }

  const handleUpdateRole = async (
    name: string,
    description: string,
    permissions: OrganizationRolePermissionsEnum[],
  ) => {
    if (roleToUpdate) {
      const success = await onUpdateRole(roleToUpdate.id, name, description, permissions)
      if (success) {
        setRoleToUpdate(null)
        setIsUpdateDialogOpen(false)
      }
      return success
    }
    return false
  }

  const handleConfirmDeleteRole = async () => {
    if (roleToDelete) {
      const success = await onDeleteRole(roleToDelete)
      if (success) {
        setRoleToDelete(null)
        setIsDeleteDialogOpen(false)
      }
      return success
    }
    return false
  }

  return (
    <>
      <div className="flex min-h-0 flex-1 flex-col pt-2">
        <div className="mb-3">
          <SearchInput
            debounced
            value={globalFilter}
            onValueChange={handleChangeFilter}
            placeholder="Search by Name, Description, or Permission"
            containerClassName="max-w-sm"
          />
        </div>
        <TableContainer
          className={cn('max-h-[550px]', {
            'min-h-[26rem]': isEmpty,
          })}
          empty={
            isEmpty ? (
              <TableEmptyState
                overlay
                colSpan={organizationRoleColumns.length}
                message={hasSearch ? 'No matching Roles found.' : 'No Roles found.'}
                icon={<Shield />}
                description={hasSearch ? null : 'Create custom roles to manage permissions in your organization.'}
                action={
                  hasSearch ? (
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
                  {headerGroup.headers.map((header) => {
                    return (
                      <TableHead
                        key={header.id}
                        sticky={header.column.getIsPinned()}
                        style={getColumnSizeStyles(header.column)}
                      >
                        {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                      </TableHead>
                    )
                  })}
                </TableRow>
              ))}
            </TableHeader>
            <TableBody>
              {loadingData ? (
                <>
                  {Array.from({ length: DEFAULT_PAGE_SIZE }).map((_, i) => (
                    <TableRow key={i} className="h-14">
                      {table.getVisibleLeafColumns().map((column) => (
                        <TableCell key={column.id} sticky={column.getIsPinned()} style={getColumnSizeStyles(column)}>
                          <Skeleton className="h-4 w-3/4" />
                        </TableCell>
                      ))}
                    </TableRow>
                  ))}
                </>
              ) : table.getRowModel().rows?.length ? (
                table.getRowModel().rows.map((row) => (
                  <TableRow
                    key={row.id}
                    data-state={row.getIsSelected() && 'selected'}
                    className={loadingRoleAction[row.original.id] ? 'opacity-50 pointer-events-none' : ''}
                  >
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
          <Pagination table={table} entityName="Roles" />
        </PageFooterPortal>
      </div>

      {roleToUpdate && (
        <UpdateOrganizationRoleDialog
          open={isUpdateDialogOpen}
          onOpenChange={(open) => {
            setIsUpdateDialogOpen(open)
            if (!open) {
              setRoleToUpdate(null)
            }
          }}
          initialData={roleToUpdate}
          onUpdateRole={handleUpdateRole}
        />
      )}

      {roleToDelete && (
        <DeleteOrganizationRoleDialog
          open={isDeleteDialogOpen}
          onOpenChange={(open) => {
            setIsDeleteDialogOpen(open)
            if (!open) {
              setRoleToDelete(null)
            }
          }}
          onDeleteRole={handleConfirmDeleteRole}
          loading={loadingRoleAction[roleToDelete]}
        />
      )}
    </>
  )
}

const organizationRoleColumns: ColumnDef<OrganizationRole>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
    cell: ({ row }) => {
      return <div className="min-w-48">{row.original.name}</div>
    },
  },
  {
    accessorKey: 'description',
    header: 'Description',
    cell: ({ row }) => {
      return (
        <Tooltip>
          <TooltipTrigger>
            <div className="truncate max-w-md cursor-text">{row.original.description}</div>
          </TooltipTrigger>
          <TooltipContent>
            <p className="max-w-[300px]">{row.original.description}</p>
          </TooltipContent>
        </Tooltip>
      )
    },
  },
  {
    accessorKey: 'permissions',
    header: () => {
      return <div className="max-w-md px-3">Permissions</div>
    },
    cell: ({ row }) => {
      const permissions = row.original.permissions.join(', ')
      return (
        <Tooltip>
          <TooltipTrigger>
            <div className="truncate max-w-md px-3 cursor-text">{permissions || '-'}</div>
          </TooltipTrigger>
          {permissions && (
            <TooltipContent>
              <p className="max-w-[300px]">{permissions}</p>
            </TooltipContent>
          )}
        </Tooltip>
      )
    },
  },
  {
    id: 'actions',
    size: 48,
    minSize: 48,
    maxSize: 48,
    cell: ({ row, table }) => {
      const { onDelete, onUpdate } = getMeta(table)

      if (row.original.isGlobal) {
        return null
      }

      return (
        <div className="text-right">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon-sm" aria-label="Open menu">
                <MoreHorizontal />
              </Button>
            </DropdownMenuTrigger>

            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => onUpdate(row.original)}>Edit</DropdownMenuItem>
              <DropdownMenuItem variant="destructive" onClick={() => onDelete(row.original.id)}>
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      )
    },
  },
]
