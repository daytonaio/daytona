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
