/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import * as React from 'react'

import { cn } from '@/lib/utils'

function Card({ ref, className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      ref={ref}
      data-slot="card"
      className={cn('rounded-lg border bg-card text-card-foreground shadow-sm', className)}
      {...props}
    />
  )
}

function CardHeader({ ref, className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div ref={ref} data-slot="card-header" className={cn('flex flex-col space-y-1.5 p-4 pb-2', className)} {...props} />
  )
}

function CardTitle({ ref, className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      ref={ref}
      data-slot="card-title"
      className={cn('text-xl font-semibold leading-none tracking-tight', className)}
      {...props}
    />
  )
}

function CardDescription({ ref, className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div ref={ref} data-slot="card-description" className={cn('text-sm text-muted-foreground', className)} {...props} />
  )
}

function CardContent({ ref, className, ...props }: React.ComponentProps<'div'>) {
  return <div ref={ref} data-slot="card-content" className={cn('p-4 w-full', className)} {...props} />
}

function CardFooter({ ref, className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      ref={ref}
      data-slot="card-footer"
      className={cn('flex items-center p-4 border-t border-border', className)}
      {...props}
    />
  )
}

export { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle }
