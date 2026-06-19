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
import { ShieldAlert } from 'lucide-react'
import { useFeatureFlagEnabled, usePostHog } from 'posthog-js/react'
import { Suspense, useEffect, type ReactNode } from 'react'
import { useAuth } from 'react-oidc-context'
import {
  createBrowserRouter,
  Navigate,
  Outlet,
  redirect,
  useLocation,
  useNavigation,
  useRouteError,
} from 'react-router'
import { RouterProvider } from 'react-router/dom'
import { BannerProvider } from './components/Banner'
import { CommandPaletteProvider } from './components/CommandPalette'
import { ErrorBoundaryFallback } from './components/ErrorBoundaryFallback'
import LoadingFallback from './components/LoadingFallback'
import { LoadingFallbackContent } from './components/LoadingFallbackContent'
import { PageContent, PageHeader, PageIntro, PageLayout } from './components/PageLayout'
import { Badge } from './components/ui/badge'
import { Button } from './components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from './components/ui/dialog'
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from './components/ui/empty'
import { DAYTONA_DOCS_URL, DAYTONA_SLACK_URL } from './constants/ExternalLinks'
import { FeatureFlags } from './enums/FeatureFlags'
import { getRouteSubPath, RoutePath, trimLeadingSlash } from './enums/RoutePath'
import { useConfig } from './hooks/useConfig'
import Dashboard from './pages/Dashboard'
import LandingPage from './pages/LandingPage'
import Logout from './pages/Logout'
import NotFound from './pages/NotFound'

import { ApiProvider } from './providers/ApiProvider'
import { SvixProvider } from './providers/SvixProvider'
import { lazyRoutes } from './routes'

function normalizeRouteError(error: unknown) {
  if (error instanceof Error) {
    return error
  }

  if (typeof error === 'string') {
    return new Error(error)
  }

  return new Error('Unknown route error')
}

function RouteErrorFallback() {
  const error = useRouteError()

  return (
    <ErrorBoundaryFallback error={normalizeRouteError(error)} resetErrorBoundary={() => window.location.reload()} />
  )
}

function AppRoot() {
  const config = useConfig()
  const location = useLocation()
  const posthog = usePostHog()

  const { error: authError, isAuthenticated, signoutRedirect, user } = useAuth()

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

  return <Outlet />
}

function DashboardOutlet() {
  const location = useLocation()
  const navigation = useNavigation()
  const isRouteLoading = navigation.state === 'loading' && navigation.location?.pathname !== location.pathname

  return (
    <Suspense fallback={<LoadingFallback source="dashboard-suspense" />}>
      <ApiProvider>
        <OrganizationsProvider>
          <SelectedOrganizationProvider>
            <UserOrganizationInvitationsProvider>
              <NotificationSocketProvider>
                <CommandPaletteProvider>
                  <BannerProvider>
                    <Dashboard>
                      {isRouteLoading ? (
                        <div className="flex min-h-screen w-full items-center justify-center bg-background p-6">
                          <LoadingFallbackContent source="route-navigation" />
                        </div>
                      ) : (
                        <Outlet />
                      )}
                    </Dashboard>
                  </BannerProvider>
                </CommandPaletteProvider>
              </NotificationSocketProvider>
            </UserOrganizationInvitationsProvider>
          </SelectedOrganizationProvider>
        </OrganizationsProvider>
      </ApiProvider>
    </Suspense>
  )
}

function DashboardIndexRedirect() {
  const location = useLocation()

  return <Navigate to={`${getRouteSubPath(RoutePath.SANDBOXES)}${location.search}`} replace />
}

function getAccessLabel(access: string) {
  return access.replace(/[:_-]+/g, ' ').toLowerCase()
}

function AccessRequiredPage({ pageTitle, requiredAccess }: { pageTitle: ReactNode; requiredAccess: string[] }) {
  return (
    <PageLayout>
      <PageHeader />
      <PageContent>
        <PageIntro title={pageTitle} />
        <Empty className="flex-none rounded-md border py-12" variant="warning">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <ShieldAlert />
            </EmptyMedia>
            <EmptyTitle>You don&apos;t have access to this page</EmptyTitle>
            <EmptyDescription>Ask your organization owner to grant you the required access.</EmptyDescription>
          </EmptyHeader>
          <EmptyContent>
            <div className="text-xs font-medium text-muted-foreground">Required access</div>
            <div className="flex flex-wrap justify-center gap-2">
              {requiredAccess.map((access) => (
                <Badge key={access} className="capitalize" title={access}>
                  {getAccessLabel(access)}
                </Badge>
              ))}
            </div>
          </EmptyContent>
        </Empty>
      </PageContent>
    </PageLayout>
  )
}

