/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationRolePermissionsEnum } from '@daytonaio/api-client'

export interface OrganizationRolePermissionGroup {
  name: string
  permissions: OrganizationRolePermissionsEnum[]
}
