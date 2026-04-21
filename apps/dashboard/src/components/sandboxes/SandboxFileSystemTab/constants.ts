/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { SandboxFileSystemNode } from './types'

export const ROOT_PATH = '/'
export const FILE_SEARCH_MIN_CHARS = 3

export const ROOT_NODE: SandboxFileSystemNode = {
  group: 'root',
  id: ROOT_PATH,
  isDir: true,
  modTime: '',
  mode: '',
  name: ROOT_PATH,
  owner: 'root',
  path: ROOT_PATH,
  permissions: '',
  size: 0,
}
