/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import Onboarding from '@/pages/Onboarding'
import OrganizationMembers from '@/pages/OrganizationMembers'
import OrganizationSettings from '@/pages/OrganizationSettings'
import UserOrganizationInvitations from '@/pages/UserOrganizationInvitations'
import { NotificationSocketProvider } from '@/providers/NotificationSocketProvider'
import { OrganizationsProvider } from '@/providers/OrganizationsProvider'
import { SelectedOrganizationProvider } from '@/providers/SelectedOrganizationProvider'
import { UserOrganizationInvitationsProvider } from '@/providers/UserOrganizationInvitationsProvider'
import { OrganizationRolePermissionsEnum, OrganizationUserRoleEnum } from '@daytonaio/api-client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { usePostHog } from 'posthog-js/react'
import React, { Suspense, useEffect } from 'react'
import { useAuth } from 'react-oidc-context'
import { Navigate, Route, Routes, useLocation } from 'react-router-dom'
import LoadingFallback from './components/LoadingFallback'
import { Button } from './components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from './components/ui/dialog'
import { DAYTONA_DOCS_URL, DAYTONA_SLACK_URL } from './constants/ExternalLinks'
import { ThemeProvider } from './contexts/ThemeContext'
import { RoutePath, getRouteSubPath } from './enums/RoutePath'
import AccountSettings from './pages/AccountSettings'
import AuditLogs from './pages/AuditLogs'
import Dashboard from './pages/Dashboard'
import EmailVerify from './pages/EmailVerify'
import Keys from './pages/Keys'
import LandingPage from './pages/LandingPage'
import Limits from './pages/Limits'
import Logout from './pages/Logout'
import NotFound from './pages/NotFound'
import Registries from './pages/Registries'
import Sandboxes from './pages/Sandboxes'
import Snapshots from './pages/Snapshots'
import Spending from './pages/Spending'
import Volumes from './pages/Volumes'
import Wallet from './pages/Wallet'
import { ApiProvider } from './providers/ApiProvider'
import { BillingProvider } from './providers/BillingProvider'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
      staleTime: 1000 * 60 * 5, // 5 minutes
    },
  },
})

// Simple redirection components for external URLs
const DocsRedirect = () => {
  React.useEffect(() => {
    window.open(DAYTONA_DOCS_URL, '_blank')
    window.location.href = RoutePath.DASHBOARD
  }, [])
  return null
}

const SlackRedirect = () => {
  React.useEffect(() => {
    window.open(DAYTONA_SLACK_URL, '_blank')
    window.location.href = RoutePath.DASHBOARD
  }, [])
  return null
}

