/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState } from 'react'
import { OrganizationRole, OrganizationRolePermissionsEnum } from '@daytonaio/api-client'
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
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ORGANIZATION_ROLE_PERMISSIONS_GROUPS } from '@/constants/OrganizationPermissionsGroups'
import { OrganizationRolePermissionGroup } from '@/types/OrganizationRolePermissionGroup'

interface UpdateOrganizationRoleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  initialData: OrganizationRole
  onUpdateRole: (name: string, description: string, permissions: OrganizationRolePermissionsEnum[]) => Promise<boolean>
}

export const UpdateOrganizationRoleDialog: React.FC<UpdateOrganizationRoleDialogProps> = ({
  open,
  onOpenChange,
  initialData,
  onUpdateRole,
}) => {
  const [name, setName] = useState(initialData.name)
  const [description, setDescription] = useState(initialData.description)
  const [permissions, setPermissions] = useState(initialData.permissions)
  const [loading, setLoading] = useState(false)

  const handleUpdateRole = async () => {
    setLoading(true)
    const success = await onUpdateRole(name, description, permissions)
    if (success) {
      onOpenChange(false)
      setName('')
      setDescription('')
      setPermissions([])
    }
    setLoading(false)
  }

  const isGroupChecked = (group: OrganizationRolePermissionGroup) => {
    return group.permissions.every((permission) => permissions.includes(permission))
  }

  // Toggle all permissions in a group
  const handleGroupToggle = (group: OrganizationRolePermissionGroup) => {
    if (isGroupChecked(group)) {
      // If all checked, uncheck all
      setPermissions((current) => current.filter((p) => !group.permissions.includes(p)))
    } else {
      // If not all checked, check all
      setPermissions((current) => {
        const newPermissions = [...current]
        group.permissions.forEach((key) => {
          if (!newPermissions.includes(key)) {
            newPermissions.push(key)
          }
        })
        return newPermissions
      })
    }
  }

  // Toggle a single permission
  const handlePermissionToggle = (permission: OrganizationRolePermissionsEnum) => {
    setPermissions((current) => {
      if (current.includes(permission)) {
        return current.filter((p) => p !== permission)
      } else {
        return [...current, permission]
      }
    })
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        onOpenChange(isOpen)
        if (!isOpen) {
          setName('')
          setDescription('')
          setPermissions([])
        }
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Role</DialogTitle>
          <DialogDescription>Modify permissions for the custom organization role.</DialogDescription>
        </DialogHeader>
        <form
          id="edit-role-form"
          className="space-y-6 overflow-y-auto px-1 pb-1"
          onSubmit={async (e) => {
            e.preventDefault()
            await handleUpdateRole()
          }}
        >
          <div className="space-y-3">
            <Label htmlFor="name">Name</Label>
            <Input id="name" value={name} onChange={(e) => setName(e.target.value)} placeholder="Name" />
          </div>
          <div className="space-y-3">
            <Label htmlFor="description">Description</Label>
            <Input
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Description"
            />
          </div>
          <div className="space-y-3">
            <Label htmlFor="permissions">Permissions</Label>
            <div className="space-y-6">
              {ORGANIZATION_ROLE_PERMISSIONS_GROUPS.map((group) => {
                const groupIsChecked = isGroupChecked(group)

                return (
                  <div key={group.name} className="space-y-3">
                    <div className="flex items-center space-x-2">
                      <Checkbox
                        id={`group-${group.name}`}
                        checked={groupIsChecked}
                        onCheckedChange={() => handleGroupToggle(group)}
                      />
                      <Label htmlFor={`group-${group.name}`} className="font-normal">
                        {group.name}
                      </Label>
                    </div>
                    <div className="ml-6 space-y-2">
                      {group.permissions.map((permission) => (
                        <div key={permission} className="flex items-center space-x-2">
                          <Checkbox
                            id={permission}
                            checked={permissions.includes(permission)}
                            onCheckedChange={() => handlePermissionToggle(permission)}
                            disabled={groupIsChecked}
                            className={`${groupIsChecked ? 'pointer-events-none' : ''}`}
                          />
                          <Label
                            htmlFor={permission}
                            className={`font-normal${groupIsChecked ? ' opacity-70 pointer-events-none' : ''}`}
                          >
                            {permission}
                          </Label>
                        </div>
                      ))}
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        </form>
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
            <Button
              type="submit"
              form="edit-role-form"
              variant="default"
              disabled={!name.trim() || !description.trim()}
            >
              Save
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
