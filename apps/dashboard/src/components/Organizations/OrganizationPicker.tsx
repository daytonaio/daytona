/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import { SidebarMenuButton, SidebarMenuItem } from '@/components/ui/sidebar'
import { useApi } from '@/hooks/useApi'
import { useCopyToClipboard } from '@/hooks/useCopyToClipboard'
import { useOrganizations } from '@/hooks/useOrganizations'
import { useRegions } from '@/hooks/useRegions'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { cn } from '@/lib/utils'
import { Organization } from '@daytona/api-client'
import { Building2, ChevronsUpDown, Copy, PlusCircle, SquareUserRound } from 'lucide-react'
import { ReactNode, useEffect, useMemo, useState } from 'react'
import { toast } from 'sonner'
import { CommandHighlight, useRegisterCommands, type CommandConfig } from '../CommandPalette'
import { CreateOrganizationSheet } from './CreateOrganizationSheet'

function useOrganizationCommands(enabled: boolean) {
  const { organizations } = useOrganizations()
  const { selectedOrganization, onSelectOrganization } = useSelectedOrganization()
  const [, copyToClipboard] = useCopyToClipboard()

  const commands: CommandConfig[] = useMemo(() => {
    if (!enabled) {
      return []
    }

    const cmds: CommandConfig[] = []

    if (selectedOrganization) {
      cmds.push({
        id: 'copy-org-id',
        label: 'Copy Organization ID',
        icon: <Copy className="w-4 h-4" />,
        onSelect: () => {
          copyToClipboard(selectedOrganization.id)
          toast.success('Organization ID copied to clipboard')
        },
      })
    }

    for (const org of organizations) {
      if (org.id === selectedOrganization?.id) continue

      cmds.push({
        id: `switch-org-${org.id}`,
        label: (
          <>
            Switch to <CommandHighlight>{org.name}</CommandHighlight>
          </>
        ),
        value: `switch to organization ${org.name}`,
        icon: <Building2 className="w-4 h-4" />,
        onSelect: () => onSelectOrganization(org.id),
      })
    }

    return cmds
  }, [enabled, organizations, selectedOrganization, copyToClipboard, onSelectOrganization])

  useRegisterCommands(commands, { groupId: 'organization', groupLabel: 'Organization', groupOrder: 5 })
}

export const OrganizationPicker: React.FC<{ variant?: 'sidebar' | 'breadcrumb' }> = ({ variant = 'sidebar' }) => {
  const { organizationsApi } = useApi()

  const { organizations, refreshOrganizations } = useOrganizations()
  const { selectedOrganization, onSelectOrganization } = useSelectedOrganization()
  const { sharedRegions: regions, loadingSharedRegions: loadingRegions } = useRegions()

  const [optimisticSelectedOrganization, setOptimisticSelectedOrganization] = useState(selectedOrganization)
  const [loadingSelectOrganization, setLoadingSelectOrganization] = useState(false)

  useOrganizationCommands(variant === 'sidebar')

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

  const [showCreateOrganizationSheet, setShowCreateOrganizationSheet] = useState(false)

  const handleCreateOrganization = async (name: string, defaultRegionId: string) => {
    try {
      const organization = (
        await organizationsApi.createOrganization({
          name: name.trim(),
          defaultRegionId,
        })
      ).data
      toast.success('Organization created successfully')
      await refreshOrganizations(organization.id)
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

  const trigger =
    variant === 'breadcrumb' ? (
      <Button
        variant="ghost"
        size="sm"
        disabled={loadingSelectOrganization}
        className="h-7 min-w-0 max-w-48 px-1.5 text-muted-foreground hover:text-foreground"
      >
        <OrganizationAvatar>{optimisticSelectedOrganization.name[0].toUpperCase()}</OrganizationAvatar>
        <span className="truncate">{optimisticSelectedOrganization.name}</span>
        <ChevronsUpDown className="size-3.5 opacity-50" />
      </Button>
    ) : (
      <SidebarMenuButton
        disabled={loadingSelectOrganization}
        className="outline outline-1 outline-border outline-offset-0 mb-2 bg-muted"
        tooltip={optimisticSelectedOrganization.name}
      >
        <OrganizationAvatar>{optimisticSelectedOrganization.name[0].toUpperCase()}</OrganizationAvatar>
        <span className="truncate text-foreground">{optimisticSelectedOrganization.name}</span>
        <ChevronsUpDown className="ml-auto w-4 h-4 opacity-50" />
      </SidebarMenuButton>
    )

  const picker = (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>{trigger}</DropdownMenuTrigger>
        <DropdownMenuContent className={cn('w-[--radix-popper-anchor-width]', variant === 'breadcrumb' && 'min-w-56')}>
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
              onClick={() => setShowCreateOrganizationSheet(true)}
            >
              <PlusCircle className="w-4 h-4 flex-shrink-0" />
              <span>Create Organization</span>
            </DropdownMenuItem>
          </div>
        </DropdownMenuContent>
      </DropdownMenu>

      <CreateOrganizationSheet
        open={showCreateOrganizationSheet}
        onOpenChange={setShowCreateOrganizationSheet}
        regions={regions}
        loadingRegions={loadingRegions}
        onCreateOrganization={handleCreateOrganization}
      />
    </>
  )

  if (variant === 'breadcrumb') {
    return picker
  }

  return <SidebarMenuItem>{picker}</SidebarMenuItem>
}

function OrganizationAvatar({ children }: { children: ReactNode }) {
  return (
    <div className="w-4 h-4 flex-shrink-0 rounded-full bg-black text-white flex items-center justify-center text-[10px] font-bold">
      {children}
    </div>
  )
}
