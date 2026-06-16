/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { ReactNode } from 'react'

import { cn } from '@/lib/utils'
import { FacetedFilter, type FacetedFilterOperator, type FacetedFilterOption } from './faceted-filter'

interface FacetFilterProps {
  title: string
  options: readonly FacetedFilterOption[]
  values: ReadonlySet<string>
  onValuesChange: (values: Set<string>) => void
  operator?: string
  operators?: readonly FacetedFilterOperator[]
  onOperatorChange?: (operator: string) => void
  maxValues?: number
  className?: string
  contentClassName?: string
  icon?: ReactNode
}

export function FacetFilter({
  title,
  options,
  values,
  onValuesChange,
  operator,
  operators,
  onOperatorChange,
  maxValues,
  className,
  contentClassName,
  icon,
}: FacetFilterProps) {
  return (
    <FacetedFilter
      title={title}
      options={options}
      values={values}
      onValuesChange={onValuesChange}
      operator={operator}
      operators={operators}
      onOperatorChange={onOperatorChange}
      maxValues={maxValues}
      className={cn('h-10', className)}
      contentClassName={contentClassName}
      icon={icon}
    />
  )
}
