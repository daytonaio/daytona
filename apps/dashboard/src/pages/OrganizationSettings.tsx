/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DeleteOrganizationDialog } from '@/components/Organizations/DeleteOrganizationDialog'
import { LeaveOrganizationDialog } from '@/components/Organizations/LeaveOrganizationDialog'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Field, FieldContent, FieldDescription, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { InputGroup, InputGroupButton, InputGroupInput } from '@/components/ui/input-group'
import { useDeleteOrganizationMutation } from '@/hooks/mutations/useDeleteOrganizationMutation'
import { useLeaveOrganizationMutation } from '@/hooks/mutations/useLeaveOrganizationMutation'
import { useOrganizations } from '@/hooks/useOrganizations'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { OrganizationUserRoleEnum } from '@daytonaio/api-client'
import { CheckIcon, CopyIcon } from 'lucide-react'
import React from 'react'
import { toast } from 'sonner'
import { useCopyToClipboard } from 'usehooks-ts'

const OrganizationSettings: React.FC = () => {
  const { refreshOrganizations } = useOrganizations()
  const { selectedOrganization, authenticatedUserOrganizationMember } = useSelectedOrganization()

  const deleteOrganizationMutation = useDeleteOrganizationMutation()
  const leaveOrganizationMutation = useLeaveOrganizationMutation()
  const [copied, copyToClipboard] = useCopyToClipboard()

  const handleDeleteOrganization = async () => {
    if (!selectedOrganization) {
      return false
    }
    try {
      await deleteOrganizationMutation.mutateAsync({ organizationId: selectedOrganization.id })
      toast.success('Organization deleted successfully')
      await refreshOrganizations()
      return true
    } catch (error) {
      handleApiError(error, 'Failed to delete organization')
      return false
    }
  }

  const handleLeaveOrganization = async () => {
    if (!selectedOrganization) {
      return false
    }
    try {
      await leaveOrganizationMutation.mutateAsync({ organizationId: selectedOrganization.id })
      toast.success('Organization left successfully')
      await refreshOrganizations()
      return true
    } catch (error) {
      handleApiError(error, 'Failed to leave organization')
      return false
    }
  }

  if (!selectedOrganization) {
    return null
  }

  const isOwner = authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER

  return (
    <div className="px-6 py-2 max-w-3xl p-5">
      <div className="mb-2 h-12 flex items-center justify-between">
        <h1 className="text-2xl font-medium">General Settings</h1>
      </div>

      <div className="flex flex-col gap-6 mt-4">
        <Card>
          <CardHeader className="p-4">
            <CardTitle>Organization Details</CardTitle>
          </CardHeader>
          <CardContent className="border-t border-border">
            <Field className="grid sm:grid-cols-2 items-center">
              <FieldContent className="flex-1">
                <FieldLabel htmlFor="organization-name">Organization Name</FieldLabel>
                <FieldDescription>The public name of your organization.</FieldDescription>
              </FieldContent>

              <Input id="organization-name" value={selectedOrganization.name} readOnly className="flex-1" />
            </Field>
          </CardContent>

          <CardContent className="border-t border-border">
            <Field className="grid sm:grid-cols-2 items-center">
              <FieldContent className="flex-1">
                <FieldLabel htmlFor="organization-id">Organization ID</FieldLabel>
                <FieldDescription>
                  The unique identifier of your organization.
                  <br />
                  Used in CLI and API calls.
                </FieldDescription>
              </FieldContent>
              <InputGroup className="pr-1 flex-1">
                <InputGroupInput id="organization-id" value={selectedOrganization.id} readOnly />
                <InputGroupButton
                  variant="ghost"
                  size="icon-xs"
                  onClick={() =>
                    copyToClipboard(selectedOrganization.id).then(() => toast.success('Copied to clipboard'))
                  }
                >
                  {copied ? <CheckIcon className="h-4 w-4" /> : <CopyIcon className="h-4 w-4" />}
                </InputGroupButton>
              </InputGroup>
            </Field>
          </CardContent>
          <CardContent className="border-t border-border">
            <Field className="grid sm:grid-cols-2 items-center">
              <FieldContent className="flex-1">
                <FieldLabel htmlFor="organization-default-region">Default Region</FieldLabel>
                <FieldDescription>The default target for creating sandboxes in this organization.</FieldDescription>
              </FieldContent>
              <Input
                id="organization-default-region"
                value={selectedOrganization.defaultRegionId}
                readOnly
                className="flex-1 uppercase"
              />
            </Field>
          </CardContent>
        </Card>

        {!selectedOrganization.personal && authenticatedUserOrganizationMember !== null && (
          <Card className="bg-destructive/5 border-destructive/30">
            <CardContent>
              <div className="flex sm:flex-row flex-col justify-between sm:items-center gap-2">
                <div className="text-sm">
                  <div className="text-muted-foreground">
                    <p className="font-semibold text-destructive">Danger Zone</p>
                    {isOwner ? (
                      <>Delete the organization and all associated data.</>
                    ) : (
                      <>Remove yourself from the organization.</>
                    )}
                  </div>
                </div>
                {isOwner ? (
                  <DeleteOrganizationDialog
                    organizationName={selectedOrganization.name}
                    onDeleteOrganization={handleDeleteOrganization}
                    loading={deleteOrganizationMutation.isPending}
                  />
                ) : (
                  <LeaveOrganizationDialog
                    onLeaveOrganization={handleLeaveOrganization}
                    loading={leaveOrganizationMutation.isPending}
                  />
                )}
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  )
}

export default OrganizationSettings
