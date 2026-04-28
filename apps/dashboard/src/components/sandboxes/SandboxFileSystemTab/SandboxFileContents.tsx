/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useVirtualizer } from '@tanstack/react-virtual'
import { AnimatePresence, motion } from 'framer-motion'
import {
  ArrowDownIcon,
  ArrowUpIcon,
  AlertTriangleIcon,
  CheckIcon,
  CopyIcon,
  DownloadIcon,
  EllipsisIcon,
  FileTextIcon,
  RefreshCwIcon,
  TextWrapIcon,
  UploadIcon,
  XIcon,
} from 'lucide-react'
import { useEffect, useMemo, useRef, useState } from 'react'

import CodeBlock from '@/components/CodeBlock'
import { MarkdownPreview } from '@/components/MarkdownPreview'
import TooltipButton from '@/components/TooltipButton'
import { Button } from '@/components/ui/button'
import { ButtonGroup } from '@/components/ui/button-group'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { FileUpload, FileUploadDropzone } from '@/components/ui/file-upload'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { Toggle } from '@/components/ui/toggle'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { useCopyToClipboard } from '@/hooks/useCopyToClipboard'
import { cn } from '@/lib/utils'

import { ROOT_PATH } from './constants'
import { useFileSystemStore } from './fileSystemStore'
import { useIsDirectoryRefreshing, useIsFilePreviewRefreshing } from './queries'
import { PathHeaderLabel } from './searchLabels'
import type { PreviewKind, SandboxFileSystemNode, SandboxInstance } from './types'
import { usePreviewState, type PreviewState } from './usePreviewState'
import { formatBytes, getCodeLanguage, getImageMimeType, getNodeMetaLine } from './utils'

const LARGE_TEXT_WRAP_THRESHOLD = 256 * 1024
const LARGE_TEXT_VIRTUALIZATION_THRESHOLD = 512 * 1024
const MotionCopyIcon = motion(CopyIcon)
const MotionCheckIcon = motion(CheckIcon)
const copyIconMotionProps = {
  initial: { opacity: 0, y: 5 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -5 },
  transition: { duration: 0.1 },
}

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

function DirectoryContentsSkeleton() {
  return (
    <Empty className="h-full min-h-[220px] rounded-md border border-dashed">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <Spinner />
        </EmptyMedia>
        <EmptyTitle>Loading directory</EmptyTitle>
      </EmptyHeader>
    </Empty>
  )
}

