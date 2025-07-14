/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState } from 'react'
import { toast } from 'sonner'
import { Copy } from 'lucide-react'
import { useApi } from '@/hooks/useApi'
import { useOrganizations } from '@/hooks/useOrganizations'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { Input } from '@/components/ui/input'
import { DeleteOrganizationDialog } from '@/components/Organizations/DeleteOrganizationDialog'
import { OrganizationUserRoleEnum } from '@daytonaio/api-client'
import { LeaveOrganizationDialog } from '@/components/Organizations/LeaveOrganizationDialog'
import { Label } from '@/components/ui/label'
import { handleApiError } from '@/lib/error-handling'

const OrganizationSettings: React.FC = () => {
  const { organizationsApi } = useApi()

  const { refreshOrganizations } = useOrganizations()
  const { selectedOrganization, authenticatedUserOrganizationMember } = useSelectedOrganization()

  const [loadingDeleteOrganization, setLoadingDeleteOrganization] = useState(false)
  const [loadingLeaveOrganization, setLoadingLeaveOrganization] = useState(false)

  const handleDeleteOrganization = async () => {
    if (!selectedOrganization) {
      return false
    }
    setLoadingDeleteOrganization(true)
    try {
      await organizationsApi.deleteOrganization(selectedOrganization.id)
      toast.success('Organization deleted successfully')
      await refreshOrganizations()
      return true
    } catch (error) {
      handleApiError(error, 'Failed to delete organization')
      return false
    } finally {
      setLoadingDeleteOrganization(false)
    }
  }

  const handleLeaveOrganization = async () => {
    if (!selectedOrganization) {
      return false
    }
    setLoadingLeaveOrganization(true)
    try {
      await organizationsApi.leaveOrganization(selectedOrganization.id)
      toast.success('Organization left successfully')
      await refreshOrganizations()
      return true
    } catch (error) {
      handleApiError(error, 'Failed to leave organization')
      return false
    } finally {
      setLoadingLeaveOrganization(false)
    }
  }

  if (!selectedOrganization) {
    return null
  }

  return (
    <div className="px-6 py-2">
      <div className="mb-2 h-12 flex items-center justify-between">
        <h1 className="text-2xl font-medium">Organization Settings</h1>
      </div>

      <div className="max-w-2xl mt-4 space-y-6">
        <div className="space-y-3">
          <Label htmlFor="organization-id">Organization ID</Label>
          <div className="relative">
            <Input id="organization-id" value={selectedOrganization.id} readOnly />
            <button
              className="absolute right-2 top-1/2 -translate-y-1/2 p-1 hover:bg-muted rounded"
              onClick={() => {
                navigator.clipboard.writeText(selectedOrganization.id)
                toast.success('Copied to clipboard')
              }}
            >
              <Copy className="h-4 w-4" />
            </button>
          </div>
        </div>

        <div className="space-y-3">
          <Label htmlFor="organization-name">Organization Name</Label>
          <Input id="organization-name" value={selectedOrganization.name} readOnly />
        </div>

        {!selectedOrganization.personal && authenticatedUserOrganizationMember !== null && (
          <div className="space-y-3">
            <h2 className="text-lg font-semibold">Danger Zone</h2>
            {authenticatedUserOrganizationMember.role === OrganizationUserRoleEnum.OWNER ? (
              <DeleteOrganizationDialog
                organizationName={selectedOrganization.name}
                onDeleteOrganization={handleDeleteOrganization}
                loading={loadingDeleteOrganization}
              />
            ) : (
              <LeaveOrganizationDialog
                onLeaveOrganization={handleLeaveOrganization}
                loading={loadingLeaveOrganization}
              />
            )}
          </div>
        )}
      </div>
    </div>
  )
}

export default OrganizationSettings
