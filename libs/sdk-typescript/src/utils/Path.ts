/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import * as _path from 'path'

export function prefixRelativePath(prefix: string, path?: string): string {
  let result = prefix

  if (path) {
    path = path.trim()
    if (path === '~') {
      result = prefix
    } else if (path.startsWith('~/')) {
      result = _path.join(prefix, path.slice(2))
    } else if (_path.isAbsolute(path)) {
      result = path
    } else {
      result = _path.join(prefix, path)
    }
  }

  return result
}
