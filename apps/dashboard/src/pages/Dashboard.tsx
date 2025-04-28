import React from 'react'
import { Outlet } from 'react-router-dom'

import { Sidebar } from '@/components/Sidebar'
import { SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'
import { Toaster } from '@/components/ui/sonner'

const Dashboard: React.FC = () => {
  return (
    <div className="relative w-full">
      <SidebarProvider>
        <Sidebar />
        <SidebarTrigger className="md:hidden" />
        <div className="w-full">
          <Outlet />
        </div>
        <Toaster />
      </SidebarProvider>
    </div>
  )
}

export default Dashboard
