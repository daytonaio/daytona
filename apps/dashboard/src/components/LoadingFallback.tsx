/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Logo, LogoText } from '@/assets/Logo'
import {
  Sidebar as SidebarComponent,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarInset,
  SidebarProvider,
} from '@/components/ui/sidebar'
import { motion } from 'framer-motion'
import { Loader2 } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Skeleton } from './ui/skeleton'

const LoadingFallback = () => {
  const [showLongLoadingMessage, setShowLongLoadingMessage] = useState(false)

  useEffect(() => {
    const timer = setTimeout(() => {
      setShowLongLoadingMessage(true)
    }, 5_000)

    return () => clearTimeout(timer)
  }, [])

  return (
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
        <div className="absolute inset-0 p-6 bg-background z-[3]">
          <div className="flex items-center justify-center h-full flex-col gap-2">
            <Loader2 className="w-8 h-8 animate-spin" />
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={showLongLoadingMessage ? { opacity: 1, y: 0 } : { opacity: 0, y: 10 }}
              transition={{ duration: 0.35 }}
            >
              <p className="text-sm text-muted-foreground text-center">This is taking longer than expected...</p>
              <p className="text-sm text-muted-foreground text-center">
                If this issue persists, contact us at{' '}
                <a href="mailto:support@daytona.io" className="text-primary underline">
                  support@daytona.io
                </a>
                .
              </p>
            </motion.div>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}

export default LoadingFallback
