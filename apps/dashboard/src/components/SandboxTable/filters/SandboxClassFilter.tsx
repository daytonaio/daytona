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
import { SandboxClass } from '@daytona/api-client'
import { Boxes } from 'lucide-react'
import { SANDBOX_CLASS_OPTIONS, getSandboxClassLabel } from '../constants'

interface SandboxClassFilterProps {
  value: string[]
  onFilterChange: (value: string[] | undefined) => void
}

export function SandboxClassFilterIndicator({ value, onFilterChange }: SandboxClassFilterProps) {
  const selectedClasses = value.map((v) => ({
    value: v,
    label: getSandboxClassLabel(v as SandboxClass),
  }))

  return (
    <FacetedFilterRoot title="Class" hasValue={value.length > 0} onClear={() => onFilterChange(undefined)}>
      <FacetedFilterAnchor>
        <FacetedFilterLabelTrigger icon={<Boxes />} aria-label="Filter by Class">
          Class
        </FacetedFilterLabelTrigger>
        <FacetedFilterOperator />
        <FacetedFilterValueTrigger
          className={cn({
            'px-1': value.length <= 2,
            'px-2': value.length > 2,
          })}
          aria-label="Edit Class filter"
        >
          <FacetedFilterValues title="Class" items={selectedClasses} maxValues={2} />
        </FacetedFilterValueTrigger>
        <FacetedFilterClear aria-label="Clear Class filter" />
      </FacetedFilterAnchor>
      <FacetedFilterContent className="p-0 w-64">
        <SandboxClassFilter value={value} onFilterChange={onFilterChange} />
      </FacetedFilterContent>
    </FacetedFilterRoot>
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
