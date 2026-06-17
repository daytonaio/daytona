/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { getColumnSizeStyles, getTableSizeStyles } from '@/lib/utils/table'
import { flexRender } from '@tanstack/react-table'
import { Receipt } from 'lucide-react'
import { Pagination } from '../Pagination'
import { Skeleton } from '../ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableEmptyState,
  TableHead,
  TableHeader,
  TableRow,
} from '../ui/table'
import { ChargesTableHeader } from './ChargesTableHeader'
import { ChargesTableProps } from './types'
import { useChargesTable } from './useChargesTable'

export function ChargesTable({ data, loading, onRowClick }: ChargesTableProps) {
  const { table } = useChargesTable({ data })
  const isEmpty = !loading && table.getRowModel().rows.length === 0
  const hasSearch = String(table.getState().globalFilter ?? '').trim().length > 0

  return (
    <div className="flex flex-col gap-3">
      <ChargesTableHeader table={table} />

      <TableContainer
        className={cn('max-h-[550px]', {
          'min-h-[20rem]': isEmpty,
        })}
        empty={
          isEmpty ? (
            <TableEmptyState
              overlay
              colSpan={table.getAllColumns().length}
              message={hasSearch ? 'No matching charges found.' : 'No charges yet.'}
              icon={<Receipt />}
              description={
                hasSearch ? (
                  <p>Try adjusting your search query.</p>
                ) : (
                  <p>Charges will appear here as payments are attempted on your organization.</p>
                )
              }
            />
          ) : null
        }
      >
        <Table className="table-fixed border-separate border-spacing-0" style={getTableSizeStyles(table)}>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead
                    key={header.id}
                    header={header}
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
                {Array.from({ length: table.getState().pagination.pageSize }).map((_, rowIndex) => (
                  <TableRow key={rowIndex}>
                    {table.getVisibleLeafColumns().map((column) => (
                      <TableCell
                        key={`${rowIndex}-${column.id}`}
                        sticky={column.getIsPinned()}
                        style={getColumnSizeStyles(column)}
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
                  className={cn('transition-colors duration-300', {
                    'cursor-pointer': onRowClick,
                  })}
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

      <div className="flex items-center justify-end">
        <Pagination className="pb-2" table={table} entityName="Charges" />
      </div>
    </div>
  )
}
