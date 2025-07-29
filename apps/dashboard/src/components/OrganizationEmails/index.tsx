/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { flexRender } from '@tanstack/react-table'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../ui/table'
import { Pagination } from '../Pagination'
import { TableEmptyState } from '../TableEmptyState'
import { OrganizationEmailsTableProps } from './types'
import { useOrganizationEmailsTable } from './useOrganizationEmailsTable'
import { OrganizationEmailsTableHeader } from './OrganizationEmailsTableHeader'
import { cn } from '@/lib/utils'
import { Mail } from 'lucide-react'

export function OrganizationEmailsTable({
  data,
  loading,
  handleDelete,
  handleResendVerification,
  handleAddEmail,
  onRowClick,
}: OrganizationEmailsTableProps) {
  // Create a loading state object for individual emails
  const loadingEmails: Record<string, boolean> = {}

  const { table } = useOrganizationEmailsTable({
    data,
    loadingEmails,
    handleDelete,
    handleResendVerification,
  })

  return (
    <>
      <OrganizationEmailsTableHeader table={table} onAddEmail={handleAddEmail} />

      <Table className="border-separate border-spacing-0">
        <TableHeader>
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id}>
              {headerGroup.headers.map((header) => {
                return (
                  <TableHead
                    key={header.id}
                    data-state={header.column.getCanSort() && 'sortable'}
                    onClick={() =>
                      header.column.getCanSort() && header.column.toggleSorting(header.column.getIsSorted() === 'asc')
                    }
                    className={cn(
                      'sticky top-0 z-[3] border-b border-border',
                      header.column.getCanSort() ? 'hover:bg-muted cursor-pointer' : '',
                    )}
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
            <TableRow>
              <TableCell colSpan={table.getAllColumns().length} className="h-10 text-center">
                Loading...
              </TableCell>
            </TableRow>
          ) : table.getRowModel().rows?.length ? (
            table.getRowModel().rows.map((row) => (
              <TableRow
                key={row.id}
                className={`transition-colors duration-300 ${onRowClick ? 'cursor-pointer' : ''}`}
                onClick={() => onRowClick?.(row.original)}
              >
                {row.getVisibleCells().map((cell) => (
                  <TableCell
                    key={cell.id}
                    onClick={(e) => {
                      if (cell.column.id === 'actions') {
                        e.stopPropagation()
                      }
                    }}
                    className="border-b border-border"
                    style={{
                      width: cell.column.id === 'email' ? '40%' : 'auto',
                      maxWidth: cell.column.getSize() + 80,
                      minWidth: cell.column.getSize(),
                    }}
                    sticky={cell.column.id === 'actions' ? 'right' : undefined}
                  >
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            ))
          ) : (
            <TableEmptyState
              colSpan={table.getAllColumns().length}
              message="No billing emails yet."
              icon={<Mail className="w-8 h-8" />}
              description={
                <div className="space-y-2">
                  <p>
                    Add billing emails which recieve important billing notifications such as invoices and credit
                    depletion notices.
                  </p>
                </div>
              }
            />
          )}
        </TableBody>
      </Table>

      <div className="flex items-center justify-end">
        <Pagination className="pb-2 pt-6" table={table} entityName="Emails" />
      </div>
    </>
  )
}
