/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQuery } from '@tanstack/react-query'
import { useEffect, useMemo, useRef } from 'react'

import { ROOT_PATH } from './constants'
import { getDirectoryChildrenQueryOptions, useFilePreviewQuery } from './queries'
import type { PreviewState, SandboxFileSystemNode, SandboxInstance } from './types'
import { getFileSystemError, isFileReadFailedError, isForbiddenFileSystemError } from './utils'

const MAX_PREVIEW_BYTES = 10 * 1024 * 1024

export function usePreviewState({
  preservePreviousPreview = true,
  sandboxInstance,
  selectedNode,
  selectedNodeError,
  selectedNodePath,
}: {
  preservePreviousPreview?: boolean
  sandboxInstance: SandboxInstance | undefined
  selectedNode: SandboxFileSystemNode | null
  selectedNodeError?: unknown
  selectedNodePath: string | null
}) {
  const previousSelectionWasFileRef = useRef(false)
  const previousFileRef = useRef<SandboxFileSystemNode | null>(null)

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
  const shouldUsePreviousPreview =
    preservePreviousPreview && previousSelectionWasFileRef.current && Boolean(selectedNode && !selectedNode.isDir)

  useEffect(() => {
    previousSelectionWasFileRef.current = Boolean(selectedNode && !selectedNode.isDir)
    previousFileRef.current = selectedNode && !selectedNode.isDir ? selectedNode : null
  }, [selectedNode?.isDir, selectedNode?.path])

  const previewState = useMemo<PreviewState>(() => {
    const selectedNodeErrorMessage = getFileSystemError(selectedNodeError)?.message
    const previewErrorMessage = getFileSystemError(previewQuery.error)?.message
    const directoryErrorMessage = getFileSystemError(selectedDirectoryQuery.error)?.message

    if (!selectedNodePath) {
      return { status: 'idle' }
    }

    if (selectedNodeError) {
      return {
        status: 'error',
        path: selectedNodePath,
        title: isForbiddenFileSystemError(selectedNodeError) ? 'Access denied' : 'Failed to load item',
        description:
          selectedNodeErrorMessage ??
          (isForbiddenFileSystemError(selectedNodeError)
            ? 'You do not have permission to access this location in the sandbox.'
            : 'Something went wrong while loading this item from the sandbox.'),
        canRetry: !isForbiddenFileSystemError(selectedNodeError),
      }
    }

    if (!selectedNode) {
      return { status: 'loading', path: selectedNodePath }
    }

    if (selectedNode.isDir) {
      if (selectedDirectoryQuery.isError) {
        return {
          status: 'error',
          path: selectedNode.path,
          title: isForbiddenFileSystemError(selectedDirectoryQuery.error)
            ? 'Access denied'
            : 'Failed to open directory',
          description:
            directoryErrorMessage ??
            (isForbiddenFileSystemError(selectedDirectoryQuery.error)
              ? 'You do not have permission to access this directory in the sandbox.'
              : 'Something went wrong while opening this directory from the sandbox.'),
          canRetry: !isForbiddenFileSystemError(selectedDirectoryQuery.error),
        }
      }

      return { status: 'directory' }
    }

    if (selectedNode.size > MAX_PREVIEW_BYTES) {
      return { status: 'too-large', path: selectedNode.path, size: selectedNode.size }
    }

    if (previewQuery.isError) {
      return {
        status: 'error',
        path: selectedNode.path,
        title: isForbiddenFileSystemError(previewQuery.error)
          ? 'Access denied'
          : isFileReadFailedError(previewQuery.error)
            ? 'Failed to read file'
            : 'Failed to load preview',
        description:
          previewErrorMessage ??
          (isForbiddenFileSystemError(previewQuery.error)
            ? 'You do not have permission to access this file in the sandbox.'
            : 'Something went wrong while reading this file from the sandbox.'),
        canRetry: !isForbiddenFileSystemError(previewQuery.error),
      }
    }

    if (previewQuery.isPending && !previewQuery.data) {
      return { status: 'loading', path: selectedNode.path }
    }

    if (previewQuery.isFetching && previewQuery.isPlaceholderData && !shouldUsePreviousPreview) {
      return { status: 'loading', path: selectedNode.path }
    }

    if (shouldUsePreviousPreview && previewQuery.isFetching && previewQuery.isPlaceholderData && previewQuery.data) {
      return {
        status: 'loading',
        path: selectedNode.path,
        previousContent: previewQuery.data.content,
        previousImageUrl: previewQuery.data.imageUrl,
        previousKind: previewQuery.data.kind,
        previousPath: previousFileRef.current?.path,
        previousSize: previousFileRef.current?.size,
      }
    }

    if (previewQuery.data) {
      return {
        status: 'ready',
        content: previewQuery.data.content,
        imageUrl: previewQuery.data.imageUrl,
        kind: previewQuery.data.kind,
        path: selectedNode.path,
      }
    }

    return { status: 'idle' }
  }, [
    previewQuery.data,
    previewQuery.isError,
    previewQuery.isFetching,
    previewQuery.isPending,
    previewQuery.isPlaceholderData,
    selectedDirectoryQuery.error,
    selectedDirectoryQuery.isError,
    selectedNode,
    selectedNodeError,
    selectedNodePath,
    preservePreviousPreview,
    shouldUsePreviousPreview,
  ])

  return {
    previewQuery,
    previewState,
  }
}
