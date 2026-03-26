/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { Column } from '@tanstack/react-table'
import { Command, CommandCheckboxItem, CommandGroup, CommandList } from './ui/command'

interface TableColumnVisibilityToggleProps {
  columns: Column<any, unknown>[]
  getColumnLabel: (id: string) => string
}

export function TableColumnVisibilityToggle({ columns, getColumnLabel }: TableColumnVisibilityToggleProps) {
  return (
    <Command>
      <CommandList>
        <CommandGroup>
          {columns
            .filter((column) => column.getCanHide())
            .map((column) => {
              return (
                <CommandCheckboxItem
                  key={column.id}
                  checked={column.getIsVisible()}
                  onSelect={() => column.toggleVisibility()}
                >
                  {getColumnLabel(column.id)}
                </CommandCheckboxItem>
              )
            })}
        </CommandGroup>
      </CommandList>
    </Command>
  )
}
