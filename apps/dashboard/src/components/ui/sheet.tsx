/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

'use client'

import * as SheetPrimitive from '@radix-ui/react-dialog'
import { cva, type VariantProps } from 'class-variance-authority'
import { GripVerticalIcon, X } from 'lucide-react'
import { animate, motion, useMotionValue } from 'motion/react'
import * as React from 'react'

import { cn } from '@/lib/utils'
import { useCallback, useEffect, useImperativeHandle, useRef, useState } from 'react'

export const sheetVariants = cva(
  'fixed z-50 gap-4 bg-background p-6 shadow-lg outline-none transition ease-out data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:duration-200 data-[state=open]:duration-250',
  {
    variants: {
      side: {
        top: 'inset-x-0 top-0 border-b data-[state=closed]:slide-out-to-top data-[state=open]:slide-in-from-top',
        bottom:
          'inset-x-0 bottom-0 border-t data-[state=closed]:slide-out-to-bottom data-[state=open]:slide-in-from-bottom',
        left: 'inset-y-0 left-0 h-full w-3/4 border-r data-[state=closed]:slide-out-to-left data-[state=open]:slide-in-from-left sm:max-w-sm',
        right:
          'inset-y-0 right-0 h-full w-3/4  border-l data-[state=closed]:slide-out-to-right data-[state=open]:slide-in-from-right',
      },
    },
    defaultVariants: {
      side: 'right',
    },
  },
)

function Sheet({ ...props }: React.ComponentProps<typeof SheetPrimitive.Root>) {
  return <SheetPrimitive.Root data-slot="sheet" {...props} />
}

function SheetTrigger({ ...props }: React.ComponentProps<typeof SheetPrimitive.Trigger>) {
  return <SheetPrimitive.Trigger data-slot="sheet-trigger" {...props} />
}

function SheetClose({ ...props }: React.ComponentProps<typeof SheetPrimitive.Close>) {
  return <SheetPrimitive.Close data-slot="sheet-close" {...props} />
}

function SheetPortal({ ...props }: React.ComponentProps<typeof SheetPrimitive.Portal>) {
  return <SheetPrimitive.Portal data-slot="sheet-portal" {...props} />
}

function SheetOverlay({ className, ...props }: React.ComponentProps<typeof SheetPrimitive.Overlay>) {
  return (
    <SheetPrimitive.Overlay
      data-slot="sheet-overlay"
      className={cn(
        'fixed inset-0 z-50 bg-black/50 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0',
        className,
      )}
      {...props}
    />
  )
}

export interface SheetContentProps
  extends React.ComponentProps<typeof SheetPrimitive.Content>,
    VariantProps<typeof sheetVariants> {
  showCloseButton?: boolean
}

function SheetContent({ className, children, side = 'right', showCloseButton = true, ...props }: SheetContentProps) {
  return (
    <SheetPortal>
      <SheetOverlay />
      <SheetPrimitive.Content
        data-slot="sheet-content"
        data-side={side}
        className={cn(sheetVariants({ side }), className)}
        {...props}
      >
        {children}
        {showCloseButton ? <SheetCloseButton /> : null}
      </SheetPrimitive.Content>
    </SheetPortal>
  )
}

function SheetHeader({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="sheet-header"
      className={cn('flex flex-col gap-2 text-center sm:text-left', className)}
      {...props}
    />
  )
}

function SheetFooter({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="sheet-footer"
      className={cn('flex flex-col-reverse gap-2 sm:flex-row sm:justify-end', className)}
      {...props}
    />
  )
}

function SheetTitle({ className, ...props }: React.ComponentProps<typeof SheetPrimitive.Title>) {
  return (
    <SheetPrimitive.Title
      data-slot="sheet-title"
      className={cn('text-lg font-medium text-foreground', className)}
      {...props}
    />
  )
}

function SheetDescription({ className, ...props }: React.ComponentProps<typeof SheetPrimitive.Description>) {
  return (
    <SheetPrimitive.Description
      data-slot="sheet-description"
      className={cn('text-sm text-muted-foreground', className)}
      {...props}
    />
  )
}

function SheetCloseButton() {
  return (
    <SheetPrimitive.Close className="absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none data-[state=open]:bg-secondary">
      <X className="h-4 w-4" />
      <span className="sr-only">Close</span>
    </SheetPrimitive.Close>
  )
}

type ResizableSheetSide = 'left' | 'right'

