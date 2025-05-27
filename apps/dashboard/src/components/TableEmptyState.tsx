/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { TableRow, TableCell } from './ui/table'

interface TableEmptyStateProps {
  /**
   * The number of columns in the table (used for colSpan)
   */
  colSpan: number
  /**
   * The message to display when no data is found
   */
  message: string
  /**
   * Optional icon to display above the message
   */
  icon?: React.ReactNode
  /**
   * Optional description text to display below the main message
   */
  description?: string
  /**
   * Additional CSS classes for the container
   */
  className?: string
}

export function TableEmptyState({ colSpan, message, icon, description, className = '' }: TableEmptyStateProps) {
  return (
    <TableRow>
      <TableCell colSpan={colSpan} className={`h-24 text-center ${className}`}>
        <div className="flex flex-col items-center justify-center space-y-2">
          {icon && <div className="text-muted-foreground">{icon}</div>}
          <p className="text-muted-foreground">{message}</p>
          {description && <p className="text-sm text-muted-foreground/80">{description}</p>}
        </div>
      </TableCell>
    </TableRow>
  )
}
