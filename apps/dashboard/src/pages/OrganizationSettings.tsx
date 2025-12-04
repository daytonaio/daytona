/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState, useEffect } from 'react'
import { toast } from 'sonner'
import { Copy } from 'lucide-react'
import { useApi } from '@/hooks/useApi'
import { useOrganizations } from '@/hooks/useOrganizations'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { DeleteOrganizationDialog } from '@/components/Organizations/DeleteOrganizationDialog'
import { OrganizationUserRoleEnum } from '@daytonaio/api-client'
import { LeaveOrganizationDialog } from '@/components/Organizations/LeaveOrganizationDialog'
import { SetDefaultRegionDialog } from '@/components/Organizations/SetDefaultRegionDialog'
import { Label } from '@/components/ui/label'
import { handleApiError } from '@/lib/error-handling'
import { useRegions } from '@/hooks/useRegions'

const OrganizationSettings: React.FC = () => {
  const { organizationsApi } = useApi()

  const { refreshOrganizations } = useOrganizations()
  const { selectedOrganization, authenticatedUserOrganizationMember } = useSelectedOrganization()
  const { getRegionName, sharedRegions, loadingRegions } = useRegions()

  const [loadingDeleteOrganization, setLoadingDeleteOrganization] = useState(false)
  const [loadingLeaveOrganization, setLoadingLeaveOrganization] = useState(false)
  const [showSetDefaultRegionDialog, setSetDefaultRegionDialog] = useState(false)

  useEffect(() => {
    if (selectedOrganization && !selectedOrganization.defaultRegionId) {
      setSetDefaultRegionDialog(true)
    }
  }, [selectedOrganization])

  const handleSetDefaultRegion = async (defaultRegionId: string): Promise<boolean> => {
    if (!selectedOrganization) {
      return false
    }

    try {
      await organizationsApi.setOrganizationDefaultRegion(selectedOrganization.id, {
        defaultRegionId,
      })
      toast.success('Default region set successfully')
      await refreshOrganizations(selectedOrganization.id)
      setSetDefaultRegionDialog(false)
      return true
    } catch (error) {
      handleApiError(error, 'Failed to set default region')
      return false
    }
  }

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

        <div className="space-y-3">
          <Label htmlFor="organization-default-region">Default Region</Label>
          {selectedOrganization.defaultRegionId ? (
            <Input
              id="organization-default-region"
              value={getRegionName(selectedOrganization.defaultRegionId) ?? selectedOrganization.defaultRegionId}
              readOnly
            />
          ) : authenticatedUserOrganizationMember !== null &&
            authenticatedUserOrganizationMember.role === OrganizationUserRoleEnum.OWNER ? (
            <div>
              <Button onClick={() => setSetDefaultRegionDialog(true)} variant="outline">
                Set Default Region
              </Button>
            </div>
          ) : null}
          <p className="text-sm text-muted-foreground mt-1 pl-1">
            The region that is used as the default target for creating sandboxes in this organization.
          </p>
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

      <SetDefaultRegionDialog
        open={showSetDefaultRegionDialog}
        onOpenChange={setSetDefaultRegionDialog}
        regions={sharedRegions}
        loadingRegions={loadingRegions}
        onSetDefaultRegion={handleSetDefaultRegion}
      />
    </div>
  )
}

export default OrganizationSettings
