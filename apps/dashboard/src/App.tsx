/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import Onboarding from '@/pages/Onboarding'
import OrganizationMembers from '@/pages/OrganizationMembers'
import OrganizationSettings from '@/pages/OrganizationSettings'
import { OrganizationsProvider } from '@/providers/OrganizationsProvider'
import { SelectedOrganizationProvider } from '@/providers/SelectedOrganizationProvider'
import { OrganizationRolePermissionsEnum, OrganizationUserRoleEnum } from '@daytonaio/api-client'
import React, { Suspense } from 'react'
import { Navigate, Route, Routes, useLocation } from 'react-router-dom'
import { BannerProvider } from './components/Banner'
import { CommandPaletteProvider } from './components/CommandPalette'
import LoadingFallback from './components/LoadingFallback'
import { DAYTONA_DOCS_URL, DAYTONA_SLACK_URL } from './constants/ExternalLinks'
import { FeatureFlags } from './enums/FeatureFlags'
import { RoutePath, getRouteSubPath } from './enums/RoutePath'
import { useConfig } from './hooks/useConfig'
import AccountSettings from './pages/AccountSettings'
import Dashboard from './pages/Dashboard'
import Experimental from './pages/Experimental'
import Keys from './pages/Keys'
import LandingPage from './pages/LandingPage'
import Limits from './pages/Limits'
import Logout from './pages/Logout'
import NotFound from './pages/NotFound'
import Playground from './pages/Playground'
import Regions from './pages/Regions'
import Registries from './pages/Registries'
import Runners from './pages/Runners'
import Sandboxes from './pages/Sandboxes'
import Snapshots from './pages/Snapshots'
import Volumes from './pages/Volumes'
import { ApiProvider } from './providers/ApiProvider'
import { NotificationSocketProvider } from './providers/NotificationSocketProvider'
import { RegionsProvider } from './providers/RegionsProvider'

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
                    <NotificationSocketProvider>
                      <CommandPaletteProvider>
                        <BannerProvider>
                          <Dashboard />
                        </BannerProvider>
                      </CommandPaletteProvider>
                    </NotificationSocketProvider>
                  </RegionsProvider>
                </SelectedOrganizationProvider>
              </OrganizationsProvider>
            </ApiProvider>
          </Suspense>
        }
      >
        <Route index element={<Navigate to={`${getRouteSubPath(RoutePath.SANDBOXES)}${location.search}`} replace />} />
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
        <Route
          path={getRouteSubPath(RoutePath.MEMBERS)}
          element={
            <NonPersonalOrganizationPageWrapper>
              <OrganizationMembers />
            </NonPersonalOrganizationPageWrapper>
          }
        />
        <Route path={getRouteSubPath(RoutePath.SETTINGS)} element={<OrganizationSettings />} />
        <Route
          path={getRouteSubPath(RoutePath.REGIONS)}
          element={
            <RequiredFeatureFlagWrapper flagKey={FeatureFlags.ORGANIZATION_INFRASTRUCTURE}>
              <Regions />
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
                <Runners />
              </RequiredPermissionsOrganizationPageWrapper>
            </RequiredFeatureFlagWrapper>
          }
        />
        <Route
          path={getRouteSubPath(RoutePath.ACCOUNT_SETTINGS)}
          element={<AccountSettings linkedAccountsEnabled={config.linkedAccountsEnabled} />}
        />
        <Route path={getRouteSubPath(RoutePath.ONBOARDING)} element={<Onboarding />} />
        <Route
          path={getRouteSubPath(RoutePath.EXPERIMENTAL)}
          element={
            <OwnerAccessOrganizationPageWrapper>
              <Experimental />
            </OwnerAccessOrganizationPageWrapper>
          }
        />
        <Route path={getRouteSubPath(RoutePath.PLAYGROUND)} element={<Playground />} />
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

// In Lite version, all feature flags are enabled by default
function RequiredFeatureFlagWrapper({ children }: { children: React.ReactNode; flagKey: FeatureFlags }) {
  return children
}

export default App