const RESIZE_HIT_AREA = 40
const RESIZE_STEP = 24
const RESIZE_PAGE_STEP = 64
const RESIZE_RELEASE_DURATION = 0.15
const RESIZE_RELEASE_EASE = [0.22, 1, 0.36, 1] as const
const DEFAULT_RESIZE_FALLOFF_DISTANCE = 24
const DEFAULT_RESIZE_FALLOFF_STIFFNESS = 3

function clampWidth(width: number, minWidth: number, maxWidth: number) {
  return Math.min(Math.max(width, minWidth), maxWidth)
}

function applyResizeFalloff(width: number, minWidth: number, maxWidth: number, distance: number, stiffness: number) {
  if (width < minWidth) {
    const overshoot = minWidth - width
    return minWidth - (distance * overshoot) / (overshoot + distance * stiffness)
  }

  if (width > maxWidth) {
    const overshoot = width - maxWidth
    return maxWidth + (distance * overshoot) / (overshoot + distance * stiffness)
  }

  return width
}

function isPointerInResizeHitArea(pointerX: number, side: ResizableSheetSide, width: number, hitArea: number) {
  if (typeof window === 'undefined') {
    return false
  }

  if (side === 'right') {
    const edgeX = window.innerWidth - width
    return pointerX >= edgeX - hitArea && pointerX <= edgeX
  }

  const edgeX = width
  return pointerX >= edgeX && pointerX <= edgeX + hitArea
}

interface ResizableSheetContentHandleProps extends Omit<React.ComponentProps<'div'>, 'ref' | 'side'> {
  ref?: React.Ref<HTMLDivElement>
  side: ResizableSheetSide
  active?: boolean
  hitArea?: number
}

function ResizableSheetContentHandle({
  ref,
  side,
  active = false,
  hitArea = RESIZE_HIT_AREA,
  className,
  ...props
}: ResizableSheetContentHandleProps) {
  const isRight = side === 'right'

  return (
    <div
      ref={ref}
      data-slot="resizable-sheet-content-handle"
      data-separator={active ? 'active' : undefined}
      className={cn(
        'group absolute top-0 z-20 hidden h-full cursor-col-resize touch-none select-none outline-none sm:block data-[disabled]:pointer-events-none',
        isRight ? 'left-0 -translate-x-full' : 'right-0 translate-x-full',
        className,
      )}
      style={{ width: hitArea }}
      {...props}
    >
      <div
        className={cn(
          'absolute inset-y-0 w-px bg-border transition-colors group-data-[separator=active]:bg-primary group-hover:bg-primary group-focus-visible:bg-primary',
          isRight ? 'right-0' : 'left-0',
        )}
      />
      <div
        className={cn(
          'absolute top-1/2 z-10 flex h-6 w-3.5 -translate-y-1/2 items-center justify-center rounded-sm border border-border bg-background text-muted-foreground transition-colors group-data-[separator=active]:border-primary group-data-[separator=active]:text-primary group-hover:border-primary group-hover:text-primary group-focus-visible:border-primary group-focus-visible:text-primary',
          isRight ? 'right-0 translate-x-1/2' : 'left-0 -translate-x-1/2',
        )}
      >
        <GripVerticalIcon className="size-3" />
      </div>
    </div>
  )
}

interface ResizableSheetContentProps extends Omit<SheetContentProps, 'ref' | 'style'> {
  ref?: React.Ref<ResizableSheetContentRef>
  defaultWidth: number
  minWidth: number
  maxWidth: number
  resizable?: boolean
  onWidthChange?: (width: number) => void
}

export interface ResizableSheetContentRef {
  resize: (width: number, options?: { duration?: number; immediate?: boolean; notify?: boolean }) => void
  getWidth: () => number
}

