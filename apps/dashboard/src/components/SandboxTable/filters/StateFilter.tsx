/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Command,
  CommandCheckboxItem,
  CommandGroup,
  CommandInput,
  CommandInputButton,
  CommandList,
} from '@/components/ui/command'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { SandboxState } from '@daytonaio/api-client'
import { X } from 'lucide-react'
import { STATUSES, getStateLabel } from '../constants'

interface StateFilterProps {
  value: string[]
  onFilterChange: (value: string[] | undefined) => void
}

export function StateFilterIndicator({ value, onFilterChange }: StateFilterProps) {
  const selectedStates = value.map((v) => getStateLabel(v as SandboxState))
  return (
    <div className="flex items-center h-6 gap-0.5 rounded-sm border border-border bg-muted/80 hover:bg-muted/50 text-sm">
      <Popover>
        <PopoverTrigger className="max-w-[240px] overflow-hidden text-ellipsis whitespace-nowrap text-muted-foreground px-2">
          States:{' '}
          <span className="text-primary font-medium">
            {selectedStates.length > 0
              ? selectedStates.length > 2
                ? `${selectedStates[0]}, ${selectedStates[1]}, +${selectedStates.length - 2}`
                : `${selectedStates.join(', ')}`
              : ''}
          </span>
        </PopoverTrigger>

        <PopoverContent className="p-0 w-72" align="start">
          <StateFilter value={value} onFilterChange={onFilterChange} />
        </PopoverContent>
      </Popover>

      <button className="h-6 w-5 p-0 border-0 hover:text-muted-foreground" onClick={() => onFilterChange(undefined)}>
        <X className="h-3 w-3" />
      </button>
    </div>
  )
}

export function StateFilter({ value, onFilterChange }: StateFilterProps) {
  return (
    <Command>
      <CommandInput placeholder="Search...">
        <CommandInputButton
          className="text-sm text-muted-foreground hover:text-primary px-2"
          onClick={() => onFilterChange(undefined)}
        >
          Clear
        </CommandInputButton>
      </CommandInput>

      <CommandList>
        <CommandGroup>
          {STATUSES.map((status) => (
            <CommandCheckboxItem
              key={status.value}
              checked={value.includes(status.value)}
              onSelect={() => {
                const newValue = value.includes(status.value)
                  ? value.filter((v) => v !== status.value)
                  : [...value, status.value]
                onFilterChange(newValue.length > 0 ? newValue : undefined)
              }}
            >
              {status.label}
            </CommandCheckboxItem>
          ))}
        </CommandGroup>
      </CommandList>
    </Command>
  )
}
