/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useEffect, useMemo } from 'react'
import { Search, X } from 'lucide-react'
import { DebouncedInput } from './DebouncedInput'
import { pluralize } from '@/lib/utils'

interface SearchInputProps {
  placeholder?: string
  value: string
  onChange: (value: string) => void
  className?: string
  showKeyboardShortcut?: boolean
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
  showKeyboardShortcut = true,
  resultCount,
  entityName = 'item',
  entityNamePlural,
  'data-testid': dataTestId,
}: SearchInputProps) {
  const platformKey = useMemo(() => {
    if (typeof navigator === 'undefined' || !navigator.userAgent) return 'Ctrl+K'
    return /Mac|iPhone|iPod|iPad/i.test(navigator.userAgent) ? '⌘K' : 'Ctrl+K'
  }, [])
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        const searchInputs = document.querySelectorAll<HTMLInputElement>('[data-search-input]')
        const searchInput = searchInputs[searchInputs.length - 1]
        if (searchInput) {
          searchInput.focus()
        }
      }
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
          className={`pl-10 ${value ? 'pr-10' : showKeyboardShortcut ? 'pr-16' : 'pr-3'}`}
          data-search-input
          data-testid={dataTestId}
        />
        {value ? (
          <button
            type="button"
            onClick={() => onChange('')}
            className="absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
            aria-label="Clear search"
          >
            <X className="h-4 w-4" />
          </button>
        ) : showKeyboardShortcut ? (
          <div className="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-muted-foreground pointer-events-none">
            <span className="hidden sm:inline">{platformKey}</span>
          </div>
        ) : null}
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
