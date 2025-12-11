/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { type ComponentProps } from 'react'

function PageLayout({ className, ...props }: ComponentProps<'div'>) {
  return <div className={cn('flex h-full flex-col', className)} {...props} />
}

function PageHeader({ className, ...props }: ComponentProps<'header'>) {
  return (
    <header className={cn('flex items-center justify-between border-b border-border p-4 px-5', className)} {...props} />
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

function PageContent({
  className,
  size = 'default',
  ...props
}: ComponentProps<'main'> & { size?: 'default' | 'full' }) {
  return (
    <main
      className={cn(
        'flex flex-col gap-4 p-4 px-5 w-full pt-6',
        {
          'max-w-5xl mx-auto': size === 'default',
        },
        className,
      )}
      {...props}
    />
  )
}

export { PageContent, PageDescription, PageHeader, PageLayout, PageTitle }
