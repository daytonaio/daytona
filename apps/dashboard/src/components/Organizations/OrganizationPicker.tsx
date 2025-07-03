/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useEffect, useMemo, useState } from 'react'
import { toast } from 'sonner'
import { Building2, ChevronsUpDown, PlusCircle, SquareUserRound } from 'lucide-react'
import { Organization } from '@daytonaio/api-client'
import { useApi } from '@/hooks/useApi'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useOrganizations } from '@/hooks/useOrganizations'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { CreateOrganizationDialog } from './CreateOrganizationDialog'
import { SidebarMenuItem, SidebarMenuButton, SidebarMenu } from '@/components/ui/sidebar'
import { handleApiError } from '@/lib/error-handling'

export const OrganizationPicker: React.FC = () => {
  const { organizationsApi } = useApi()

  const { organizations, refreshOrganizations } = useOrganizations()
  const { selectedOrganization, onSelectOrganization } = useSelectedOrganization()

  const [optimisticSelectedOrganization, setOptimisticSelectedOrganization] = useState(selectedOrganization)
  const [loadingSelectOrganization, setLoadingSelectOrganization] = useState(false)

  useEffect(() => {
    setOptimisticSelectedOrganization(selectedOrganization)
  }, [selectedOrganization])

  const handleSelectOrganization = async (organizationId: string) => {
    const organization = organizations.find((org) => org.id === organizationId)
    if (!organization) {
      return
    }

    setOptimisticSelectedOrganization(organization)
    setLoadingSelectOrganization(true)
    const success = await onSelectOrganization(organizationId)
    if (!success) {
      setOptimisticSelectedOrganization(selectedOrganization)
    }
    setLoadingSelectOrganization(false)
  }

  const [showCreateOrganizationDialog, setShowCreateOrganizationDialog] = useState(false)

  const handleCreateOrganization = async (name: string) => {
    try {
      const organization = (
        await organizationsApi.createOrganization({
          name: name.trim(),
        })
      ).data
      toast.success('Organization created successfully')
      await refreshOrganizations()
      await onSelectOrganization(organization.id)
      return organization
    } catch (error) {
      handleApiError(error, 'Failed to create organization')
      return null
    }
  }

  const getOrganizationIcon = (organization: Organization) => {
    if (organization.personal) {
      return <SquareUserRound className="w-5 h-5" />
    }
    return <Building2 className="w-5 h-5" />
  }

  // personal first, then alphabetical
  const sortedOrganizations = useMemo(() => {
    return organizations.sort((a, b) => {
      if (a.personal && !b.personal) {
        return -1
      } else if (!a.personal && b.personal) {
        return 1
      } else {
        return a.name.localeCompare(b.name)
      }
    })
  }, [organizations])

  if (!optimisticSelectedOrganization) {
    return null
  }

  return (
    <SidebarMenu>
      <SidebarMenuItem className={`mb-1 ${loadingSelectOrganization ? 'cursor-progress' : ''}`}>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              disabled={loadingSelectOrganization}
              className="outline outline-1 outline-border outline-offset-0 mb-2 bg-muted px-3"
            >
              <div className="w-4 h-4 flex-shrink-0 bg-black rounded-full text-white flex items-center justify-center text-[10px] font-bold">
                {optimisticSelectedOrganization.name[0].toUpperCase()}
              </div>
              <span className="truncate text-foreground">{optimisticSelectedOrganization.name}</span>
              <ChevronsUpDown className="ml-auto w-4 h-4 opacity-50" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="w-[--radix-popper-anchor-width]">
            <div className="max-h-44 overflow-y-auto">
              {sortedOrganizations.map((org) => (
                <DropdownMenuItem
                  key={org.id}
                  onClick={() => handleSelectOrganization(org.id)}
                  className="cursor-pointer flex items-center gap-2"
                >
                  {getOrganizationIcon(org)}
                  <span className="truncate">{org.name}</span>
                </DropdownMenuItem>
              ))}
            </div>
            <DropdownMenuSeparator />
            <div>
              <DropdownMenuItem
                className="cursor-pointer text-primary flex items-center gap-2"
                onClick={() => setShowCreateOrganizationDialog(true)}
              >
                <PlusCircle className="w-4 h-4 flex-shrink-0" />
                <span>Create Organization</span>
              </DropdownMenuItem>
            </div>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>

      <CreateOrganizationDialog
        open={showCreateOrganizationDialog}
        onOpenChange={setShowCreateOrganizationDialog}
        onCreateOrganization={handleCreateOrganization}
      />
    </SidebarMenu>
  )
}
