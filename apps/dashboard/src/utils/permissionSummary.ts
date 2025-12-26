/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiKeyListPermissionsEnum } from '@daytonaio/api-client'

export interface PermissionSummary {
  type: 'full' | 'readonly' | 'custom'
  label: string
  variant: 'default' | 'secondary' | 'outline'
  categories: PermissionCategory[]
}

export interface PermissionCategory {
  name: string
  permissions: string[]
  level: 'read' | 'write' | 'delete' | 'full'
}

// All available permissions for comparison
const ALL_PERMISSIONS = Object.values(ApiKeyListPermissionsEnum) as string[]

// Permission categories and their mappings
const PERMISSION_CATEGORIES = {
  sandboxes: {
    name: 'Sandboxes',
    permissions: [ApiKeyListPermissionsEnum.WRITE_SANDBOXES, ApiKeyListPermissionsEnum.DELETE_SANDBOXES] as string[],
  },
  images: {
    name: 'Images',
    permissions: [ApiKeyListPermissionsEnum.WRITE_IMAGES, ApiKeyListPermissionsEnum.DELETE_IMAGES] as string[],
  },
  registries: {
    name: 'Registries',
    permissions: [ApiKeyListPermissionsEnum.WRITE_REGISTRIES, ApiKeyListPermissionsEnum.DELETE_REGISTRIES] as string[],
  },
  volumes: {
    name: 'Volumes',
    permissions: [
      ApiKeyListPermissionsEnum.READ_VOLUMES,
      ApiKeyListPermissionsEnum.WRITE_VOLUMES,
      ApiKeyListPermissionsEnum.DELETE_VOLUMES,
    ] as string[],
  },
}

/**
 * Determine the permission level for a category based on granted permissions
 */
function getCategoryLevel(
  categoryPermissions: string[],
  grantedPermissions: string[],
): 'read' | 'write' | 'delete' | 'full' | null {
  const granted = grantedPermissions.filter((p) => categoryPermissions.includes(p))

  if (granted.length === 0) return null
  if (granted.length === categoryPermissions.length) return 'full'

  // Check specific patterns
  const hasRead = granted.some((p) => p.includes('read:'))
  const hasWrite = granted.some((p) => p.includes('write:'))
  const hasDelete = granted.some((p) => p.includes('delete:'))

  // For resources with read permissions (like volumes)
  if (hasRead && !hasWrite && !hasDelete) return 'read'

  // For delete permissions (highest level)
  if (hasDelete && hasWrite) return 'full'
  if (hasDelete) return 'delete'

  // For write permissions
  if (hasWrite) return 'write'

  // Fallback
  return hasRead ? 'read' : 'write'
}

/**
 * Summarize API key permissions into a more readable format
 */
export function summarizePermissions(permissions: string[]): PermissionSummary {
  // Check if user has all permissions (full access)
  const hasAllPermissions = ALL_PERMISSIONS.every((permission) => permissions.includes(permission))

  if (hasAllPermissions) {
    return {
      type: 'full',
      label: 'Full Access',
      variant: 'default',
      categories: [],
    }
  }

  // Check if user has only read permissions (only applies to volumes currently)
  const hasOnlyReadPermissions =
    permissions.length === 1 && permissions.includes(ApiKeyListPermissionsEnum.READ_VOLUMES)

  if (hasOnlyReadPermissions) {
    return {
      type: 'readonly',
      label: 'Read-only',
      variant: 'secondary',
      categories: [],
    }
  }

  // Categorize permissions
  const categories: PermissionCategory[] = []

  Object.entries(PERMISSION_CATEGORIES).forEach(([key, category]) => {
    const level = getCategoryLevel(category.permissions, permissions)
    if (level) {
      categories.push({
        name: category.name,
        permissions: permissions.filter((p) => category.permissions.includes(p)),
        level,
      })
    }
  })

  return {
    type: 'custom',
    label: 'Custom',
    variant: 'outline',
    categories,
  }
}

/**
 * Get a display-friendly label for a permission category level
 */
export function getCategoryLevelLabel(level: 'read' | 'write' | 'delete' | 'full'): string {
  switch (level) {
    case 'read':
      return 'Read'
    case 'write':
      return 'Write'
    case 'delete':
      return 'Delete'
    case 'full':
      return 'Full'
    default:
      return level
  }
}

/**
 * Get a color variant for a permission category level
 */
export function getCategoryLevelVariant(
  level: 'read' | 'write' | 'delete' | 'full',
): 'default' | 'secondary' | 'outline' {
  switch (level) {
    case 'read':
      return 'secondary'
    case 'write':
      return 'outline'
    case 'delete':
    case 'full':
      return 'default'
    default:
      return 'outline'
  }
}
