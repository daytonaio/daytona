/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState } from 'react'
import {
  CreateOrganizationInvitationRoleEnum,
  OrganizationUserRoleEnum,
  UpdateOrganizationMemberRoleRoleEnum,
} from '@daytonaio/api-client'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'

interface UpdateOrganizationMemberRoleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  initialRole: OrganizationUserRoleEnum
  onUpdateMemberRole: (role: UpdateOrganizationMemberRoleRoleEnum) => Promise<boolean>
  loading: boolean
}

export const UpdateOrganizationMemberRoleDialog: React.FC<UpdateOrganizationMemberRoleDialogProps> = ({
  open,
  onOpenChange,
  initialRole,
  onUpdateMemberRole,
  loading,
}) => {
  const [role, setRole] = useState<CreateOrganizationInvitationRoleEnum>(initialRole)

  const handleUpdateMemberRole = async () => {
    const success = await onUpdateMemberRole(role)
    if (success) {
      onOpenChange(false)
      setRole(initialRole)
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        onOpenChange(isOpen)
        if (!isOpen) {
          setRole(initialRole)
        }
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Change Role</DialogTitle>
          <DialogDescription>Choose an appropriate role for the organization member.</DialogDescription>
        </DialogHeader>
        <div className="space-y-6 overflow-y-auto px-1 pb-1">
          <RadioGroup
            className="gap-6"
            value={role}
            onValueChange={(value: CreateOrganizationInvitationRoleEnum) => setRole(value)}
          >
            <div className="flex items-center space-x-4">
              <RadioGroupItem value={CreateOrganizationInvitationRoleEnum.OWNER} id="role-owner" />
              <div className="space-y-1">
                <Label htmlFor="role-owner" className="font-normal">
                  Owner
                </Label>
                <p className="text-sm text-gray-500">
                  Full administrative access to the organization and its resources
                </p>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <RadioGroupItem value={CreateOrganizationInvitationRoleEnum.MEMBER} id="role-member" />
              <div className="space-y-1">
                <Label htmlFor="role-member" className="font-normal">
                  Member
                </Label>
                <p className="text-sm text-gray-500">Access to organization resources is based on assignments</p>
              </div>
            </div>
          </RadioGroup>
        </div>

        <DialogFooter>
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
            <Button type="button" variant="default" onClick={handleUpdateMemberRole}>
              Save
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
