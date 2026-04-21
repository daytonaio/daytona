/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Search } from 'lucide-react'
import React from 'react'
import { DebouncedInput } from '../DebouncedInput'
import { ChargesTableHeaderProps } from './types'

export function ChargesTableHeader({ table }: ChargesTableHeaderProps) {
  const [globalFilter, setGlobalFilter] = React.useState('')

  return (
    <div className="flex items-center justify-between pb-4">
      <div className="flex flex-1 items-center space-x-2">
        <div className="relative w-full max-w-sm">
          <Search className="absolute left-2 top-2 h-4 w-4 text-muted-foreground" />
          <DebouncedInput
            placeholder="Search charges..."
            value={globalFilter ?? ''}
            onChange={(value) => {
              setGlobalFilter(String(value))
              table.setGlobalFilter(String(value))
            }}
            className="pl-8"
          />
        </div>
      </div>
    </div>
  )
}
