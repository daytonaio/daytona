/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RefreshCwIcon, SearchIcon, XIcon } from 'lucide-react'
import { useEffect, useImperativeHandle, useRef, useState } from 'react'

import TooltipButton from '@/components/TooltipButton'
import { InputGroup, InputGroupAddon, InputGroupInput, InputGroupText } from '@/components/ui/input-group'
import { cn } from '@/lib/utils'
import { useFileSystemStore } from './fileSystemStore'

export type FileSearchHeaderHandle = {
  clear: () => void
}

export function FileSearchHeader({
  isRefreshing,
  onRefresh,
  ref,
}: {
  isRefreshing: boolean
  onRefresh: () => void | Promise<void>
  ref?: React.Ref<FileSearchHeaderHandle>
}) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [inputValue, setInputValue] = useState('')
  const isOpen = useFileSystemStore((state) => state.isSearchOpen)
  const { resetSearch, setSearchOpen, setSearchQuery } = useFileSystemStore((state) => state.actions)

  useImperativeHandle(
    ref,
    () => ({
      clear: () => {
        setInputValue('')
        setSearchQuery('')
      },
    }),
    [setSearchQuery],
  )

  useEffect(() => {
    if (!isOpen) {
      setInputValue('')
      return
    }

    const frame = window.requestAnimationFrame(() => {
      inputRef.current?.focus()
    })

    return () => window.cancelAnimationFrame(frame)
  }, [isOpen])

  useEffect(() => {
    const timeout = window.setTimeout(() => {
      setSearchQuery(inputValue.trim())
    }, 200)

    return () => window.clearTimeout(timeout)
  }, [inputValue, setSearchQuery])

  return (
    <div className="relative h-11 shrink-0 overflow-hidden border-b border-border">
      <div
        className={cn('absolute inset-0 flex items-center gap-2 px-3 transition-all duration-200', {
          'pointer-events-auto z-10 translate-x-0 opacity-100': !isOpen,
          'pointer-events-none z-0 -translate-x-6 opacity-0': isOpen,
        })}
      >
        <span className="text-sm font-medium">Files</span>
        <TooltipButton
          tooltipText="Refresh files"
          variant="ghost"
          size="icon-sm"
          onClick={onRefresh}
          className="ml-auto"
          disabled={isRefreshing}
        >
          <RefreshCwIcon
            className={cn('size-4', {
              'animate-spin': isRefreshing,
            })}
          />
        </TooltipButton>
        <TooltipButton tooltipText="Search files" variant="ghost" size="icon-sm" onClick={() => setSearchOpen(true)}>
          <SearchIcon className="size-4" />
        </TooltipButton>
      </div>

      <div
        className={cn('absolute inset-0 flex items-center gap-2 px-2 transition-all duration-200', {
          'pointer-events-auto z-10 translate-x-0 opacity-100': isOpen,
          'pointer-events-none z-0 translate-x-6 opacity-0': !isOpen,
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
                  className="shrink-0 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none"
                  onClick={() => {
                    setInputValue('')
                    setSearchQuery('')
                  }}
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
          onClick={() => {
            setSearchOpen(false)
            resetSearch()
          }}
          className="shrink-0"
        >
          <XIcon className="size-4" />
        </TooltipButton>
      </div>
    </div>
  )
}
