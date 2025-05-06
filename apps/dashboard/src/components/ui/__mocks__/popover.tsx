/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'

// Mock implementation of the Popover components
export const Popover = ({ children, open }: { children: React.ReactNode; open?: boolean }) => (
  <div data-state={open ? 'open' : 'closed'}>{children}</div>
)
export const PopoverTrigger = ({ children, asChild }: { children: React.ReactNode; asChild?: boolean }) => (
  <div>{children}</div>
)
export const PopoverContent = ({
  children,
  side,
  className,
}: {
  children: React.ReactNode
  side?: string
  className?: string
}) => (
  <div className={className} data-side={side}>
    {children}
  </div>
)
