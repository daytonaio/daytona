/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useEffect, useMemo, useRef } from 'react'

import { useFilePreviewQuery } from './queries'
import type { PreviewState, SandboxFileSystemNode, SandboxInstance } from './types'

const MAX_PREVIEW_BYTES = 10 * 1024 * 1024

export function usePreviewState({
  preservePreviousPreview = true,
  sandboxInstance,
  selectedNode,
  selectedNodePath,
}: {
  preservePreviousPreview?: boolean
  sandboxInstance: SandboxInstance | undefined
  selectedNode: SandboxFileSystemNode | null
  selectedNodePath: string | null
}) {
  const previousSelectionWasFileRef = useRef(false)
  const previousFileRef = useRef<SandboxFileSystemNode | null>(null)

  const previewQuery = useFilePreviewQuery({
    enabled: Boolean(sandboxInstance && selectedNode && !selectedNode.isDir && selectedNode.size <= MAX_PREVIEW_BYTES),
    path: selectedNodePath ?? '/',
    sandboxInstance,
  })
  const shouldUsePreviousPreview =
    preservePreviousPreview && previousSelectionWasFileRef.current && Boolean(selectedNode && !selectedNode.isDir)

  useEffect(() => {
    previousSelectionWasFileRef.current = Boolean(selectedNode && !selectedNode.isDir)
    previousFileRef.current = selectedNode && !selectedNode.isDir ? selectedNode : null
  }, [selectedNode?.isDir, selectedNode?.path])

  const previewState = useMemo<PreviewState>(() => {
    if (!selectedNodePath) {
      return { status: 'idle' }
    }

    if (!selectedNode) {
      return { status: 'loading', path: selectedNodePath }
    }

    if (selectedNode.isDir) {
      return { status: 'directory' }
    }

    if (selectedNode.size > MAX_PREVIEW_BYTES) {
      return { status: 'too-large', path: selectedNode.path, size: selectedNode.size }
    }

    if (previewQuery.isError) {
      return { status: 'error', path: selectedNode.path }
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
    selectedNode,
    selectedNodePath,
    preservePreviousPreview,
    shouldUsePreviousPreview,
  ])

  return {
    previewQuery,
    previewState,
  }
}
