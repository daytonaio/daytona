/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationRolePermissionsEnum } from '@daytonaio/api-client'

export const ORGANIZATION_ROLE_PERMISSIONS_GROUPS: { name: string; permissions: OrganizationRolePermissionsEnum[] }[] =
  [
    {
      name: 'Sandboxes',
      permissions: [OrganizationRolePermissionsEnum.WRITE_SANDBOXES, OrganizationRolePermissionsEnum.DELETE_SANDBOXES],
    },
    {
      name: 'Snapshots',
      permissions: [OrganizationRolePermissionsEnum.WRITE_SNAPSHOTS, OrganizationRolePermissionsEnum.DELETE_SNAPSHOTS],
    },
    {
      name: 'Registries',
      permissions: [
        OrganizationRolePermissionsEnum.WRITE_REGISTRIES,
        OrganizationRolePermissionsEnum.DELETE_REGISTRIES,
      ],
    },
    {
      name: 'Volumes',
      permissions: [
        OrganizationRolePermissionsEnum.READ_VOLUMES,
        OrganizationRolePermissionsEnum.WRITE_VOLUMES,
        OrganizationRolePermissionsEnum.DELETE_VOLUMES,
      ],
    },
  ]
