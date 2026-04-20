/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogOverlay, DialogTitle } from '@/components/ui/dialog'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { Field, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Spinner } from '@/components/ui/spinner'
import { useCopyToClipboard } from '@/hooks/useCopyToClipboard'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { cn, downloadBlob } from '@/lib/utils'
import { isStoppable } from '@/lib/utils/sandbox'
import { Sandbox } from '@daytona/api-client'
import { Daytona } from '@daytona/sdk'
import { Buffer } from 'buffer'
import { HardDriveIcon } from 'lucide-react'
import { memo, startTransition, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { toast } from 'sonner'

import { useConfig } from '@/hooks/useConfig'
import { CONTENTS_OVERLAY_MIN_WIDTH, FILES_COLUMN_MAX_WIDTH, MAX_PREVIEW_BYTES, ROOT_PATH } from './constants'
import { FileContentsPanel } from './FileContentsPanel'
import { FileNodeActions } from './FileNodeActions'
import { FilePreview } from './FilePreview'
import { FileSystemStoreProvider, useFileSystemStore } from './fileSystemStore'
import { FileTreePane, type FileTreePaneHandle } from './FileTreePane'
import { useCreateFolderMutation, useDeleteNodeMutation, useUploadFilesMutation } from './mutations'
import { useFilePreviewQuery, useSandboxInstanceQuery } from './queries'
import type { PreviewState, SandboxFileSystemNode } from './types'
import { createFallbackNode, getCodeLanguage, getParentPath, isProbablyBinary, joinSandboxPath } from './utils'

const MemoizedFileTreePane = memo(FileTreePane)
const MemoizedFileContentsPanel = memo(FileContentsPanel)
const MemoizedFilePreview = memo(FilePreview)
const MemoizedContentsActions = memo(FileNodeActions)

function SandboxFileSystem({ sandbox }: { sandbox: Sandbox }) {
  const { user } = useAuth()
  const { apiUrl } = useConfig()
  const { selectedOrganization } = useSelectedOrganization()
  const [copiedContents, copyContents] = useCopyToClipboard()
  const [canNavigateNext, setCanNavigateNext] = useState(false)
  const [canNavigatePrevious, setCanNavigatePrevious] = useState(false)
  const [isSelectedDirectoryRefreshing, setIsSelectedDirectoryRefreshing] = useState(false)
  const previousSelectionWasFileRef = useRef(false)
  const previousFileRef = useRef<SandboxFileSystemNode | null>(null)
  const filesystemRootRef = useRef<HTMLDivElement>(null)
  const createFolderInputRef = useRef<HTMLInputElement>(null)
  const fileTreePaneRef = useRef<FileTreePaneHandle>(null)

  const selectedNode = useFileSystemStore((state) => state.selectedNode)
  const deleteTarget = useFileSystemStore((state) => state.deleteTarget)
  const folderCreationParentPath = useFileSystemStore((state) => state.folderCreationParentPath)
  const isContentsOverlayMode = useFileSystemStore((state) => state.isContentsOverlayMode)
  const isContentsOverlayOpen = useFileSystemStore((state) => state.isContentsOverlayOpen)
  const lastOpenedNodePath = useFileSystemStore((state) => state.lastOpenedNodePath)
  const newFolderName = useFileSystemStore((state) => state.newFolderName)
  const openDropdownPath = useFileSystemStore((state) => state.openDropdownPath)
  const {
    clearSelectedNode,
    closeCreateFolder,
    closeDeleteDialog,
    openNode,
    setContentsOverlayMode,
    setContentsOverlayOpen,
    setDeleteTarget,
    setFolderCreationParentPath,
    setNewFolderName,
    setOpenDropdownPath,
  } = useFileSystemStore((state) => state.actions)

  const client = useMemo(() => {
    if (!user?.access_token || !selectedOrganization?.id) {
      return null
    }

    return new Daytona({
      jwtToken: user.access_token,
      apiUrl,
      organizationId: selectedOrganization.id,
    })
  }, [selectedOrganization?.id, user?.access_token, apiUrl])

  const sandboxInstanceQuery = useSandboxInstanceQuery({
    client,
    sandboxId: sandbox.id,
  })
  const sandboxInstance = sandboxInstanceQuery.data

  const previewQuery = useFilePreviewQuery({
    enabled: Boolean(sandboxInstance && selectedNode && !selectedNode.isDir && selectedNode.size <= MAX_PREVIEW_BYTES),
    path: selectedNode?.path ?? ROOT_PATH,
    sandboxInstance,
  })
  const shouldUsePreviousPreview = previousSelectionWasFileRef.current && Boolean(selectedNode && !selectedNode.isDir)

  useEffect(() => {
    previousSelectionWasFileRef.current = Boolean(selectedNode && !selectedNode.isDir)
    previousFileRef.current = selectedNode && !selectedNode.isDir ? selectedNode : null
  }, [selectedNode?.isDir, selectedNode?.path])

  const previewState = useMemo<PreviewState>(() => {
    if (!selectedNode) {
      return { status: 'idle' }
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
    shouldUsePreviousPreview,
  ])

  const createFolderMutation = useCreateFolderMutation({ sandboxInstance })
  const deleteNodeMutation = useDeleteNodeMutation({ sandboxInstance })
  const uploadFilesMutation = useUploadFilesMutation({ sandboxInstance })
  const isCreateFolderPending = createFolderMutation.isPending
  const isDeletePending = deleteNodeMutation.isPending
  const isSelectedFileRefreshing = Boolean(
    selectedNode && previewState.status === 'loading' && previewState.path === selectedNode.path,
  )
  const isContentsRefreshing = isSelectedDirectoryRefreshing || isSelectedFileRefreshing
  const selectedCodeLanguage = selectedNode ? getCodeLanguage(selectedNode.path) : null
  const showWrapToggle =
    Boolean(selectedNode && !selectedNode.isDir) &&
    (previewState.status === 'ready' || previewState.status === 'loading') &&
    (previewState.status === 'ready' ? previewState.kind === 'text' : previewState.previousKind === 'text') &&
    !selectedCodeLanguage

  const handleDownloadNode = useCallback(
    async (node: SandboxFileSystemNode) => {
      if (node.isDir) {
        return
      }

      try {
        if (!sandboxInstance) {
          throw new Error('Sandbox instance is not available')
        }

        const fileContents = Buffer.from(await sandboxInstance.fs.downloadFile(node.path))
        downloadBlob(new Blob([fileContents]), node.name || 'download')
        toast.success(`Downloaded ${node.name || node.path}`)
      } catch (error) {
        handleApiError(error, `Failed to download ${node.path}`)
      }
    },
    [sandboxInstance],
  )

  const handleDownloadSelected = useCallback(async () => {
    if (!selectedNode) {
      return
    }

    await handleDownloadNode(selectedNode)
  }, [handleDownloadNode, selectedNode])

  const handleCopyNodeContents = useCallback(
    async (node: SandboxFileSystemNode) => {
      if (node.isDir) {
        return
      }

      try {
        if (!sandboxInstance) {
          throw new Error('Sandbox instance is not available')
        }

        const fileContents = Buffer.from(await sandboxInstance.fs.downloadFile(node.path))

        if (isProbablyBinary(fileContents)) {
          toast.error('Binary file contents cannot be copied as text')
          return
        }

        await navigator.clipboard.writeText(fileContents.toString('utf-8'))
        toast.success(`Copied contents of ${node.name || node.path}`)
      } catch (error) {
        handleApiError(error, `Failed to copy ${node.path}`)
      }
    },
    [sandboxInstance],
  )

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

          return {
            batchLabel,
            files: uploadedFiles,
            firstUploadedNode:
              fileTreePaneRef.current?.getNode(firstUploadedPath) ?? createFallbackNode(firstUploadedPath),
            targetPath: uploadedTargetPath,
          }
        })

      void uploadPromise.then(({ firstUploadedNode }) => {
        openNode(firstUploadedNode)
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

  const handleRefreshSelected = useCallback(() => {
    if (!selectedNode) {
      return
    }

    if (selectedNode.isDir) {
      void fileTreePaneRef.current?.refreshPath(selectedNode.path)
      return
    }

    void previewQuery.refetch()
  }, [previewQuery, selectedNode])

  const handleCloseContents = useCallback(() => {
    clearSelectedNode()
  }, [clearSelectedNode])

  const handleOpenCreateFolderDialog = useCallback(
    (parentPath: string) => {
      setOpenDropdownPath(null)
      setFolderCreationParentPath(parentPath)
      setNewFolderName('')
    },
    [setFolderCreationParentPath, setNewFolderName, setOpenDropdownPath],
  )

  const handleCreateFolderSheetOpenChange = useCallback(
    (open: boolean) => {
      if (!open && !isCreateFolderPending) {
        closeCreateFolder()
      }
    },
    [closeCreateFolder, isCreateFolderPending],
  )

  useEffect(() => {
    if (!folderCreationParentPath) {
      return
    }

    const frameId = requestAnimationFrame(() => {
      createFolderInputRef.current?.focus()
    })

    return () => cancelAnimationFrame(frameId)
  }, [folderCreationParentPath])

  const handleDeleteNode = useCallback(
    async (node: SandboxFileSystemNode) => {
      try {
        const parentPath = getParentPath(node.path)
        await deleteNodeMutation.mutateAsync(node)

        if (selectedNode?.path === node.path) {
          openNode(fileTreePaneRef.current?.getNode(parentPath) ?? createFallbackNode(parentPath))
        }

        toast.success(`Deleted ${node.path}`)
      } catch (error) {
        handleApiError(error, `Failed to delete ${node.path}`)
      }
    },
    [deleteNodeMutation, openNode, selectedNode?.path],
  )

  const handleConfirmDelete = useCallback(async () => {
    if (!deleteTarget) {
      return
    }

    await handleDeleteNode(deleteTarget)
    closeDeleteDialog()
  }, [closeDeleteDialog, deleteTarget, handleDeleteNode])

  const handleCreateFolder = useCallback(
    async (parentPath: string) => {
      const trimmedName = newFolderName.trim()
      if (!trimmedName) {
        return
      }

      const newFolderPath = joinSandboxPath(parentPath, trimmedName)

      try {
        await createFolderMutation.mutateAsync({
          path: newFolderPath,
          permissions: '0755',
        })

        await fileTreePaneRef.current?.expandPathAncestors(newFolderPath)
        closeCreateFolder()
        setOpenDropdownPath(null)
        openNode(
          fileTreePaneRef.current?.getNode(newFolderPath) ??
            ({
              ...createFallbackNode(newFolderPath),
              isDir: true,
              permissions: '0755',
              mode: '0755',
            } satisfies SandboxFileSystemNode),
        )
        toast.success(`Created folder ${newFolderPath}`)
      } catch (error) {
        handleApiError(error, `Failed to create ${newFolderPath}`)
      }
    },
    [closeCreateFolder, createFolderMutation, newFolderName, openNode, setOpenDropdownPath],
  )

  const handleNavigateFile = useCallback(
    (direction: -1 | 1) => {
      if (!selectedNode || selectedNode.isDir) {
        return
      }

      const nextNode = fileTreePaneRef.current?.getAdjacentFileNode(selectedNode.path, direction)
      if (!nextNode) {
        return
      }

      startTransition(() => {
        openNode(nextNode)
      })
    },
    [openNode, selectedNode],
  )

  const handleCopySelectedContents = useCallback(async () => {
    if (previewState.status !== 'ready' || previewState.kind !== 'text') {
      return
    }

    await copyContents(previewState.content || '')
  }, [copyContents, previewState])

  const handleRetryPreview = useCallback(() => {
    void previewQuery.refetch()
  }, [previewQuery])

  const handleContentsDropdownOpenChange = useCallback(
    (open: boolean) => {
      setOpenDropdownPath(open && selectedNode ? `contents:${selectedNode.path}` : null)
    },
    [selectedNode, setOpenDropdownPath],
  )

  const handleDeleteSelectedNode = useCallback(() => {
    if (!selectedNode) {
      return
    }

    setOpenDropdownPath(null)
    setDeleteTarget(selectedNode)
  }, [selectedNode, setDeleteTarget, setOpenDropdownPath])

  const handleStartCreateFolderFromContents = useCallback(() => {
    if (!selectedNode) {
      return
    }

    handleOpenCreateFolderDialog(selectedNode.path)
  }, [handleOpenCreateFolderDialog, selectedNode])

  const handleVisibleFileNavigationChange = useCallback(
    ({ canNavigateNext, canNavigatePrevious }: { canNavigateNext: boolean; canNavigatePrevious: boolean }) => {
      setCanNavigateNext(canNavigateNext)
      setCanNavigatePrevious(canNavigatePrevious)
    },
    [],
  )

  useEffect(() => {
    const element = filesystemRootRef.current
    if (!element) {
      return
    }

    const observer = new ResizeObserver(([entry]) => {
      const nextOverlayMode = entry.contentRect.width < FILES_COLUMN_MAX_WIDTH + CONTENTS_OVERLAY_MIN_WIDTH
      setContentsOverlayMode(nextOverlayMode)
    })

    observer.observe(element)

    return () => observer.disconnect()
  }, [setContentsOverlayMode])

  useEffect(() => {
    if (!isContentsOverlayMode) {
      setContentsOverlayOpen(false)
      return
    }

    const hasActiveContents = Boolean(selectedNode && previewState.status !== 'idle')

    if (hasActiveContents) {
      setContentsOverlayOpen(true)
    }
  }, [isContentsOverlayMode, previewState.status, selectedNode, setContentsOverlayOpen])

  const contentsPrimaryAction = useMemo(() => {
    if (!selectedNode) {
      return undefined
    }

    if (
      !selectedNode.isDir &&
      previewState.status === 'ready' &&
      previewState.kind === 'text' &&
      previewState.content
    ) {
      return {
        kind: 'copy' as const,
        copied: copiedContents === (previewState.content || ''),
        onClick: async () => {
          await copyContents(previewState.content || '')
        },
      }
    }

    if (!selectedNode.isDir) {
      return {
        kind: 'download' as const,
        onClick: () => {
          void handleDownloadSelected()
        },
      }
    }

    return undefined
  }, [copiedContents, copyContents, handleDownloadSelected, previewState, selectedNode])

  const createFolderParentNode = folderCreationParentPath
    ? (fileTreePaneRef.current?.getNode(folderCreationParentPath) ?? createFallbackNode(folderCreationParentPath))
    : null
  const contentsActions = selectedNode ? (
    <MemoizedContentsActions
      variant="compound"
      node={selectedNode}
      canDelete={selectedNode.path !== ROOT_PATH}
      isDropdownOpen={openDropdownPath === `contents:${selectedNode.path}`}
      isRefreshing={isContentsRefreshing}
      onRefresh={handleRefreshSelected}
      onDownload={!selectedNode.isDir ? handleDownloadSelected : undefined}
      onCopy={
        !selectedNode.isDir && previewState.status === 'ready' && previewState.kind === 'text'
          ? handleCopySelectedContents
          : undefined
      }
      onDelete={handleDeleteSelectedNode}
      onDropdownOpenChange={handleContentsDropdownOpenChange}
      onStartCreateFolder={selectedNode.isDir ? handleStartCreateFolderFromContents : undefined}
      primaryAction={contentsPrimaryAction}
    />
  ) : undefined

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
      <Sheet open={Boolean(folderCreationParentPath)} onOpenChange={handleCreateFolderSheetOpenChange}>
        <SheetContent side="right" className="w-dvw flex flex-col gap-0 p-0 sm:w-[400px]">
          <SheetHeader className="flex flex-row items-center border-b border-border p-4 px-5 text-left">
            <SheetTitle className="text-2xl">Create folder</SheetTitle>
            <SheetDescription className="sr-only">Create a folder in {createFolderParentNode?.path}</SheetDescription>
          </SheetHeader>
          <div className="flex-1 overflow-y-auto p-5">
            <p className="mb-4 break-all text-sm text-muted-foreground">{createFolderParentNode?.path}</p>
            <form
              id="create-folder-form"
              onSubmit={(event) => {
                event.preventDefault()
                event.stopPropagation()

                if (folderCreationParentPath) {
                  void handleCreateFolder(folderCreationParentPath)
                }
              }}
            >
              <Field>
                <FieldLabel htmlFor="create-folder-name">Folder name</FieldLabel>
                <Input
                  id="create-folder-name"
                  ref={createFolderInputRef}
                  autoFocus
                  value={newFolderName}
                  onChange={(event) => setNewFolderName(event.target.value)}
                  placeholder="Name your folder"
                />
              </Field>
            </form>
          </div>
          <SheetFooter className="mt-auto border-t border-border p-4 px-5">
            <SheetClose asChild>
              <Button variant="secondary" disabled={isCreateFolderPending}>
                Close
              </Button>
            </SheetClose>
            <Button
              type="submit"
              form="create-folder-form"
              disabled={!newFolderName.trim() || isCreateFolderPending || !folderCreationParentPath}
            >
              {isCreateFolderPending && <Spinner />}
              Create
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>

      <AlertDialog
        open={Boolean(deleteTarget)}
        onOpenChange={(open) => {
          if (!open && !isDeletePending) {
            closeDeleteDialog()
          }
        }}
      >
        <AlertDialogContent className="max-w-sm sm:max-w-sm">
          <AlertDialogHeader>
            <AlertDialogTitle>Delete {deleteTarget?.isDir ? 'directory' : 'file'}?</AlertDialogTitle>
            <AlertDialogDescription className="break-words">
              {deleteTarget ? (
                <>
                  <span>This will permanently delete</span>
                  <span className="mt-2 block break-all whitespace-normal text-foreground">{deleteTarget.path}</span>
                  {deleteTarget.isDir ? <span className="mt-2 block">Its contents will be removed too.</span> : null}
                </>
              ) : null}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeletePending}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              disabled={isDeletePending}
              onClick={() => void handleConfirmDelete()}
            >
              {isDeletePending ? 'Deleting…' : 'Delete'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <div className="flex flex-1 min-h-0 flex-col overflow-hidden p-4">
        <div
          ref={filesystemRootRef}
          className={cn('relative grid min-h-0 flex-1 overflow-hidden rounded-md border border-border', {
            'grid-cols-1': isContentsOverlayMode,
            'grid-cols-[minmax(280px,360px)_minmax(0,1fr)]': !isContentsOverlayMode,
          })}
        >
          <div
            className={cn('flex min-h-0 flex-col overflow-hidden', {
              'border-r border-border': !isContentsOverlayMode,
            })}
          >
            <MemoizedFileTreePane
              ref={fileTreePaneRef}
              onCopyNodeContents={handleCopyNodeContents}
              onDownloadNode={handleDownloadNode}
              onSelectedDirectoryRefreshingChange={setIsSelectedDirectoryRefreshing}
              onVisibleFileNavigationChange={handleVisibleFileNavigationChange}
              previewState={previewState}
              sandboxInstance={sandboxInstance}
            />
          </div>

          {!isContentsOverlayMode ? (
            <div className="flex min-h-0 flex-col overflow-hidden">
              <MemoizedFileContentsPanel
                actions={contentsActions}
                canNavigateNext={canNavigateNext}
                canNavigatePrevious={canNavigatePrevious}
                isContentsRefreshing={isContentsRefreshing}
                onNavigatePrevious={() => handleNavigateFile(-1)}
                onNavigateNext={() => handleNavigateFile(1)}
                onClose={handleCloseContents}
                showWrapToggle={showWrapToggle}
              >
                {({ isWrapEnabled }) => (
                  <MemoizedFilePreview
                    isWrapEnabled={isWrapEnabled}
                    previewState={previewState}
                    onRetry={handleRetryPreview}
                    onUploadFiles={handleUploadFiles}
                  />
                )}
              </MemoizedFileContentsPanel>
            </div>
          ) : null}

          {isContentsOverlayMode ? (
            <Dialog
              open={isContentsOverlayOpen}
              onOpenChange={(open) => {
                if (!open) {
                  handleCloseContents()
                  return
                }

                setContentsOverlayOpen(true)
              }}
            >
              <DialogContent
                container={filesystemRootRef.current}
                overlay={
                  <DialogOverlay className="absolute inset-0 z-20 bg-background/80 backdrop-blur-[1px] data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:animate-in data-[state=open]:fade-in-0" />
                }
                className="absolute inset-0 top-0 left-0 z-30 max-h-none max-w-none origin-center translate-x-0 translate-y-0 overflow-hidden rounded-none border-0 bg-background p-0 outline-none data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:slide-out-to-left-0 data-[state=closed]:slide-out-to-top-0 data-[state=closed]:zoom-out-95 data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=open]:slide-in-from-left-0 data-[state=open]:slide-in-from-top-0 data-[state=open]:zoom-in-95"
                showCloseButton={false}
                onCloseAutoFocus={(event) => {
                  event.preventDefault()
                  fileTreePaneRef.current?.restoreFocus(lastOpenedNodePath)
                }}
              >
                <DialogTitle className="sr-only">{selectedNode?.path ?? 'Contents'}</DialogTitle>
                <DialogDescription className="sr-only">
                  Previewing sandbox filesystem contents inside the filesystem tab.
                </DialogDescription>
                <MemoizedFileContentsPanel
                  overlay
                  actions={contentsActions}
                  canNavigateNext={canNavigateNext}
                  canNavigatePrevious={canNavigatePrevious}
                  isContentsRefreshing={isContentsRefreshing}
                  onNavigatePrevious={() => handleNavigateFile(-1)}
                  onNavigateNext={() => handleNavigateFile(1)}
                  onClose={handleCloseContents}
                  showWrapToggle={showWrapToggle}
                >
                  {({ isWrapEnabled }) => (
                    <MemoizedFilePreview
                      isWrapEnabled={isWrapEnabled}
                      previewState={previewState}
                      onRetry={handleRetryPreview}
                      onUploadFiles={handleUploadFiles}
                    />
                  )}
                </MemoizedFileContentsPanel>
              </DialogContent>
            </Dialog>
          ) : null}
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
