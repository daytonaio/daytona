/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'

// Mock implementation of the DropdownMenu components
export const DropdownMenu = ({ children }: { children: React.ReactNode }) => <div>{children}</div>
export const DropdownMenuTrigger = ({ children }: { children: React.ReactNode }) => <div>{children}</div>
export const DropdownMenuContent = ({ children }: { children: React.ReactNode }) => <div>{children}</div>
export const DropdownMenuItem = ({
  children,
  onClick,
  disabled,
}: {
  children: React.ReactNode
  onClick?: () => void
  disabled?: boolean
}) => (
  <button onClick={onClick} disabled={disabled}>
    {children}
  </button>
)
export const DropdownMenuSeparator = () => <hr />
export const DropdownMenuLabel = ({ children }: { children: React.ReactNode }) => <div>{children}</div>
