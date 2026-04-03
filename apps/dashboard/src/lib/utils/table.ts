/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { Column } from '@tanstack/react-table'
import { CSSProperties } from 'react'

export function getColumnPinningStyles<T>(column: Column<T>, fixedColumnIds?: string[]): CSSProperties {
  const isPinned = column.getIsPinned()
  const isFixed = fixedColumnIds?.includes(column.id)
  const width = isFixed ? column.getSize() : undefined

  if (!isPinned) {
    return {
      width,
      minWidth: width,
      maxWidth: width,
    }
  }

  return {
    width,
    minWidth: width,
    maxWidth: width,
    left: isPinned === 'left' ? `${column.getStart('left')}px` : undefined,
    right: isPinned === 'right' ? `${column.getAfter('right')}px` : undefined,
  }
}

export function getColumnPinningClasses<T>(column: Column<T>, isHeader = false): string {
  const isPinned = column.getIsPinned()
  if (!isPinned) return ''
  return cn('md:sticky', isHeader ? 'md:z-[2]' : 'md:z-[1]')
}

export function getColumnPinningBorderClasses<T>(
  column: Column<T>,
  leftPinnedCount: number,
  columnIndex: number,
): string {
  const isPinned = column.getIsPinned()

  if (isPinned === 'left') {
    return cn(column.getIsLastColumn('left') && 'md:border-r')
  }
  if (isPinned === 'right') {
    return cn(column.getIsFirstColumn('right') && 'md:border-l')
  }
  return ''
}

export function getExplicitColumnSize(col: Column<any> | { column: Column<any> }): CSSProperties {
  const column = 'column' in col ? col.column : col
  if (column.columnDef.size === undefined) return {}
  const size = column.getSize()
  return { width: size, minWidth: size }
}
