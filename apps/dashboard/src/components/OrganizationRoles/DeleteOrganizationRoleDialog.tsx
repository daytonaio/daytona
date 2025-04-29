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

interface DeleteOrganizationRoleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onDeleteRole: () => Promise<boolean>
  loading: boolean
}

export const DeleteOrganizationRoleDialog: React.FC<DeleteOrganizationRoleDialogProps> = ({
  open,
  onOpenChange,
  onDeleteRole,
  loading,
}) => {
  const handleDeleteRole = async () => {
    const success = await onDeleteRole()
    if (success) {
      onOpenChange(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete Role</DialogTitle>
          <DialogDescription>Are you sure you want to delete this role?</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="secondary" disabled={loading}>
              Cancel
            </Button>
          </DialogClose>
          {loading ? (
            <Button type="button" variant="default" disabled>
              Deleting...
            </Button>
          ) : (
            <Button variant="destructive" onClick={handleDeleteRole}>
              Delete
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
