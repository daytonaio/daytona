/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Command, CommandList, CommandGroup, CommandCheckboxItem } from '@/components/ui/command'
import {
  FacetedFilterAnchor,
  FacetedFilterClear,
  FacetedFilterContent,
  FacetedFilterLabelTrigger,
  FacetedFilterRoot,
  FacetedFilterValueTrigger,
  FacetedFilterValues,
} from '@/components/ui/faceted-filter'
import { Label } from '@/components/ui/label'
import type { ReactNode } from 'react'

type BooleanValueLabels = Record<'true' | 'false', ReactNode>

interface BooleanFilterProps {
  label: string
  value: boolean | undefined
  onFilterChange: (value: boolean | undefined) => void
  valueLabels?: BooleanValueLabels
}

interface BooleanFilterIndicatorProps extends BooleanFilterProps {
  icon?: ReactNode
}

const OPTIONS = [
  { value: true, labelKey: 'true' },
  { value: false, labelKey: 'false' },
] as const

const DEFAULT_VALUE_LABELS = {
  true: 'Yes',
  false: 'No',
} satisfies BooleanValueLabels

function getBooleanLabel(value: boolean, valueLabels: BooleanValueLabels) {
  return valueLabels[String(value) as keyof BooleanValueLabels]
}

export function BooleanFilterIndicator({
  label,
  value,
  onFilterChange,
  valueLabels = DEFAULT_VALUE_LABELS,
  icon,
}: BooleanFilterIndicatorProps) {
  const selectedValue =
    value === undefined
      ? []
      : [
          {
            value: String(value),
            label: getBooleanLabel(value, valueLabels),
          },
        ]

  return (
    <FacetedFilterRoot title={label} hasValue={value !== undefined} onClear={() => onFilterChange(undefined)}>
      <FacetedFilterAnchor>
        <FacetedFilterLabelTrigger icon={icon} aria-label={`Filter by ${label}`}>
          {label}
        </FacetedFilterLabelTrigger>
        <FacetedFilterValueTrigger className="px-1" aria-label={`Edit ${label} filter`}>
          <FacetedFilterValues title={label} items={selectedValue} maxValues={1} />
        </FacetedFilterValueTrigger>
        <FacetedFilterClear aria-label={`Clear ${label} filter`} />
      </FacetedFilterAnchor>
      <FacetedFilterContent className="p-2 w-48">
        <BooleanFilter label={label} onFilterChange={onFilterChange} value={value} valueLabels={valueLabels} />
      </FacetedFilterContent>
    </FacetedFilterRoot>
  )
}

export function BooleanFilter({
  label,
  onFilterChange,
  value,
  valueLabels = DEFAULT_VALUE_LABELS,
}: BooleanFilterProps) {
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
              {valueLabels[option.labelKey]}
            </CommandCheckboxItem>
          ))}
        </CommandGroup>
      </CommandList>
    </Command>
  )
}
