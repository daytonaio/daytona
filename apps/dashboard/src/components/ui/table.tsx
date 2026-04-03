/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import * as React from 'react'

import { cn } from '@/lib/utils'
import { CSSProperties, useEffect, useRef } from 'react'
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from './empty'

function getStickyState({
  className,
  style,
  sticky,
}: {
  className?: string
  style?: CSSProperties
  sticky?: 'left' | 'right'
}): 'left' | 'right' | undefined {
  if (sticky) return sticky
  if (!className?.includes('sticky')) return undefined
  if (style?.right !== undefined) return 'right'
  if (style?.left !== undefined) return 'left'
  return undefined
}

function useScrollStateRef() {
  const containerRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    const el = containerRef.current
    if (!el) return

    const update = () => {
      const scrollLeft = el.scrollLeft
      const maxScroll = el.scrollWidth - el.clientWidth

      if (scrollLeft > 1) {
        el.dataset.scrolledLeft = ''
      } else {
        delete el.dataset.scrolledLeft
      }

      if (maxScroll - scrollLeft > 1) {
        el.dataset.scrolledRight = ''
      } else {
        delete el.dataset.scrolledRight
      }
    }

    update()
    el.addEventListener('scroll', update, { passive: true })

    const observer = new ResizeObserver(update)
    observer.observe(el)

    return () => {
      el.removeEventListener('scroll', update)
      observer.disconnect()
    }
  }, [])

  return containerRef
}

function TableContainer({
  className,
  children,
  empty,
  ...props
}: React.ComponentProps<'div'> & { empty?: React.ReactNode }) {
  const containerRef = useScrollStateRef()

  return (
    <div
      ref={containerRef}
      data-slot="table-container"
      className={cn(
        'relative w-full overflow-auto border border-border bg-table-cell overscroll-none rounded-md scrollbar-sm has-[[data-slot=empty]]:overflow-hidden',
        className,
      )}
      {...props}
    >
      {children}
      {empty}
    </div>
  )
}

function Table({ className, ...props }: React.ComponentProps<'table'>) {
  return (
    <table
      data-slot="table"
      className={cn('w-full caption-bottom border-separate border-spacing-0 text-sm', className)}
      {...props}
    />
  )
}

function TableHeader({ className, ...props }: React.ComponentProps<'thead'>) {
  return (
    <thead
      data-slot="table-header"
      className={cn(
        'sticky top-0 z-20 [&_tr:first-child_th:first-child]:rounded-tl-[5px] [&_tr:first-child_th:last-child]:rounded-tr-[5px]',
        className,
      )}
      {...props}
    />
  )
}

function TableBody({ className, ...props }: React.ComponentProps<'tbody'>) {
  return (
    <tbody
      data-slot="table-body"
      className={cn(
        '[&_tr:last-child_td]:border-b-0 [&_tr:last-child_td:first-child]:rounded-bl-[5px] [&_tr:last-child_td:last-child]:rounded-br-[5px]',
        className,
      )}
      {...props}
    />
  )
}

function TableFooter({ className, ...props }: React.ComponentProps<'tfoot'>) {
  return (
    <tfoot
      data-slot="table-footer"
      className={cn('bg-muted/60 border-t font-medium [&>tr]:last:border-b-0', className)}
      {...props}
    />
  )
}

function TableRow({ className, ...props }: React.ComponentProps<'tr'>) {
  return <tr data-slot="table-row" className={cn('group/row transition-colors', className)} {...props} />
}

function TableHead({ className, style, sticky, ...props }: React.ComponentProps<'th'> & { sticky?: 'left' | 'right' }) {
  const stickyState = getStickyState({ className, style, sticky })

  return (
    <th
      data-slot="table-head"
      data-sticky-state={stickyState}
      className={cn(
        'text-muted-foreground border-b h-8 px-3 text-left align-middle font-medium whitespace-nowrap text-sm [&_*]:text-sm bg-table-header [&:has([data-sort])]:text-foreground',
        className,
      )}
      style={style}
      {...props}
    />
  )
}

function TableCell({ className, style, sticky, ...props }: React.ComponentProps<'td'> & { sticky?: 'left' | 'right' }) {
  const stickyState = getStickyState({ className, style, sticky })

  return (
    <td
      data-slot="table-cell"
      data-sticky-state={stickyState}
      className={cn(
        'px-3 py-2 align-middle whitespace-nowrap border-b bg-table-cell group-hover/row:bg-table-cell-hover group-data-[state=selected]/row:bg-table-cell-active has-[[data-slot=empty]]:group-hover/row:bg-table-cell',
        className,
      )}
      style={style}
      {...props}
    />
  )
}

function TableCaption({ className, ...props }: React.ComponentProps<'caption'>) {
  return (
    <caption data-slot="table-caption" className={cn('text-muted-foreground mt-4 text-sm', className)} {...props} />
  )
}

interface TableEmptyStateProps {
  colSpan: number
  message: string
  icon?: React.ReactNode
  description?: React.ReactNode
  action?: React.ReactNode
  className?: string
  overlay?: boolean
}

function TableEmptyState({
  colSpan,
  message,
  icon,
  description,
  action,
  className = '',
  overlay = false,
}: TableEmptyStateProps) {
  if (overlay) {
    return (
      <div
        data-slot="empty"
        className={cn('absolute inset-x-0 top-8 bottom-0 flex items-center justify-center p-4', className)}
      >
        <Empty className="w-full max-w-xl border-none py-8 bg-transparent">
          <EmptyHeader className="w-full">
            {icon && <EmptyMedia variant="icon">{icon}</EmptyMedia>}
            <EmptyTitle>{message}</EmptyTitle>
            {description && <EmptyDescription>{description}</EmptyDescription>}
          </EmptyHeader>
          {action && <EmptyContent>{action}</EmptyContent>}
        </Empty>
      </div>
    )
  }

  return (
    <TableRow>
      <TableCell colSpan={colSpan} className={cn('h-24 text-center', className)}>
        <Empty className="border-none py-8 bg-transparent">
          <EmptyHeader>
            {icon && <EmptyMedia variant="icon">{icon}</EmptyMedia>}
            <EmptyTitle>{message}</EmptyTitle>
            {description && <EmptyDescription>{description}</EmptyDescription>}
          </EmptyHeader>
          {action && <EmptyContent>{action}</EmptyContent>}
        </Empty>
      </TableCell>
    </TableRow>
  )
}

export {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableContainer,
  TableEmptyState,
  TableFooter,
  TableHead,
  TableHeader,
  TableRow,
}
