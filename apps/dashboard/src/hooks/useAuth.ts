/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AdminAuthContext } from '@/contexts/AdminAuthContext'
import { RoutePath } from '@/enums/RoutePath'
import { useContext } from 'react'
import { useNavigate } from 'react-router-dom'

export function useAuth() {
  const { user, isAuthenticated, isLoading, login, logout } = useContext(AdminAuthContext)
  const navigate = useNavigate()

  const signinRedirect = (options?: { state?: { returnTo?: string } }) => {
    navigate(RoutePath.LANDING, { state: options?.state })
  }

  const signoutRedirect = () => {
    logout()
    navigate(RoutePath.LANDING)
  }

  return {
    user,
    isAuthenticated,
    isLoading,
    login,
    signinRedirect,
    signoutRedirect,
  }
}
