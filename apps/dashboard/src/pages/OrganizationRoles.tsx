/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateOrganizationRoleSheet } from '@/components/OrganizationRoles/CreateOrganizationRoleSheet'
import { type CommandConfig, useRegisterCommands } from '@/components/CommandPalette'
import { OrganizationRoleTable } from '@/components/OrganizationRoles/OrganizationRoleTable'
import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { useApi } from '@/hooks/useApi'
import { useOrganizationRoles } from '@/hooks/useOrganizationRoles'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { OrganizationRolePermissionsEnum } from '@daytona/api-client'
import { PlusIcon } from 'lucide-react'
import React, { useMemo, useRef, useState } from 'react'
import { toast } from 'sonner'

const OrganizationRoles: React.FC = () => {
  const { organizationsApi } = useApi()

  const { selectedOrganization } = useSelectedOrganization()
  const { roles, loadingRoles, refreshRoles } = useOrganizationRoles()

  const [loadingRoleAction, setLoadingRoleAction] = useState<Record<string, boolean>>({})
  const createRoleSheetRef = useRef<{ open: () => void }>(null)

  const handleCreateRole = async (
    name: string,
    description: string,
    permissions: OrganizationRolePermissionsEnum[],
  ): Promise<boolean> => {
    if (!selectedOrganization) {
      return false
    }
    try {
      await organizationsApi.createOrganizationRole(selectedOrganization.id, {
        name: name.trim(),
        description: description?.trim(),
        permissions,
      })
      toast.success('Role created successfully')
      await refreshRoles(false)
      return true
    } catch (error) {
      handleApiError(error, 'Failed to create role')
      return false
    }
  }

  const handleUpdateRole = async (
    roleId: string,
    name: string,
    description: string,
    permissions: OrganizationRolePermissionsEnum[],
  ): Promise<boolean> => {
    if (!selectedOrganization) {
      return false
    }
    setLoadingRoleAction((prev) => ({ ...prev, [roleId]: true }))
    try {
      await organizationsApi.updateOrganizationRole(selectedOrganization.id, roleId, {
        name: name.trim(),
        description: description?.trim(),
        permissions,
      })
      toast.success('Role updated successfully')
      await refreshRoles(false)
      return true
    } catch (error) {
      handleApiError(error, 'Failed to update role')
      return false
    } finally {
      setLoadingRoleAction((prev) => ({ ...prev, [roleId]: false }))
    }
  }

  const handleDeleteRole = async (roleId: string): Promise<boolean> => {
    if (!selectedOrganization) {
      return false
    }
    setLoadingRoleAction((prev) => ({ ...prev, [roleId]: true }))
    try {
      await organizationsApi.deleteOrganizationRole(selectedOrganization.id, roleId)
      toast.success('Role deleted successfully')
      await refreshRoles(false)
      return true
    } catch (error) {
      handleApiError(error, 'Failed to delete role')
      return false
    } finally {
      setLoadingRoleAction((prev) => ({ ...prev, [roleId]: false }))
    }
  }

  const rootCommands: CommandConfig[] = useMemo(
    () => [
      {
        id: 'create-role',
        label: 'Create Role',
        icon: <PlusIcon className="w-4 h-4" />,
        onSelect: () => createRoleSheetRef.current?.open(),
      },
    ],
    [],
  )

  useRegisterCommands(rootCommands, { groupId: 'role-actions', groupLabel: 'Role actions', groupOrder: 0 })

  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>Roles</PageTitle>
        <CreateOrganizationRoleSheet className="ml-auto" onCreateRole={handleCreateRole} ref={createRoleSheetRef} />
      </PageHeader>

      <PageContent>
        <OrganizationRoleTable
          data={roles}
          loadingData={loadingRoles}
          onUpdateRole={handleUpdateRole}
          onDeleteRole={handleDeleteRole}
          loadingRoleAction={loadingRoleAction}
        />
      </PageContent>
    </PageLayout>
  )
}

export default OrganizationRoles
