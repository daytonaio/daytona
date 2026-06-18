/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import * as React from 'react'

import { cn } from '@/lib/utils'
import { getColumnResizeSizeBounds } from '@/lib/utils/table'
import type { Header } from '@tanstack/react-table'
import { useEffect, useRef } from 'react'
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from './empty'
import './table.css'

function useScrollStateRef() {
  const containerRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    const el = containerRef.current
    if (!el) return

    const scrollThreshold = 1

    const update = () => {
      const scrollTop = el.scrollTop
      const scrollLeft = el.scrollLeft
      const maxScrollLeft = el.scrollWidth - el.clientWidth

      if (scrollTop > scrollThreshold) {
        el.dataset.scrolledTop = ''
      } else {
        delete el.dataset.scrolledTop
      }

      if (scrollLeft > scrollThreshold) {
        el.dataset.scrolledLeft = ''
      } else {
        delete el.dataset.scrolledLeft
      }

      if (maxScrollLeft - scrollLeft > scrollThreshold) {
        el.dataset.overflowRight = ''
      } else {
        delete el.dataset.overflowRight
      }
    }

    update()
    el.addEventListener('scroll', update, { passive: true })

    const observer = new ResizeObserver(update)
    observer.observe(el)
    if (el.firstElementChild) {
      observer.observe(el.firstElementChild)
    }

    return () => {
      el.removeEventListener('scroll', update)
      observer.disconnect()
    }
  }, [])

  return containerRef
}

type TableContainerProps = React.ComponentProps<'div'> & {
  empty?: React.ReactNode
}

