/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { FileTextIcon, RefreshCwIcon, UploadIcon } from 'lucide-react'
import { useMemo, useRef } from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'

import CodeBlock from '@/components/CodeBlock'
import { Button } from '@/components/ui/button'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { FileUpload, FileUploadDropzone } from '@/components/ui/file-upload'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'

import { LARGE_TEXT_VIRTUALIZATION_THRESHOLD } from './constants'
import { useFileSystemStore } from './fileSystemStore'
import { MarkdownPreview } from './MarkdownPreview'
import type { PreviewKind, PreviewState } from './types'
import { formatBytes, getCodeLanguage, getImageMimeType } from './utils'

function VirtualizedTextPreview({ content, isRefreshing }: { content: string; isRefreshing: boolean }) {
  const scrollRef = useRef<HTMLDivElement>(null)
  const lines = useMemo(() => content.split('\n'), [content])
  const rowVirtualizer = useVirtualizer({
    count: lines.length,
    getScrollElement: () => scrollRef.current,
    estimateSize: () => 24,
    overscan: 20,
  })

  return (
    <div
      ref={scrollRef}
      className={cn(
        'scrollbar-sm h-full min-h-0 overflow-auto rounded-md border border-border bg-muted/20 text-sm leading-6 transition-opacity contain-layout contain-paint font-mono',
        {
          'opacity-50': isRefreshing,
        },
      )}
    >
      <div className="relative w-max min-w-full px-4 py-4" style={{ height: `${rowVirtualizer.getTotalSize()}px` }}>
        {rowVirtualizer.getVirtualItems().map((virtualRow) => (
          <div
            key={virtualRow.key}
            className="absolute left-0 top-0 w-full whitespace-pre"
            style={{ transform: `translateY(${virtualRow.start}px)` }}
          >
            {lines[virtualRow.index] || ' '}
          </div>
        ))}
      </div>
    </div>
  )
}

function FilePreviewSkeleton({ kind }: { kind: PreviewKind | null }) {
  if (kind === 'image') {
    return <Skeleton className="h-full min-h-0 w-full rounded-md" />
  }

  return (
    <div className="space-y-3 rounded-md border border-border p-4">
      <Skeleton className="h-4 w-40 rounded-sm" />
      <Skeleton className="h-4 w-56 rounded-sm" />
      <Skeleton className="h-4 w-48 rounded-sm" />
      <Skeleton className="h-4 w-full rounded-sm" />
      <Skeleton className="h-4 w-[92%] rounded-sm" />
      <Skeleton className="h-4 w-[84%] rounded-sm" />
      <Skeleton className="h-4 w-[96%] rounded-sm" />
      <Skeleton className="h-4 w-[78%] rounded-sm" />
      <Skeleton className="h-4 w-[88%] rounded-sm" />
      <Skeleton className="h-4 w-[70%] rounded-sm" />
    </div>
  )
}

