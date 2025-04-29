/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
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

interface CancelOrganizationInvitationDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onCancelInvitation: () => Promise<boolean>
  loading: boolean
}

export const CancelOrganizationInvitationDialog: React.FC<CancelOrganizationInvitationDialogProps> = ({
  open,
  onOpenChange,
  onCancelInvitation,
  loading,
}) => {
  const handleCancelInvitation = async () => {
    const success = await onCancelInvitation()
    if (success) {
      onOpenChange(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Cancel Invitation</DialogTitle>
          <DialogDescription>
            Are you sure you want to cancel this invitation to join the organization?
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="secondary" disabled={loading}>
              Cancel
            </Button>
          </DialogClose>
          {loading ? (
            <Button type="button" variant="default" disabled>
              Confirming...
            </Button>
          ) : (
            <Button variant="destructive" onClick={handleCancelInvitation}>
              Confirm
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
