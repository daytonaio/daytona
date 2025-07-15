/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState } from 'react'
import { Plus } from 'lucide-react'
import { OrganizationRolePermissionsEnum } from '@daytonaio/api-client'
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
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ORGANIZATION_ROLE_PERMISSIONS_GROUPS } from '@/constants/OrganizationPermissionsGroups'
import { OrganizationRolePermissionGroup } from '@/types/OrganizationRolePermissionGroup'

interface CreateOrganizationRoleDialogProps {
  onCreateRole: (name: string, description: string, permissions: OrganizationRolePermissionsEnum[]) => Promise<boolean>
}

export const CreateOrganizationRoleDialog: React.FC<CreateOrganizationRoleDialogProps> = ({ onCreateRole }) => {
  const [open, setOpen] = useState(false)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [permissions, setPermissions] = useState<OrganizationRolePermissionsEnum[]>([])
  const [loading, setLoading] = useState(false)

  const handleCreateRole = async () => {
    setLoading(true)
    const success = await onCreateRole(name, description, permissions)
    if (success) {
      setOpen(false)
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
        setOpen(isOpen)
        if (!isOpen) {
          setName('')
          setDescription('')
          setPermissions([])
        }
      }}
    >
      <DialogTrigger asChild>
        <Button variant="default" size="sm" className="w-auto px-4" title="Add Registry">
          <Plus className="w-4 h-4" />
          Create Role
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Role</DialogTitle>
          <DialogDescription>Define a custom role for managing access to the organization.</DialogDescription>
        </DialogHeader>
        <form
          id="create-role-form"
          className="space-y-6 overflow-y-auto px-1 pb-1"
          onSubmit={async (e) => {
            e.preventDefault()
            await handleCreateRole()
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
              Creating...
            </Button>
          ) : (
            <Button
              type="submit"
              form="create-role-form"
              variant="default"
              disabled={!name.trim() || !description.trim() || permissions.length === 0}
            >
              Create
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
