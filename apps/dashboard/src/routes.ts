/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { ComponentType } from 'react'
import type { LazyRouteFunction, RouteObject } from 'react-router'

type LazyRoute = LazyRouteFunction<RouteObject>

type PageModule<TComponent extends ComponentType<any> = ComponentType<any>> = {
  default: TComponent
}

function createRouteLazy<TComponent extends ComponentType<any>>(
  loadModule: () => Promise<PageModule<TComponent>>,
): LazyRoute {
  let routePromise: ReturnType<LazyRoute> | null = null

  return () => {
    routePromise ??= loadModule()
      .then((module) => ({
        Component: module.default,
      }))
      .catch((error) => {
        routePromise = null
        throw error
      })

    return routePromise
  }
}

export const lazyRoutes = {
  AccountSettings: createRouteLazy(() => import('@/pages/AccountSettings')),
  AuditLogs: createRouteLazy(() => import('@/pages/AuditLogs')),
  EmailVerify: createRouteLazy(() => import('@/pages/EmailVerify')),
  Keys: createRouteLazy(() => import('@/pages/Keys')),
  Limits: createRouteLazy(() => import('@/pages/Limits')),
  Onboarding: createRouteLazy(() => import('@/pages/Onboarding')),
  OrganizationMembers: createRouteLazy(() => import('@/pages/OrganizationMembers')),
  OrganizationSettings: createRouteLazy(() => import('@/pages/OrganizationSettings')),
  Playground: createRouteLazy(() => import('@/pages/Playground')),
  Regions: createRouteLazy(() => import('@/pages/Regions')),
  Registries: createRouteLazy(() => import('@/pages/Registries')),
  Runners: createRouteLazy(() => import('@/pages/Runners')),
  SandboxDetails: createRouteLazy(() => import('@/components/sandboxes/SandboxDetails')),
  Sandboxes: createRouteLazy(() => import('@/pages/Sandboxes')),
  Snapshots: createRouteLazy(() => import('@/pages/Snapshots')),
  Spending: createRouteLazy(() => import('@/pages/Spending')),
  UserOrganizationInvitations: createRouteLazy(() => import('@/pages/UserOrganizationInvitations')),
  Volumes: createRouteLazy(() => import('@/pages/Volumes')),
  Wallet: createRouteLazy(() => import('@/pages/Wallet')),
  WebhookEndpointDetails: createRouteLazy(() => import('@/pages/WebhookEndpointDetails')),
  Webhooks: createRouteLazy(() => import('@/pages/Webhooks')),
}
