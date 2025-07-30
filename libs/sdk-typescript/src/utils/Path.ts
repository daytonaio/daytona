/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import * as pathe from 'pathe'

export function prefixRelativePath(prefix: string, path?: string): string {
  let result = prefix

  if (path) {
    path = path.trim()
    if (path === '~') {
      result = prefix
    } else if (path.startsWith('~/')) {
      result = pathe.join(prefix, path.slice(2))
    } else if (pathe.isAbsolute(path)) {
      result = path
    } else {
      result = pathe.join(prefix, path)
    }
  }

  return result
}
