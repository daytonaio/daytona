/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState } from 'react'
import { OrganizationRole } from '@daytonaio/api-client'
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
import { Checkbox } from '@/components/ui/checkbox'
import { ViewerOrganizationRoleCheckbox } from '@/components/OrganizationMembers/ViewerOrganizationRoleCheckbox'
import { Label } from '@/components/ui/label'

interface UpdateAssignedOrganizationRolesDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  initialData: OrganizationRole[]
  availableRoles: OrganizationRole[]
  loadingAvailableRoles: boolean
  onUpdateAssignedRoles: (roleIds: string[]) => Promise<boolean>
  loading: boolean
}

export const UpdateAssignedOrganizationRolesDialog: React.FC<UpdateAssignedOrganizationRolesDialogProps> = ({
  open,
  onOpenChange,
  initialData,
  availableRoles,
  loadingAvailableRoles,
  onUpdateAssignedRoles,
  loading,
}) => {
  const [roleIds, setRoleIds] = useState(initialData.map((role) => role.id))

  const handleUpdateAssignedRoles = async () => {
    const success = await onUpdateAssignedRoles(roleIds)
    if (success) {
      onOpenChange(false)
      setRoleIds(initialData.map((role) => role.id))
    }
  }

  const handleRoleToggle = (roleId: string) => {
    setRoleIds((current) => {
      if (current.includes(roleId)) {
        return current.filter((p) => p !== roleId)
      } else {
        return [...current, roleId]
      }
    })
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        onOpenChange(isOpen)
        if (!isOpen) {
          setRoleIds(initialData.map((role) => role.id))
        }
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Assign Roles</DialogTitle>
          <DialogDescription>
            {availableRoles.length > 0
              ? 'Select the roles you want to assign to this user.'
              : 'No roles are available for assignment.'}
          </DialogDescription>
        </DialogHeader>
        {!loadingAvailableRoles && availableRoles.length > 0 && (
          <div className="space-y-6 overflow-y-auto px-1 pb-1">
            <ViewerOrganizationRoleCheckbox />
            {availableRoles.map((role) => (
              <div key={role.id} className="flex items-center space-x-4">
                <Checkbox
                  id={role.id}
                  checked={roleIds.includes(role.id)}
                  onCheckedChange={() => handleRoleToggle(role.id)}
                />
                <div className="space-y-1">
                  <Label htmlFor={role.id} className="font-normal">
                    {role.name}
                  </Label>
                  {role.description && <p className="text-sm text-gray-500">{role.description}</p>}
                </div>
              </div>
            ))}
          </div>
        )}
        <DialogFooter>
          {!loadingAvailableRoles && availableRoles.length > 0 ? (
            <>
              <DialogClose asChild>
                <Button type="button" variant="secondary" disabled={loading}>
                  Cancel
                </Button>
              </DialogClose>
              {loading ? (
                <Button type="button" variant="default" disabled>
                  Saving...
                </Button>
              ) : (
                <Button type="button" variant="default" onClick={handleUpdateAssignedRoles}>
                  Save
                </Button>
              )}
            </>
          ) : (
            <DialogClose asChild>
              <Button type="button" variant="secondary">
                Close
              </Button>
            </DialogClose>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
