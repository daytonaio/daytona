/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from './ui/empty'
import { TableCell, TableRow } from './ui/table'

interface TableEmptyStateProps {
  colSpan: number
  message: string
  icon?: React.ReactNode
  description?: React.ReactNode
  className?: string
}

export function TableEmptyState({ colSpan, message, icon, description, className = '' }: TableEmptyStateProps) {
  return (
    <TableRow>
      <TableCell colSpan={colSpan} className={cn(`h-24 text-center ${className}`)}>
        <Empty className="border-none py-8">
          <EmptyHeader>
            {icon && <EmptyMedia variant="icon">{icon}</EmptyMedia>}
            <EmptyTitle>{message}</EmptyTitle>
            {description && <EmptyDescription>{description}</EmptyDescription>}
          </EmptyHeader>
        </Empty>
      </TableCell>
    </TableRow>
  )
}
