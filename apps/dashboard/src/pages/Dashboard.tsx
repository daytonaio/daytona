/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useEffect, useState } from 'react'
import { Outlet } from 'react-router-dom'

import { Sidebar } from '@/components/Sidebar'
import { SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'
import { Toaster } from '@/components/ui/sonner'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { VerifyEmailDialog } from '@/components/VerifyEmailDialog'
import { TableSortingProvider } from '@/providers/TableSortingProvider'

type SortingState = {
  [key: string]: {
    field: string
    direction: 'asc' | 'desc'
  }
}

const Dashboard: React.FC = () => {
  const { selectedOrganization } = useSelectedOrganization()
  const [showVerifyEmailDialog, setShowVerifyEmailDialog] = useState(false)

  useEffect(() => {
    if (
      selectedOrganization?.suspended &&
      selectedOrganization.suspensionReason === 'Please verify your email address'
    ) {
      setShowVerifyEmailDialog(true)
    }
  }, [selectedOrganization])

  return (
    <div className="relative w-full">
      <SidebarProvider>
        <TableSortingProvider>
          <Sidebar />
          <SidebarTrigger className="md:hidden" />
          <div className="w-full">
            <Outlet />
          </div>
          <Toaster />
          <VerifyEmailDialog open={showVerifyEmailDialog} onOpenChange={setShowVerifyEmailDialog} />
        </TableSortingProvider>
      </SidebarProvider>
    </div>
  )
}

export default Dashboard
