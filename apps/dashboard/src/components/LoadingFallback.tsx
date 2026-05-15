/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Sidebar as SidebarComponent,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuItem,
  SidebarProvider,
  SidebarSeparator,
} from '@/components/ui/sidebar'
import { cn } from '@/lib/utils'
import { AnimatePresence, motion } from 'framer-motion'
import { Loader2 } from 'lucide-react'
import { useEffect, useState } from 'react'
import { AnimatedLogo } from './AnimatedLogo'
import { Separator } from './ui/separator'
import { Skeleton } from './ui/skeleton'

const sidebarSkeletonGroups = [
  ['w-24', 'w-20', 'w-24', 'w-20'],
  ['w-24'],
  ['w-20', 'w-16', 'w-24', 'w-20'],
  ['w-20', 'w-16'],
]

function SidebarSkeletonItem({ width }: { width: string }) {
  return (
    <SidebarMenuItem>
      <div className="flex h-8 items-center gap-2 rounded-md px-2">
        <Skeleton className="size-4 shrink-0 rounded-sm" />
        <Skeleton className={`h-4 ${width} group-data-[collapsible=icon]:hidden`} />
      </div>
    </SidebarMenuItem>
  )
}

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
        <SidebarHeader>
          <div className="flex h-[46px] items-center justify-between gap-2 px-2 pt-2">
            <AnimatePresence initial={false}>
              <AnimatedLogo className={cn('w-[117px] text-primary')} />
            </AnimatePresence>
          </div>
        </SidebarHeader>
        <Separator className="mx-0 w-full" />
        <SidebarContent className="pt-4">
          <SidebarMenu className="px-2 pb-2">
            <SidebarMenuItem>
              <div className="flex h-8 items-center gap-2 rounded-md px-2">
                <Skeleton className="size-4 shrink-0 rounded-sm" />
                <Skeleton className="h-4 w-28 group-data-[collapsible=icon]:hidden" />
              </div>
            </SidebarMenuItem>
            <SidebarMenuItem>
              <div className="flex h-8 items-center gap-2 rounded-md px-2">
                <Skeleton className="size-4 shrink-0 rounded-sm" />
                <Skeleton className="h-4 w-20 group-data-[collapsible=icon]:hidden" />
              </div>
            </SidebarMenuItem>
          </SidebarMenu>

          {sidebarSkeletonGroups.map((group, groupIndex) => (
            <div key={groupIndex}>
              {groupIndex > 0 && <SidebarSeparator />}
              <SidebarGroup>
                <SidebarMenu>
                  {group.map((width, itemIndex) => (
                    <SidebarSkeletonItem key={`${groupIndex}-${itemIndex}`} width={width} />
                  ))}
                </SidebarMenu>
              </SidebarGroup>
            </div>
          ))}
        </SidebarContent>

        <SidebarFooter className="pb-4">
          <Skeleton className="mx-2 h-3 w-20 group-data-[collapsible=icon]:hidden" />
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