export function FilePreview({
  isWrapEnabled,
  onRetry,
  onUploadFiles,
  previewState,
}: {
  isWrapEnabled: boolean
  onRetry: () => void
  onUploadFiles: (files: File[]) => void
  previewState: PreviewState
}) {
  const selectedNode = useFileSystemStore((state) => state.selectedNode)

  if (previewState.status === 'idle') {
    return (
      <Empty className="min-h-[280px] border-0">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <FileTextIcon className="size-4" />
          </EmptyMedia>
          <EmptyTitle>Select a file</EmptyTitle>
          <EmptyDescription>Choose a file from the tree to preview its contents.</EmptyDescription>
        </EmptyHeader>
      </Empty>
    )
  }

  if (!selectedNode) {
    return null
  }

  const previewPath =
    previewState.status === 'loading' ? (previewState.previousPath ?? selectedNode.path) : selectedNode.path
  const previewSize =
    previewState.status === 'loading' ? (previewState.previousSize ?? selectedNode.size) : selectedNode.size
  const codeLanguage = getCodeLanguage(previewPath)
  const imageMimeType = getImageMimeType(previewPath)
  const loadingPreviewKind =
    previewState.status === 'loading' ? (previewState.previousKind ?? (imageMimeType ? 'image' : 'text')) : null

  if (previewState.status === 'directory') {
    return (
      <FileUpload className="flex h-full min-h-0 flex-1" multiple onFilesSelected={onUploadFiles}>
        <FileUploadDropzone className="flex min-h-0 flex-1 justify-center px-6 py-8 text-center">
          <Empty variant="neutral" className="pointer-events-none h-full border-0 bg-transparent p-0">
            <EmptyHeader className="gap-1">
              <EmptyMedia variant="icon">
                <UploadIcon className="size-4" />
              </EmptyMedia>
              <EmptyTitle>Click to upload or drop files</EmptyTitle>
              <EmptyDescription className="text-xs">{selectedNode.path}</EmptyDescription>
            </EmptyHeader>
          </Empty>
        </FileUploadDropzone>
      </FileUpload>
    )
  }

  if (
    previewState.status === 'loading' &&
    !previewState.previousContent &&
    !previewState.previousImageUrl &&
    !previewState.previousKind
  ) {
    return <FilePreviewSkeleton kind={loadingPreviewKind} />
  }

  if (previewState.status === 'error') {
    return (
      <Empty className="h-full min-h-[220px] rounded-md border border-dashed">
        <EmptyHeader>
          <EmptyTitle>Failed to read file</EmptyTitle>
          <EmptyDescription>Something went wrong while reading this file from the sandbox.</EmptyDescription>
        </EmptyHeader>
        <Button variant="outline" size="sm" onClick={onRetry}>
          <RefreshCwIcon className="size-4" />
          Retry
        </Button>
      </Empty>
    )
  }

  if (previewState.status === 'too-large') {
    return (
      <Empty className="h-full min-h-[220px] rounded-md border border-dashed">
        <EmptyHeader>
          <EmptyTitle>Preview skipped</EmptyTitle>
          <EmptyDescription>
            This file is {formatBytes(previewState.size)}. Preview is only available for files up to 10 MB.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    )
  }

  const activeKind = previewState.status === 'ready' ? previewState.kind : previewState.previousKind
  const content = previewState.status === 'ready' ? (previewState.content ?? '') : (previewState.previousContent ?? '')
  const imageUrl = previewState.status === 'ready' ? previewState.imageUrl : previewState.previousImageUrl
  const isRefreshing = previewState.status === 'loading'
  const isShowingPreviousLargeText =
    previewState.status === 'loading' && previewSize >= LARGE_TEXT_VIRTUALIZATION_THRESHOLD

  if (!activeKind) {
    return null
  }

  if (activeKind === 'binary') {
    return (
      <Empty className="h-full min-h-[220px] rounded-md border border-dashed">
        <EmptyHeader>
          <EmptyTitle>Binary file</EmptyTitle>
          <EmptyDescription>This file looks binary, so a text preview is not shown.</EmptyDescription>
        </EmptyHeader>
      </Empty>
    )
  }

  if (activeKind === 'image') {
    return (
      <div
        className={cn(
          'flex h-full min-h-0 items-center justify-center overflow-auto rounded-md border border-border bg-muted/20 p-4 transition-opacity',
          {
            'opacity-50': isRefreshing,
          },
        )}
      >
        <img
          src={imageUrl}
          alt={selectedNode.name || selectedNode.path}
          className="mx-auto max-h-[60vh] max-w-full rounded-md object-contain"
        />
      </div>
    )
  }

  if (codeLanguage === 'markdown') {
    return <MarkdownPreview content={content} isLoading={isRefreshing} />
  }

  if (codeLanguage) {
    return (
      <CodeBlock
        code={content}
        language={codeLanguage}
        showCopy={false}
        className={cn('h-full min-h-0 border border-border transition-opacity', {
          'opacity-50': isRefreshing,
        })}
        codeAreaClassName="scrollbar-sm h-full min-h-0 overflow-auto rounded-lg text-sm leading-6"
      />
    )
  }

  const shouldVirtualizeLargeText =
    !codeLanguage &&
    previewSize >= LARGE_TEXT_VIRTUALIZATION_THRESHOLD &&
    (isShowingPreviousLargeText || !isWrapEnabled)

  if (shouldVirtualizeLargeText) {
    return <VirtualizedTextPreview content={content} isRefreshing={isRefreshing} />
  }

  return (
    <pre
      className={cn(
        'scrollbar-sm h-full min-h-0 overflow-auto rounded-md border border-border bg-muted/20 p-4 text-sm leading-6 transition-opacity contain-layout contain-paint font-mono',
        {
          'opacity-50': isRefreshing,
          'whitespace-pre': !isWrapEnabled,
          'whitespace-pre-wrap break-words': isWrapEnabled,
        },
      )}
    >
      {content}
    </pre>
  )
}
