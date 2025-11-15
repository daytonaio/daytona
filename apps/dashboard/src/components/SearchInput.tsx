/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useEffect } from 'react'
import { Search, X } from 'lucide-react'
import { DebouncedInput } from './DebouncedInput'

interface SearchInputProps {
  placeholder?: string
  value: string
  onChange: (value: string) => void
  className?: string
  showKeyboardShortcut?: boolean
  resultCount?: number
  entityName?: string
  'data-testid'?: string
}

export function SearchInput({
  placeholder = 'Search...',
  value,
  onChange,
  className = '',
  showKeyboardShortcut = true,
  resultCount,
  entityName = 'items',
  'data-testid': dataTestId,
}: SearchInputProps) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        const searchInput = document.querySelector('[data-search-input]') as HTMLInputElement
        if (searchInput) {
          searchInput.focus()
        }
      }
      if (e.key === 'Escape' && document.activeElement?.getAttribute('data-search-input') !== null) {
        onChange('')
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
            onClick={() => onChange('')}
            className="absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
            aria-label="Clear search"
          >
            <X className="h-4 w-4" />
          </button>
        ) : showKeyboardShortcut ? (
          <div className="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-muted-foreground pointer-events-none">
            <span className="hidden sm:inline">{navigator.platform.includes('Mac') ? '⌘K' : 'Ctrl+K'}</span>
          </div>
        ) : null}
      </div>
      {value && resultCount !== undefined && (
        <div className="mt-2 text-sm text-muted-foreground">
          {resultCount === 0
            ? `No ${entityName} match your search.`
            : `Found ${resultCount} ${entityName}${resultCount === 1 ? '' : 's'}`}
        </div>
      )}
    </div>
  )
}
