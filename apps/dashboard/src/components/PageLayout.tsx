/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { ArrowUpRight } from 'lucide-react'
import { type ComponentProps, type ReactNode, useLayoutEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { BannerStack } from './Banner'
import { OrganizationPicker } from './Organizations/OrganizationPicker'
import { Button } from './ui/button'
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

function PageBreadcrumbs({
  current,
  parent,
  className,
  ...props
}: ComponentProps<'nav'> & { current: string; parent?: ReactNode }) {
  return (
    <nav aria-label="Breadcrumb" className={cn('min-w-0', className)} {...props}>
      <ol className="flex min-w-0 items-center gap-1.5 text-sm text-muted-foreground">
        <li className="min-w-0 truncate">{parent ?? <OrganizationPicker variant="breadcrumb" />}</li>
        <li className="shrink-0">/</li>
        <li className="ml-2 truncate font-medium text-foreground">{current}</li>
      </ol>
    </nav>
  )
}

function PageDocsLink({ href, label, className, ...props }: ComponentProps<'a'> & { href: string; label: string }) {
  return (
    <Button variant="link" size="sm" className={cn('ml-auto px-0 text-muted-foreground', className)} asChild>
      <a href={href} target="_blank" rel="noopener noreferrer" {...props}>
        {label}
        <ArrowUpRight className="size-4" />
      </a>
    </Button>
  )
}

function PageIntro({
  title,
  description,
  titleMeta,
  titleActions,
  actions,
  className,
}: {
  title: ReactNode
  description: ReactNode
  titleMeta?: ReactNode
  titleActions?: ReactNode
  actions?: ReactNode
  className?: string
}) {
  return (
    <div
      className={cn(
        'mb-8 shrink-0',
        actions ? 'grid gap-4 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-start' : 'flex flex-col gap-1',
        className,
      )}
    >
      <div className="flex min-w-0 flex-col gap-1">
        <div className="flex min-w-0 flex-wrap items-center justify-between gap-3">
          <div className="flex min-w-0 flex-wrap items-center gap-x-4 gap-y-2">
            <PageTitle>{title}</PageTitle>
            {titleMeta}
          </div>
          {titleActions ? <div className="ml-auto flex items-center gap-2">{titleActions}</div> : null}
        </div>
        <PageDescription>{description}</PageDescription>
      </div>
      {actions ? <div className="flex flex-wrap items-center gap-2 sm:justify-end">{actions}</div> : null}
    </div>
  )
}

type PageStatItem = {
  label: ReactNode
  value: ReactNode
  markerClassName?: string
}

function PageStats({
  items,
  loadingText,
  className,
}: {
  items: PageStatItem[]
  loadingText?: ReactNode
  className?: string
}) {
  return (
    <div className={cn('flex max-w-full flex-wrap gap-x-5 gap-y-1 text-xs text-muted-foreground', className)}>
      {loadingText ? (
        <span>{loadingText}</span>
      ) : (
        items.map((item, index) => (
          <span key={index} className="inline-flex items-center gap-1.5">
            {item.markerClassName ? <span className={cn('size-2 rounded-[2px]', item.markerClassName)} /> : null}
            <span className="font-medium text-foreground">{item.value}</span> {item.label}
          </span>
        ))
      )}
    </div>
  )
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

export {
  PageBreadcrumbs,
  PageContent,
  PageDescription,
  PageDocsLink,
  PageFooter,
  PageFooterPortal,
  PageHeader,
  PageIntro,
  PageLayout,
  PageStats,
  PageTitle,
}
