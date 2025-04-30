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

const Dashboard: React.FC = () => {
  const { selectedOrganization } = useSelectedOrganization()
  const [showVerifyEmailDialog, setShowVerifyEmailDialog] = useState(false)
  const [tableSortState, setTableSortState] = useState(() => {
    const stored = localStorage.getItem('tableSortState')
    return stored ? JSON.parse(stored) : {}
  })

  useEffect(() => {
    if (
      selectedOrganization?.suspended &&
      selectedOrganization.suspensionReason === 'Please verify your email address'
    ) {
      setShowVerifyEmailDialog(true)
    }
  }, [selectedOrganization])

  useEffect(() => {
    localStorage.setItem('tableSortState', JSON.stringify(tableSortState))
  }, [tableSortState])

  return (
    <div className="relative w-full">
      <SidebarProvider>
        <Sidebar />
        <SidebarTrigger className="md:hidden" />
        <div className="w-full">
          <Outlet context={{ tableSortState, setTableSortState }} />
        </div>
        <Toaster />
        <VerifyEmailDialog open={showVerifyEmailDialog} onOpenChange={setShowVerifyEmailDialog} />
      </SidebarProvider>
    </div>
  )
}

export default Dashboard
