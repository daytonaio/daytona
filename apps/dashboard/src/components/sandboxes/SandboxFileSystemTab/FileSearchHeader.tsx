/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SearchIcon, XIcon } from 'lucide-react'
import { Ref, useCallback, useEffect, useImperativeHandle, useRef, useState, type ReactNode } from 'react'

import TooltipButton from '@/components/TooltipButton'
import { InputGroup, InputGroupAddon, InputGroupInput, InputGroupText } from '@/components/ui/input-group'
import { cn } from '@/lib/utils'

export type FileSearchHeaderHandle = {
  clear: () => void
}

export function FileSearchHeader({
  actions,
  onSearchQueryChange,
  ref,
}: {
  actions?: ReactNode
  onSearchQueryChange: (value: string) => void
  ref?: Ref<FileSearchHeaderHandle>
}) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [inputValue, setInputValue] = useState('')
  const [isOpen, setIsOpen] = useState(false)

  const clearSearch = useCallback(() => {
    setInputValue('')
    onSearchQueryChange('')
  }, [onSearchQueryChange])

  const closeSearch = useCallback(() => {
    setIsOpen(false)
    clearSearch()
  }, [clearSearch])

  useImperativeHandle(
    ref,
    () => ({
      clear: clearSearch,
    }),
    [clearSearch],
  )

  useEffect(() => {
    if (!isOpen) {
      return
    }

    const frame = window.requestAnimationFrame(() => {
      inputRef.current?.focus()
    })

    return () => window.cancelAnimationFrame(frame)
  }, [isOpen])

  useEffect(() => {
    const timeout = window.setTimeout(() => {
      onSearchQueryChange(inputValue.trim())
    }, 200)

    return () => window.clearTimeout(timeout)
  }, [inputValue, onSearchQueryChange])

  return (
    <div className="relative h-11 shrink-0 overflow-hidden border-b border-border">
      <div
        aria-hidden={isOpen}
        inert={isOpen}
        className={cn('absolute inset-0 flex items-center px-2 transition-all duration-200', {
          'z-10 translate-x-0 opacity-100': !isOpen,
          'z-0 -translate-x-6 opacity-0': isOpen,
        })}
      >
        <span className="mr-2 text-sm font-medium">Files</span>
        <div className="ml-auto flex items-center">{actions}</div>
        <TooltipButton tooltipText="Search files" variant="ghost" size="icon-sm" onClick={() => setIsOpen(true)}>
          <SearchIcon className="size-4" />
        </TooltipButton>
      </div>

      <div
        aria-hidden={!isOpen}
        inert={!isOpen}
        className={cn('absolute inset-0 flex items-center gap-2 px-2 transition-all duration-200', {
          'z-10 translate-x-0 opacity-100': isOpen,
          'z-0 translate-x-6 opacity-0': !isOpen,
        })}
      >
        <InputGroup className="h-8 min-w-0 flex-1 overflow-hidden border-0 bg-transparent shadow-none has-[[data-slot=input-group-control]:focus-visible]:border-transparent has-[[data-slot=input-group-control]:focus-visible]:ring-0">
          <InputGroupAddon align="inline-start" className="pl-2 pr-0">
            <InputGroupText>
              <SearchIcon className="size-4" />
            </InputGroupText>
          </InputGroupAddon>
          <InputGroupInput
            ref={inputRef}
            value={inputValue}
            onChange={(event) => setInputValue(event.target.value)}
            placeholder="Search files..."
            className="h-8 px-2"
          />
          {inputValue ? (
            <InputGroupAddon align="inline-end" className="pr-3">
              <span className="shrink-0">
                <button
                  type="button"
                  className="shrink-0 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground disabled:pointer-events-none disabled:opacity-50 focus-visible:outline-none"
                  onClick={clearSearch}
                >
                  Clear
                </button>
              </span>
            </InputGroupAddon>
          ) : null}
        </InputGroup>
        <TooltipButton
          tooltipText="Close search"
          variant="ghost"
          size="icon-sm"
          onClick={closeSearch}
          className="shrink-0"
        >
          <XIcon className="size-4" />
        </TooltipButton>
      </div>
    </div>
  )
}
