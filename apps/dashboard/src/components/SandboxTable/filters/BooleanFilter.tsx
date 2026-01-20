/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover'
import { Command, CommandList, CommandGroup, CommandCheckboxItem } from '@/components/ui/command'
import { X } from 'lucide-react'
import { Label } from '@/components/ui/label'

interface BooleanFilterProps {
  label: string
  value: boolean | undefined
  onFilterChange: (value: boolean | undefined) => void
}

const OPTIONS = [
  { value: true, label: 'Yes' },
  { value: false, label: 'No' },
]

export function BooleanFilterIndicator({ label, value, onFilterChange }: BooleanFilterProps) {
  return (
    <div className="flex items-center h-6 gap-0.5 rounded-sm border border-border bg-muted/80 hover:bg-muted/50 text-sm">
      <Popover>
        <PopoverTrigger className="max-w-[180px] overflow-hidden text-ellipsis whitespace-nowrap text-muted-foreground px-2">
          {label}:{' '}
          <span className="text-primary font-medium">{value === true ? 'Yes' : value === false ? 'No' : 'Any'}</span>
        </PopoverTrigger>
        <PopoverContent className="p-0 w-48" align="start">
          <BooleanFilter label={label} onFilterChange={onFilterChange} value={value} />
        </PopoverContent>
      </Popover>

      <button className="h-6 w-5 p-0 border-0 hover:text-muted-foreground" onClick={() => onFilterChange(undefined)}>
        <X className="h-3 w-3" />
      </button>
    </div>
  )
}

export function BooleanFilter({ label, onFilterChange, value }: BooleanFilterProps) {
  return (
    <Command>
      <div className="flex items-center gap-2 justify-between mb-2">
        <Label>{label}</Label>
        <button
          className="text-sm text-muted-foreground hover:text-primary px-2"
          onClick={() => onFilterChange(undefined)}
        >
          Clear
        </button>
      </div>
      <CommandList>
        <CommandGroup className="p-0">
          {OPTIONS.map((option) => (
            <CommandCheckboxItem
              checked={value === option.value}
              key={String(option.value)}
              onSelect={() => {
                onFilterChange(value === option.value ? undefined : option.value)
              }}
            >
              {option.label}
            </CommandCheckboxItem>
          ))}
        </CommandGroup>
      </CommandList>
    </Command>
  )
}
