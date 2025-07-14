/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useEffect, useState } from 'react'
import { Plus } from 'lucide-react'
import { CreateOrganizationInvitationRoleEnum, OrganizationRole } from '@daytonaio/api-client'
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
  DialogTrigger,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { ViewerOrganizationRoleCheckbox } from '@/components/OrganizationMembers/ViewerOrganizationRoleCheckbox'

interface CreateOrganizationInvitationDialogProps {
  availableRoles: OrganizationRole[]
  loadingAvailableRoles: boolean
  onCreateInvitation: (
    email: string,
    role: CreateOrganizationInvitationRoleEnum,
    assignedRoleIds: string[],
  ) => Promise<boolean>
}

export const CreateOrganizationInvitationDialog: React.FC<CreateOrganizationInvitationDialogProps> = ({
  availableRoles,
  loadingAvailableRoles,
  onCreateInvitation,
}) => {
  const [open, setOpen] = useState(false)
  const [email, setEmail] = useState('')
  const [role, setRole] = useState<CreateOrganizationInvitationRoleEnum>(CreateOrganizationInvitationRoleEnum.MEMBER)
  const [assignedRoleIds, setAssignedRoleIds] = useState<string[]>([])
  const [loading, setLoading] = useState(false)

  const [developerRole, setDeveloperRole] = useState<OrganizationRole | null>(null)

  useEffect(() => {
    if (!loadingAvailableRoles) {
      const developerRole = availableRoles.find((r) => r.name === 'Developer')
      if (developerRole) {
        setDeveloperRole(developerRole)
        setAssignedRoleIds([developerRole.id])
      }
    }
  }, [loadingAvailableRoles, availableRoles])

  const handleRoleAssignmentToggle = (roleId: string) => {
    setAssignedRoleIds((current) => {
      if (current.includes(roleId)) {
        return current.filter((p) => p !== roleId)
      } else {
        return [...current, roleId]
      }
    })
  }

  const handleCreateInvitation = async () => {
    setLoading(true)
    const success = await onCreateInvitation(
      email,
      role,
      role === CreateOrganizationInvitationRoleEnum.OWNER ? [] : assignedRoleIds,
    )
    if (success) {
      setOpen(false)
      setEmail('')
      setRole(CreateOrganizationInvitationRoleEnum.MEMBER)
      if (developerRole) {
        setAssignedRoleIds([developerRole.id])
      } else {
        setAssignedRoleIds([])
      }
    }
    setLoading(false)
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        setOpen(isOpen)
        if (!isOpen) {
          setEmail('')
          setRole(CreateOrganizationInvitationRoleEnum.MEMBER)
          if (developerRole) {
            setAssignedRoleIds([developerRole.id])
          } else {
            setAssignedRoleIds([])
          }
        }
      }}
    >
      <DialogTrigger asChild>
        <Button variant="default" size="sm" className="w-auto px-4" title="Add Registry">
          <Plus className="w-4 h-4" />
          Invite Member
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Invite Member</DialogTitle>
          <DialogDescription>
            Give them access to the organization with an appropriate role and assignments.
          </DialogDescription>
        </DialogHeader>
        <form
          id="invitation-form"
          className="space-y-6 overflow-y-auto px-1 pb-1"
          onSubmit={async (e) => {
            e.preventDefault()
            await handleCreateInvitation()
          }}
        >
          <div className="space-y-3">
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              value={email}
              type="email"
              onChange={(e) => setEmail(e.target.value)}
              placeholder="mail@example.com"
            />
          </div>

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

          {role === CreateOrganizationInvitationRoleEnum.MEMBER && !loadingAvailableRoles && (
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
        </form>

        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="secondary" disabled={loading}>
              Cancel
            </Button>
          </DialogClose>
          {loading ? (
            <Button type="button" variant="default" disabled>
              Inviting...
            </Button>
          ) : (
            <Button type="submit" form="invitation-form" variant="default" disabled={!email.trim()}>
              Invite
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
