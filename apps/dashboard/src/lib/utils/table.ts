/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column } from '@tanstack/react-table'
import { CSSProperties } from 'react'

export function getColumnSizeStyles<T>(column: Column<T>): CSSProperties {
  const pinned = column.getIsPinned()
  const hasMaxSize = column.columnDef.maxSize !== Number.MAX_SAFE_INTEGER

  return {
    width: hasMaxSize ? column.getSize() : undefined,
    minWidth: column.columnDef.minSize,
    left: pinned === 'left' ? column.getStart('left') : undefined,
    right: pinned === 'right' ? column.getAfter('right') : undefined,
  }
}
