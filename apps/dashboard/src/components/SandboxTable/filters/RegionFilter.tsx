/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandInputButton,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { cn } from '@/lib/utils'
import { Check, Loader2, X } from 'lucide-react'
import { FacetedFilterOption } from '../types'

interface RegionFilterProps {
  value: string[]
  onFilterChange: (value: string[] | undefined) => void
  options?: FacetedFilterOption[]
  isLoading?: boolean
}

export function RegionFilterIndicator({ value, onFilterChange, options, isLoading }: RegionFilterProps) {
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
          <RegionFilter value={value} onFilterChange={onFilterChange} options={options} isLoading={isLoading} />
        </PopoverContent>
      </Popover>

      <button className="h-6 w-5 p-0 border-0 hover:text-muted-foreground" onClick={() => onFilterChange(undefined)}>
        <X className="h-3 w-3" />
      </button>
    </div>
  )
}

export function RegionFilter({ value, onFilterChange, options, isLoading }: RegionFilterProps) {
  return (
    <Command>
      <CommandInput placeholder="Search..." className="">
        <CommandInputButton onClick={() => onFilterChange(undefined)}>Clear</CommandInputButton>
      </CommandInput>
      <CommandList>
        {isLoading ? (
          <div className="flex items-center justify-center py-6">
            <Loader2 className="h-4 w-4 animate-spin mr-2" />
            <span className="text-sm text-muted-foreground">Loading regions...</span>
          </div>
        ) : (
          <>
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
          </>
        )}
      </CommandList>
    </Command>
  )
}
