/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateApiKeyPermissionsEnum } from '@daytonaio/api-client'

export const CREATE_API_KEY_PERMISSIONS_GROUPS: { name: string; permissions: CreateApiKeyPermissionsEnum[] }[] = [
  {
    name: 'Sandboxes',
    permissions: [CreateApiKeyPermissionsEnum.WRITE_SANDBOXES, CreateApiKeyPermissionsEnum.DELETE_SANDBOXES],
  },
  {
    name: 'Snapshots',
    permissions: [CreateApiKeyPermissionsEnum.WRITE_SNAPSHOTS, CreateApiKeyPermissionsEnum.DELETE_SNAPSHOTS],
  },
  {
    name: 'Registries',
    permissions: [CreateApiKeyPermissionsEnum.WRITE_REGISTRIES, CreateApiKeyPermissionsEnum.DELETE_REGISTRIES],
  },
  {
    name: 'Volumes',
    permissions: [
      CreateApiKeyPermissionsEnum.READ_VOLUMES,
      CreateApiKeyPermissionsEnum.WRITE_VOLUMES,
      CreateApiKeyPermissionsEnum.DELETE_VOLUMES,
    ],
  },
]
