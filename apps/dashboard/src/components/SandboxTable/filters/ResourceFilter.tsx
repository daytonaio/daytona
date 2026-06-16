/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

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
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useMemo, type ReactNode } from 'react'

export interface ResourceFilterValue {
  cpu?: { min?: number; max?: number }
  memory?: { min?: number; max?: number }
  disk?: { min?: number; max?: number }
}

interface ResourceFilterProps {
  value: ResourceFilterValue
  onFilterChange: (value: ResourceFilterValue | undefined) => void
  resourceType?: 'cpu' | 'memory' | 'disk'
}

interface ResourceFilterIndicatorProps extends ResourceFilterProps {
  icon?: ReactNode
}

const RESOURCE_CONFIG = {
  cpu: { label: 'vCPU', displayLabel: 'CPU' },
  memory: { label: 'Memory (GiB)', displayLabel: 'Memory' },
  disk: { label: 'Disk (GiB)', displayLabel: 'Disk' },
} as const

export function ResourceFilterIndicator({ value, onFilterChange, resourceType, icon }: ResourceFilterIndicatorProps) {
  const { title, label, hasValue } = useMemo(() => {
    let title = 'All'
    let label = 'Resources'
    let hasValue = false

    if (resourceType) {
      const resourceValue = value[resourceType]
      hasValue = Boolean(resourceValue)

      if (resourceValue?.min !== undefined || resourceValue?.max !== undefined) {
        const config = RESOURCE_CONFIG[resourceType]
        const unit = resourceType === 'cpu' ? 'vCPU' : 'GiB'
        title = `${resourceValue.min ?? 'Any'} - ${resourceValue.max ?? 'Any'} ${unit}`
        label = config.displayLabel
      }
    } else {
      const filters: string[] = []
      Object.entries(RESOURCE_CONFIG).forEach(([type, config]) => {
        const resourceValue = value[type as keyof ResourceFilterValue]
        if (resourceValue?.min !== undefined || resourceValue?.max !== undefined) {
          const unit = type === 'cpu' ? 'vCPU' : 'GiB'
          filters.push(`${config.displayLabel}: ${resourceValue.min ?? 'any'}-${resourceValue.max ?? 'any'} ${unit}`)
        }
      })
      hasValue = filters.length > 0
      title = filters.length > 0 ? filters.join('; ') : 'All'
    }

    return { title, label, hasValue }
  }, [value, resourceType])

  const handleClear = () => {
    if (resourceType) {
      const newFilterValue = { ...value }
      delete newFilterValue[resourceType]
      onFilterChange(Object.keys(newFilterValue).length > 0 ? newFilterValue : undefined)
      return
    }

    onFilterChange(undefined)
  }

  return (
    <FacetedFilterRoot title={label} hasValue={hasValue} onClear={handleClear}>
      <FacetedFilterAnchor>
        <FacetedFilterLabelTrigger icon={icon} aria-label={`Filter by ${label}`}>
          {label}
        </FacetedFilterLabelTrigger>
        <FacetedFilterOperator />
        <FacetedFilterValueTrigger className="px-1" aria-label={`Edit ${label} filter`}>
          <FacetedFilterValues
            title={label}
            items={[{ value: resourceType ?? 'resources', label: title }]}
            maxValues={1}
          />
        </FacetedFilterValueTrigger>
        <FacetedFilterClear aria-label={`Clear ${label} filter`} />
      </FacetedFilterAnchor>
      <FacetedFilterContent className="w-72 p-4">
        <ResourceFilter value={value} onFilterChange={onFilterChange} resourceType={resourceType} />
      </FacetedFilterContent>
    </FacetedFilterRoot>
  )
}

export function ResourceFilter({ value, onFilterChange, resourceType }: ResourceFilterProps) {
  const handleValueChange = (
    resource: keyof ResourceFilterValue,
    field: 'min' | 'max',
    newValue: number | undefined,
  ) => {
    const currentResourceValue = value[resource] || {}
    const updatedResourceValue = { ...currentResourceValue, [field]: newValue }

    if (newValue === undefined && !currentResourceValue.min && !currentResourceValue.max) {
      const newFilterValue = { ...value }
      delete newFilterValue[resource]
      onFilterChange(Object.keys(newFilterValue).length > 0 ? newFilterValue : undefined)
      return
    }

    const newFilterValue = { ...value, [resource]: updatedResourceValue }
    onFilterChange(newFilterValue)
  }

  const handleClear = (resource: keyof ResourceFilterValue) => {
    const newFilterValue = { ...value }
    delete newFilterValue[resource]
    onFilterChange(Object.keys(newFilterValue).length > 0 ? newFilterValue : undefined)
  }

  if (resourceType) {
    const config = RESOURCE_CONFIG[resourceType]
    const currentValues = value[resourceType] || {}

    return (
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between gap-2">
          <Label>{config.label}</Label>
          <button
            className="text-sm text-muted-foreground hover:text-primary"
            onClick={() => handleClear(resourceType)}
          >
            Clear
          </button>
        </div>
        <div className="flex items-center gap-2">
          <Input
            type="number"
            placeholder="Min"
            min={0}
            value={currentValues.min ?? ''}
            onChange={(e) => {
              const newValue = e.target.value ? Number(e.target.value) : undefined
              handleValueChange(resourceType, 'min', newValue)
            }}
            className="w-full"
          />
          <div className="w-8 h-[1px] bg-border"></div>
          <Input
            type="number"
            placeholder="Max"
            min={0}
            value={currentValues.max ?? ''}
            onChange={(e) => {
              const newValue = e.target.value ? Number(e.target.value) : undefined
              handleValueChange(resourceType, 'max', newValue)
            }}
            className="w-full"
          />
        </div>
      </div>
    )
  }

  return null
}
