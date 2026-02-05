/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { Slot } from '@radix-ui/react-slot'
import { useRef, useState } from 'react'
import { Tooltip, TooltipContent, TooltipTrigger } from './ui/tooltip'

export function EllipsisWithTooltip({
  children,
  asChild,
  className,
  ...props
}: {
  children: React.ReactNode
  className?: string
  asChild?: boolean
}) {
  const [isOpen, setIsOpen] = useState(false)
  const triggerRef = useRef<HTMLDivElement>(null)

  const Comp = asChild ? Slot : 'div'

  return (
    <Tooltip
      open={isOpen}
      onOpenChange={(shouldOpen) => {
        if (shouldOpen) {
          const isTruncated = triggerRef.current && triggerRef.current.scrollWidth > triggerRef.current.clientWidth
          if (isTruncated) {
            setIsOpen(true)
          }
        } else {
          setIsOpen(false)
        }
      }}
      delayDuration={300}
    >
      <TooltipTrigger asChild>
        <Comp ref={triggerRef} className={cn('truncate', className)} {...props}>
          {children}
        </Comp>
      </TooltipTrigger>
      <TooltipContent>{children}</TooltipContent>
    </Tooltip>
  )
}
