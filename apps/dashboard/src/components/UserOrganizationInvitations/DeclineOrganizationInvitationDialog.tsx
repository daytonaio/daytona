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

interface DeclineOrganizationInvitationDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onDeclineInvitation: () => Promise<boolean>
  loading: boolean
}

export const DeclineOrganizationInvitationDialog: React.FC<DeclineOrganizationInvitationDialogProps> = ({
  open,
  onOpenChange,
  onDeclineInvitation,
  loading,
}) => {
  const handleDeclineInvitation = async () => {
    const success = await onDeclineInvitation()
    if (success) {
      onOpenChange(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Decline Invitation</DialogTitle>
          <DialogDescription>
            Are you sure you want to decline this invitation to join the organization?
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
              Declining...
            </Button>
          ) : (
            <Button variant="destructive" onClick={handleDeclineInvitation}>
              Decline
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