function ImagePreview({
  imageBlob,
  isRefreshing,
  selectedNode,
}: {
  imageBlob: Blob | undefined
  isRefreshing: boolean
  selectedNode: SandboxFileSystemNode
}) {
  const [imageUrl, setImageUrl] = useState<string | null>(null)

  useEffect(() => {
    if (!imageBlob) {
      setImageUrl(null)
      return
    }

    const objectUrl = URL.createObjectURL(imageBlob)
    setImageUrl(objectUrl)

    return () => {
      URL.revokeObjectURL(objectUrl)
    }
  }, [imageBlob])

  if (!imageUrl) {
    return <FilePreviewSkeleton kind="image" />
  }

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

function SandboxFileContentsBody({
  isWrapEnabled,
  isRefreshing,
  onRetry,
  onUploadFiles,
  previewState,
  selectedNode,
}: {
  isWrapEnabled: boolean
  isRefreshing: boolean
  onRetry: () => void
  onUploadFiles: (files: File[]) => void
  previewState: PreviewState
  selectedNode: SandboxFileSystemNode | null
}) {
  if (previewState.status === 'idle') {
    return (
      <Empty className="h-full min-h-0 border-0">
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

  if (previewState.status === 'error') {
    return (
      <Empty className="h-full min-h-[220px] rounded-md border border-dashed">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <AlertTriangleIcon className="size-4" />
          </EmptyMedia>
          <EmptyTitle>{previewState.title}</EmptyTitle>
          <EmptyDescription>{previewState.description}</EmptyDescription>
        </EmptyHeader>
        {previewState.canRetry ? (
          <Button variant="outline" size="sm" onClick={onRetry}>
            <RefreshCwIcon className="size-4" />
            Retry
          </Button>
        ) : null}
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

  const codeLanguage = getCodeLanguage(selectedNode.path)
  const imageMimeType = getImageMimeType(selectedNode.path)

  if (previewState.status === 'loading') {
    if (selectedNode.isDir) {
      return <DirectoryContentsSkeleton />
    }

    return <FilePreviewSkeleton kind={imageMimeType ? 'image' : 'text'} />
  }

  const activeKind = previewState.kind
  const content = previewState.content ?? ''
  const imageBlob = previewState.imageBlob

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
    return <ImagePreview imageBlob={imageBlob} isRefreshing={isRefreshing} selectedNode={selectedNode} />
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

  const shouldVirtualizeLargeText = selectedNode.size >= LARGE_TEXT_VIRTUALIZATION_THRESHOLD && !isWrapEnabled

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

export function SandboxFileContents({
  onClose,
  onDelete,
  onDownload,
  onNavigateNext,
  onNavigatePrevious,
  onRefresh,
  onStartCreateFolder,
  onUploadFiles,
  sandboxInstance,
  selectedNode,
  selectedNodeError,
  selectedNodePath,
}: {
  onClose: () => void
  onDelete: (node: SandboxFileSystemNode) => void
  onDownload: () => void | Promise<void>
  onNavigateNext: () => void
  onNavigatePrevious: () => void
  onRefresh: () => void | Promise<void>
  onStartCreateFolder: (parentPath: string) => void
  onUploadFiles: (files: File[]) => void
  sandboxInstance: SandboxInstance | undefined
  selectedNode: SandboxFileSystemNode | null
  selectedNodeError?: unknown
  selectedNodePath: string | null
}) {
  const nextFilePath = useFileSystemStore((state) => state.nextFilePath)
  const previousFilePath = useFileSystemStore((state) => state.previousFilePath)
  const isSelectedDirectoryRefreshing = useIsDirectoryRefreshing({
    path: selectedNode?.isDir ? selectedNode.path : null,
    sandboxInstance,
  })
  const isSelectedFileRefreshing = useIsFilePreviewRefreshing({
    path: selectedNode && !selectedNode.isDir ? selectedNode.path : null,
    sandboxInstance,
  })
  const { previewState, retryPreviewState } = usePreviewState({
    refetchSelectedNode: onRefresh,
    sandboxInstance,
    selectedNode,
    selectedNodeError,
    selectedNodePath,
  })
  const [copiedText, copyToClipboard] = useCopyToClipboard()
  const isContentsRefreshing = isSelectedDirectoryRefreshing || isSelectedFileRefreshing
  const nodePath = selectedNode?.path
  const canCreateFolderInSelectedDirectory =
    Boolean(selectedNode?.isDir) && !(previewState.status === 'error' && !previewState.canRetry)
  const selectedCodeLanguage = nodePath ? getCodeLanguage(nodePath) : null
  const headerText = selectedNode?.path ?? 'Contents'
  const [isWrapEnabled, setIsWrapEnabled] = useState(true)
  const showWrapToggle =
    Boolean(selectedNode && nodePath && !selectedNode.isDir) &&
    previewState.status === 'ready' &&
    previewState.kind === 'text' &&
    !selectedCodeLanguage

  useEffect(() => {
    setIsWrapEnabled((selectedNode?.size ?? 0) <= LARGE_TEXT_WRAP_THRESHOLD)
  }, [selectedNode?.path, selectedNode?.size])

  return (
    <div className="flex h-full min-h-0 w-full flex-col bg-background">
      <div className="flex h-11 shrink-0 items-center border-b border-border px-3">
        <PathHeaderLabel text={headerText} className="flex-1 text-sm font-medium" />
        {selectedNode ? (
          <>
            <TooltipButton
              tooltipText="Previous file"
              variant="ghost"
              size="icon-sm"
              onClick={onNavigatePrevious}
              disabled={!previousFilePath}
            >
              <ArrowUpIcon className="size-4" />
            </TooltipButton>
            <TooltipButton
              tooltipText="Next file"
              variant="ghost"
              size="icon-sm"
              onClick={onNavigateNext}
              disabled={!nextFilePath}
            >
              <ArrowDownIcon className="size-4" />
            </TooltipButton>
            <TooltipButton tooltipText="Close contents" variant="ghost" size="icon-sm" onClick={onClose}>
              <XIcon className="size-4" />
            </TooltipButton>
          </>
        ) : null}
      </div>
      {selectedNode ? (
        <div className="flex shrink-0 items-center gap-3 px-3 pt-3 pb-0 text-xs text-muted-foreground">
          <span>{getNodeMetaLine(selectedNode)}</span>
          <div className="ml-auto flex items-center gap-1">
            {showWrapToggle ? (
              <Tooltip delayDuration={0}>
                <TooltipTrigger asChild>
                  <span className="inline-flex">
                    <Toggle
                      size="sm"
                      variant="outline"
                      pressed={isWrapEnabled}
                      onPressedChange={setIsWrapEnabled}
                      aria-label="Toggle wrapped lines"
                    >
                      <TextWrapIcon className="size-4" />
                    </Toggle>
                  </span>
                </TooltipTrigger>
                <TooltipContent>
                  <div>{isWrapEnabled ? 'Disable wrapped lines' : 'Enable wrapped lines'}</div>
                </TooltipContent>
              </Tooltip>
            ) : null}
            <ButtonGroup>
              {previewState.status === 'ready' && previewState.kind === 'text' ? (
                <TooltipButton
                  tooltipText="Copy contents"
                  variant="outline"
                  size="icon-sm"
                  onClick={async () => {
                    await copyToClipboard(previewState.content || '')
                  }}
                >
                  <AnimatePresence initial={false} mode="wait">
                    {copiedText === (previewState.content || '') ? (
                      <MotionCheckIcon key="copied" className="size-4 text-success" {...copyIconMotionProps} />
                    ) : (
                      <MotionCopyIcon key="copy" className="size-4" {...copyIconMotionProps} />
                    )}
                  </AnimatePresence>
                </TooltipButton>
              ) : !selectedNode.isDir ? (
                <TooltipButton tooltipText="Download" variant="outline" size="icon-sm" onClick={onDownload}>
                  <DownloadIcon className="size-4" />
                </TooltipButton>
              ) : null}
              {previewState.status === 'ready' && previewState.kind === 'text' && !selectedNode.isDir ? (
                <TooltipButton tooltipText="Download" variant="outline" size="icon-sm" onClick={onDownload}>
                  <DownloadIcon className="size-4" />
                </TooltipButton>
              ) : null}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" size="icon-sm" className="text-muted-foreground" aria-label="More actions">
                    <EllipsisIcon className="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  align="end"
                  side="bottom"
                  className={cn({
                    'w-44': selectedNode.isDir,
                    'w-48': !selectedNode.isDir,
                  })}
                  onCloseAutoFocus={(event) => event.preventDefault()}
                >
                  <DropdownMenuItem onClick={onRefresh} disabled={isContentsRefreshing}>
                    Refresh
                  </DropdownMenuItem>
                  {!selectedNode.isDir ? <DropdownMenuItem onClick={onDownload}>Download</DropdownMenuItem> : null}
                  {previewState.status === 'ready' && previewState.kind === 'text' ? (
                    <DropdownMenuItem
                      onClick={async () => {
                        await copyToClipboard(previewState.content || '')
                      }}
                    >
                      Copy contents
                    </DropdownMenuItem>
                  ) : null}
                  {canCreateFolderInSelectedDirectory ? (
                    <DropdownMenuItem onSelect={() => onStartCreateFolder(selectedNode.path)}>
                      Create folder
                    </DropdownMenuItem>
                  ) : null}
                  {selectedNode.path !== ROOT_PATH ? <DropdownMenuSeparator /> : null}
                  {selectedNode.path !== ROOT_PATH ? (
                    <DropdownMenuItem variant="destructive" onClick={() => onDelete(selectedNode)}>
                      Delete
                    </DropdownMenuItem>
                  ) : null}
                </DropdownMenuContent>
              </DropdownMenu>
            </ButtonGroup>
          </div>
        </div>
      ) : null}
      <div className="flex-1 min-h-0 overflow-hidden p-3">
        <div
          key={selectedNode?.path ?? 'empty'}
          className={
            isContentsRefreshing ? 'h-full min-h-0 opacity-60 transition-opacity' : 'h-full min-h-0 transition-opacity'
          }
        >
          <SandboxFileContentsBody
            isWrapEnabled={isWrapEnabled}
            isRefreshing={isContentsRefreshing}
            onRetry={retryPreviewState}
            onUploadFiles={onUploadFiles}
            previewState={previewState}
            selectedNode={selectedNode}
          />
        </div>
      </div>
    </div>
  )
}
