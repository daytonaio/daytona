/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
import {
  SidebarGroup,
  SidebarInset,
  SidebarProvider,
  Sidebar as SidebarComponent,
  SidebarContent,
  SidebarFooter,
} from '@/components/ui/sidebar'
import { Logo, LogoText } from '@/assets/Logo'
import { Loader2 } from 'lucide-react'
import { Skeleton } from './ui/skeleton'

const LoadingFallback = () => (
  <SidebarProvider isBannerVisible={false}>
    <SidebarComponent isBannerVisible={false} collapsible="icon">
      <SidebarContent>
        <SidebarGroup>
          <div className="flex justify-between items-center gap-2 px-2 mb-2 h-12">
            <div className="flex items-center gap-2 group-data-[state=collapsed]:hidden text-primary">
              <Logo />
              <LogoText />
            </div>
          </div>

          <Skeleton className="w-full h-8 mb-2" />

          <div className="flex flex-col gap-2">
            <Skeleton className="w-full h-8" />
            <Skeleton className="w-full h-8" />
            <Skeleton className="w-full h-8" />
            <Skeleton className="w-full h-8" />
            <Skeleton className="w-full h-8" />
          </div>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        <div className="flex flex-col gap-2">
          <Skeleton className="w-full h-8" />
          <Skeleton className="w-full h-8" />
          <Skeleton className="w-full h-8" />
        </div>
      </SidebarFooter>
    </SidebarComponent>
    <SidebarInset className="overflow-hidden">
      <div className="fixed top-0 left-0 w-full h-full p-6 bg-background z-[3]">
        <div className="flex items-center justify-center h-full">
          <Loader2 className="w-8 h-8 animate-spin" />
        </div>
      </div>
    </SidebarInset>
  </SidebarProvider>
)

export default LoadingFallback
