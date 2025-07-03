/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover'
import { Command, CommandList, CommandGroup, CommandItem, CommandInput } from '@/components/ui/command'
import { cn } from '@/lib/utils'
import { Check } from 'lucide-react'
import { X } from 'lucide-react'
import { FacetedFilterOption } from '@/components/ui/data-table-faceted-filter'

interface LabelFilterProps {
  value: string[]
  onFilterChange: (value: string[] | undefined) => void
  options: FacetedFilterOption[]
}

export function LabelFilterIndicator({
  value,
  onFilterChange,
  options,
}: Pick<LabelFilterProps, 'value' | 'onFilterChange' | 'options'>) {
  return (
    <div className="flex items-center h-6 gap-0.5 rounded-sm border border-border bg-muted/80 hover:bg-muted/50 text-sm">
      <Popover>
        <PopoverTrigger className="max-w-[160px] overflow-hidden text-ellipsis whitespace-nowrap text-muted-foreground px-2">
          Labels: <span className="text-primary font-medium">{value.length} selected</span>
        </PopoverTrigger>

        <PopoverContent className="p-0 w-[240px]" align="start">
          <LabelFilter value={value} onFilterChange={onFilterChange} options={options} />
        </PopoverContent>
      </Popover>

      <button className="h-6 w-5 p-0 border-0 hover:text-muted-foreground" onClick={() => onFilterChange(undefined)}>
        <X className="h-3 w-3" />
      </button>
    </div>
  )
}

export function LabelFilter({ value, onFilterChange, options }: LabelFilterProps) {
  return (
    <Command>
      <div className="flex items-center gap-2 justify-between p-2">
        <CommandInput placeholder="Filter by label..." className="border border-border rounded-md h-8" />
        <button
          className="text-sm text-muted-foreground hover:text-primary px-2"
          onClick={() => onFilterChange(undefined)}
        >
          Clear
        </button>
      </div>
      <CommandList>
        <CommandGroup>
          {options.map((option) => (
            <CommandItem
              key={option.value}
              onSelect={() => {
                const newValue = value.includes(option.value)
                  ? value.filter((v) => v !== option.value)
                  : [...value, option.value]
                onFilterChange(newValue.length > 0 ? newValue : undefined)
              }}
            >
              <div className="flex items-center">
                <div
                  className={cn(
                    'mr-2 flex h-4 w-4 items-center justify-center rounded-sm border border-primary',
                    value.includes(option.value)
                      ? 'bg-primary text-primary-foreground'
                      : 'opacity-50 [&_svg]:invisible',
                  )}
                >
                  <Check className={cn('h-4 w-4')} />
                </div>
                <div className="truncate max-w-md rounded-sm bg-blue-100 dark:bg-blue-950 text-blue-800 dark:text-blue-200 px-1">
                  {option.label.split(':')[0]}
                </div>

                <span className="ml-2 text-muted-foreground">{option.label.split(':')[1]}</span>
              </div>
            </CommandItem>
          ))}
        </CommandGroup>
      </CommandList>
    </Command>
  )
}
