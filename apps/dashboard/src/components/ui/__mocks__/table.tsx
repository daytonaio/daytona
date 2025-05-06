/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'

// Mock implementation of the Table components
export const Table = ({ children }: { children: React.ReactNode }) => <table>{children}</table>
export const TableHeader = ({ children }: { children: React.ReactNode }) => <thead>{children}</thead>
export const TableBody = ({ children }: { children: React.ReactNode }) => <tbody>{children}</tbody>
export const TableRow = ({
  children,
  className,
  ...props
}: {
  children: React.ReactNode
  className?: string
  [key: string]: any
}) => (
  <tr className={className} {...props}>
    {children}
  </tr>
)
export const TableHead = ({ children, className }: { children: React.ReactNode; className?: string }) => (
  <th className={className}>{children}</th>
)
export const TableCell = ({ children, className }: { children: React.ReactNode; className?: string }) => (
  <td className={className}>{children}</td>
)
