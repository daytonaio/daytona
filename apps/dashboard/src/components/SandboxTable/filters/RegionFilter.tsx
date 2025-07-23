/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover'
import { Command, CommandList, CommandGroup, CommandItem, CommandInput, CommandEmpty } from '@/components/ui/command'
import { cn } from '@/lib/utils'
import { Check } from 'lucide-react'
import { X } from 'lucide-react'
import { FacetedFilterOption } from '../types'

interface RegionFilterProps {
  value: string[]
  onFilterChange: (value: string[] | undefined) => void
  options?: FacetedFilterOption[]
}

export function RegionFilterIndicator({ value, onFilterChange, options }: RegionFilterProps) {
  const selectedRegionLabels = value
    .map((v) => options?.find((r) => r.value === v)?.label)
    .filter(Boolean)
    .join(', ')

  return (
    <div className="flex items-center h-6 gap-0.5 rounded-sm border border-border bg-muted/80 hover:bg-muted/50 text-sm">
      <Popover>
        <PopoverTrigger className="max-w-[160px] overflow-hidden text-ellipsis whitespace-nowrap text-muted-foreground px-2">
          Region:{' '}
          <span className="text-primary font-medium">
            {selectedRegionLabels.length > 0 ? selectedRegionLabels : 'All'}
          </span>
        </PopoverTrigger>

        <PopoverContent className="p-0 w-72" align="start">
          <RegionFilter value={value} onFilterChange={onFilterChange} options={options} />
        </PopoverContent>
      </Popover>

      <button className="h-6 w-5 p-0 border-0 hover:text-muted-foreground" onClick={() => onFilterChange(undefined)}>
        <X className="h-3 w-3" />
      </button>
    </div>
  )
}

export function RegionFilter({ value, onFilterChange, options }: RegionFilterProps) {
  return (
    <Command>
      <div className="flex items-center gap-2 justify-between p-2">
        <CommandInput placeholder="Search..." className="border border-border rounded-md h-8" />
        <button
          className="text-sm text-muted-foreground hover:text-primary px-2"
          onClick={() => onFilterChange(undefined)}
        >
          Clear
        </button>
      </div>
      <CommandList>
        <CommandEmpty>No regions found.</CommandEmpty>
        <CommandGroup>
          {options?.map((region) => (
            <CommandItem
              key={region.value}
              onSelect={() => {
                const newValue = value.includes(region.value)
                  ? value.filter((v) => v !== region.value)
                  : [...value, region.value]
                onFilterChange(newValue.length > 0 ? newValue : undefined)
              }}
            >
              <div className="flex items-center">
                <div
                  className={cn(
                    'mr-2 flex h-4 w-4 items-center justify-center rounded-sm border border-primary',
                    value.includes(region.value)
                      ? 'bg-primary text-primary-foreground'
                      : 'opacity-50 [&_svg]:invisible',
                  )}
                >
                  <Check className={cn('h-4 w-4')} />
                </div>
                {region.label}
              </div>
            </CommandItem>
          ))}
        </CommandGroup>
      </CommandList>
    </Command>
  )
}
