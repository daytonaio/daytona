/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { ChevronDown, ChevronRight } from 'lucide-react'
import { FC, Fragment, ReactNode, useState } from 'react'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table'

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
    <div
      className={cn(
        'w-full rounded-lg border border-border bg-card overflow-x-auto scrollbar-thin scrollbar-thumb-border scrollbar-track-background',
        className,
      )}
    >
      <Table>
        <TableHeader>
          <TableRow className="border-b border-border">
            <TableHead className="py-2 px-4 text-muted-foreground border-r border-border sticky left-0 bg-background z-10">
              {headerLabel}
            </TableHead>
            {columns.map((column, index) => (
              <TableHead
                key={index}
                className={cn(
                  'py-2 px-4 text-xs text-nowrap',
                  index === currentColumn ? 'text-foreground bg-muted/50' : 'text-muted-foreground',
                )}
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
    </div>
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
          className={cn('cursor-pointer border-b border-border group select-none hover:bg-muted')}
          aria-expanded={isOpen}
        >
          <TableCell colSpan={10} className="py-2 px-4 bg-muted">
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
            className={cn('border-b border-border group hover:bg-muted transition-none', {
              'bg-muted': currentRow === rowIdx,
            })}
          >
            <TableCell
              className={cn(
                'py-2 px-4 text-muted-foreground text-xs border-r border-border align-top w-40 sticky left-0 bg-background z-10 group-hover:bg-muted',
                {
                  'bg-muted': currentRow === rowIdx,
                },
              )}
            >
              {row.label}
            </TableCell>

            {row.values.map((val, colIdx) => {
              const isActive = colIdx === currentColumn

              return (
                <TableCell
                  key={colIdx}
                  className={cn(
                    'py-2 px-4 text-xs align-top text-right tabular-nums',
                    isActive ? 'bg-muted text-foreground' : 'text-muted-foreground',
                  )}
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
