/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AdminAuthContext, AdminUser } from '@/contexts/AdminAuthContext'
import { ReactNode, useCallback, useEffect, useState } from 'react'

const TOKEN_KEY = 'admin_access_token'

function decodeJwtPayload(token: string): Record<string, unknown> | null {
  try {
    const base64Url = token.split('.')[1]
    if (!base64Url) return null
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    return JSON.parse(atob(base64))
  } catch {
    return null
  }
}

function isTokenExpired(payload: Record<string, unknown>): boolean {
  if (typeof payload.exp !== 'number') return false
  return Date.now() / 1000 > payload.exp
}

function buildUserFromToken(token: string): AdminUser | null {
  const payload = decodeJwtPayload(token)
  if (!payload || isTokenExpired(payload)) return null
  return {
    access_token: token,
    profile: {
      sub: (payload.sub as string) || 'admin',
      name: (payload.name as string) || 'Admin',
      email: (payload.email as string) || '',
      picture: payload.picture as string | undefined,
    },
  }
}

type Props = {
  children: ReactNode
}

export function AdminAuthProvider({ children }: Props) {
  const [user, setUser] = useState<AdminUser | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    const token = localStorage.getItem(TOKEN_KEY)
    if (token) {
      setUser(buildUserFromToken(token))
    }
    setIsLoading(false)
  }, [])

  const login = useCallback((token: string) => {
    localStorage.setItem(TOKEN_KEY, token)
    setUser(buildUserFromToken(token))
  }, [])

  const logout = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY)
    setUser(null)
  }, [])

  return (
    <AdminAuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        isLoading,
        login,
        logout,
      }}
    >
      {children}
    </AdminAuthContext.Provider>
  )
}
