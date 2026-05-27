/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { NotificationSocketProvider } from '@/providers/NotificationSocketProvider'
import { OrganizationsProvider } from '@/providers/OrganizationsProvider'
import { SelectedOrganizationProvider } from '@/providers/SelectedOrganizationProvider'
import { UserOrganizationInvitationsProvider } from '@/providers/UserOrganizationInvitationsProvider'
import { initPylon } from '@/vendor/pylon'
import { OrganizationRolePermissionsEnum, OrganizationUserRoleEnum } from '@daytona/api-client'
import { useFeatureFlagEnabled, usePostHog } from 'posthog-js/react'
import React, { Suspense, useEffect } from 'react'
import { useAuth } from 'react-oidc-context'
import { Navigate, Route, Routes, useLocation } from 'react-router-dom'
import { BannerProvider } from './components/Banner'
import { CommandPaletteProvider } from './components/CommandPalette'
import LoadingFallback from './components/LoadingFallback'
import { PageSuspense } from './components/PageSuspense'
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
import { FeatureFlags } from './enums/FeatureFlags'
import { RoutePath, getRouteSubPath } from './enums/RoutePath'
import { useConfig } from './hooks/useConfig'
import Dashboard from './pages/Dashboard'
import LandingPage from './pages/LandingPage'
import Logout from './pages/Logout'
import NotFound from './pages/NotFound'

import { ApiProvider } from './providers/ApiProvider'
import { RegionsProvider } from './providers/RegionsProvider'
import { SvixProvider } from './providers/SvixProvider'

