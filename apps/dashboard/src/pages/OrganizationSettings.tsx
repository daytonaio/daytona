/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CopyButton } from '@/components/CopyButton'
import { DeleteOrganizationDialog } from '@/components/Organizations/DeleteOrganizationDialog'
import { LeaveOrganizationDialog } from '@/components/Organizations/LeaveOrganizationDialog'
import { OtelConfigCard } from '@/components/Organizations/OtelConfigCard'
import {
  SetDefaultRegionDialog,
  type SetDefaultRegionDialogRef,
} from '@/components/Organizations/SetDefaultRegionDialog'
import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Field, FieldContent, FieldDescription, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { InputGroup, InputGroupInput } from '@/components/ui/input-group'
import { useDeleteOrganizationMutation } from '@/hooks/mutations/useDeleteOrganizationMutation'
import { useLeaveOrganizationMutation } from '@/hooks/mutations/useLeaveOrganizationMutation'
import { useOrganizations } from '@/hooks/useOrganizations'
import { useRegions } from '@/hooks/useRegions'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { OrganizationUserRoleEnum } from '@daytona/api-client'
import React, { useRef } from 'react'
import { toast } from 'sonner'

const OrganizationSettings: React.FC = () => {
  const { refreshOrganizations } = useOrganizations()
  const { selectedOrganization, authenticatedUserOrganizationMember } = useSelectedOrganization()
  const { getRegionName } = useRegions()

  const deleteOrganizationMutation = useDeleteOrganizationMutation()
  const leaveOrganizationMutation = useLeaveOrganizationMutation()
  const setDefaultRegionDialogRef = useRef<SetDefaultRegionDialogRef>(null)

  if (!selectedOrganization) {
    return null
  }

  const handleDeleteOrganization = async () => {
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

  const isOwner = authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER

  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>Settings</PageTitle>
      </PageHeader>

      <PageContent>
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
                <InputGroupInput
                  id="organization-id"
                  value={selectedOrganization.id}
                  readOnly
                  className="font-mono text-sm"
                />
                <CopyButton value={selectedOrganization.id} size="icon-xs" tooltipText="Copy Organization ID" />
              </InputGroup>
            </Field>
          </CardContent>
          <CardContent className="border-t border-border">
            <Field className="grid sm:grid-cols-2 items-center">
              <FieldContent className="flex-1">
                <FieldLabel htmlFor="organization-default-region">Default Region</FieldLabel>
                <FieldDescription>The default target for creating sandboxes in this organization.</FieldDescription>
              </FieldContent>
              {selectedOrganization.defaultRegionId ? (
                <Input
                  id="organization-default-region"
                  value={getRegionName(selectedOrganization.defaultRegionId) ?? selectedOrganization.defaultRegionId}
                  readOnly
                  className="flex-1 uppercase"
                />
              ) : isOwner ? (
                <div className="flex sm:justify-end">
                  <Button onClick={() => setDefaultRegionDialogRef.current?.open()} variant="secondary">
                    Set Region
                  </Button>
                </div>
              ) : null}
            </Field>
          </CardContent>
        </Card>

        {isOwner && <OtelConfigCard />}

        {!selectedOrganization.personal && authenticatedUserOrganizationMember !== null && (
          <Card className="bg-destructive-background border-destructive-separator">
            <CardContent>
              <div className="flex sm:flex-row flex-col justify-between sm:items-center gap-2">
                <div className="text-sm">
                  <div className="text-muted-foreground">
                    <p className="font-semibold text-destructive-foreground">Danger Zone</p>
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
        <SetDefaultRegionDialog ref={setDefaultRegionDialogRef} />
      </PageContent>
    </PageLayout>
  )
}

export default OrganizationSettings
