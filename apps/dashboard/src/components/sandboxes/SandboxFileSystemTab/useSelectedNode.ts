/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ROOT_PATH } from './constants'
import { useFileSystemStore } from './fileSystemStore'
import { useFileDetailsQuery } from './queries'
import type { SandboxInstance } from './types'

export function useSelectedNode({ sandboxInstance }: { sandboxInstance: SandboxInstance | undefined }) {
  const selectedNodePath = useFileSystemStore((state) => state.selectedNodePath)

  const selectedNodeQuery = useFileDetailsQuery({
    enabled: Boolean(sandboxInstance && selectedNodePath),
    path: selectedNodePath ?? ROOT_PATH,
    sandboxInstance,
  })

  return {
    selectedNode: selectedNodePath ? (selectedNodeQuery.data ?? null) : null,
    selectedNodePath,
    selectedNodeQuery,
  }
}
