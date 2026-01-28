/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { flexRender } from '@tanstack/react-table'
import { FileText } from 'lucide-react'
import { Pagination } from '../Pagination'
import { TableEmptyState } from '../TableEmptyState'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../ui/table'
import { InvoicesTableHeader } from './InvoicesTableHeader'
import { InvoicesTableProps } from './types'
import { useInvoicesTable } from './useInvoicesTable'

export function InvoicesTable({
  data,
  pagination,
  pageCount,
  onPaginationChange,
  loading,
  onViewInvoice,
  onVoidInvoice,
  onRowClick,
}: InvoicesTableProps) {
  const { table } = useInvoicesTable({
    data,
    pagination,
    pageCount,
    onPaginationChange,
    onViewInvoice,
    onVoidInvoice,
  })

  return (
    <>
      <InvoicesTableHeader table={table} />

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
                      width: cell.column.id === 'number' ? '20%' : 'auto',
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
              message="No invoices yet."
              icon={<FileText className="w-8 h-8" />}
              description={
                <div className="space-y-2">
                  <p>Invoices will appear here once they are generated.</p>
                </div>
              }
            />
          )}
        </TableBody>
      </Table>

      <div className="flex items-center justify-end">
        <Pagination className="pb-2 pt-6" table={table} entityName="Invoices" />
      </div>
    </>
  )
}
