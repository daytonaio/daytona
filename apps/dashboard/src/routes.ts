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

const ROUTE_LAZY_LOAD_TIMEOUT_MS = 8_000

function loadRouteModule<TComponent extends ComponentType<any>>(
  routeName: string,
  loadModule: () => Promise<PageModule<TComponent>>,
): Promise<PageModule<TComponent>> {
  let timeout: ReturnType<typeof setTimeout> | undefined

  return Promise.race([
    loadModule(),
    new Promise<never>((_, reject) => {
      timeout = setTimeout(() => {
        reject(new Error(`Timed out loading route "${routeName}" after ${ROUTE_LAZY_LOAD_TIMEOUT_MS}ms`))
      }, ROUTE_LAZY_LOAD_TIMEOUT_MS)
    }),
  ]).finally(() => {
    if (timeout) {
      clearTimeout(timeout)
    }
  })
}

function createRouteLazy<TComponent extends ComponentType<any>>(
  routeName: string,
  loadModule: () => Promise<PageModule<TComponent>>,
): LazyRoute {
  let routePromise: ReturnType<LazyRoute> | null = null

  return () => {
    if (!routePromise) {
      const startedAt = Date.now()

      routePromise = loadRouteModule(routeName, loadModule)
        .then((module) => {
          console.info('[dashboard] Route lazy module loaded', {
            routeName,
            durationMs: Date.now() - startedAt,
          })

          return { Component: module.default }
        })
        .catch((error) => {
          console.error('[dashboard] Route lazy module failed to load', { routeName, error })
          routePromise = null
          throw error
        })
    }

    return routePromise
  }
}

export const lazyRoutes = {
  AccountSettings: createRouteLazy('AccountSettings', () => import('@/pages/AccountSettings')),
  AuditLogs: createRouteLazy('AuditLogs', () => import('@/pages/AuditLogs')),
  EmailVerify: createRouteLazy('EmailVerify', () => import('@/pages/EmailVerify')),
  Keys: createRouteLazy('Keys', () => import('@/pages/Keys')),
  Limits: createRouteLazy('Limits', () => import('@/pages/Limits')),
  Onboarding: createRouteLazy('Onboarding', () => import('@/pages/Onboarding')),
  OrganizationMembers: createRouteLazy('OrganizationMembers', () => import('@/pages/OrganizationMembers')),
  OrganizationSettings: createRouteLazy('OrganizationSettings', () => import('@/pages/OrganizationSettings')),
  Playground: createRouteLazy('Playground', () => import('@/pages/Playground')),
  Regions: createRouteLazy('Regions', () => import('@/pages/Regions')),
  Registries: createRouteLazy('Registries', () => import('@/pages/Registries')),
  Runners: createRouteLazy('Runners', () => import('@/pages/Runners')),
  SandboxDetails: createRouteLazy('SandboxDetails', () => import('@/components/sandboxes/SandboxDetails')),
  Sandboxes: createRouteLazy('Sandboxes', () => import('@/pages/Sandboxes')),
  Snapshots: createRouteLazy('Snapshots', () => import('@/pages/Snapshots')),
  Spending: createRouteLazy('Spending', () => import('@/pages/Spending')),
  UserOrganizationInvitations: createRouteLazy(
    'UserOrganizationInvitations',
    () => import('@/pages/UserOrganizationInvitations'),
  ),
  Volumes: createRouteLazy('Volumes', () => import('@/pages/Volumes')),
  Wallet: createRouteLazy('Wallet', () => import('@/pages/Wallet')),
  WebhookEndpointDetails: createRouteLazy('WebhookEndpointDetails', () => import('@/pages/WebhookEndpointDetails')),
  Webhooks: createRouteLazy('Webhooks', () => import('@/pages/Webhooks')),
}
