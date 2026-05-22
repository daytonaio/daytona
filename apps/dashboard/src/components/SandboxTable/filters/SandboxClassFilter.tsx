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
import { SandboxClass } from '@daytona/api-client'
import { X } from 'lucide-react'
import { SANDBOX_CLASS_OPTIONS, getSandboxClassLabel } from '../constants'

interface SandboxClassFilterProps {
  value: string[]
  onFilterChange: (value: string[] | undefined) => void
}

export function SandboxClassFilterIndicator({ value, onFilterChange }: SandboxClassFilterProps) {
  const selectedLabels = value.map((v) => getSandboxClassLabel(v as SandboxClass))
  return (
    <div className="flex items-center h-6 gap-0.5 rounded-sm border border-border bg-muted/80 hover:bg-muted/50 text-sm">
      <Popover>
        <PopoverTrigger className="max-w-[240px] overflow-hidden text-ellipsis whitespace-nowrap text-muted-foreground px-2">
          Class:{' '}
          <span className="text-primary font-medium">
            {selectedLabels.length > 0
              ? selectedLabels.length > 2
                ? `${selectedLabels[0]}, ${selectedLabels[1]}, +${selectedLabels.length - 2}`
                : `${selectedLabels.join(', ')}`
              : ''}
          </span>
        </PopoverTrigger>

        <PopoverContent className="p-0 w-64" align="start">
          <SandboxClassFilter value={value} onFilterChange={onFilterChange} />
        </PopoverContent>
      </Popover>

      <button className="h-6 w-5 p-0 border-0 hover:text-muted-foreground" onClick={() => onFilterChange(undefined)}>
        <X className="h-3 w-3" />
      </button>
    </div>
  )
}

export function SandboxClassFilter({ value, onFilterChange }: SandboxClassFilterProps) {
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
          {SANDBOX_CLASS_OPTIONS.map((option) => {
            const Icon = option.icon
            return (
              <CommandCheckboxItem
                key={option.value}
                checked={value.includes(option.value)}
                onSelect={() => {
                  const newValue = value.includes(option.value)
                    ? value.filter((v) => v !== option.value)
                    : [...value, option.value]
                  onFilterChange(newValue.length > 0 ? newValue : undefined)
                }}
              >
                <span className="flex items-center gap-2">
                  {Icon ? <Icon className="size-4 text-muted-foreground" /> : null}
                  {option.label}
                </span>
              </CommandCheckboxItem>
            )
          })}
        </CommandGroup>
      </CommandList>
    </Command>
  )
}