const AccountSettings = React.lazy(() => import('./pages/AccountSettings'))
const AuditLogs = React.lazy(() => import('./pages/AuditLogs'))
const EmailVerify = React.lazy(() => import('./pages/EmailVerify'))
const Keys = React.lazy(() => import('./pages/Keys'))
const Limits = React.lazy(() => import('./pages/Limits'))
const Onboarding = React.lazy(() => import('./pages/Onboarding'))
const OrganizationMembers = React.lazy(() => import('./pages/OrganizationMembers'))
const OrganizationSettings = React.lazy(() => import('./pages/OrganizationSettings'))
const Playground = React.lazy(() => import('./pages/Playground'))
const Regions = React.lazy(() => import('./pages/Regions'))
const Registries = React.lazy(() => import('./pages/Registries'))
const Runners = React.lazy(() => import('./pages/Runners'))
const SandboxDetails = React.lazy(() => import('./components/sandboxes/SandboxDetails'))
const Sandboxes = React.lazy(() => import('./pages/Sandboxes'))
const Snapshots = React.lazy(() => import('./pages/Snapshots'))
const Spending = React.lazy(() => import('./pages/Spending'))
const UserOrganizationInvitations = React.lazy(() => import('./pages/UserOrganizationInvitations'))
const Volumes = React.lazy(() => import('./pages/Volumes'))
const Wallet = React.lazy(() => import('./pages/Wallet'))
const WebhookEndpointDetails = React.lazy(() => import('./pages/WebhookEndpointDetails'))
const Webhooks = React.lazy(() => import('./pages/Webhooks'))

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
  const config = useConfig()
  const location = useLocation()
  const posthog = usePostHog()

  const { error: authError, isAuthenticated, user, signoutRedirect } = useAuth()

  useEffect(() => {
    if (isAuthenticated && user && posthog?.get_distinct_id() !== user.profile.sub) {
      posthog?.identify(user.profile.sub, {
        email: user.profile.email,
        name: user.profile.name,
      })
    }
    if (import.meta.env.PROD && config.pylonAppId && isAuthenticated && user) {
      initPylon(config.pylonAppId, {
        chat_settings: {
          app_id: config.pylonAppId,
          email: user.profile.email || '',
          name: user.profile.name || '',
          avatar_url: user.profile.picture,
          email_hash: user.profile?.email_hash as string | undefined,
        },
      })
    }
  }, [isAuthenticated, user, posthog, config.pylonAppId])

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
    <Routes>
      <Route path={RoutePath.LANDING} element={<LandingPage />} />
      <Route path={RoutePath.LOGOUT} element={<Logout />} />
      <Route path={RoutePath.DOCS} element={<DocsRedirect />} />
      <Route path={RoutePath.SLACK} element={<SlackRedirect />} />
      <Route
        path={RoutePath.DASHBOARD}
        element={
          <Suspense fallback={<LoadingFallback />}>
            <ApiProvider>
              <OrganizationsProvider>
                <SelectedOrganizationProvider>
                  <RegionsProvider>
                    <UserOrganizationInvitationsProvider>
                      <NotificationSocketProvider>
                        <CommandPaletteProvider>
                          <BannerProvider>
                            <Dashboard />
                          </BannerProvider>
                        </CommandPaletteProvider>
                      </NotificationSocketProvider>
                    </UserOrganizationInvitationsProvider>
                  </RegionsProvider>
                </SelectedOrganizationProvider>
              </OrganizationsProvider>
            </ApiProvider>
          </Suspense>
        }
      >
        <Route index element={<Navigate to={`${getRouteSubPath(RoutePath.SANDBOXES)}${location.search}`} replace />} />
        <Route
          path={getRouteSubPath(RoutePath.KEYS)}
          element={
            <PageSuspense>
              <Keys />
            </PageSuspense>
          }
        />
        <Route
          path={getRouteSubPath(RoutePath.SANDBOXES)}
          element={
            <PageSuspense>
              <Sandboxes />
            </PageSuspense>
          }
        />
        <Route
          path={getRouteSubPath(RoutePath.SANDBOX_DETAILS)}
          element={
            <PageSuspense>
              <SandboxDetails />
            </PageSuspense>
          }
        />
        <Route
          path={getRouteSubPath(RoutePath.SNAPSHOTS)}
          element={
            <PageSuspense>
              <Snapshots />
            </PageSuspense>
          }
        />
        <Route
          path={getRouteSubPath(RoutePath.REGISTRIES)}
          element={
            <PageSuspense>
              <Registries />
            </PageSuspense>
          }
        />
        <Route
          path={getRouteSubPath(RoutePath.VOLUMES)}
          element={
            <RequiredPermissionsOrganizationPageWrapper
              requiredPermissions={[OrganizationRolePermissionsEnum.READ_VOLUMES]}
            >
              <PageSuspense>
                <Volumes />
              </PageSuspense>
            </RequiredPermissionsOrganizationPageWrapper>
          }
        />
        <Route
          path={getRouteSubPath(RoutePath.LIMITS)}
          element={
            <OwnerAccessOrganizationPageWrapper>
              <PageSuspense>
                <Limits />
              </PageSuspense>
            </OwnerAccessOrganizationPageWrapper>
          }
        />
        {config.billingApiUrl && (
          <>
            <Route
              path={getRouteSubPath(RoutePath.BILLING_SPENDING)}
              element={
                <OwnerAccessOrganizationPageWrapper>
                  <PageSuspense>
                    <Spending />
                  </PageSuspense>
                </OwnerAccessOrganizationPageWrapper>
              }
            />
            <Route
              path={getRouteSubPath(RoutePath.BILLING_WALLET)}
              element={
                <OwnerAccessOrganizationPageWrapper>
                  <PageSuspense>
                    <Wallet />
                  </PageSuspense>
                </OwnerAccessOrganizationPageWrapper>
              }
            />
            <Route
              path={getRouteSubPath(RoutePath.EMAIL_VERIFY)}
              element={
                <PageSuspense>
                  <EmailVerify />
                </PageSuspense>
              }
            />
          </>
        )}
        <Route
          path={getRouteSubPath(RoutePath.MEMBERS)}
          element={
            <PageSuspense>
              <OrganizationMembers />
            </PageSuspense>
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
              <PageSuspense>
                <AuditLogs />
              </PageSuspense>
            </RequiredPermissionsOrganizationPageWrapper>
          }
        />
        <Route
          path={getRouteSubPath(RoutePath.SETTINGS)}
          element={
            <PageSuspense>
              <OrganizationSettings />
            </PageSuspense>
          }
        />
        <Route
          path={getRouteSubPath(RoutePath.REGIONS)}
          element={
            <RequiredFeatureFlagWrapper flagKey={FeatureFlags.ORGANIZATION_INFRASTRUCTURE}>
              <PageSuspense>
                <Regions />
              </PageSuspense>
            </RequiredFeatureFlagWrapper>
          }
        />
        <Route
          path={getRouteSubPath(RoutePath.RUNNERS)}
          element={
            <RequiredFeatureFlagWrapper flagKey={FeatureFlags.ORGANIZATION_INFRASTRUCTURE}>
              <RequiredPermissionsOrganizationPageWrapper
                requiredPermissions={[OrganizationRolePermissionsEnum.READ_RUNNERS]}
              >
                <PageSuspense>
                  <Runners />
                </PageSuspense>
              </RequiredPermissionsOrganizationPageWrapper>
            </RequiredFeatureFlagWrapper>
          }
        />
        <Route
          path={getRouteSubPath(RoutePath.ACCOUNT_SETTINGS)}
          element={
            <PageSuspense>
              <AccountSettings linkedAccountsEnabled={config.linkedAccountsEnabled} />
            </PageSuspense>
          }
        />
        <Route
          path={getRouteSubPath(RoutePath.USER_INVITATIONS)}
          element={
            <PageSuspense>
              <UserOrganizationInvitations />
            </PageSuspense>
          }
        />
        <Route
          path={getRouteSubPath(RoutePath.ONBOARDING)}
          element={
            <PageSuspense>
              <Onboarding />
            </PageSuspense>
          }
        />
        <Route
          path={getRouteSubPath(RoutePath.PLAYGROUND)}
          element={
            <PageSuspense>
              <Playground />
            </PageSuspense>
          }
        />
        <Route
          path={getRouteSubPath(RoutePath.WEBHOOKS)}
          element={
            <SvixProvider>
              <PageSuspense>
                <Webhooks />
              </PageSuspense>
            </SvixProvider>
          }
        />
        <Route
          path={getRouteSubPath(RoutePath.WEBHOOK_ENDPOINT_DETAILS)}
          element={
            <SvixProvider>
              <PageSuspense>
                <WebhookEndpointDetails />
              </PageSuspense>
            </SvixProvider>
          }
        />
      </Route>
      <Route path="*" element={<NotFound />} />
    </Routes>
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

function RequiredFeatureFlagWrapper({ children, flagKey }: { children: React.ReactNode; flagKey: FeatureFlags }) {
  const flagEnabled = useFeatureFlagEnabled(flagKey)

  if (!flagEnabled) {
    return <Navigate to={RoutePath.DASHBOARD} replace />
  }

  return children
}

export default App
