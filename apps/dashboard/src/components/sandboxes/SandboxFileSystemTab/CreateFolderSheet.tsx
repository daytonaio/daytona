/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { Field, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Spinner } from '@/components/ui/spinner'
import { Ref, useEffect, useImperativeHandle, useRef, useState, type FormEvent } from 'react'

import type { SandboxFileSystemNode } from './types'

export type CreateFolderSheetHandle = {
  close: () => void
  open: (parentPath: string) => void
}

export function CreateFolderSheet({
  getParentNode,
  isPending,
  onCreate,
  ref,
}: {
  getParentNode: (parentPath: string) => SandboxFileSystemNode
  isPending: boolean
  onCreate: (value: { name: string; parentPath: string }) => Promise<void>
  ref?: Ref<CreateFolderSheetHandle>
}) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [isOpen, setIsOpen] = useState(false)
  const [name, setName] = useState('')
  const [parentPath, setParentPath] = useState<string | null>(null)

  const resetState = () => {
    setIsOpen(false)
    setName('')
    setParentPath(null)
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
      open: (nextParentPath: string) => {
        setParentPath(nextParentPath)
        setName('')
        setIsOpen(true)
      },
    }),
    [isPending],
  )

  useEffect(() => {
    if (!isOpen) {
      return
    }

    const frameId = requestAnimationFrame(() => {
      inputRef.current?.focus()
    })

    return () => cancelAnimationFrame(frameId)
  }, [isOpen])

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

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    event.stopPropagation()

    const trimmedName = name.trim()
    if (!parentPath || !trimmedName || isPending) {
      return
    }

    await onCreate({
      name: trimmedName,
      parentPath,
    })
    resetState()
  }

  const parentNode = parentPath ? getParentNode(parentPath) : null

  return (
    <Sheet open={isOpen} onOpenChange={handleOpenChange}>
      <SheetContent side="right" className="w-dvw flex flex-col gap-0 p-0 sm:w-[400px]">
        <SheetHeader className="flex flex-row items-center border-b border-border p-4 px-5 text-left">
          <SheetTitle className="text-2xl">Create folder</SheetTitle>
          <SheetDescription className="sr-only">Create a folder in {parentNode?.path}</SheetDescription>
        </SheetHeader>
        <div className="flex-1 overflow-y-auto p-5">
          <p className="mb-4 break-all text-sm text-muted-foreground">{parentNode?.path}</p>
          <form id="create-folder-form" onSubmit={handleSubmit}>
            <Field>
              <FieldLabel htmlFor="create-folder-name">Folder name</FieldLabel>
              <Input
                id="create-folder-name"
                ref={inputRef}
                autoFocus
                value={name}
                onChange={(event) => setName(event.target.value)}
                placeholder="Name your folder"
              />
            </Field>
          </form>
        </div>
        <SheetFooter className="mt-auto border-t border-border p-4 px-5">
          <Button type="button" variant="secondary" disabled={isPending} onClick={() => handleOpenChange(false)}>
            Close
          </Button>
          <Button type="submit" form="create-folder-form" disabled={!name.trim() || isPending || !parentPath}>
            {isPending && <Spinner />}
            Create
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
