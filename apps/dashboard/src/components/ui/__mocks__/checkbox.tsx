/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'

// Mock implementation of the Checkbox component
export const Checkbox = React.forwardRef<
  HTMLInputElement,
  React.InputHTMLAttributes<HTMLInputElement> & { indeterminate?: boolean }
>(({ className, checked, indeterminate, ...props }, ref) => {
  const inputRef = React.useRef<HTMLInputElement>(null)

  React.useImperativeHandle(ref, () => inputRef.current as HTMLInputElement)

  React.useEffect(() => {
    if (inputRef.current) {
      inputRef.current.indeterminate = indeterminate === true
    }
  }, [indeterminate])

  return <input type="checkbox" ref={inputRef} className={className} checked={checked} {...props} />
})
