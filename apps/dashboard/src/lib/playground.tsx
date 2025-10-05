/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ReactNode } from 'react'

export const createErrorMessageOutput = (error: unknown): ReactNode => {
  return (
    <span>
      <span className="text-red-500">Error: </span>
      <span>{error instanceof Error ? error.message : String(error)}</span>
    </span>
  )
}
