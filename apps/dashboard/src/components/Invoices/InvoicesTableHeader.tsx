/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Search } from 'lucide-react'
import React from 'react'
import { DebouncedInput } from '../DebouncedInput'
import { InvoicesTableHeaderProps } from './types'

export function InvoicesTableHeader({ table }: InvoicesTableHeaderProps) {
  const [globalFilter, setGlobalFilter] = React.useState('')

  React.useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        // Focus search input
        const searchInput = document.querySelector('[data-search-input]') as HTMLInputElement
        if (searchInput) {
          searchInput.focus()
        }
      }
    }

    document.addEventListener('keydown', down)
    return () => document.removeEventListener('keydown', down)
  }, [])

  return (
    <div className="flex items-center justify-between pb-4">
      <div className="flex flex-1 items-center space-x-2">
        <div className="relative w-full max-w-sm">
          <Search className="absolute left-2 top-2 h-4 w-4 text-muted-foreground" />
          <DebouncedInput
            placeholder="Search invoices..."
            value={globalFilter ?? ''}
            onChange={(value) => {
              setGlobalFilter(String(value))
              table.setGlobalFilter(String(value))
            }}
            className="pl-8"
            data-search-input
          />
        </div>
      </div>
    </div>
  )
}
