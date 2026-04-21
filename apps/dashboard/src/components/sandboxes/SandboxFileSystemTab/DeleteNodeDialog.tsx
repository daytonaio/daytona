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
import { Ref, useImperativeHandle, useState } from 'react'

import type { SandboxFileSystemNode } from './types'

export type DeleteNodeDialogHandle = {
  close: () => void
  open: (node: SandboxFileSystemNode) => void
}

export function DeleteNodeDialog({
  isPending,
  onDelete,
  ref,
}: {
  isPending: boolean
  onDelete: (node: SandboxFileSystemNode) => Promise<void>
  ref?: Ref<DeleteNodeDialogHandle>
}) {
  const [isOpen, setIsOpen] = useState(false)
  const [target, setTarget] = useState<SandboxFileSystemNode | null>(null)

  const resetState = () => {
    setIsOpen(false)
    setTarget(null)
  }

  useImperativeHandle(
    ref,
    () => ({
      close: () => {
        if (isPending) {
          return
        }

        resetState()
      },
      open: (node: SandboxFileSystemNode) => {
        setTarget(node)
        setIsOpen(true)
      },
    }),
    [isPending],
  )

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      if (isPending) {
        return
      }

      resetState()
      return
    }

    setIsOpen(true)
  }

  const handleConfirmDelete = async () => {
    if (!target || isPending) {
      return
    }

    await onDelete(target)
    resetState()
  }

  return (
    <AlertDialog open={isOpen} onOpenChange={handleOpenChange}>
      <AlertDialogContent className="max-w-sm sm:max-w-sm">
        <AlertDialogHeader>
          <AlertDialogTitle>Delete {target?.isDir ? 'directory' : 'file'}?</AlertDialogTitle>
          <AlertDialogDescription className="break-words">
            {target ? (
              <>
                <span>This will permanently delete</span>
                <span className="mt-2 block break-all whitespace-normal text-foreground">{target.path}</span>
                {target.isDir ? <span className="mt-2 block">Its contents will be removed too.</span> : null}
              </>
            ) : null}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isPending}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            disabled={isPending}
            onClick={async (event) => {
              event.preventDefault()
              await handleConfirmDelete()
            }}
          >
            {isPending ? 'Deleting…' : 'Delete'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
