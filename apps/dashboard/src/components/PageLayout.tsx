/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { type ComponentProps } from 'react'
import { BannerStack } from './Banner'
import { SidebarTrigger } from './ui/sidebar'

function PageLayout({ className, ...props }: ComponentProps<'div'>) {
  return <div className={cn('flex h-full flex-col group/page', className)} {...props} />
}

function PageHeader({ className, children, ...props }: ComponentProps<'header'>) {
  return (
    <header
      className={cn(
        'flex gap-2 sm:gap-4 items-center border-b border-border p-4 sm:px-5 bg-background z-10 group-[:has([data-slot=page-banner]:not(:empty))]/page:border-b-transparent',
        className,
      )}
      {...props}
    >
      <SidebarTrigger className="[&_svg]:size-5 md:hidden" />
      {children}
    </header>
  )
}

function PageTitle({ className, children, ...props }: ComponentProps<'h1'>) {
  return (
    <h1 className={cn('text-2xl font-medium tracking-tight', className)} {...props}>
      {children}
    </h1>
  )
}

function PageDescription({ className, ...props }: ComponentProps<'p'>) {
  return <p className={cn('text-sm text-muted-foreground', className)} {...props} />
}

function PageBanner({ className, children, ...props }: ComponentProps<'div'>) {
  return (
    <div data-slot="page-banner" className={cn('w-full relative z-30 empty:hidden', className)} {...props}>
      {children}
    </div>
  )
}

function PageContent({
  className,
  size = 'default',
  ...props
}: ComponentProps<'main'> & { size?: 'default' | 'full' }) {
  return (
    <>
      <PageBanner>
        <BannerStack />
      </PageBanner>
      <main
        className={cn(
          'flex flex-col gap-4 p-4 sm:px-5 w-full pt-6',
          {
            'max-w-5xl mx-auto': size === 'default',
          },
          className,
        )}
        {...props}
      />
    </>
  )
}

export { PageContent, PageDescription, PageHeader, PageLayout, PageTitle }
