/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

'use client'

import * as ScrollAreaPrimitive from '@radix-ui/react-scroll-area'
import { RefObject, useCallback, useEffect, useRef } from 'react'
import { useResizeObserver } from 'usehooks-ts'

import { cn } from '@/lib/utils'

const updateScrollOffsets = (viewport: HTMLElement | null, root: HTMLElement | null) => {
  if (!viewport || !root) {
    return
  }

  const { scrollTop, scrollHeight, clientHeight } = viewport
  const top = scrollTop
  const bottom = scrollHeight - clientHeight - scrollTop
  const { scrollLeft, scrollWidth, clientWidth } = viewport
  const left = scrollLeft
  const right = Math.max(0, scrollWidth - clientWidth - scrollLeft)

  root.style.setProperty('--offset-y-top', `${top}`)
  root.style.setProperty('--offset-y-bottom', `${bottom}`)
  root.style.setProperty('--offset-x-left', `${left}`)
  root.style.setProperty('--offset-x-right', `${right}`)
}

function ScrollArea({
  className,
  children,
  fade,
  fadeOrientation,
  fadeSide,
  horizontal,
  vertical = true,
  fadeOffset = 25,
  viewportRef: externalViewportRef,
  ...props
}: React.ComponentProps<typeof ScrollAreaPrimitive.Root> & {
  fade?: 'mask' | 'shadow'
  fadeOffset?: number
  fadeOrientation?: 'vertical' | 'horizontal'
  fadeSide?: 'start' | 'end'
  horizontal?: boolean
  vertical?: boolean
  viewportRef?: React.Ref<HTMLDivElement>
}) {
  const rootRef = useRef<HTMLDivElement>(null)
  const viewportRef = useRef<HTMLDivElement>(null)
  const resolvedFadeOrientation = fadeOrientation ?? (horizontal ? 'horizontal' : 'vertical')
  const setViewportRef = useCallback(
    (element: HTMLDivElement | null) => {
      viewportRef.current = element

      if (typeof externalViewportRef === 'function') {
        externalViewportRef(element)
        return
      }

      if (externalViewportRef) {
        externalViewportRef.current = element
      }
    },
    [externalViewportRef],
  )

  useResizeObserver({
    ref: viewportRef as RefObject<HTMLElement>,
    onResize: () => updateScrollOffsets(viewportRef.current, rootRef.current),
  })

  useEffect(() => {
    updateScrollOffsets(viewportRef.current, rootRef.current)
  }, [])

  return (
    <ScrollAreaPrimitive.Root
      ref={rootRef}
      data-slot="scroll-area"
      style={
        {
          '--fade-offset': fadeOffset !== undefined ? `${fadeOffset}px` : '30px',
          ...props.style,
        } as React.CSSProperties
      }
      className={cn(
        'relative group/scroll-area',
        {
          'before:pointer-events-none before:absolute before:top-0 before:left-0 before:right-0 before:z-10 before:[height:var(--fade-offset)] before:bg-gradient-to-b dark:before:from-black/20 before:from-black/10 before:to-transparent before:transition-opacity before:duration-150 before:opacity-[min(1,calc(var(--offset-y-top)/20))]':
            fade === 'shadow' && resolvedFadeOrientation === 'vertical' && fadeSide !== 'end',
          'after:pointer-events-none after:absolute after:bottom-0 after:left-0 after:right-0 after:z-10 after:[height:var(--fade-offset)] after:bg-gradient-to-t dark:after:from-black/20 after:from-black/10 after:to-transparent after:transition-opacity after:duration-150 after:opacity-[min(1,calc(var(--offset-y-bottom)/20))]':
            fade === 'shadow' && resolvedFadeOrientation === 'vertical' && fadeSide !== 'start',
          'before:pointer-events-none before:absolute before:top-0 before:bottom-0 before:left-0 before:z-10 before:[width:var(--fade-offset)] before:bg-gradient-to-r dark:before:from-black/20 before:from-black/10 before:to-transparent before:transition-opacity before:duration-150 before:opacity-[min(1,calc(var(--offset-x-left)/20))]':
            fade === 'shadow' && resolvedFadeOrientation === 'horizontal' && fadeSide !== 'end',
          'after:pointer-events-none after:absolute after:top-0 after:bottom-0 after:right-0 after:z-10 after:[width:var(--fade-offset)] after:bg-gradient-to-l dark:after:from-black/20 after:from-black/10 after:to-transparent after:transition-opacity after:duration-150 after:opacity-[min(1,calc(var(--offset-x-right)/20))]':
            fade === 'shadow' && resolvedFadeOrientation === 'horizontal' && fadeSide !== 'start',
        },
        className,
      )}
      {...props}
    >
      <ScrollAreaPrimitive.Viewport
        ref={setViewportRef}
        onScroll={(e) => {
          updateScrollOffsets(e.currentTarget, rootRef.current)
        }}
        data-slot="scroll-area-viewport"
        className={cn(
          'focus-visible:ring-ring/50 size-full rounded-[inherit] transition-[color,box-shadow] outline-none focus-visible:ring-[3px] focus-visible:outline-1 [&>div]:!block',
          {
            '!overflow-y-hidden': !vertical,
            '[mask-image:linear-gradient(to_bottom,transparent,black_min(var(--offset-y-top)*1px,var(--fade-offset)),black_calc(100%-min(var(--offset-y-bottom)*1px,var(--fade-offset))),transparent)]':
              fade === 'mask' && resolvedFadeOrientation === 'vertical' && fadeSide === undefined,
            '[mask-image:linear-gradient(to_bottom,transparent,black_min(var(--offset-y-top)*1px,var(--fade-offset)),black_100%)]':
              fade === 'mask' && resolvedFadeOrientation === 'vertical' && fadeSide === 'start',
            '[mask-image:linear-gradient(to_bottom,black_0,black_calc(100%-min(var(--offset-y-bottom)*1px,var(--fade-offset))),transparent)]':
              fade === 'mask' && resolvedFadeOrientation === 'vertical' && fadeSide === 'end',
            '[mask-image:linear-gradient(to_right,transparent,black_min(var(--offset-x-left)*1px,var(--fade-offset)),black_calc(100%-min(var(--offset-x-right)*1px,var(--fade-offset))),transparent)]':
              fade === 'mask' && resolvedFadeOrientation === 'horizontal' && fadeSide === undefined,
            '[mask-image:linear-gradient(to_right,transparent,black_min(var(--offset-x-left)*1px,var(--fade-offset)),black_100%)]':
              fade === 'mask' && resolvedFadeOrientation === 'horizontal' && fadeSide === 'start',
            '[mask-image:linear-gradient(to_right,black_0,black_calc(100%-min(var(--offset-x-right)*1px,var(--fade-offset))),transparent)]':
              fade === 'mask' && resolvedFadeOrientation === 'horizontal' && fadeSide === 'end',
          },
        )}
      >
        {children}
      </ScrollAreaPrimitive.Viewport>
      {vertical && <ScrollBar />}
      {horizontal && <ScrollBar orientation="horizontal" />}
      <ScrollAreaPrimitive.Corner />
    </ScrollAreaPrimitive.Root>
  )
}

function ScrollBar({
  className,
  orientation = 'vertical',
  ...props
}: React.ComponentProps<typeof ScrollAreaPrimitive.ScrollAreaScrollbar>) {
  return (
    <ScrollAreaPrimitive.ScrollAreaScrollbar
      data-slot="scroll-area-scrollbar"
      orientation={orientation}
      className={cn(
        'flex touch-none p-px transition-colors select-none',
        {
          'h-full w-2.5 border-l border-l-transparent': orientation === 'vertical',
          'h-2.5 flex-col border-t border-t-transparent': orientation === 'horizontal',
        },
        className,
      )}
      {...props}
    >
      <ScrollAreaPrimitive.ScrollAreaThumb
        data-slot="scroll-area-thumb"
        className="bg-border relative flex-1 rounded-full"
      />
    </ScrollAreaPrimitive.ScrollAreaScrollbar>
  )
}

export { ScrollArea, ScrollBar }
