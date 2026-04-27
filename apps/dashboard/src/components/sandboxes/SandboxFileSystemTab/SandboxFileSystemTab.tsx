/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Dialog, DialogContent, DialogDescription, DialogOverlay, DialogTitle } from '@/components/ui/dialog'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { Spinner } from '@/components/ui/spinner'
import { handleApiError } from '@/lib/error-handling'
import { cn } from '@/lib/utils'
import { isStoppable } from '@/lib/utils/sandbox'
import { Sandbox } from '@daytona/api-client'
import { useQueryClient } from '@tanstack/react-query'
import { HardDriveIcon } from 'lucide-react'
import { memo, startTransition, useCallback, useEffect, useRef, useState } from 'react'
import { toast } from 'sonner'

import { CreateFolderSheet, type CreateFolderSheetHandle } from './CreateFolderSheet'
import { DeleteNodeDialog, type DeleteNodeDialogHandle } from './DeleteNodeDialog'
import { FileSystemStoreProvider, useFileSystemStore } from './fileSystemStore'
import { FileTreePane, type FileTreePaneRef } from './FileTreePane'
import { useCreateFolderMutation, useDeleteNodeMutation, useUploadFilesMutation } from './mutations'
import { invalidateFileDetailsQuery, invalidateFilePreviewQuery, useIsFilePreviewRefreshing } from './queries'
import { SandboxFileContents } from './SandboxFileContents'
import type { SandboxFileSystemNode } from './types'
import { useSandboxInstance } from './useSandboxInstance'
import { useSelectedNode } from './useSelectedNode'
import {
  createFallbackNode,
  downloadSandboxFile,
  getParentPath,
  isSameOrDescendantPath,
  joinSandboxPath,
} from './utils'

const FILES_COLUMN_MAX_WIDTH = 360
const CONTENTS_OVERLAY_MIN_WIDTH = 350
const DEFAULT_DIRECTORY_PERMISSIONS = '0755'

const MemoizedFileTreePane = memo(FileTreePane)
const MemoizedSandboxFileContents = memo(SandboxFileContents)