function TableContainer({ className, children, empty, ...props }: TableContainerProps) {
  const containerRef = useScrollStateRef()

  return (
    <div
      ref={containerRef}
      data-slot="table-container"
      className={cn(
        'relative w-full overflow-auto border border-border bg-table-cell overscroll-x-none rounded-md scrollbar-sm has-[[data-slot=empty]]:overflow-x-hidden',
        className,
      )}
      {...props}
    >
      {children}
      <div aria-hidden="true" data-slot="table-column-resize-indicator" />
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

type StickyState = boolean | 'left' | 'right'

function getStickySide(sticky?: StickyState) {
  return sticky === 'left' || sticky === 'right' ? sticky : undefined
}

function getHeaderResizeLabel<TData>(header: Header<TData, unknown>) {
  const columnHeader = header.column.columnDef.header

  return typeof columnHeader === 'string' ? columnHeader : header.column.id
}

function getHeaderResizeBounds<TData>(header: Header<TData, unknown>) {
  return getColumnResizeSizeBounds(header.column, header.getContext().table.options.defaultColumn)
}

function getHeaderCanResize<TData>(header?: Header<TData, unknown>) {
  if (!header || header.isPlaceholder || !header.column.getCanResize()) return false

  const { minSize, maxSize } = getHeaderResizeBounds(header)

  return minSize < maxSize
}

type ColumnResizePreviewState = {
  directionMultiplier: 1 | -1
  initialSize: number
  initialX: number
  maxSize: number
  minSize: number
}

type ColumnResizePreviewPosition = {
  boundaryX: number
  isPastBounds: boolean
  size: number
  x: number
}

const COLUMN_RESIZE_LIMIT_RELEASE_MS = 150
const COLUMN_RESIZE_FALLOFF_DISTANCE = 24
const COLUMN_RESIZE_FALLOFF_STIFFNESS = 3
const columnResizeReleaseTimeouts = new WeakMap<HTMLElement, number>()

function clearColumnResizeRelease(container: HTMLElement) {
  const timeout = columnResizeReleaseTimeouts.get(container)

  if (timeout) {
    window.clearTimeout(timeout)
    columnResizeReleaseTimeouts.delete(container)
  }

  delete container.dataset.columnResizeReleasing
}

function getColumnResizeAriaValueText(size: number) {
  return `${Math.round(size)} pixels`
}

function resetColumnResizeIndicator(indicator: HTMLElement) {
  indicator.style.transform = ''
  indicator.style.removeProperty('--table-column-resize-indicator-height')
}

function getColumnResizePointerPosition(container: HTMLElement, clientX: number) {
  const rect = container.getBoundingClientRect()

  return clientX - rect.left - container.clientLeft + container.scrollLeft
}

function clampColumnResizeSize(size: number, minSize: number, maxSize: number) {
  return Math.min(Math.max(size, minSize), maxSize)
}

function applyColumnResizeFalloff(size: number, minSize: number, maxSize: number, distance: number, stiffness: number) {
  if (size < minSize) {
    const overshoot = minSize - size
    return minSize - (distance * overshoot) / (overshoot + distance * stiffness)
  }

  if (size > maxSize) {
    const overshoot = size - maxSize
    return maxSize + (distance * overshoot) / (overshoot + distance * stiffness)
  }

  return size
}

function getColumnResizePreviewPosition(
  container: HTMLElement,
  clientX: number,
  { directionMultiplier, initialSize, initialX, maxSize, minSize }: ColumnResizePreviewState,
): ColumnResizePreviewPosition {
  const pointerX = getColumnResizePointerPosition(container, clientX)
  const rawSize = initialSize + (pointerX - initialX) * directionMultiplier
  const nextSize = clampColumnResizeSize(rawSize, minSize, maxSize)
  const isPastBounds = rawSize < minSize || rawSize > maxSize
  const displaySize = applyColumnResizeFalloff(
    rawSize,
    minSize,
    maxSize,
    COLUMN_RESIZE_FALLOFF_DISTANCE,
    COLUMN_RESIZE_FALLOFF_STIFFNESS,
  )
  const boundaryX = initialX + (nextSize - initialSize) * directionMultiplier

  return {
    boundaryX,
    isPastBounds,
    size: nextSize,
    x: initialX + (displaySize - initialSize) * directionMultiplier,
  }
}

function startColumnResize<TData>(
  target: HTMLElement,
  initialClientX: number,
  pointerId: number,
  header: Header<TData, unknown>,
) {
  const container = target.closest<HTMLElement>('[data-slot="table-container"]')
  const indicator = container?.querySelector<HTMLElement>('[data-slot="table-column-resize-indicator"]')

  if (!container || !indicator) return

  const table = header.getContext().table
  const { minSize, maxSize } = getHeaderResizeBounds(header)
  const previewState: ColumnResizePreviewState = {
    directionMultiplier: table.options.columnResizeDirection === 'rtl' ? -1 : 1,
    initialSize: header.getSize(),
    initialX: getColumnResizePointerPosition(container, initialClientX),
    maxSize,
    minSize,
  }

  let frame = 0
  let latestClientX = initialClientX
  let lastPosition: ColumnResizePreviewPosition | undefined
  let stopped = false

  const commitSize = (nextSize: number) => {
    table.setColumnSizing((old) => ({
      ...old,
      [header.column.id]: nextSize,
    }))
  }

  const update = () => {
    frame = 0
    const position = getColumnResizePreviewPosition(container, latestClientX, previewState)
    lastPosition = position

    indicator.style.transform = `translate3d(${position.x}px, 0, 0)`
    target.setAttribute('aria-valuenow', String(Math.round(position.size)))
    target.setAttribute('aria-valuetext', getColumnResizeAriaValueText(position.size))
  }

  const scheduleUpdate = (clientX: number) => {
    latestClientX = clientX
    if (frame) return

    frame = requestAnimationFrame(update)
  }

  const handlePointerMove = (event: PointerEvent) => {
    if (event.pointerId !== pointerId) return

    event.preventDefault()
    scheduleUpdate(event.clientX)
  }

  const stop = () => {
    if (stopped) return

    stopped = true
    clearColumnResizeRelease(container)
    if (frame) {
      cancelAnimationFrame(frame)
      frame = 0
    }

    if (target.hasPointerCapture(pointerId)) {
      target.releasePointerCapture(pointerId)
    }

    document.removeEventListener('pointermove', handlePointerMove)
    document.removeEventListener('pointerup', handlePointerEnd)
    document.removeEventListener('pointercancel', handlePointerEnd)
    target.removeEventListener('lostpointercapture', stop)
    window.removeEventListener('blur', stop)

    commitSize(lastPosition?.size ?? previewState.initialSize)
    delete target.dataset.resizing

    if (lastPosition?.isPastBounds) {
      container.dataset.columnResizeReleasing = 'true'
      indicator.style.transform = `translate3d(${lastPosition.boundaryX}px, 0, 0)`

      const releaseTimeout = window.setTimeout(() => {
        delete container.dataset.columnResizing
        delete container.dataset.columnResizeReleasing
        resetColumnResizeIndicator(indicator)
        columnResizeReleaseTimeouts.delete(container)
      }, COLUMN_RESIZE_LIMIT_RELEASE_MS)
      columnResizeReleaseTimeouts.set(container, releaseTimeout)
      return
    }

    delete container.dataset.columnResizing
    delete container.dataset.columnResizeReleasing
    resetColumnResizeIndicator(indicator)
  }

  const handlePointerEnd = (event: PointerEvent) => {
    if (event.pointerId !== pointerId) return

    stop()
  }

  container.dataset.columnResizing = 'true'
  clearColumnResizeRelease(container)
  indicator.style.setProperty('--table-column-resize-indicator-height', `${container.scrollHeight}px`)
  target.dataset.resizing = 'true'
  target.setPointerCapture(pointerId)
  update()
  document.addEventListener('pointermove', handlePointerMove, { passive: false })
  document.addEventListener('pointerup', handlePointerEnd)
  document.addEventListener('pointercancel', handlePointerEnd)
  target.addEventListener('lostpointercapture', stop)
  window.addEventListener('blur', stop)
}

function TableColumnResizeHandle<TData>({ header }: { header: Header<TData, unknown> }) {
  if (!getHeaderCanResize(header)) return null

  const table = header.getContext().table
  const { minSize, maxSize } = getHeaderResizeBounds(header)
  const size = header.column.getSize()
  const label = getHeaderResizeLabel(header)

  const updateColumnSize = (nextSize: number) => {
    const size = clampColumnResizeSize(nextSize, minSize, maxSize)

    table.setColumnSizing((old) => ({
      ...old,
      [header.column.id]: size,
    }))
  }

  const handleKeyDown = (event: React.KeyboardEvent<HTMLDivElement>) => {
    if (event.key === 'Home') {
      event.preventDefault()
      event.stopPropagation()
      updateColumnSize(minSize)
      return
    }

    if (event.key === 'End' && maxSize !== Number.MAX_SAFE_INTEGER) {
      event.preventDefault()
      event.stopPropagation()
      updateColumnSize(maxSize)
      return
    }

    if (event.key === 'Enter') {
      event.preventDefault()
      event.stopPropagation()
      header.column.resetSize()
      return
    }

    if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return

    event.preventDefault()
    event.stopPropagation()

    const step = event.shiftKey ? 25 : 10
    const direction = event.key === 'ArrowRight' ? 1 : -1
    const directionMultiplier = table.options.columnResizeDirection === 'rtl' ? -1 : 1

    updateColumnSize(size + step * direction * directionMultiplier)
  }

  const handlePointerDown = (event: React.PointerEvent<HTMLDivElement>) => {
    if (!event.isPrimary || event.button !== 0) return

    event.preventDefault()
    event.stopPropagation()
    startColumnResize(event.currentTarget, event.clientX, event.pointerId, header)
  }

  return (
    <div
      aria-label={`Resize ${label} column`}
      aria-orientation="vertical"
      aria-valuemax={maxSize === Number.MAX_SAFE_INTEGER ? undefined : maxSize}
      aria-valuemin={minSize}
      aria-valuenow={Math.round(size)}
      aria-valuetext={getColumnResizeAriaValueText(size)}
      data-resizing={header.column.getIsResizing() ? 'true' : undefined}
      data-slot="table-column-resize-handle"
      onDoubleClick={(event) => {
        event.preventDefault()
        event.stopPropagation()
        header.column.resetSize()
      }}
      onKeyDown={handleKeyDown}
      onPointerDown={handlePointerDown}
      role="separator"
      tabIndex={0}
      title={`Resize ${label} column`}
    />
  )
}

function TableHead<TData = unknown>({
  children,
  className,
  header,
  style,
  sticky,
  ...props
}: React.ComponentProps<'th'> & { header?: Header<TData, unknown>; sticky?: StickyState }) {
  const stickySide = getStickySide(sticky)
  const canResize = getHeaderCanResize(header)

  return (
    <th
      data-slot="table-head"
      data-sticky-state={stickySide}
      data-resizable={canResize ? 'true' : undefined}
      className={cn(
        'text-muted-foreground border-b h-8 px-3 text-left align-middle !font-mono font-medium !uppercase whitespace-nowrap !text-xs [&_*]:!font-mono [&_*]:!uppercase [&_*]:!text-xs bg-table-header [&:has([data-sort])]:text-foreground',
        {
          'sticky z-[2]': sticky,
          relative: !sticky,
          'left-0': stickySide === 'left',
          'right-0': stickySide === 'right',
        },
        className,
      )}
      style={style}
      {...props}
    >
      {children}
      {header ? <TableColumnResizeHandle header={header} /> : null}
    </th>
  )
}

function TableCell({ className, style, sticky, ...props }: React.ComponentProps<'td'> & { sticky?: StickyState }) {
  const stickySide = getStickySide(sticky)

  return (
    <td
      data-slot="table-cell"
      data-sticky-state={stickySide}
      className={cn(
        'px-3 py-2.5 align-middle whitespace-nowrap border-b bg-table-cell group-hover/row:bg-table-cell-hover group-data-[state=selected]/row:bg-table-cell-active group-data-[selected=true]/row:bg-table-cell-active has-[[data-slot=empty]]:group-hover/row:bg-table-cell',
        {
          'sticky z-[1]': sticky,
          'left-0': stickySide === 'left',
          'right-0': stickySide === 'right',
        },
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
        className={cn('absolute inset-x-0 top-8 bottom-0 flex justify-center p-4 items-start', className)}
      >
        <Empty className="w-full max-w-xl border-none py-8 bg-transparent">
          <EmptyHeader className="w-full">
            {icon && (
              <EmptyMedia variant="icon" className="[&_svg]:size-4">
                {icon}
              </EmptyMedia>
            )}
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
            {icon && (
              <EmptyMedia variant="icon" className="[&_svg]:size-4">
                {icon}
              </EmptyMedia>
            )}
            <EmptyTitle>{message}</EmptyTitle>
            {description && <EmptyDescription>{description}</EmptyDescription>}
          </EmptyHeader>
          {action && <EmptyContent>{action}</EmptyContent>}
        </Empty>
      </TableCell>
    </TableRow>
  )
}

function TableFillHead() {
  return <th className="p-0 border-b bg-table-header" aria-hidden="true" />
}

function TableFillCell() {
  return (
    <td
      className="p-0 border-b bg-table-cell group-hover/row:bg-table-cell-hover group-data-[state=selected]/row:bg-table-cell-active group-data-[selected=true]/row:bg-table-cell-active"
      aria-hidden="true"
    />
  )
}

export {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableContainer,
  TableEmptyState,
  TableFillCell,
  TableFillHead,
  TableFooter,
  TableHead,
  TableHeader,
  TableRow,
}
