/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { Column } from '@tanstack/react-table'
import type { ReactNode } from 'react'

import { FacetedFilter, type FacetedFilterOperator, type FacetedFilterOption } from './faceted-filter'

interface DataTableFacetedFilterProps<TData, TValue> {
  column?: Column<TData, TValue>
  title?: string
  options: readonly FacetedFilterOption[]
  operator?: string
  operators?: readonly FacetedFilterOperator[]
  onOperatorChange?: (operator: string) => void
  maxValues?: number
  className?: string
  contentClassName?: string
  icon?: ReactNode
}

export type { FacetedFilterOperator, FacetedFilterOption } from './faceted-filter'

export function DataTableFacetedFilter<TData, TValue>({
  column,
  title,
  options,
  operator,
  operators,
  onOperatorChange,
  maxValues,
  className,
  contentClassName,
  icon,
}: DataTableFacetedFilterProps<TData, TValue>) {
  const facets = column?.getFacetedUniqueValues()
  const values = new Set((column?.getFilterValue() as string[] | undefined) ?? [])

  const handleValuesChange = (nextValues: Set<string>) => {
    const filterValues = Array.from(nextValues)
    column?.setFilterValue(filterValues.length ? filterValues : undefined)
  }

  return (
    <FacetedFilter
      title={title ?? 'Filter'}
      options={options}
      values={values}
      onValuesChange={handleValuesChange}
      operator={operator}
      operators={operators}
      onOperatorChange={onOperatorChange}
      facets={facets}
      maxValues={maxValues}
      className={className}
      contentClassName={contentClassName}
      icon={icon}
    />
  )
}
