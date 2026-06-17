/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { Column, Table } from '@tanstack/react-table'
import type { CSSProperties } from 'react'

export function getColumnSizeStyles<T>(column: Column<T>): CSSProperties {
  const pinned = column.getIsPinned()
  const maxSize = column.columnDef.maxSize
  const hasMaxSize = maxSize !== undefined && maxSize !== Number.MAX_SAFE_INTEGER

  return {
    width: column.getSize(),
    minWidth: column.columnDef.minSize,
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
