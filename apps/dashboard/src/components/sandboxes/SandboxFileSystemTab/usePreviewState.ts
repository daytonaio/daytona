/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery, useQueryClient } from '@tanstack/react-query'

import { ROOT_PATH } from './constants'
import { fileSystemQueryKeys, getDirectoryChildrenQueryOptions, useFilePreviewQuery } from './queries'
import type { PreviewKind, SandboxFileSystemNode, SandboxInstance } from './types'
import { getFileSystemError, isFileReadFailedError, isForbiddenFileSystemError } from './utils'

const MAX_PREVIEW_BYTES = 10 * 1024 * 1024

export type PreviewState =
  | { status: 'directory' }
  | { status: 'error'; path: string; title: string; description: string; canRetry?: boolean }
  | { status: 'idle' }
  | { status: 'loading'; path: string }
  | { status: 'ready'; content?: string; imageBlob?: Blob; kind: PreviewKind; path: string }
  | { status: 'too-large'; path: string; size: number }

function createPreviewErrorState({
  accessDeniedDescription,
  error,
  fallbackDescription,
  fallbackTitle,
  path,
}: {
  accessDeniedDescription: string
  error: unknown
  fallbackDescription: string
  fallbackTitle: string
  path: string
}): PreviewState {
  const errorMessage = getFileSystemError(error)?.message
  const isForbidden = isForbiddenFileSystemError(error)

  return {
    status: 'error',
    path,
    title: isForbidden ? 'Access denied' : fallbackTitle,
    description: errorMessage ?? (isForbidden ? accessDeniedDescription : fallbackDescription),
    canRetry: !isForbidden,
  }
}

export function usePreviewState({
  refetchSelectedNode,
  sandboxInstance,
  selectedNode,
  selectedNodeError,
  selectedNodePath,
}: {
  refetchSelectedNode: () => void | Promise<unknown>
  sandboxInstance: SandboxInstance | undefined
  selectedNode: SandboxFileSystemNode | null
  selectedNodeError?: unknown
  selectedNodePath: string | null
}) {
  const queryClient = useQueryClient()

  const previewQuery = useFilePreviewQuery({
    enabled: Boolean(sandboxInstance && selectedNode && !selectedNode.isDir && selectedNode.size <= MAX_PREVIEW_BYTES),
    notifyOnError: false,
    path: selectedNodePath ?? '/',
    sandboxInstance,
  })
  const selectedDirectoryQuery = useQuery({
    ...(sandboxInstance && selectedNode?.isDir
      ? getDirectoryChildrenQueryOptions({
          notifyOnError: false,
          path: selectedNode.path,
          sandboxInstance,
        })
      : {
          queryKey: ['sandbox-file-system', 'unknown', 'directory', selectedNodePath ?? ROOT_PATH],
          queryFn: async () => [] as SandboxFileSystemNode[],
        }),
    enabled: Boolean(sandboxInstance && selectedNode?.isDir),
  })

  let previewState: PreviewState

  if (!selectedNodePath) {
    previewState = { status: 'idle' }
  } else if (selectedNodeError) {
    previewState = createPreviewErrorState({
      accessDeniedDescription: 'You do not have permission to access this location in the sandbox.',
      error: selectedNodeError,
      fallbackDescription: 'Something went wrong while loading this item from the sandbox.',
      fallbackTitle: 'Failed to load item',
      path: selectedNodePath,
    })
  } else if (!selectedNode) {
    previewState = { status: 'loading', path: selectedNodePath }
  } else if (selectedNode.isDir) {
    const cachedDirectoryData =
      sandboxInstance && selectedNode.path
        ? queryClient.getQueryData<SandboxFileSystemNode[]>(
            fileSystemQueryKeys.directory(sandboxInstance.id, selectedNode.path),
          )
        : undefined
    const hasDirectoryData = selectedDirectoryQuery.data !== undefined || cachedDirectoryData !== undefined

    if (selectedDirectoryQuery.isPending && !hasDirectoryData) {
      previewState = { status: 'loading', path: selectedNode.path }
    } else if (selectedDirectoryQuery.isError) {
      previewState = createPreviewErrorState({
        accessDeniedDescription: 'You do not have permission to access this directory in the sandbox.',
        error: selectedDirectoryQuery.error,
        fallbackDescription: 'Something went wrong while opening this directory from the sandbox.',
        fallbackTitle: 'Failed to open directory',
        path: selectedNode.path,
      })
    } else {
      previewState = { status: 'directory' }
    }
  } else if (selectedNode.size > MAX_PREVIEW_BYTES) {
    previewState = { status: 'too-large', path: selectedNode.path, size: selectedNode.size }
  } else if (previewQuery.isError) {
    previewState = createPreviewErrorState({
      accessDeniedDescription: 'You do not have permission to access this file in the sandbox.',
      error: previewQuery.error,
      fallbackDescription: 'Something went wrong while reading this file from the sandbox.',
      fallbackTitle: isFileReadFailedError(previewQuery.error) ? 'Failed to read file' : 'Failed to load preview',
      path: selectedNode.path,
    })
  } else if (previewQuery.isPending && !previewQuery.data) {
    previewState = { status: 'loading', path: selectedNode.path }
  } else if (previewQuery.data) {
    previewState = {
      status: 'ready',
      content: previewQuery.data.content,
      imageBlob: previewQuery.data.imageBlob,
      kind: previewQuery.data.kind,
      path: selectedNode.path,
    }
  } else {
    previewState = { status: 'idle' }
  }

  async function retryPreviewState() {
    if (selectedNodeError) {
      if (sandboxInstance && selectedNodePath) {
        await queryClient.invalidateQueries({
          queryKey: fileSystemQueryKeys.details(sandboxInstance.id, selectedNodePath),
        })
      } else {
        await refetchSelectedNode()
      }
      return
    }

    if (!selectedNode) {
      return
    }

    if (selectedNode.isDir) {
      await selectedDirectoryQuery.refetch()
      return
    }

    await previewQuery.refetch()
  }

  return {
    previewQuery,
    previewState,
    retryPreviewState,
  }
}
