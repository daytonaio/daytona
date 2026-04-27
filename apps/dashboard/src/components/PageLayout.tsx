/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { type ComponentProps, type ReactNode, useLayoutEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { BannerStack } from './Banner'
import { SidebarTrigger } from './ui/sidebar'

function PageLayout({ className, contained = false, ...props }: ComponentProps<'div'> & { contained?: boolean }) {
  return (
    <div
      className={cn('flex h-full flex-col group/page', { 'max-h-screen overflow-hidden': contained }, className)}
      {...props}
    />
  )
}

function PageHeader({ className, children, ...props }: ComponentProps<'header'>) {
  return (
    <header
      className={cn(
        'flex gap-2 sm:gap-4 items-center border-b border-border p-4 bg-background z-10 group-[:has([data-slot=page-banner]:not(:empty))]/page:border-b-transparent min-h-[57px]',
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
        <BannerStack bannerClassName={cn({ 'max-w-5xl mx-auto': size === 'default' })} />
      </PageBanner>
      <main
        className={cn(
          'flex flex-col gap-4 p-4 w-full flex-1 min-h-0 overflow-auto',
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

function PageFooterPortal({ children }: { children: ReactNode }): ReactNode {
  const [container, setContainer] = useState<Element | null>(null)

  useLayoutEffect(() => {
    setContainer(document.querySelector('[data-slot="page-footer"]'))
  }, [])

  if (!container) return children

  return <>{createPortal(children, container)}</>
}

function PageFooter({ className, children, ...props }: ComponentProps<'footer'>) {
  return (
    <footer
      data-slot="page-footer"
      className={cn(
        'flex gap-2 sm:gap-4 items-center border-t border-border p-4 bg-background z-10 empty:hidden',
        className,
      )}
      {...props}
    >
      {children}
    </footer>
  )
}

export { PageContent, PageDescription, PageFooter, PageFooterPortal, PageHeader, PageLayout, PageTitle }
