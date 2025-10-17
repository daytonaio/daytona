/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Check } from 'lucide-react'
import { Command, CommandList, CommandGroup, CommandItem } from './ui/command'
import { cn } from '@/lib/utils'
import type { Column } from '@tanstack/react-table'

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
                <CommandItem key={column.id} onSelect={() => column.toggleVisibility(!column.getIsVisible())}>
                  <div className="flex items-center">
                    <div
                      className={cn(
                        'mr-2 flex h-4 w-4 items-center justify-center rounded-sm border border-primary',
                        column.getIsVisible() ? 'bg-primary text-primary-foreground' : 'opacity-50 [&_svg]:invisible',
                      )}
                    >
                      <Check className={cn('h-4 w-4')} />
                    </div>
                    {getColumnLabel(column.id)}
                  </div>
                </CommandItem>
              )
            })}
        </CommandGroup>
      </CommandList>
    </Command>
  )
}
