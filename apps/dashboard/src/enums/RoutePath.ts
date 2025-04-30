/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export enum RoutePath {
  Dashboard = '/dashboard',
  Workspaces = '/workspaces',
  Keys = '/keys',
  Images = '/images',
  Registries = '/registries',
  Organizations = '/organizations',
  OrganizationMembers = '/organizations/:organizationId/members',
  OrganizationRoles = '/organizations/:organizationId/roles',
  Usage = '/usage',
  Billing = '/billing',
  LandingPage = '/',
  Logout = '/logout',
  Members = '/members',
  Settings = '/settings',
  Sandboxes = '/sandboxes',
  UserInvitations = '/user-invitations',
  Docs = '/docs',
}
