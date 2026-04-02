/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import {
  getColumnPinningBorderClasses,
  getColumnPinningClasses,
  getColumnPinningStyles,
  getExplicitColumnSize,
} from '@/lib/utils/table'
import { Column, flexRender } from '@tanstack/react-table'
import { FileText } from 'lucide-react'
import { CSSProperties } from 'react'
import { Pagination } from '../Pagination'
import { TableEmptyState } from '../TableEmptyState'
import { Skeleton } from '../ui/skeleton'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow } from '../ui/table'
import { InvoicesTableHeader } from './InvoicesTableHeader'
import { InvoicesTableProps } from './types'
import { useInvoicesTable } from './useInvoicesTable'

function getColumnStyles(column: Column<any>): CSSProperties {
  if (column.id !== 'actions') {
    return {
      width: column.id === 'number' ? '20%' : 'auto',
      maxWidth: column.getSize() + 80,
      minWidth: column.getSize(),
      ...getColumnPinningStyles(column),
    }
  }

  return {
    width: column.getSize(),
    minWidth: column.getSize(),
    maxWidth: column.getSize(),
    ...getColumnPinningStyles(column, ['actions']),
  }
}

export function InvoicesTable({
  data,
  pagination,
  totalItems,
  pageCount,
  onPaginationChange,
  loading,
  onViewInvoice,
  onVoidInvoice,
  onRowClick,
  onPayInvoice,
}: InvoicesTableProps) {
  const { table } = useInvoicesTable({
    data,
    pagination,
    pageCount,
    onPaginationChange,
    onViewInvoice,
    onVoidInvoice,
    onPayInvoice,
  })

  const isEmpty = !loading && table.getRowModel().rows.length === 0

  return (
    <>
      <InvoicesTableHeader table={table} />

      <TableContainer
        className={isEmpty ? 'min-h-[20rem]' : undefined}
        empty={
          isEmpty ? (
            <TableEmptyState
              overlay
              colSpan={table.getAllColumns().length}
              message="No invoices yet."
              icon={<FileText className="w-8 h-8" />}
              description={
                <div className="space-y-2">
                  <p>Invoices will appear here once they are generated.</p>
                </div>
              }
            />
          ) : undefined
        }
      >
        <Table className="border-separate border-spacing-0">
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead
                      key={header.id}
                      onClick={() =>
                        header.column.getCanSort() && header.column.toggleSorting(header.column.getIsSorted() === 'asc')
                      }
                      className={cn(
                        'border-b border-border',
                        getColumnPinningBorderClasses(header.column, 0, 0),
                        getColumnPinningClasses(header.column, true),
                        {
                          '!px-2': header.column.id === 'actions',
                          'hover:bg-muted cursor-pointer': header.column.getCanSort(),
                        },
                      )}
                      style={getColumnStyles(header.column)}
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
                {Array.from({ length: 10 }).map((_, rowIndex) => (
                  <TableRow key={rowIndex}>
                    {table.getVisibleLeafColumns().map((column) => (
                      <TableCell
                        key={`${rowIndex}-${column.id}`}
                        className={cn(getColumnPinningClasses(column), {
                          '!px-2': column.id === 'actions',
                        })}
                        style={getColumnStyles(column)}
                      >
                        <Skeleton className="h-4 w-10/12" />
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
              </>
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
                      className={cn(
                        'border-b border-border',
                        getColumnPinningBorderClasses(cell.column, 0, 0),
                        getColumnPinningClasses(cell.column),
                        {
                          '!px-2': cell.column.id === 'actions',
                        },
                      )}
                      style={getColumnStyles(cell.column)}
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

      <div className="flex items-center justify-end">
        <Pagination className="pb-2 pt-6" table={table} entityName="Invoices" totalItems={totalItems} />
      </div>
    </>
  )
}