export function ResizableSheetContent({
  ref,
  className,
  children,
  side = 'right',
  showCloseButton = true,
  defaultWidth,
  minWidth,
  maxWidth,
  resizable = true,
  onWidthChange,
  ...props
}: ResizableSheetContentProps) {
  const resizeSide: ResizableSheetSide = side === 'left' ? 'left' : 'right'
  const canResize = resizable && (side === 'left' || side === 'right')
  const initialWidth = clampWidth(defaultWidth, minWidth, maxWidth)
  const widthValue = useMotionValue(initialWidth)
  const animationRef = useRef<ReturnType<typeof animate> | null>(null)
  const handleRef = useRef<HTMLDivElement>(null)
  const resizeStateRef = useRef<{ pointerId: number; startX: number; startWidth: number } | null>(null)
  const rawWidthRef = useRef<number | null>(null)
  const [committedWidth, setCommittedWidth] = useState(initialWidth)
  const [isResizing, setIsResizing] = useState(false)
  const [isHandleTemporarilyDisabled, setIsHandleTemporarilyDisabled] = useState(false)

  const updateHandleAria = useCallback(
    (value: number) => {
      if (!handleRef.current) {
        return
      }

      handleRef.current.setAttribute('aria-valuemin', String(minWidth))
      handleRef.current.setAttribute('aria-valuemax', String(maxWidth))
      handleRef.current.setAttribute('aria-valuenow', String(value))
      handleRef.current.setAttribute('aria-valuetext', `${value}px wide`)
    },
    [maxWidth, minWidth],
  )

  const animateWidth = useCallback(
    (target: number, immediate = false, duration = RESIZE_RELEASE_DURATION) => {
      animationRef.current?.stop()

      if (immediate) {
        widthValue.set(target)
        return
      }

      animationRef.current = animate(widthValue, target, {
        duration,
        ease: RESIZE_RELEASE_EASE,
      })
    },
    [widthValue],
  )

  useEffect(() => {
    return () => animationRef.current?.stop()
  }, [])

  const setWidth = useCallback(
    (nextWidth: number, options?: { duration?: number; immediate?: boolean; notify?: boolean }) => {
      const clampedWidth = clampWidth(nextWidth, minWidth, maxWidth)

      resizeStateRef.current = null
      rawWidthRef.current = null
      setCommittedWidth(clampedWidth)
      setIsResizing(false)
      updateHandleAria(clampedWidth)
      animateWidth(clampedWidth, options?.immediate, options?.duration)

      if (options?.notify !== false) {
        onWidthChange?.(clampedWidth)
      }
    },
    [animateWidth, maxWidth, minWidth, onWidthChange, updateHandleAria],
  )

  useEffect(() => {
    if (isResizing) {
      return
    }

    const clampedCommittedWidth = clampWidth(committedWidth, minWidth, maxWidth)

    if (clampedCommittedWidth === committedWidth) {
      return
    }

    setWidth(clampedCommittedWidth, { notify: false })
  }, [committedWidth, isResizing, maxWidth, minWidth, setWidth])

  useEffect(() => {
    if (!isHandleTemporarilyDisabled) {
      return
    }

    const reenableHandleIfPointerLeftHitArea = (event: PointerEvent) => {
      if (!isPointerInResizeHitArea(event.clientX, resizeSide, committedWidth, RESIZE_HIT_AREA)) {
        setIsHandleTemporarilyDisabled(false)
      }
    }

    window.addEventListener('pointermove', reenableHandleIfPointerLeftHitArea)

    return () => {
      window.removeEventListener('pointermove', reenableHandleIfPointerLeftHitArea)
    }
  }, [committedWidth, isHandleTemporarilyDisabled, resizeSide])

  const commitResize = useCallback(() => {
    if (!resizeStateRef.current) {
      return
    }

    const nextCommittedWidth = clampWidth(rawWidthRef.current ?? committedWidth, minWidth, maxWidth)

    setIsHandleTemporarilyDisabled(true)
    handleRef.current?.blur()
    setWidth(nextCommittedWidth)
  }, [committedWidth, maxWidth, minWidth, setWidth])

  useEffect(() => {
    if (!isResizing) {
      return
    }

    const handleWindowPointerMove = (event: PointerEvent) => {
      if (!resizeStateRef.current || resizeStateRef.current.pointerId !== event.pointerId) {
        return
      }

      event.preventDefault()

      const delta =
        resizeSide === 'right'
          ? resizeStateRef.current.startX - event.clientX
          : event.clientX - resizeStateRef.current.startX
      const rawWidth = resizeStateRef.current.startWidth + delta
      const displayWidth = applyResizeFalloff(
        rawWidth,
        minWidth,
        maxWidth,
        DEFAULT_RESIZE_FALLOFF_DISTANCE,
        DEFAULT_RESIZE_FALLOFF_STIFFNESS,
      )

      rawWidthRef.current = rawWidth
      widthValue.set(displayWidth)
      updateHandleAria(clampWidth(rawWidth, minWidth, maxWidth))
    }

    const handleWindowPointerEnd = (event: PointerEvent) => {
      if (resizeStateRef.current && resizeStateRef.current.pointerId !== event.pointerId) {
        return
      }

      commitResize()
    }

    const handleWindowBlur = () => {
      commitResize()
    }

    window.addEventListener('pointermove', handleWindowPointerMove, { passive: false })
    window.addEventListener('pointerup', handleWindowPointerEnd)
    window.addEventListener('pointercancel', handleWindowPointerEnd)
    window.addEventListener('blur', handleWindowBlur)

    return () => {
      window.removeEventListener('pointermove', handleWindowPointerMove)
      window.removeEventListener('pointerup', handleWindowPointerEnd)
      window.removeEventListener('pointercancel', handleWindowPointerEnd)
      window.removeEventListener('blur', handleWindowBlur)
    }
  }, [commitResize, isResizing, maxWidth, minWidth, resizeSide, updateHandleAria, widthValue])

  useImperativeHandle(
    ref,
    () => ({
      resize: (nextWidth, options) => {
        setWidth(nextWidth, options)
      },
      getWidth: () => committedWidth,
    }),
    [committedWidth, setWidth],
  )

  const handleResizePointerDown = (event: React.PointerEvent<HTMLDivElement>) => {
    if (!canResize) {
      return
    }

    event.preventDefault()

    resizeStateRef.current = {
      pointerId: event.pointerId,
      startX: event.clientX,
      startWidth: committedWidth,
    }

    animationRef.current?.stop()
    setIsResizing(true)
    event.currentTarget.focus()
    event.currentTarget.setPointerCapture(event.pointerId)
  }

  const handleResizePointerMove = (event: React.PointerEvent<HTMLDivElement>) => {
    if (resizeStateRef.current?.pointerId !== event.pointerId) {
      return
    }

    event.preventDefault()
  }

  const handleResizePointerUp = (event: React.PointerEvent<HTMLDivElement>) => {
    if (resizeStateRef.current?.pointerId !== event.pointerId) {
      return
    }

    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId)
    }

    commitResize()
  }

  const handleResizeLostPointerCapture = () => {
    commitResize()
  }

  const handleResizeKeyDown = (event: React.KeyboardEvent<HTMLDivElement>) => {
    if (!canResize) {
      return
    }

    const currentWidth = committedWidth
    const expandKey = side === 'right' ? 'ArrowLeft' : 'ArrowRight'
    const shrinkKey = side === 'right' ? 'ArrowRight' : 'ArrowLeft'

    let nextWidth: number | null = null

    switch (event.key) {
      case expandKey:
        nextWidth = clampWidth(currentWidth + RESIZE_STEP, minWidth, maxWidth)
        break
      case shrinkKey:
        nextWidth = clampWidth(currentWidth - RESIZE_STEP, minWidth, maxWidth)
        break
      case 'PageUp':
        nextWidth = clampWidth(currentWidth + RESIZE_PAGE_STEP, minWidth, maxWidth)
        break
      case 'PageDown':
        nextWidth = clampWidth(currentWidth - RESIZE_PAGE_STEP, minWidth, maxWidth)
        break
      case 'Home':
        nextWidth = minWidth
        break
      case 'End':
        nextWidth = maxWidth
        break
      case 'Escape':
        nextWidth = committedWidth
        break
      default:
        break
    }

    if (nextWidth === null) {
      return
    }

    event.preventDefault()
    rawWidthRef.current = null
    setCommittedWidth(nextWidth)
    updateHandleAria(nextWidth)
    animateWidth(nextWidth)
    onWidthChange?.(nextWidth)
  }

  return (
    <SheetPortal>
      <SheetOverlay />
      <SheetPrimitive.Content asChild data-slot="sheet-content" data-side={side} {...props}>
        <motion.div
          className={cn(sheetVariants({ side }), 'overflow-visible', className)}
          style={{ width: widthValue }}
        >
          <div className="relative flex min-h-0 flex-1 flex-col overflow-visible">
            {canResize && (
              <ResizableSheetContentHandle
                ref={handleRef}
                side={resizeSide}
                active={isResizing}
                data-disabled={isHandleTemporarilyDisabled ? '' : undefined}
                hitArea={RESIZE_HIT_AREA}
                role="separator"
                tabIndex={0}
                aria-label="Resize sheet"
                aria-orientation="vertical"
                onPointerDown={handleResizePointerDown}
                onPointerMove={handleResizePointerMove}
                onPointerUp={handleResizePointerUp}
                onPointerCancel={handleResizePointerUp}
                onLostPointerCapture={handleResizeLostPointerCapture}
                onKeyDown={handleResizeKeyDown}
              />
            )}
            {children}
            {showCloseButton ? <SheetCloseButton /> : null}
          </div>
        </motion.div>
      </SheetPrimitive.Content>
    </SheetPortal>
  )
}

export {
  Sheet,
  SheetClose,
  SheetCloseButton,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetOverlay,
  SheetPortal,
  SheetTitle,
  SheetTrigger,
}
