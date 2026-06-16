/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Command,
  CommandCheckboxItem,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandInputButton,
  CommandList,
} from '@/components/ui/command'
import {
  FacetedFilterAnchor,
  FacetedFilterClear,
  FacetedFilterContent,
  FacetedFilterLabelTrigger,
  FacetedFilterOperator,
  FacetedFilterRoot,
  FacetedFilterValueTrigger,
  FacetedFilterValues,
} from '@/components/ui/faceted-filter'
import { cn } from '@/lib/utils'
import { Globe, Loader2 } from 'lucide-react'
import { FacetedFilterOption } from '../types'

interface RegionFilterProps {
  value: string[]
  onFilterChange: (value: string[] | undefined) => void
  options?: FacetedFilterOption[]
  isLoading?: boolean
}

export function RegionFilterIndicator({ value, onFilterChange, options, isLoading }: RegionFilterProps) {
  const selectedRegions = value.map((v) => ({
    value: v,
    label: options?.find((region) => region.value === v)?.label ?? v,
  }))

  return (
    <FacetedFilterRoot title="Region" hasValue={value.length > 0} onClear={() => onFilterChange(undefined)}>
      <FacetedFilterAnchor>
        <FacetedFilterLabelTrigger icon={<Globe />} aria-label="Filter by Region">
          Region
        </FacetedFilterLabelTrigger>
        <FacetedFilterOperator />
        <FacetedFilterValueTrigger
          className={cn({
            'px-1': value.length <= 1,
            'px-2': value.length > 1,
          })}
          aria-label="Edit Region filter"
        >
          <FacetedFilterValues title="Region" items={selectedRegions} maxValues={1} />
        </FacetedFilterValueTrigger>
        <FacetedFilterClear aria-label="Clear Region filter" />
      </FacetedFilterAnchor>
      <FacetedFilterContent className="p-0 w-72">
        <RegionFilter value={value} onFilterChange={onFilterChange} options={options} isLoading={isLoading} />
      </FacetedFilterContent>
    </FacetedFilterRoot>
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
                <CommandCheckboxItem
                  checked={value.includes(region.value)}
                  key={region.value}
                  onSelect={() => {
                    const newValue = value.includes(region.value)
                      ? value.filter((v) => v !== region.value)
                      : [...value, region.value]
                    onFilterChange(newValue.length > 0 ? newValue : undefined)
                  }}
                >
                  {region.label}
                </CommandCheckboxItem>
              ))}
            </CommandGroup>
          </>
        )}
      </CommandList>
    </Command>
  )
}
