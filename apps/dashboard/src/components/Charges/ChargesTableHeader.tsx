/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
import { SearchInput } from '../SearchInput'
import { ChargesTableHeaderProps } from './types'

export function ChargesTableHeader({ table }: ChargesTableHeaderProps) {
  const [globalFilter, setGlobalFilter] = React.useState('')

  return (
    <div className="flex items-center">
      <SearchInput
        debounced
        placeholder="Search charges..."
        value={globalFilter ?? ''}
        onValueChange={(value) => {
          setGlobalFilter(value)
          table.setGlobalFilter(value)
        }}
        containerClassName="w-full max-w-sm"
        data-search-input
      />
    </div>
  )
}
