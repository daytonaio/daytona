/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

'use client'

import { SearchIcon, XIcon } from 'lucide-react'
import { ChangeEvent, ComponentProps, useEffect, useRef, useState } from 'react'

import { cn } from '@/lib/utils'

import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from './ui/input-group'

interface SearchInputProps extends Omit<ComponentProps<'input'>, 'defaultValue' | 'onChange' | 'type' | 'value'> {
  containerClassName?: string
  clearButtonAriaLabel?: string
  debounced?: boolean
  debounceMs?: number
  onClear?: () => void
  onValueChange: (value: string) => void
  value: string
}

function SearchInput({
  className,
  containerClassName,
  clearButtonAriaLabel = 'Clear search',
  debounced = false,
  debounceMs = 500,
  onClear,
  onValueChange,
  value,
  ...props
}: SearchInputProps) {
  const inputRef = useRef<HTMLInputElement>(null)
  const hasMountedRef = useRef(false)
  const onValueChangeRef = useRef(onValueChange)
  const skipDebounceRef = useRef(false)
  const [internalValue, setInternalValue] = useState(value)

  const currentValue = debounced ? internalValue : value
  const hasValue = currentValue !== ''

  useEffect(() => {
    setInternalValue(value)
  }, [value])

  useEffect(() => {
    onValueChangeRef.current = onValueChange
  }, [onValueChange])

  useEffect(() => {
    if (!debounced) {
      return
    }

    if (!hasMountedRef.current) {
      hasMountedRef.current = true
      return
    }

    if (skipDebounceRef.current) {
      skipDebounceRef.current = false
      return
    }

    const timeout = window.setTimeout(() => {
      onValueChangeRef.current(internalValue)
    }, debounceMs)

    return () => window.clearTimeout(timeout)
  }, [debounceMs, debounced, internalValue])

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    const nextValue = event.target.value

    if (debounced) {
      setInternalValue(nextValue)
    } else {
      onValueChange(nextValue)
    }
  }

  const handleClear = () => {
    const input = inputRef.current
    if (props.disabled || props.readOnly) {
      return
    }

    skipDebounceRef.current = true
    setInternalValue('')
    onValueChangeRef.current('')

    if (input) {
      input.focus()
    }

    onClear?.()
  }

  return (
    <InputGroup className={cn('h-8', containerClassName)} data-disabled={props.disabled ? true : undefined}>
      <InputGroupAddon align="inline-start">
        <SearchIcon className="size-4" />
      </InputGroupAddon>
      <InputGroupInput
        ref={inputRef}
        type="search"
        role="searchbox"
        autoComplete="off"
        value={currentValue}
        onChange={handleChange}
        className={cn(
          'h-8',
          '[&::-webkit-search-cancel-button]:hidden [&::-webkit-search-decoration]:hidden text-sm',
          className,
        )}
        {...props}
      />
      {hasValue ? (
        <InputGroupAddon align="inline-end">
          <InputGroupButton
            size="icon-xs"
            aria-label={clearButtonAriaLabel}
            onClick={handleClear}
            disabled={props.disabled || props.readOnly}
          >
            <XIcon className="size-4" />
          </InputGroupButton>
        </InputGroupAddon>
      ) : null}
    </InputGroup>
  )
}

export { SearchInput }