function OwnerAccessOrganizationPageWrapper({ children, pageTitle }: { children: ReactNode; pageTitle: ReactNode }) {
  const { authenticatedUserOrganizationMember } = useSelectedOrganization()

  if (authenticatedUserOrganizationMember?.role !== OrganizationUserRoleEnum.OWNER) {
    return <AccessRequiredPage pageTitle={pageTitle} requiredAccess={['owner role']} />
  }

  return children
}

function RequiredPermissionsOrganizationPageWrapper({
  children,
  pageTitle,
  requiredPermissions,
}: {
  children: ReactNode
  pageTitle: ReactNode
  requiredPermissions: OrganizationRolePermissionsEnum[]
}) {
  const { authenticatedUserHasPermission } = useSelectedOrganization()
  const missingPermissions = requiredPermissions.filter((permission) => {
    return !authenticatedUserHasPermission(permission)
  })

  if (missingPermissions.length > 0) {
    return <AccessRequiredPage pageTitle={pageTitle} requiredAccess={missingPermissions} />
  }

  return children
}

function RequiredFeatureFlagWrapper({ children, flagKey }: { children: ReactNode; flagKey: FeatureFlags }) {
  const flagEnabled = useFeatureFlagEnabled(flagKey)

  if (!flagEnabled) {
    return <Navigate to={RoutePath.DASHBOARD} replace />
  }

  return children
}

function OwnerAccessOrganizationOutlet({ pageTitle }: { pageTitle: ReactNode }) {
  return (
    <OwnerAccessOrganizationPageWrapper pageTitle={pageTitle}>
      <Outlet />
    </OwnerAccessOrganizationPageWrapper>
  )
}

function RequiredPermissionsOrganizationOutlet({
  pageTitle,
  requiredPermissions,
}: {
  pageTitle: ReactNode
  requiredPermissions: OrganizationRolePermissionsEnum[]
}) {
  return (
    <RequiredPermissionsOrganizationPageWrapper pageTitle={pageTitle} requiredPermissions={requiredPermissions}>
      <Outlet />
    </RequiredPermissionsOrganizationPageWrapper>
  )
}

function RequiredFeatureFlagOutlet({ flagKey }: { flagKey: FeatureFlags }) {
  return (
    <RequiredFeatureFlagWrapper flagKey={flagKey}>
      <Outlet />
    </RequiredFeatureFlagWrapper>
  )
}

function BillingEnabledOutlet() {
  const config = useConfig()

  if (!config.billingApiUrl) {
    return <Navigate to={RoutePath.DASHBOARD} replace />
  }

  return <Outlet />
}

function BillingOwnerAccessOutlet({ pageTitle }: { pageTitle: ReactNode }) {
  const config = useConfig()

  if (!config.billingApiUrl) {
    return <Navigate to={RoutePath.DASHBOARD} replace />
  }

  return (
    <OwnerAccessOrganizationPageWrapper pageTitle={pageTitle}>
      <Outlet />
    </OwnerAccessOrganizationPageWrapper>
  )
}

function RunnersAccessOutlet() {
  return (
    <RequiredFeatureFlagWrapper flagKey={FeatureFlags.ORGANIZATION_INFRASTRUCTURE}>
      <RequiredPermissionsOrganizationPageWrapper
        pageTitle="Runners"
        requiredPermissions={[OrganizationRolePermissionsEnum.READ_RUNNERS]}
      >
        <Outlet />
      </RequiredPermissionsOrganizationPageWrapper>
    </RequiredFeatureFlagWrapper>
  )
}

function WebhooksOutlet() {
  return (
    <SvixProvider>
      <Outlet />
    </SvixProvider>
  )
}

