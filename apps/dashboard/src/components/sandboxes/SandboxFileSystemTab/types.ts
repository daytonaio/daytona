/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { Daytona } from '@daytona/sdk'

export type PreviewKind = 'binary' | 'image' | 'text'

export type SandboxInstance = Awaited<ReturnType<Daytona['get']>>

export type SandboxFileSystemNode = {
  group: string
  id: string
  isDir: boolean
  modTime: string
  mode: string
  name: string
  owner: string
  path: string
  permissions: string
  size: number
}

export type PreviewState =
  | { status: 'directory' }
  | { status: 'error'; path: string }
  | { status: 'idle' }
  | {
      status: 'loading'
      path: string
      previousContent?: string
      previousImageUrl?: string
      previousKind?: PreviewKind
      previousPath?: string
      previousSize?: number
    }
  | { status: 'ready'; content?: string; imageUrl?: string; kind: PreviewKind; path: string }
  | { status: 'too-large'; path: string; size: number }