function App() {
  const location = useLocation()
  const posthog = usePostHog()
  const { error: authError, isAuthenticated, user, signoutRedirect } = useAuth()

  useEffect(() => {
    if (import.meta.env.PROD && isAuthenticated && user && posthog?.get_distinct_id() !== user.profile.sub) {
      posthog?.identify(user.profile.sub, {
        email: user.profile.email,
        name: user.profile.name,
      })
    }
    if (import.meta.env.PROD && import.meta.env.VITE_PYLON_APP_ID && isAuthenticated && user) {
      window.pylon = {
        chat_settings: {
          app_id: import.meta.env.VITE_PYLON_APP_ID,
          email: user.profile.email || '',
          name: user.profile.name || '',
          avatar_url: user.profile.picture,
          email_hash: user.profile?.email_hash as string | undefined,
        },
      }
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

  if (authError) {
    return (
      <Dialog open>
        <DialogContent className="[&>button]:hidden">
          <DialogHeader>
            <DialogTitle>Authentication Error</DialogTitle>
            <DialogDescription>{authError.message}</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button onClick={() => signoutRedirect()}>Go Back</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    )
  }

  return (
    <ThemeProvider>
      <Routes>
        <Route path={RoutePath.LANDING} element={<LandingPage />} />
        <Route path={RoutePath.LOGOUT} element={<Logout />} />
        <Route path={RoutePath.DOCS} element={<DocsRedirect />} />
        <Route path={RoutePath.SLACK} element={<SlackRedirect />} />
        <Route
          path={RoutePath.DASHBOARD}
          element={
            <Suspense fallback={<LoadingFallback />}>
              <QueryClientProvider client={queryClient}>
                <ApiProvider>
                  <OrganizationsProvider>
                    <SelectedOrganizationProvider>
                      <BillingProvider>
                        <UserOrganizationInvitationsProvider>
                          <NotificationSocketProvider>
                            <Dashboard />
                          </NotificationSocketProvider>
                        </UserOrganizationInvitationsProvider>
                      </BillingProvider>
                    </SelectedOrganizationProvider>
                  </OrganizationsProvider>
                </ApiProvider>
              </QueryClientProvider>
            </Suspense>
          }
        >
          <Route
            index
            element={<Navigate to={`${getRouteSubPath(RoutePath.SANDBOXES)}${location.search}`} replace />}
          />
          <Route path={getRouteSubPath(RoutePath.KEYS)} element={<Keys />} />
          <Route path={getRouteSubPath(RoutePath.SANDBOXES)} element={<Sandboxes />} />
          <Route path={getRouteSubPath(RoutePath.SNAPSHOTS)} element={<Snapshots />} />
          <Route path={getRouteSubPath(RoutePath.REGISTRIES)} element={<Registries />} />
          <Route
            path={getRouteSubPath(RoutePath.VOLUMES)}
            element={
              <RequiredPermissionsOrganizationPageWrapper
                requiredPermissions={[OrganizationRolePermissionsEnum.READ_VOLUMES]}
              >
                <Volumes />
              </RequiredPermissionsOrganizationPageWrapper>
            }
          />
          <Route
            path={getRouteSubPath(RoutePath.LIMITS)}
            element={
              <OwnerAccessOrganizationPageWrapper>
                <Limits />
              </OwnerAccessOrganizationPageWrapper>
            }
          />
          {import.meta.env.VITE_BILLING_API_URL && (
            <>
              <Route
                path={getRouteSubPath(RoutePath.BILLING_SPENDING)}
                element={
                  <OwnerAccessOrganizationPageWrapper>
                    <Spending />
                  </OwnerAccessOrganizationPageWrapper>
                }
              />
              <Route
                path={getRouteSubPath(RoutePath.BILLING_WALLET)}
                element={
                  <OwnerAccessOrganizationPageWrapper>
                    <Wallet />
                  </OwnerAccessOrganizationPageWrapper>
                }
              />
              <Route path={getRouteSubPath(RoutePath.EMAIL_VERIFY)} element={<EmailVerify />} />
            </>
          )}
          <Route
            path={getRouteSubPath(RoutePath.MEMBERS)}
            element={
              <NonPersonalOrganizationPageWrapper>
                <OrganizationMembers />
              </NonPersonalOrganizationPageWrapper>
            }
          />
          {
            // TODO: uncomment when we allow creating custom roles
            /* <Route
            path={getRouteSubPath(RoutePath.ROLES)}
            element={
              <NonPersonalOrganizationPageWrapper>
                <OwnerAccessOrganizationPageWrapper>
                  <OrganizationRoles />
                </OwnerAccessOrganizationPageWrapper>
              </NonPersonalOrganizationPageWrapper>
            }
          /> */
          }
          <Route
            path={getRouteSubPath(RoutePath.AUDIT_LOGS)}
            element={
              <RequiredPermissionsOrganizationPageWrapper
                requiredPermissions={[OrganizationRolePermissionsEnum.READ_AUDIT_LOGS]}
              >
                <AuditLogs />
              </RequiredPermissionsOrganizationPageWrapper>
            }
          />
          <Route path={getRouteSubPath(RoutePath.SETTINGS)} element={<OrganizationSettings />} />
          <Route path={getRouteSubPath(RoutePath.ACCOUNT_SETTINGS)} element={<AccountSettings />} />
          <Route path={getRouteSubPath(RoutePath.USER_INVITATIONS)} element={<UserOrganizationInvitations />} />
          <Route path={getRouteSubPath(RoutePath.ONBOARDING)} element={<Onboarding />} />
        </Route>
        <Route path="*" element={<NotFound />} />
      </Routes>
    </ThemeProvider>
  )
}

function NonPersonalOrganizationPageWrapper({ children }: { children: React.ReactNode }) {
  const { selectedOrganization } = useSelectedOrganization()

  if (selectedOrganization?.personal) {
    return <Navigate to={RoutePath.DASHBOARD} replace />
  }

  return children
}

function OwnerAccessOrganizationPageWrapper({ children }: { children: React.ReactNode }) {
  const { authenticatedUserOrganizationMember } = useSelectedOrganization()

  if (authenticatedUserOrganizationMember?.role !== OrganizationUserRoleEnum.OWNER) {
    return <Navigate to={RoutePath.DASHBOARD} replace />
  }

  return children
}

function RequiredPermissionsOrganizationPageWrapper({
  children,
  requiredPermissions,
}: {
  children: React.ReactNode
  requiredPermissions: OrganizationRolePermissionsEnum[]
}) {
  const { authenticatedUserHasPermission } = useSelectedOrganization()

  if (!requiredPermissions.every((permission) => authenticatedUserHasPermission(permission))) {
    return <Navigate to={RoutePath.DASHBOARD} replace />
  }

  return children
}

export default App
