/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useState } from 'react'
import { toast } from 'sonner'
import { Field, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Spinner } from '@/components/ui/spinner'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useRevokeSshAccessMutation } from '@/hooks/mutations/useRevokeSshAccessMutation'
import { handleApiError } from '@/lib/error-handling'

interface RevokeSshAccessDialogProps {
  sandboxId: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function RevokeSshAccessDialog({ sandboxId, open, onOpenChange }: RevokeSshAccessDialogProps) {
  const [token, setToken] = useState('')
  const revokeMutation = useRevokeSshAccessMutation()

  const handleOpenChange = (isOpen: boolean) => {
    onOpenChange(isOpen)
    if (!isOpen) {
      setToken('')
      revokeMutation.reset()
    }
  }

  const handleRevoke = async () => {
    if (!token.trim()) {
      toast.error('Please enter a token to revoke')
      return
    }
    try {
      await revokeMutation.mutateAsync({ sandboxId, token })
      toast.success('SSH access revoked successfully')
      handleOpenChange(false)
    } catch (error) {
      handleApiError(error, 'Failed to revoke SSH access')
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>Revoke SSH Access</DialogTitle>
          <DialogDescription>Enter the SSH access token you want to revoke.</DialogDescription>
        </DialogHeader>
        <Field>
          <FieldLabel htmlFor="ssh-revoke-token">SSH Token</FieldLabel>
          <Input
            id="ssh-revoke-token"
            value={token}
            onChange={(e) => setToken(e.target.value)}
            placeholder="Paste token here"
          />
        </Field>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="secondary">Cancel</Button>
          </DialogClose>
          <Button variant="destructive" onClick={handleRevoke} disabled={!token.trim() || revokeMutation.isPending}>
            {revokeMutation.isPending && <Spinner />}
            Revoke
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
