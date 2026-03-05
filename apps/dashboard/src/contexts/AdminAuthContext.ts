/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { createContext } from 'react'

export interface AdminUser {
  access_token: string
  profile: { sub: string; name: string; email: string; picture?: string }
}

export interface AdminAuthContextValue {
  user: AdminUser | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (token: string) => void
  logout: () => void
}

export const AdminAuthContext = createContext<AdminAuthContextValue>({
  user: null,
  isAuthenticated: false,
  isLoading: true,
  login: () => {},
  logout: () => {},
})
