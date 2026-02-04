/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

interface DeleteAccountDialogProps {
  onDeleteAccount: () => Promise<{ success: boolean; reasons: string[] }>
  loading: boolean
}

const CONFIRM_TEXT = 'DELETE'

export const DeleteAccountDialog: React.FC<DeleteAccountDialogProps> = ({ onDeleteAccount, loading }) => {
  const [open, setOpen] = useState(false)
  const [confirmText, setConfirmText] = useState('')
  const [errorReasons, setErrorReasons] = useState<string[]>([])

  const handleDeleteAccount = async () => {
    setErrorReasons([])
    const result = await onDeleteAccount()
    if (result.success) {
      setOpen(false)
      setConfirmText('')
      setErrorReasons([])
    } else {
      setErrorReasons(result.reasons)
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        setOpen(isOpen)
        if (!isOpen) {
          setConfirmText('')
          setErrorReasons([])
        }
      }}
    >
      <DialogTrigger asChild>
        <Button variant="destructive" className="w-auto px-4">
          Delete Account
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete Account</DialogTitle>
          <DialogDescription>
            This will permanently delete your account and all associated data including sandboxes, snapshots, and
            organizations you created. This action cannot be undone.
          </DialogDescription>
        </DialogHeader>
        {errorReasons.length > 0 && (
          <div className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive-foreground">
            <ul className="mt-2 list-disc space-y-1 pl-5">
              {errorReasons.map((reason, index) => (
                <li key={`${reason}-${index}`}>{reason}</li>
              ))}
            </ul>
          </div>
        )}
        <form
          id="delete-account-form"
          className="space-y-6 overflow-y-auto px-1 pb-1"
          onSubmit={async (e) => {
            e.preventDefault()
            await handleDeleteAccount()
          }}
        >
          <div className="space-y-6">
            <div className="space-y-3">
              <Label htmlFor="confirm-action">
                Please type <span className="font-bold cursor-text select-all">{CONFIRM_TEXT}</span> to confirm
              </Label>
              <Input
                id="confirm-action"
                value={confirmText}
                onChange={(e) => setConfirmText(e.target.value)}
                placeholder={CONFIRM_TEXT}
              />
            </div>
          </div>
        </form>
        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="secondary" disabled={loading}>
              Cancel
            </Button>
          </DialogClose>
          {loading ? (
            <Button type="button" variant="destructive" disabled>
              Deleting...
            </Button>
          ) : (
            <Button
              type="submit"
              form="delete-account-form"
              variant="destructive"
              disabled={confirmText !== CONFIRM_TEXT}
            >
              Delete Account
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
