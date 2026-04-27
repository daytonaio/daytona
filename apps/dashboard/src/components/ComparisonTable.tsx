/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { ChevronDown, ChevronRight } from 'lucide-react'
import { FC, Fragment, ReactNode, useState } from 'react'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow } from './ui/table'

interface ComparisonRow {
  label: ReactNode
  values: ReactNode[]
}

export interface ComparisonSection {
  id?: string
  title: ReactNode
  collapsed?: boolean
  collapsible?: boolean
  rows: ComparisonRow[]
}

interface Props {
  columns: string[]
  headerLabel?: string
  currentColumn?: number
  currentRow?: number
  data: ComparisonSection[]
  className?: string
}

export function ComparisonTable({ columns = [], headerLabel, currentColumn, currentRow, data = [], className }: Props) {
  return (
    <TableContainer className={cn('rounded-none', className)}>
      <Table>
        <TableHeader>
          <TableRow className="rounded-none">
            <TableHead sticky="left" className="py-2 px-4 text-muted-foreground">
              {headerLabel}
            </TableHead>
            {columns.map((column, index) => (
              <TableHead
                key={index}
                className={cn('py-2 px-4 text-xs text-nowrap', {
                  'text-foreground': index === currentColumn,
                })}
              >
                {column}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody className="[&_tr]:h-auto">
          {data.map((section, idx) => (
            <CollapsibleSection
              key={section.id || idx}
              section={section}
              currentColumn={currentColumn}
              currentRow={currentRow}
            />
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  )
}

interface CollapsibleSectionProps {
  section: ComparisonSection
  currentColumn?: number
  currentRow?: number
}

const CollapsibleSection: FC<CollapsibleSectionProps> = ({ section, currentColumn, currentRow }) => {
  const [isOpen, setIsOpen] = useState(section.collapsible ? !section.collapsed : true)

  return (
    <Fragment>
      {section.collapsible && (
        <TableRow
          onClick={() => setIsOpen(!isOpen)}
          className={cn('cursor-pointer group select-none')}
          aria-expanded={isOpen}
        >
          <TableCell colSpan={10} className="py-2 px-4">
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              {isOpen ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
              {section.title}
            </div>
          </TableCell>
        </TableRow>
      )}

      {isOpen &&
        section.rows.map((row, rowIdx) => (
          <TableRow
            key={rowIdx}
            className={cn('border-b border-border group transition-none', {
              'text-foreground': rowIdx === currentRow,
            })}
          >
            <TableCell
              sticky="left"
              className={cn('py-2 px-4 text-xs align-top w-40 text-muted-foreground md:border-r', {
                'text-foreground': rowIdx === currentRow || currentColumn === 0,
              })}
            >
              {row.label}
              {rowIdx === currentRow && (
                <span className="inline-flex ml-2 text-muted-foreground text-xs">(Current)</span>
              )}
            </TableCell>

            {row.values.map((val, colIdx) => {
              const isActive = rowIdx === currentRow || colIdx === currentColumn

              return (
                <TableCell
                  key={colIdx}
                  className={cn('py-2 px-4 text-xs align-top text-right tabular-nums text-muted-foreground', {
                    'text-foreground': isActive,
                  })}
                >
                  {val}
                </TableCell>
              )
            })}
          </TableRow>
        ))}
    </Fragment>
  )
}
