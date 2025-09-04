/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState } from 'react'
import { CreateOrganizationInvitationRoleEnum, OrganizationRole, OrganizationUserRoleEnum } from '@daytonaio/api-client'
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
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { ViewerOrganizationRoleCheckbox } from '@/components/OrganizationMembers/ViewerOrganizationRoleCheckbox'

interface UpdateOrganizationMemberAccessProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  initialRole: OrganizationUserRoleEnum
  initialAssignments: OrganizationRole[]
  availableAssignments: OrganizationRole[]
  loadingAvailableAssignments: boolean
  onUpdateAccess: (role: OrganizationUserRoleEnum, assignedRoleIds: string[]) => Promise<boolean>
  processingUpdateAccess: boolean
}

export const UpdateOrganizationMemberAccess: React.FC<UpdateOrganizationMemberAccessProps> = ({
  open,
  onOpenChange,
  initialRole,
  initialAssignments,
  availableAssignments,
  loadingAvailableAssignments,
  onUpdateAccess,
  processingUpdateAccess,
}) => {
  const [role, setRole] = useState<OrganizationUserRoleEnum>(initialRole)
  const [assignedRoleIds, setAssignedRoleIds] = useState<string[]>(initialAssignments.map((a) => a.id))

  const handleRoleAssignmentToggle = (roleId: string) => {
    setAssignedRoleIds((current) => {
      if (current.includes(roleId)) {
        return current.filter((p) => p !== roleId)
      } else {
        return [...current, roleId]
      }
    })
  }

  const handleUpdateAccess = async () => {
    const success = await onUpdateAccess(role, role === OrganizationUserRoleEnum.OWNER ? [] : assignedRoleIds)
    if (success) {
      onOpenChange(false)
      setRole(initialRole)
      setAssignedRoleIds(initialAssignments.map((a) => a.id))
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        onOpenChange(isOpen)
        if (!isOpen) {
          setRole(initialRole)
          setAssignedRoleIds(initialAssignments.map((a) => a.id))
        }
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Update Access</DialogTitle>
          <DialogDescription>
            Manage access to the organization with an appropriate role and assignments.
          </DialogDescription>
          {role !== OrganizationUserRoleEnum.OWNER && (
            <DialogDescription className="text-yellow-600 dark:text-yellow-400">
              Removing assignments will automatically revoke any API keys this member created using permissions based on
              those assignments.
            </DialogDescription>
          )}
        </DialogHeader>
        <form
          id="update-access-form"
          className="space-y-6 overflow-y-auto px-1 pb-1"
          onSubmit={async (e) => {
            e.preventDefault()
            await handleUpdateAccess()
          }}
        >
          <div className="space-y-3">
            <Label htmlFor="role">Role</Label>
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

          {role === CreateOrganizationInvitationRoleEnum.MEMBER &&
            !loadingAvailableAssignments &&
            availableAssignments.length > 0 && (
              <div className="space-y-3">
                <Label htmlFor="assignments">Assignments</Label>
                <div className="space-y-6">
                  <ViewerOrganizationRoleCheckbox />
                  {availableAssignments.map((assignment) => (
                    <div key={assignment.id} className="flex items-center space-x-4">
                      <Checkbox
                        id={`role-${assignment.id}`}
                        checked={assignedRoleIds.includes(assignment.id)}
                        onCheckedChange={() => handleRoleAssignmentToggle(assignment.id)}
                      />
                      <div className="space-y-1">
                        <Label htmlFor={`role-${assignment.id}`} className="font-normal">
                          {assignment.name}
                        </Label>
                        {assignment.description && <p className="text-sm text-gray-500">{assignment.description}</p>}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
        </form>
        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="secondary" disabled={processingUpdateAccess}>
              Cancel
            </Button>
          </DialogClose>
          {processingUpdateAccess ? (
            <Button type="button" variant="default" disabled>
              Saving...
            </Button>
          ) : (
            <Button type="submit" form="update-access-form" variant="default" disabled={loadingAvailableAssignments}>
              Save
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
