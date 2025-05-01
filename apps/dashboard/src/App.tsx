/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { Suspense, useEffect, useState } from 'react'
import { Routes, Route, Navigate, useLocation } from 'react-router-dom'
import Dashboard from './pages/Dashboard'
import Workspaces from './pages/Workspaces'
import Keys from './pages/Keys'
import { ThemeProvider } from './contexts/ThemeContext'
import { useAuth } from 'react-oidc-context'
import LoadingFallback from './components/LoadingFallback'
import Images from './pages/Images'
import Registries from './pages/Registries'
import { usePostHog } from 'posthog-js/react'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from './components/ui/dialog'
import { OrganizationsProvider } from '@/providers/OrganizationsProvider'
import { SelectedOrganizationProvider } from '@/providers/SelectedOrganizationProvider'
import { UserOrganizationInvitationsProvider } from '@/providers/UserOrganizationInvitationsProvider'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import OrganizationMembers from '@/pages/OrganizationMembers'
import OrganizationSettings from '@/pages/OrganizationSettings'
import UserOrganizationInvitations from '@/pages/UserOrganizationInvitations'
import { OrganizationUserRoleEnum } from '@daytonaio/api-client'
import Usage from './pages/Usage'
import Billing from './pages/Billing'
import { NotificationSocketProvider } from '@/providers/NotificationSocketProvider'
import { ApiProvider } from './providers/ApiProvider'
import LandingPage from './pages/LandingPage'
import Logout from './pages/Logout'
import { RoutePath } from './enums/RoutePath'

// Docs redirect component
const DocsRedirect = () => {
  React.useEffect(() => {
    window.open('https://www.daytona.io/docs/', '_blank')
    // Navigate back to dashboard after opening docs
    window.location.href = '/dashboard'
  }, [])

  return null
}

function App() {
  const location = useLocation()
  const posthog = usePostHog()
  const { error, isAuthenticated, user } = useAuth()
  const [errorDialogOpen, setErrorDialogOpen] = useState(false)

  useEffect(() => {
    if (import.meta.env.PROD && isAuthenticated && user && posthog?.get_distinct_id() !== user.profile.sub) {
      posthog?.identify(user.profile.sub, {
        email: user.profile.email,
        name: user.profile.name,
      })
    }
  }, [isAuthenticated, user, posthog])

  // Hack for tracking PostHog pageviews in SPAs
  useEffect(() => {
    if (import.meta.env.PROD) {
      posthog?.capture('$pageview', {
        $current_url: window.location.href,
      })
    }
  }, [location, posthog])

  useEffect(() => {
    if (error) {
      if (error.message === 'User not found in waitlist.') {
        window.location.href = 'https://www.daytona.io/failed-signup'
        return
      }

      setErrorDialogOpen(true)
    }
  }, [error])

  const [errorTitle, errorDescription] = error?.message.startsWith(`You're currently`)
    ? [error.message.split('\n')[0], error.message.split('\n').slice(1).join('\n')]
    : ['Authentication Error', error?.message]

  return (
    <ThemeProvider>
      <Dialog open={errorDialogOpen} onOpenChange={setErrorDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{errorTitle}</DialogTitle>
            <DialogDescription>{errorDescription}</DialogDescription>
          </DialogHeader>
        </DialogContent>
      </Dialog>
      <Routes>
        <Route path={RoutePath.LandingPage} element={<LandingPage />} />
        <Route path={RoutePath.Logout} element={<Logout />} />
        <Route
          path={RoutePath.Dashboard}
          element={
            <Suspense fallback={<LoadingFallback />}>
              <ApiProvider>
                <NotificationSocketProvider>
                  <OrganizationsProvider>
                    <SelectedOrganizationProvider>
                      <UserOrganizationInvitationsProvider>
                        <Dashboard />
                      </UserOrganizationInvitationsProvider>
                    </SelectedOrganizationProvider>
                  </OrganizationsProvider>
                </NotificationSocketProvider>
              </ApiProvider>
            </Suspense>
          }
        >
          <Route index element={<Navigate to={RoutePath.Sandboxes} replace />} />
          <Route path={RoutePath.Keys} element={<Keys />} />
          <Route path={RoutePath.Sandboxes} element={<Workspaces />} />
          <Route path={RoutePath.Images} element={<Images />} />
          <Route path={RoutePath.Registries} element={<Registries />} />
          <Route path={RoutePath.Usage} element={<Usage />} />
          {import.meta.env.VITE_BILLING_API_URL && (
            <Route
              path={RoutePath.Billing}
              element={
                <OwnerAccessOrganizationPageWrapper>
                  <Billing />
                </OwnerAccessOrganizationPageWrapper>
              }
            />
          )}
          <Route
            path={RoutePath.Organizations}
            element={
              <NonPersonalOrganizationPageWrapper>
                <OrganizationMembers />
              </NonPersonalOrganizationPageWrapper>
            }
          />
          {
            // TODO: uncomment when we allow creating custom roles
            /* <Route
            path="roles"
            element={
              <NonPersonalOrganizationPageWrapper>
                <OwnerAccessOrganizationPageWrapper>
                  <OrganizationRoles />
                </OwnerAccessOrganizationPageWrapper>
              </NonPersonalOrganizationPageWrapper>
            }
          /> */
          }
          <Route path={RoutePath.Settings} element={<OrganizationSettings />} />
          <Route path={RoutePath.UserOrganizationInvitations} element={<UserOrganizationInvitations />} />
        </Route>
        <Route path={RoutePath.DocsRedirect} element={<DocsRedirect />} />
        {/* Add other routes as needed */}
      </Routes>
    </ThemeProvider>
  )
}

function NonPersonalOrganizationPageWrapper({ children }: { children: React.ReactNode }) {
  const { selectedOrganization } = useSelectedOrganization()

  if (selectedOrganization?.personal) {
    return <Navigate to={RoutePath.Dashboard} replace />
  }

  return <>{children}</>
}

function OwnerAccessOrganizationPageWrapper({ children }: { children: React.ReactNode }) {
  const { authenticatedUserOrganizationMember } = useSelectedOrganization()

  if (authenticatedUserOrganizationMember?.role !== OrganizationUserRoleEnum.OWNER) {
    return <Navigate to={RoutePath.Dashboard} replace />
  }

  return <>{children}</>
}

export default App
