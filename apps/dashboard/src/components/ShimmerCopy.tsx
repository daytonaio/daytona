/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useCopyToClipboard } from '@/hooks/useCopyToClipboard'
import { cn } from '@/lib/utils'
import { ComponentProps, createContext, useContext, useMemo } from 'react'
import { CopyButton } from './CopyButton'

interface ShimmerCopyContextValue {
  value: string
  copied: string | null
  copy: (text: string) => void
}

const ShimmerCopyContext = createContext<ShimmerCopyContextValue | null>(null)

function useShimmerCopyContext() {
  const context = useContext(ShimmerCopyContext)
  if (!context) {
    throw new Error('ShimmerCopy must be used within ShimmerCopy')
  }
  return context
}

interface ShimmerCopyProps extends ComponentProps<'span'> {
  value: string
  children: React.ReactNode
}

function ShimmerCopy({ value, children, className, ...props }: ShimmerCopyProps) {
  const [copied, copy] = useCopyToClipboard()

  const contextValue = useMemo(
    () => ({
      value,
      copied,
      copy,
    }),
    [value, copied, copy],
  )

  return (
    <ShimmerCopyContext.Provider value={contextValue}>
      <span className={cn('inline-flex gap-2 items-center group/copy-button', className)} {...props}>
        {children}
      </span>
    </ShimmerCopyContext.Provider>
  )
}

interface ShimmerCopyLabelProps extends Omit<ComponentProps<'span'>, 'children'> {
  children?: React.ReactNode
}

function ShimmerCopyLabel({ children, className, ...props }: ShimmerCopyLabelProps) {
  const { value, copied } = useShimmerCopyContext()

  return (
    <span
      className={cn(
        'text-sm whitespace-nowrap z-[1] relative transition-colors',
        {
          'animate-shimmer-text': copied,
        },
        className,
      )}
      {...props}
    >
      {children ?? value}
    </span>
  )
}

function ShimmerCopyButton({
  autoHide = false,
  size = 'icon-xs',
  ...props
}: Omit<ComponentProps<typeof CopyButton>, 'value' | 'copied' | 'copy'>) {
  const { value, copied, copy } = useShimmerCopyContext()

  return (
    <CopyButton
      value={value}
      autoHide={autoHide}
      size={size}
      data-copied={copied}
      copied={copied}
      copy={copy}
      {...props}
    />
  )
}

export { ShimmerCopy, ShimmerCopyButton, ShimmerCopyLabel }
