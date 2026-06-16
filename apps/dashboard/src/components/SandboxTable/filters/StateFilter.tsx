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
import { SandboxState } from '@daytona/api-client'
import { Square } from 'lucide-react'
import { STATUSES, getStateLabel } from '../constants'

interface StateFilterProps {
  value: string[]
  onFilterChange: (value: string[] | undefined) => void
}

function StateFilterLabel({ colorClassName, label }: { colorClassName: string; label: string }) {
  return (
    <span className="inline-flex min-w-0 items-center gap-2">
      <span className={cn('size-2 shrink-0 rounded-full', colorClassName)} aria-hidden="true" />
      <span className="truncate">{label}</span>
    </span>
  )
}

function getStateFilterColorClass(state: SandboxState) {
  switch (state) {
    case SandboxState.STARTED:
      return 'bg-success-foreground'
    case SandboxState.ERROR:
    case SandboxState.BUILD_FAILED:
      return 'bg-destructive'
    case SandboxState.STARTING:
    case SandboxState.STOPPING:
    case SandboxState.DESTROYING:
    case SandboxState.ARCHIVING:
      return 'bg-warning-foreground'
    case SandboxState.STOPPED:
    case SandboxState.ARCHIVED:
    default:
      return 'bg-muted-foreground'
  }
}

export function StateFilterIndicator({ value, onFilterChange }: StateFilterProps) {
  const selectedStates = value.map((v) => ({
    value: v,
    label: (
      <StateFilterLabel
        colorClassName={getStateFilterColorClass(v as SandboxState)}
        label={getStateLabel(v as SandboxState)}
      />
    ),
  }))

  return (
    <FacetedFilterRoot title="State" hasValue={value.length > 0} onClear={() => onFilterChange(undefined)}>
      <FacetedFilterAnchor>
        <FacetedFilterLabelTrigger icon={<Square />} aria-label="Filter by State">
          State
        </FacetedFilterLabelTrigger>
        <FacetedFilterOperator />
        <FacetedFilterValueTrigger
          className={cn({
            'px-1': value.length <= 2,
            'px-2': value.length > 2,
          })}
          aria-label="Edit State filter"
        >
          <FacetedFilterValues title="State" items={selectedStates} maxValues={2} />
        </FacetedFilterValueTrigger>
        <FacetedFilterClear aria-label="Clear State filter" />
      </FacetedFilterAnchor>
      <FacetedFilterContent className="p-0 w-72">
        <StateFilter value={value} onFilterChange={onFilterChange} />
      </FacetedFilterContent>
    </FacetedFilterRoot>
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
              <StateFilterLabel
                colorClassName={getStateFilterColorClass(status.value as SandboxState)}
                label={status.label}
              />
            </CommandCheckboxItem>
          ))}
        </CommandGroup>
      </CommandList>
    </Command>
  )
}
