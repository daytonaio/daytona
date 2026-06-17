/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Sidebar as SidebarComponent,
  SidebarContent,
  SidebarGroup,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuItem,
  SidebarProvider,
  SidebarSeparator,
} from '@/components/ui/sidebar'
import { cn } from '@/lib/utils'
import { AnimatePresence } from 'framer-motion'
import { AnimatedLogo } from './AnimatedLogo'
import { LoadingFallbackContent } from './LoadingFallbackContent'
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
        <Skeleton className={cn('h-4 group-data-[collapsible=icon]:hidden', width)} />
      </div>
    </SidebarMenuItem>
  )
}

type LoadingFallbackProps = {
  source?: string
}

const LoadingFallback = ({ source = 'unknown' }: LoadingFallbackProps) => {
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
      </SidebarComponent>
      <SidebarInset className="overflow-hidden">
        <div className="absolute inset-0 p-6 bg-background z-[3]">
          <LoadingFallbackContent className="h-full" source={source} />
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}

export default LoadingFallback
