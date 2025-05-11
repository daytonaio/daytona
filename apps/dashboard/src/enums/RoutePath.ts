/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * Enum for all route paths in the application
 * Use this for consistent route references across the app
 */
export enum RoutePath {
  // Main routes
  LANDING = '/',
  LOGOUT = '/logout',
  DASHBOARD = '/dashboard',
  DOCS = '/docs',
  SLACK = '/slack',

  // Dashboard sub-routes
  KEYS = '/dashboard/keys',
  SANDBOXES = '/dashboard/sandboxes',
  IMAGES = '/dashboard/images',
  REGISTRIES = '/dashboard/registries',
  USAGE = '/dashboard/usage',
  BILLING = '/dashboard/billing',
  MEMBERS = '/dashboard/members',
  ROLES = '/dashboard/roles',
  SETTINGS = '/dashboard/settings',
  ONBOARDING = '/dashboard/onboarding',

  // User routes
  USER_INVITATIONS = '/dashboard/user/invitations',
}

/**
 * Returns only the path segment for dashboard sub-routes
 * Useful for nested routes under the dashboard
 */
export const getRouteSubPath = (path: RoutePath): string => {
  return path.replace('/dashboard/', '')
}