const router = createBrowserRouter([
  {
    path: RoutePath.LANDING,
    element: <AppRoot />,
    hydrateFallbackElement: <LoadingFallback source="app-root-hydrate" />,
    errorElement: <RouteErrorFallback />,
    children: [
      { index: true, element: <LandingPage /> },
      { path: trimLeadingSlash(RoutePath.LOGOUT), element: <Logout /> },
      { path: trimLeadingSlash(RoutePath.DOCS), loader: () => redirect(DAYTONA_DOCS_URL) },
      { path: trimLeadingSlash(RoutePath.SLACK), loader: () => redirect(DAYTONA_SLACK_URL) },
      {
        path: trimLeadingSlash(RoutePath.DASHBOARD),
        element: <DashboardOutlet />,
        children: [
          { index: true, element: <DashboardIndexRedirect /> },
          { path: getRouteSubPath(RoutePath.KEYS), lazy: lazyRoutes.Keys },
          { path: getRouteSubPath(RoutePath.SANDBOXES), lazy: lazyRoutes.Sandboxes },
          { path: getRouteSubPath(RoutePath.SANDBOX_DETAILS), lazy: lazyRoutes.SandboxDetails },
          { path: getRouteSubPath(RoutePath.SNAPSHOTS), lazy: lazyRoutes.Snapshots },
          { path: getRouteSubPath(RoutePath.REGISTRIES), lazy: lazyRoutes.Registries },
          {
            path: getRouteSubPath(RoutePath.VOLUMES),
            element: (
              <RequiredPermissionsOrganizationOutlet
                pageTitle="Volumes"
                requiredPermissions={[OrganizationRolePermissionsEnum.READ_VOLUMES]}
              />
            ),
            children: [{ index: true, lazy: lazyRoutes.Volumes }],
          },
          {
            path: getRouteSubPath(RoutePath.LIMITS),
            element: <OwnerAccessOrganizationOutlet pageTitle="Limits" />,
            children: [{ index: true, lazy: lazyRoutes.Limits }],
          },
          {
            path: getRouteSubPath(RoutePath.BILLING_SPENDING),
            element: <BillingOwnerAccessOutlet pageTitle="Spending" />,
            children: [{ index: true, lazy: lazyRoutes.Spending }],
          },
          {
            path: getRouteSubPath(RoutePath.BILLING_WALLET),
            element: <BillingOwnerAccessOutlet pageTitle="Wallet" />,
            children: [{ index: true, lazy: lazyRoutes.Wallet }],
          },
          {
            path: getRouteSubPath(RoutePath.EMAIL_VERIFY),
            element: <BillingEnabledOutlet />,
            children: [{ index: true, lazy: lazyRoutes.EmailVerify }],
          },
          { path: getRouteSubPath(RoutePath.MEMBERS), lazy: lazyRoutes.OrganizationMembers },
          {
            path: getRouteSubPath(RoutePath.AUDIT_LOGS),
            element: (
              <RequiredPermissionsOrganizationOutlet
                pageTitle="Audit Logs"
                requiredPermissions={[OrganizationRolePermissionsEnum.READ_AUDIT_LOGS]}
              />
            ),
            children: [{ index: true, lazy: lazyRoutes.AuditLogs }],
          },
          { path: getRouteSubPath(RoutePath.SETTINGS), lazy: lazyRoutes.OrganizationSettings },
          {
            path: getRouteSubPath(RoutePath.REGIONS),
            element: <RequiredFeatureFlagOutlet flagKey={FeatureFlags.ORGANIZATION_INFRASTRUCTURE} />,
            children: [{ index: true, lazy: lazyRoutes.Regions }],
          },
          {
            path: getRouteSubPath(RoutePath.RUNNERS),
            element: <RunnersAccessOutlet />,
            children: [{ index: true, lazy: lazyRoutes.Runners }],
          },
          { path: getRouteSubPath(RoutePath.ACCOUNT_SETTINGS), lazy: lazyRoutes.AccountSettings },
          { path: getRouteSubPath(RoutePath.USER_INVITATIONS), lazy: lazyRoutes.UserOrganizationInvitations },
          { path: getRouteSubPath(RoutePath.ONBOARDING), lazy: lazyRoutes.Onboarding },
          { path: getRouteSubPath(RoutePath.PLAYGROUND), lazy: lazyRoutes.Playground },
          {
            path: getRouteSubPath(RoutePath.WEBHOOKS),
            element: <WebhooksOutlet />,
            children: [
              { index: true, lazy: lazyRoutes.Webhooks },
              { path: ':endpointId', lazy: lazyRoutes.WebhookEndpointDetails },
            ],
          },
        ],
      },
      { path: '*', element: <NotFound /> },
    ],
  },
])

function App() {
  return <RouterProvider router={router} />
}

export default App
