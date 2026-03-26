/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState } from 'react'
import {
  UpdateOrganizationInvitationRoleEnum,
  OrganizationRole,
  OrganizationInvitation,
  OrganizationInvitationRoleEnum,
} from '@daytonaio/api-client'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { ViewerOrganizationRoleCheckbox } from '@/components/OrganizationMembers/ViewerOrganizationRoleCheckbox'

interface UpdateOrganizationInvitationDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  invitation: OrganizationInvitation
  availableRoles: OrganizationRole[]
  loadingAvailableRoles: boolean
  onUpdateInvitation: (role: UpdateOrganizationInvitationRoleEnum, assignedRoleIds: string[]) => Promise<boolean>
}

export const UpdateOrganizationInvitationDialog: React.FC<UpdateOrganizationInvitationDialogProps> = ({
  open,
  onOpenChange,
  invitation,
  availableRoles,
  loadingAvailableRoles,
  onUpdateInvitation,
}) => {
  const [role, setRole] = useState<OrganizationInvitationRoleEnum>(invitation.role)
  const [assignedRoleIds, setAssignedRoleIds] = useState<string[]>(invitation.assignedRoles.map((role) => role.id))
  const [loading, setLoading] = useState(false)

  const handleRoleAssignmentToggle = (roleId: string) => {
    setAssignedRoleIds((current) => {
      if (current.includes(roleId)) {
        return current.filter((p) => p !== roleId)
      } else {
        return [...current, roleId]
      }
    })
  }

  const handleUpdateInvitation = async () => {
    setLoading(true)
    const success = await onUpdateInvitation(role, role === OrganizationInvitationRoleEnum.OWNER ? [] : assignedRoleIds)
    if (success) {
      onOpenChange(false)
      setRole(invitation.role)
      setAssignedRoleIds(invitation.assignedRoles.map((role) => role.id))
    }
    setLoading(false)
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        onOpenChange(isOpen)
        if (!isOpen) {
          setRole(invitation.role)
          setRole(invitation.role)
          setAssignedRoleIds(invitation.assignedRoles.map((role) => role.id))
        }
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Update Invitation</DialogTitle>
          <DialogDescription>Modify organization access for the invited member.</DialogDescription>
        </DialogHeader>
        <div className="space-y-6 overflow-y-auto px-1 pb-1">
          <div className="space-y-3">
            <Label htmlFor="email">Email</Label>
            <Input value={invitation.email} type="email" disabled readOnly />
          </div>

          <div className="space-y-3">
            <Label htmlFor="role">Role</Label>
            <RadioGroup
              className="gap-6"
              value={role}
              onValueChange={(value: OrganizationInvitationRoleEnum) => setRole(value)}
            >
              <div className="flex items-center space-x-4">
                <RadioGroupItem value={OrganizationInvitationRoleEnum.OWNER} id="role-owner" />
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
                <RadioGroupItem value={OrganizationInvitationRoleEnum.MEMBER} id="role-member" />
                <div className="space-y-1">
                  <Label htmlFor="role-member" className="font-normal">
                    Member
                  </Label>
                  <p className="text-sm text-gray-500">Access to organization resources is based on assignments</p>
                </div>
              </div>
            </RadioGroup>
          </div>

          {role === OrganizationInvitationRoleEnum.MEMBER && !loadingAvailableRoles && (
            <div className="space-y-3">
              <Label htmlFor="assignments">Assignments</Label>
              <div className="space-y-6">
                <ViewerOrganizationRoleCheckbox />
                {availableRoles.map((availableRole) => (
                  <div key={availableRole.id} className="flex items-center space-x-4">
                    <Checkbox
                      id={`role-${availableRole.id}`}
                      checked={assignedRoleIds.includes(availableRole.id)}
                      onCheckedChange={() => handleRoleAssignmentToggle(availableRole.id)}
                    />
                    <div className="space-y-1">
                      <Label htmlFor={`role-${availableRole.id}`} className="font-normal">
                        {availableRole.name}
                      </Label>
                      {availableRole.description && (
                        <p className="text-sm text-gray-500">{availableRole.description}</p>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="secondary" disabled={loading}>
              Cancel
            </Button>
          </DialogClose>
          {loading ? (
            <Button type="button" variant="default" disabled>
              Updating...
            </Button>
          ) : (
            <Button type="button" variant="default" onClick={handleUpdateInvitation}>
              Update
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
