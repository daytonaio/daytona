/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
import { SandboxListClientPaginatedProvider } from './SandboxListClientPaginatedProvider'
import { SandboxListServerPaginatedProvider } from './SandboxListServerPaginatedProvider'
export { useSandboxListContext } from './useSandboxListContext'

const useClientPagination = import.meta.env.VITE_CLIENT_SIDE_SANDBOX_PAGINATION === 'true'

export const SandboxListProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  if (useClientPagination) {
    return <SandboxListClientPaginatedProvider>{children}</SandboxListClientPaginatedProvider>
  }

  return <SandboxListServerPaginatedProvider>{children}</SandboxListServerPaginatedProvider>
}