function SandboxFileSystem({ sandbox }: { sandbox: Sandbox }) {
  const queryClient = useQueryClient()
  const filesystemRootRef = useRef<HTMLDivElement>(null)
  const createFolderSheetRef = useRef<CreateFolderSheetHandle>(null)
  const deleteNodeDialogRef = useRef<DeleteNodeDialogHandle>(null)
  const fileTreePaneRef = useRef<FileTreePaneRef>(null)
  const [isContentsOverlayMode, setIsContentsOverlayMode] = useState(false)
  const [isContentsOverlayOpen, setIsContentsOverlayOpen] = useState(false)
  const [filesystemRootElement, setFilesystemRootElement] = useState<HTMLDivElement | null>(null)
  const [filesPaneElement, setFilesPaneElement] = useState<HTMLDivElement | null>(null)
  const [inlineContentsPaneElement, setInlineContentsPaneElement] = useState<HTMLDivElement | null>(null)

  const nextFilePath = useFileSystemStore((state) => state.nextFilePath)
  const lastOpenedNodePath = useFileSystemStore((state) => state.lastOpenedNodePath)
  const previousFilePath = useFileSystemStore((state) => state.previousFilePath)
  const { clearSelectedNode, openNode } = useFileSystemStore((state) => state.actions)

  const sandboxInstanceQuery = useSandboxInstance(sandbox.id)
  const sandboxInstance = sandboxInstanceQuery.data
  const { selectedNode, selectedNodePath, selectedNodeQuery } = useSelectedNode({ sandboxInstance })
  const isSelectedFileRefreshing = useIsFilePreviewRefreshing({
    path: selectedNode && !selectedNode.isDir ? selectedNode.path : null,
    sandboxInstance,
  })

  const createFolderMutation = useCreateFolderMutation({ sandboxInstance })
  const deleteNodeMutation = useDeleteNodeMutation({ sandboxInstance })
  const uploadFilesMutation = useUploadFilesMutation({ sandboxInstance })

  const handleDownloadNode = useCallback(
    async (node: SandboxFileSystemNode) => {
      if (!sandboxInstance) {
        return
      }

      await downloadSandboxFile({
        node,
        sandboxInstance,
      })
    },
    [sandboxInstance],
  )

  const handleDownloadSelected = useCallback(async () => {
    if (!selectedNode) {
      return
    }

    await handleDownloadNode(selectedNode)
  }, [handleDownloadNode, selectedNode])

  const handleUploadFiles = useCallback(
    (files: File[]) => {
      if (!selectedNode?.isDir || files.length === 0) {
        return
      }

      const targetPath = selectedNode.path
      const batchLabel = files.length === 1 ? files[0].name : `${files.length} files`
      const firstUploadedPath = joinSandboxPath(targetPath, files[0].name)
      const uploadPromise = uploadFilesMutation
        .mutateAsync({
          files,
          targetPath,
        })
        .then(async ({ files: uploadedFiles, targetPath: uploadedTargetPath }) => {
          await fileTreePaneRef.current?.refreshPath(uploadedTargetPath)
          openNode(firstUploadedPath)

          return {
            batchLabel,
            files: uploadedFiles,
            targetPath: uploadedTargetPath,
          }
        })

      toast.promise(uploadPromise, {
        loading: `Uploading ${batchLabel}…`,
        success: ({ files: uploadedFiles, targetPath: uploadedTargetPath }) =>
          uploadedFiles.length === 1
            ? `Uploaded ${uploadedFiles[0].name} to ${uploadedTargetPath}`
            : `Uploaded ${uploadedFiles.length} files to ${uploadedTargetPath}`,
        error: (error) =>
          error instanceof Error
            ? `Upload failed: ${error.message}`
            : `Failed to upload ${batchLabel} to ${targetPath}`,
      })
    },
    [openNode, selectedNode, uploadFilesMutation],
  )

  const handleRefreshSelected = useCallback(async () => {
    if (!selectedNode || !sandboxInstance) {
      return
    }

    await invalidateFileDetailsQuery({
      path: selectedNode.path,
      queryClient,
      sandboxInstance,
    })

    if (selectedNode.isDir) {
      await fileTreePaneRef.current?.refreshPath(selectedNode.path)
      return
    }

    await invalidateFilePreviewQuery({
      path: selectedNode.path,
      queryClient,
      sandboxInstance,
    })
  }, [queryClient, sandboxInstance, selectedNode])

  const handleCloseContents = useCallback(() => {
    setIsContentsOverlayOpen(false)
    clearSelectedNode()
  }, [clearSelectedNode])

  const handleDeleteNode = useCallback(
    async (node: SandboxFileSystemNode) => {
      const parentPath = getParentPath(node.path)
      const deletedSelectedNode = selectedNode?.path ? isSameOrDescendantPath(selectedNode.path, node.path) : false

      if (deletedSelectedNode) {
        openNode((fileTreePaneRef.current?.getNode(parentPath) ?? createFallbackNode(parentPath)).path)
      }

      try {
        await deleteNodeMutation.mutateAsync(node)
        await fileTreePaneRef.current?.refreshPath(parentPath)

        toast.success(`Deleted ${node.path}`)
      } catch (error) {
        if (deletedSelectedNode) {
          openNode(selectedNode?.path ?? node.path)
        }

        handleApiError(error, `Failed to delete ${node.path}`)
        throw error
      }
    },
    [deleteNodeMutation, openNode, selectedNode?.path],
  )

  const handleRequestDelete = useCallback((node: SandboxFileSystemNode) => {
    deleteNodeDialogRef.current?.open(node)
  }, [])

  const handleRequestCreateFolder = useCallback((parentPath: string) => {
    createFolderSheetRef.current?.open(parentPath)
  }, [])

  const handleCreateFolder = useCallback(
    async ({ name, parentPath }: { name: string; parentPath: string }) => {
      const newFolderPath = joinSandboxPath(parentPath, name)

      try {
        await createFolderMutation.mutateAsync({
          path: newFolderPath,
          permissions: DEFAULT_DIRECTORY_PERMISSIONS,
        })

        await fileTreePaneRef.current?.expandPathAncestors(newFolderPath)
        openNode(
          (
            fileTreePaneRef.current?.getNode(newFolderPath) ??
            ({
              ...createFallbackNode(newFolderPath),
              isDir: true,
              permissions: DEFAULT_DIRECTORY_PERMISSIONS,
              mode: DEFAULT_DIRECTORY_PERMISSIONS,
            } satisfies SandboxFileSystemNode)
          ).path,
        )
        toast.success(`Created folder ${newFolderPath}`)
      } catch (error) {
        handleApiError(error, `Failed to create ${newFolderPath}`)
        throw error
      }
    },
    [createFolderMutation, openNode],
  )

  const handleNavigatePrevious = useCallback(() => {
    if (!previousFilePath) {
      return
    }

    fileTreePaneRef.current?.revealPath(previousFilePath)
    startTransition(() => {
      openNode(previousFilePath)
    })
  }, [openNode, previousFilePath])

  const handleNavigateNext = useCallback(() => {
    if (!nextFilePath) {
      return
    }

    fileTreePaneRef.current?.revealPath(nextFilePath)
    startTransition(() => {
      openNode(nextFilePath)
    })
  }, [nextFilePath, openNode])

  const setFilesystemRootRef = useCallback((element: HTMLDivElement | null) => {
    filesystemRootRef.current = element
    setFilesystemRootElement(element)
  }, [])

  const setFilesPaneRef = useCallback((element: HTMLDivElement | null) => {
    setFilesPaneElement(element)
  }, [])

  const setInlineContentsPaneRef = useCallback((element: HTMLDivElement | null) => {
    setInlineContentsPaneElement(element)
  }, [])

  const updateContentsOverlayMode = useCallback(() => {
    if (!filesystemRootElement) {
      return
    }

    const inlineContentsWidth = inlineContentsPaneElement?.getBoundingClientRect().width
    if (inlineContentsWidth && inlineContentsWidth > 0) {
      setIsContentsOverlayMode(inlineContentsWidth < CONTENTS_OVERLAY_MIN_WIDTH)
      return
    }

    const filesPaneWidth = Math.min(filesPaneElement?.getBoundingClientRect().width ?? 0, FILES_COLUMN_MAX_WIDTH)
    const rootWidth = filesystemRootElement.getBoundingClientRect().width
    const contentsWidth = Math.max(0, rootWidth - filesPaneWidth)
    setIsContentsOverlayMode(contentsWidth < CONTENTS_OVERLAY_MIN_WIDTH)
  }, [filesystemRootElement, filesPaneElement, inlineContentsPaneElement])

  useEffect(() => {
    if (!filesystemRootElement) {
      return
    }

    updateContentsOverlayMode()

    const observer = new ResizeObserver(() => {
      updateContentsOverlayMode()
    })

    observer.observe(filesystemRootElement)
    if (filesPaneElement) {
      observer.observe(filesPaneElement)
    }
    if (inlineContentsPaneElement) {
      observer.observe(inlineContentsPaneElement)
    }

    return () => observer.disconnect()
  }, [filesystemRootElement, filesPaneElement, inlineContentsPaneElement, updateContentsOverlayMode])

  useEffect(() => {
    if (!isContentsOverlayMode) {
      setIsContentsOverlayOpen(false)
      return
    }

    const hasActiveContents = Boolean(selectedNodePath)

    if (hasActiveContents) {
      setIsContentsOverlayOpen(true)
      return
    }

    setIsContentsOverlayOpen(false)
  }, [isContentsOverlayMode, selectedNodePath])

  const hasActiveContents = Boolean(selectedNodePath)
  const previewLoadingPath = isSelectedFileRefreshing ? (selectedNodePath ?? undefined) : undefined

  const contentsView = (
    <MemoizedSandboxFileContents
      onNavigatePrevious={handleNavigatePrevious}
      onNavigateNext={handleNavigateNext}
      onClose={handleCloseContents}
      onDelete={handleRequestDelete}
      onDownload={handleDownloadSelected}
      onRefresh={handleRefreshSelected}
      onStartCreateFolder={handleRequestCreateFolder}
      onUploadFiles={handleUploadFiles}
      sandboxInstance={sandboxInstance}
      selectedNode={selectedNode}
      selectedNodeError={selectedNodeQuery.error}
      selectedNodePath={selectedNodePath}
    />
  )

  if (!isStoppable(sandbox)) {
    return (
      <div className="flex flex-1 flex-col p-4">
        <div className="flex min-h-0 flex-1 rounded-md border border-border">
          <Empty className="border-0">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <HardDriveIcon className="size-4" />
              </EmptyMedia>
              <EmptyTitle>Sandbox is not running</EmptyTitle>
              <EmptyDescription>Start the sandbox to browse and inspect its filesystem.</EmptyDescription>
            </EmptyHeader>
          </Empty>
        </div>
      </div>
    )
  }

  if (sandboxInstanceQuery.isPending) {
    return (
      <div className="flex flex-1 flex-col p-4">
        <div className="flex min-h-0 flex-1 rounded-md border border-border">
          <Empty className="border-0">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Spinner />
              </EmptyMedia>
              <EmptyTitle>Loading filesystem</EmptyTitle>
              <EmptyDescription>Preparing the sandbox filesystem for browsing.</EmptyDescription>
            </EmptyHeader>
          </Empty>
        </div>
      </div>
    )
  }

  if (!sandboxInstance || sandboxInstanceQuery.isError) {
    return (
      <div className="flex flex-1 flex-col p-4">
        <div className="flex min-h-0 flex-1 rounded-md border border-border">
          <Empty className="border-0">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <HardDriveIcon className="size-4" />
              </EmptyMedia>
              <EmptyTitle>Sandbox is unavailable</EmptyTitle>
              <EmptyDescription>The filesystem can’t be shown because the sandbox is not available.</EmptyDescription>
            </EmptyHeader>
          </Empty>
        </div>
      </div>
    )
  }

  return (
    <>
      <CreateFolderSheet
        getParentNode={(parentPath) => fileTreePaneRef.current?.getNode(parentPath) ?? createFallbackNode(parentPath)}
        isPending={createFolderMutation.isPending}
        onCreate={handleCreateFolder}
        ref={createFolderSheetRef}
      />
      <DeleteNodeDialog
        isPending={deleteNodeMutation.isPending}
        onDelete={handleDeleteNode}
        ref={deleteNodeDialogRef}
      />

      <div className="flex flex-1 min-h-0 flex-col overflow-hidden p-4">
        <div
          ref={setFilesystemRootRef}
          className={cn('relative grid min-h-0 flex-1 overflow-hidden rounded-md border border-border', {
            'grid-cols-1': isContentsOverlayMode,
          })}
          style={
            isContentsOverlayMode
              ? undefined
              : {
                  gridTemplateColumns: `minmax(280px, ${FILES_COLUMN_MAX_WIDTH}px) minmax(0, 1fr)`,
                }
          }
        >
          <div
            ref={setFilesPaneRef}
            className={cn('flex min-h-0 flex-col overflow-hidden', {
              'border-r border-border': !isContentsOverlayMode,
            })}
          >
            <MemoizedFileTreePane
              onRequestCreateFolder={handleRequestCreateFolder}
              onRequestDelete={handleRequestDelete}
              ref={fileTreePaneRef}
              previewLoadingPath={previewLoadingPath}
              sandboxId={sandbox.id}
            />
          </div>

          {isContentsOverlayMode ? (
            <Dialog
              open={isContentsOverlayOpen && hasActiveContents}
              onOpenChange={(open) => {
                if (!open) {
                  handleCloseContents()
                  return
                }

                setIsContentsOverlayOpen(true)
              }}
            >
              <DialogContent
                animate={false}
                container={filesystemRootRef.current}
                overlay={<DialogOverlay className="absolute inset-0 z-20" />}
                className="absolute inset-0 z-30 max-h-none max-w-none gap-0 overflow-hidden rounded-none border-0 p-0 shadow-none outline-none sm:max-w-none translate-y-0 translate-x-0 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95"
                showCloseButton={false}
                onCloseAutoFocus={(event) => {
                  event.preventDefault()
                  fileTreePaneRef.current?.restoreFocus(lastOpenedNodePath)
                }}
              >
                <DialogTitle className="sr-only">{selectedNodePath ?? 'Contents'}</DialogTitle>
                <DialogDescription className="sr-only">
                  Previewing sandbox filesystem contents inside the filesystem tab.
                </DialogDescription>
                {hasActiveContents ? contentsView : null}
              </DialogContent>
            </Dialog>
          ) : (
            <div ref={setInlineContentsPaneRef} className="flex min-h-0 flex-col overflow-hidden">
              {contentsView}
            </div>
          )}
        </div>
      </div>
    </>
  )
}

export function SandboxFileSystemTab({ sandbox }: { sandbox: Sandbox }) {
  return (
    <FileSystemStoreProvider>
      <SandboxFileSystem sandbox={sandbox} />
    </FileSystemStoreProvider>
  )
}
