/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { Column, Table } from '@tanstack/react-table'
import type { CSSProperties } from 'react'

export const DEFAULT_TABLE_COLUMN_MAX_RESIZE_SIZE = 350
export const DEFAULT_TABLE_COLUMN_MIN_SIZE = 80
export const TABLE_COLUMN_MAX_RESIZE_SIZE_OFFSET = 100
export const DEFAULT_TABLE_COLUMN = {
  minSize: DEFAULT_TABLE_COLUMN_MIN_SIZE,
}

type ColumnSizingDefaults = {
  maxSize?: number
  minSize?: number
}

function getColumnMaxSize(maxSize: number | undefined, fallbackMaxSize: number) {
  return maxSize === undefined || maxSize === Number.MAX_SAFE_INTEGER ? fallbackMaxSize : maxSize
}

export function getTableColumnMaxResizeSize(maxSize: number) {
  return maxSize + TABLE_COLUMN_MAX_RESIZE_SIZE_OFFSET
}

export function getColumnSizeBounds<T>(column: Column<T>, defaultColumn?: ColumnSizingDefaults) {
  const maxSize = column.columnDef.maxSize ?? defaultColumn?.maxSize ?? Number.MAX_SAFE_INTEGER
  const configuredMinSize = column.columnDef.minSize ?? defaultColumn?.minSize ?? DEFAULT_TABLE_COLUMN_MIN_SIZE
  const minSize = Math.min(Math.max(configuredMinSize, DEFAULT_TABLE_COLUMN_MIN_SIZE), maxSize)

  return {
    maxSize,
    minSize,
  }
}

export function getColumnResizeSizeBounds<T>(column: Column<T>, defaultColumn?: ColumnSizingDefaults) {
  const maxSize = getColumnMaxSize(
    column.columnDef.maxSize ?? defaultColumn?.maxSize,
    DEFAULT_TABLE_COLUMN_MAX_RESIZE_SIZE,
  )
  const configuredMinSize = column.columnDef.minSize ?? defaultColumn?.minSize ?? DEFAULT_TABLE_COLUMN_MIN_SIZE
  const minSize = Math.min(Math.max(configuredMinSize, DEFAULT_TABLE_COLUMN_MIN_SIZE), maxSize)

  return {
    maxSize,
    minSize,
  }
}

export function getColumnSizeStyles<T>(column: Column<T>): CSSProperties {
  const pinned = column.getIsPinned()
  const { maxSize, minSize } = getColumnSizeBounds(column)
  const hasMaxSize = maxSize !== Number.MAX_SAFE_INTEGER

  return {
    width: column.getSize(),
    minWidth: minSize,
    maxWidth: hasMaxSize ? maxSize : undefined,
    left: pinned === 'left' ? column.getStart('left') : undefined,
    right: pinned === 'right' ? column.getAfter('right') : undefined,
  }
}

export function getTableSizeStyles<T>(table: Table<T>): CSSProperties {
  return {
    width: table.getTotalSize(),
    minWidth: '100%',
  }
}
