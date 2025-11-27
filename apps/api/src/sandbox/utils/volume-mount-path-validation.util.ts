/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxVolume } from '../dto/sandbox.dto'

/**
 * Validates mount paths for sandbox volumes to ensure they are safe and valid
 * @param volumes - Array of SandboxVolume objects to validate
 * @throws Error with descriptive message if any mount path is invalid
 */
export function validateMountPaths(volumes: SandboxVolume[]): void {
  const errors: string[] = []

  for (const volume of volumes) {
    const value = volume.mountPath

    if (typeof value !== 'string') {
      errors.push(`Invalid mount path ${value} (must be a string)`)
      continue
    }

    if (!value.startsWith('/')) {
      errors.push(`Invalid mount path ${value} (must be absolute)`)
      continue
    }

    if (value === '/' || value === '//') {
      errors.push(`Invalid mount path ${value} (cannot mount to the root directory)`)
      continue
    }

    if (value.includes('/../') || value.includes('/./') || value.endsWith('/..') || value.endsWith('/.')) {
      errors.push(`Invalid mount path ${value} (cannot contain relative path components)`)
      continue
    }

    if (/\/\/+/.test(value.slice(1))) {
      errors.push(`Invalid mount path ${value} (cannot contain consecutive slashes)`)
      continue
    }

    const invalidPaths = ['/proc', '/sys', '/dev', '/boot', '/etc', '/bin', '/sbin', '/lib', '/lib64']
    const matchedInvalid = invalidPaths.find((invalid) => value === invalid || value.startsWith(invalid + '/'))
    if (matchedInvalid) {
      errors.push(`Invalid mount path ${value} (cannot mount to system directory)`)
    }
  }

  if (errors.length > 0) {
    throw new Error(errors.join(', '))
  }
}

/**
 * Validates subpaths for sandbox volumes to ensure they are safe S3 key prefixes
 * @param volumes - Array of SandboxVolume objects to validate
 * @throws Error with descriptive message if any subpath is invalid
 */
export function validateSubpaths(volumes: SandboxVolume[]): void {
  const errors: string[] = []

  for (const volume of volumes) {
    const subpath = volume.subpath

    // Empty/undefined subpath is valid (means mount entire volume)
    if (!subpath) {
      continue
    }

    if (typeof subpath !== 'string') {
      errors.push(`Invalid subpath ${subpath} (must be a string)`)
      continue
    }

    // S3 keys should not start with /
    if (subpath.startsWith('/')) {
      errors.push(`Invalid subpath "${subpath}" (S3 key prefixes cannot start with /)`)
      continue
    }

    // Prevent path traversal
    if (subpath.includes('..')) {
      errors.push(`Invalid subpath "${subpath}" (cannot contain .. for security)`)
      continue
    }

    // No consecutive slashes
    if (subpath.includes('//')) {
      errors.push(`Invalid subpath "${subpath}" (cannot contain consecutive slashes)`)
      continue
    }
  }

  if (errors.length > 0) {
    throw new Error(errors.join(', '))
  }
}
