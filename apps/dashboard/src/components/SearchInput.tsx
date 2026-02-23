/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useEffect } from 'react'
import { Search, X } from 'lucide-react'
import { DebouncedInput } from './DebouncedInput'
import { pluralize } from '@/lib/utils'

interface SearchInputProps {
  placeholder?: string
  value: string
  onChange: (value: string) => void
  className?: string
  resultCount?: number
  entityName?: string
  entityNamePlural?: string
  'data-testid'?: string
}

export function SearchInput({
  placeholder = 'Search...',
  value,
  onChange,
  className = '',
  resultCount,
  entityName = 'item',
  entityNamePlural,
  'data-testid': dataTestId,
}: SearchInputProps) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && document.activeElement?.hasAttribute('data-search-input')) {
        const hasOpenDialog = document.querySelector('[role="dialog"]')
        if (!hasOpenDialog) {
          onChange('')
        }
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [onChange])

  return (
    <div className={`mb-4 ${className}`}>
      <div className="relative w-full max-w-md">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <DebouncedInput
          placeholder={placeholder}
          value={value}
          onChange={(inputValue) => onChange(String(inputValue))}
          className={`pl-10 ${value ? 'pr-10' : 'pr-3'}`}
          data-search-input
          data-testid={dataTestId}
        />
        {value && (
          <button
            type="button"
            onClick={() => onChange('')}
            className="absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
            aria-label="Clear search"
          >
            <X className="h-4 w-4" />
          </button>
        )}
      </div>
      {value && resultCount !== undefined && (
        <div className="mt-2 text-sm text-muted-foreground">
          {resultCount === 0
            ? `No ${entityNamePlural ?? `${entityName}s`} match your search.`
            : `Found ${pluralize(resultCount, entityName, entityNamePlural ?? `${entityName}s`)}`}
        </div>
      )}
    </div>
  )
}
