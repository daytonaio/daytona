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

interface RemoveOrganizationMemberDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onRemoveMember: () => Promise<boolean>
  loading: boolean
}

export const RemoveOrganizationMemberDialog: React.FC<RemoveOrganizationMemberDialogProps> = ({
  open,
  onOpenChange,
  onRemoveMember,
  loading,
}) => {
  const handleRemoveMember = async () => {
    const success = await onRemoveMember()
    if (success) {
      onOpenChange(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Remove Member</DialogTitle>
          <DialogDescription>Are you sure you want to remove this member from the organization?</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="secondary" disabled={loading}>
              Cancel
            </Button>
          </DialogClose>
          {loading ? (
            <Button type="button" variant="default" disabled>
              Removing...
            </Button>
          ) : (
            <Button variant="destructive" onClick={handleRemoveMember}>
              Remove
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
